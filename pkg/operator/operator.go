package operator

import (
	"context"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd-backup-operator/pkg/client"
	"github.com/coreos/etcd-backup-operator/pkg/util/constants"
	"github.com/coreos/etcd-backup-operator/pkg/util/k8sutil"
	"github.com/coreos/etcd-operator/pkg/util/retryutil"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Backup struct {
	namespace      string
	name           string
	serviceAccount string
	// k8s workqueue pattern
	indexer  cache.Indexer
	informer cache.Controller
	queue    workqueue.RateLimitingInterface

	kubecli       kubernetes.Interface
	backupCRCli   client.BackupCR
	kubeExtClient apiextensionsclient.Interface
}

// New creates a backup operator.
func New() *Backup {
	return &Backup{
		namespace:     os.Getenv(constants.EnvOperatorPodNamespace),
		name:          os.Getenv(constants.EnvOperatorPodName),
		kubecli:       k8sutil.MustNewKubeClient(),
		backupCRCli:   client.MustNewInCluster(),
		kubeExtClient: k8sutil.MustNewKubeExtClient(),
	}
}

// Start starts the Backup operator.
func (b *Backup) Start(ctx context.Context) error {
	err := b.init(ctx)
	if err != nil {
		return err
	}
	go b.run(ctx)
	<-ctx.Done()
	return ctx.Err()
}

func (b *Backup) init(ctx context.Context) error {
	err := k8sutil.CreateBackupCRD(b.kubeExtClient)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	b.serviceAccount, err = b.getMyPodServiceAccount()
	return err
}

func (b *Backup) getMyPodServiceAccount() (string, error) {
	var sa string
	name := os.Getenv(constants.EnvOperatorPodName)
	err := retryutil.Retry(5*time.Second, 100, func() (bool, error) {
		pod, err := b.kubecli.CoreV1().Pods(b.namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			logrus.Errorf("fail to get operator pod (%s): %v", name, err)
			return false, nil
		}
		sa = pod.Spec.ServiceAccountName
		return true, nil
	})
	return sa, err
}
