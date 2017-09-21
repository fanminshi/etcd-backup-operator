package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/Sirupsen/logrus"
	api "github.com/coreos/etcd-backup-operator/pkg/apis/backup/v1alpha1"
	"github.com/coreos/etcd-backup-operator/pkg/backup"
	"github.com/coreos/etcd-backup-operator/pkg/util/k8sutil"
)

var (
	clusterName string
	namespace   string
)

func init() {
	flag.StringVar(&clusterName, "etcd-cluster", "", "")
	flag.Parse()

	namespace = os.Getenv("MY_POD_NAMESPACE")
	if len(namespace) == 0 {
		namespace = "default"
	}
}

func main() {
	if len(clusterName) == 0 {
		panic("clusterName not set")
	}

	var ebs api.EtcdBackupSpec
	bss := os.Getenv("BACKUP_SPEC")
	if err := json.Unmarshal([]byte(bss), &ebs); err != nil {
		logrus.Fatalf("fail to parse backup policy (%s): %v", bss, err)
	}

	kclient := k8sutil.MustNewKubeClient()
	bk, err := backup.New(kclient, ebs, clusterName, namespace)
	if err != nil {
		logrus.Fatalf("failed to create backup sidecar: %v", err)
	}

	bk.Run()
}
