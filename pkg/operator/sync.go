package operator

import (
	"fmt"

	api "github.com/coreos/etcd-backup-operator/pkg/apis/backup/v1alpha1"

	"k8s.io/client-go/kubernetes/scheme"
)

const (
	// Copy from deployment_controller.go:
	// maxRetries is the number of times a Vault will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// a Vault is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
)

func (b *Backup) runWorker() {
	for b.processNextItem() {
	}
}

func (b *Backup) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := b.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer b.queue.Done(key)

	err := b.processItem(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	b.handleErr(err, key)
	return true
}

func (b *Backup) processItem(key string) error {
	obj, exists, err := b.indexer.GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	cobj, err := scheme.Scheme.DeepCopy(obj)
	if err != nil {
		return err
	}
	eb := cobj.(*api.EtcdBackup)
	b.handleBackup(&eb.Spec)
	return nil
}

func (b *Backup) handleBackup(spec *api.EtcdBackupSpec) error {
	if spec.StorageType != "s3" {
		return fmt.Errorf("unknown storage type: %v", spec.StorageType)
	}
	if spec.StorageSource.S3 == nil {
		return fmt.Errorf("s3 spec is not initialized")
	}
	// TODO
	return nil
}

func (b *Backup) handleErr(err error, key interface{}) {
	// TODO
}
