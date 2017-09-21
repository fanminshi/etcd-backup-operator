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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd-backup-operator/pkg/backup/s3"
)

const (
	backupFilePerm       = 0600
	backupFilenameSuffix = "etcd.backup"
)

type s3Backend struct {
	S3 *s3.S3
	// dir to temporarily store backup files before upload it to S3.
	dir string
}

func (sb *s3Backend) save(version string, snapRev int64, rc io.Reader) (int64, error) {
	// make a local file copy of the backup first, since s3 requires io.ReadSeeker.
	key := makeBackupName(version, snapRev)
	tmpfile, err := os.OpenFile(filepath.Join(sb.dir, key), os.O_RDWR|os.O_CREATE, backupFilePerm)
	if err != nil {
		return -1, fmt.Errorf("failed to create snapshot tempfile: %v", err)
	}
	defer func() {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
	}()

	n, err := io.Copy(tmpfile, rc)
	if err != nil {
		return -1, fmt.Errorf("failed to save snapshot: %v", err)
	}
	_, err = tmpfile.Seek(0, os.SEEK_SET)
	if err != nil {
		return -1, err
	}
	// S3 put is atomic, so let's go ahead and put the key directly.
	err = sb.S3.Put(key, tmpfile)
	if err != nil {
		return -1, err
	}
	logrus.Infof("saved backup %s (size: %d) successfully", key, n)
	return n, nil
}

func makeBackupName(ver string, rev int64) string {
	return fmt.Sprintf("%s_%016x_%s", ver, rev, backupFilenameSuffix)
}

func (sb *s3Backend) getLatest() (string, error) {
	keys, err := sb.S3.List()
	if err != nil {
		return "", fmt.Errorf("failed to list s3 bucket: %v", err)
	}

	return getLatestBackupName(keys), nil
}

func getLatestBackupName(names []string) string {
	bnames := filterAndSortBackups(names)
	if len(bnames) == 0 {
		return ""
	}
	return bnames[len(bnames)-1]
}

type backupNames []string

func (bn backupNames) Len() int { return len(bn) }

func (bn backupNames) Less(i, j int) bool {
	ri, err := getRev(bn[i])
	if err != nil {
		panic(err)
	}
	rj, err := getRev(bn[j])
	if err != nil {
		panic(err)
	}

	return ri < rj
}

func (bn backupNames) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}

func filterAndSortBackups(names []string) []string {
	bnames := make(backupNames, 0)
	for _, n := range names {
		if !isBackup(n) {
			continue
		}
		_, err := getRev(n)
		if err != nil {
			logrus.Errorf("fail to get rev from backup (%s): %v", n, err)
			continue
		}
		bnames = append(bnames, n)
	}

	sort.Sort(bnames)
	return []string(bnames)
}

func isBackup(name string) bool {
	return strings.HasSuffix(name, backupFilenameSuffix)
}
