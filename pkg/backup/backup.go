// Copyright 2016 The etcd-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backup

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	api "github.com/coreos/etcd-backup-operator/pkg/apis/backup/v1alpha1"
	"github.com/coreos/etcd-backup-operator/pkg/backup/s3"
	"github.com/coreos/etcd-backup-operator/pkg/util/constants"
	"github.com/coreos/etcd-backup-operator/pkg/util/etcdutil"
	k8sutil "github.com/coreos/etcd-backup-operator/pkg/util/k8sutil"
	"github.com/coreos/etcd/clientv3"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	backupTmpDir = "tmp"
	AWSS3Bucket  = "AWS_S3_BUCKET"
)

type Backup struct {
	kclient     kubernetes.Interface
	spec        api.EtcdBackupSpec
	clusterName string
	namespace   string
	be          *s3Backend
}

func New(kclient kubernetes.Interface, sp api.EtcdBackupSpec, clusterName, namespace string) (*Backup, error) {
	bdir := path.Join(constants.BackupMountDir, "v1", clusterName)
	tmpDir := path.Join(bdir, backupTmpDir)
	err := os.MkdirAll(tmpDir, 0700)
	if err != nil {

		return nil, err
	}

	if sp.StorageType != "s3" {
		return nil, fmt.Errorf("unsupported storage type: %v", sp.StorageType)
	}

	s3cli, err := s3.New(os.Getenv(AWSS3Bucket), ToS3Prefix(sp.S3.Prefix, namespace, clusterName))
	if err != nil {
		return nil, err
	}

	s3be := &s3Backend{
		dir: tmpDir,
		S3:  s3cli,
	}

	return &Backup{
		kclient:     kclient,
		spec:        sp,
		clusterName: clusterName,
		namespace:   namespace,
		be:          s3be,
	}, nil
}

const S3V1 = "v1"

func ToS3Prefix(s3Prefix, namespace, clusterName string) string {
	return path.Join(s3Prefix, S3V1, namespace, clusterName)
}

func (b *Backup) Run() {
	lastSnapRev := b.getLatestBackupRev()
	interval := constants.DefaultSnapshotInterval
	if b.spec.BackupIntervalInSecond != 0 {
		interval = time.Duration(b.spec.BackupIntervalInSecond) * time.Second
	}
	for {
		<-time.After(interval)
		rev, err := b.saveSnap(lastSnapRev)
		if err != nil {
			logrus.Errorf("failed to save snapshot: %v", err)
		}
		lastSnapRev = rev
	}
}

func (b *Backup) saveSnap(lastSnapRev int64) (int64, error) {
	podList, err := b.kclient.Core().Pods(b.namespace).List(k8sutil.ClusterListOpt(b.clusterName))
	if err != nil {
		return lastSnapRev, err
	}

	var pods []*v1.Pod
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Status.Phase == v1.PodRunning {
			pods = append(pods, pod)
		}
	}

	if len(pods) == 0 {
		msg := "no running etcd pods found"
		logrus.Warning(msg)
		return lastSnapRev, fmt.Errorf(msg)
	}
	member, rev := getMemberWithMaxRev(pods, nil)
	if member == nil {
		logrus.Warning("no reachable member")
		return lastSnapRev, fmt.Errorf("no reachable member")
	}

	if rev <= lastSnapRev {
		logrus.Info("skipped creating new backup: no change since last time")
		return lastSnapRev, nil
	}

	log.Printf("saving backup for cluster (%s)", b.clusterName)
	if err := b.writeSnap(member, rev); err != nil {
		err = fmt.Errorf("write snapshot failed: %v", err)
		return lastSnapRev, err
	}
	return rev, nil
}

func (b *Backup) writeSnap(m *etcdutil.Member, rev int64) error {
	cfg := clientv3.Config{
		Endpoints:   []string{m.ClientURL()},
		DialTimeout: constants.DefaultDialTimeout,
		TLS:         nil,
	}
	etcdcli, err := clientv3.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create etcd client (%v)", err)
	}
	defer etcdcli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultRequestTimeout)
	resp, err := etcdcli.Maintenance.Status(ctx, m.ClientURL())
	cancel()
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), constants.DefaultSnapshotTimeout)
	rc, err := etcdcli.Maintenance.Snapshot(ctx)
	if err != nil {
		return fmt.Errorf("failed to receive snapshot (%v)", err)
	}
	defer cancel()
	defer rc.Close()

	_, err = b.be.save(resp.Version, rev, rc)
	if err != nil {
		return err
	}

	return nil
}

func getMemberWithMaxRev(pods []*v1.Pod, tc *tls.Config) (*etcdutil.Member, int64) {
	var member *etcdutil.Member
	maxRev := int64(0)
	for _, pod := range pods {
		m := &etcdutil.Member{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			SecureClient: tc != nil,
		}
		cfg := clientv3.Config{
			Endpoints:   []string{m.ClientURL()},
			DialTimeout: constants.DefaultDialTimeout,
			TLS:         tc,
		}
		etcdcli, err := clientv3.New(cfg)
		if err != nil {
			logrus.Warningf("failed to create etcd client for pod (%v): %v", pod.Name, err)
			continue
		}
		defer etcdcli.Close()

		ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultRequestTimeout)
		resp, err := etcdcli.Get(ctx, "/", clientv3.WithSerializable())
		cancel()
		if err != nil {
			logrus.Warningf("getMaxRev: failed to get revision from member %s (%s)", m.Name, m.ClientURL())
			continue
		}

		logrus.Infof("getMaxRev: member %s revision (%d)", m.Name, resp.Header.Revision)
		if resp.Header.Revision > maxRev {
			maxRev = resp.Header.Revision
			member = m
		}
	}
	return member, maxRev
}

func (b *Backup) getLatestBackupRev() int64 {
	// If there is any error, we just exit backup sidecar because we can't serve the backup any way.
	name, err := b.be.getLatest()
	if err != nil {
		logrus.Fatal(err)
	}
	if len(name) == 0 {
		return 0
	}
	rev, err := getRev(name)
	if err != nil {
		logrus.Fatal(err)
	}
	return rev
}

func getRev(name string) (int64, error) {
	parts := strings.SplitN(name, "_", 3)
	if len(parts) != 3 {
		return 0, fmt.Errorf("bad backup name: %s", name)
	}

	return strconv.ParseInt(parts[1], 16, 64)
}
