package cluster

import (
	"fmt"

	api "github.com/coreos/etcd-backup-operator/pkg/apis/backup/v1alpha1"
	k8sutil "github.com/coreos/etcd-backup-operator/pkg/util/k8sutil"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/client-go/kubernetes"
)

type backupManager struct {
	kubeCli        kubernetes.Interface
	backup         *api.EtcdBackup
	serviceAccount string
}

func New(kubeCli kubernetes.Interface, serviceAccount string, backup *api.EtcdBackup) (*backupManager, error) {
	if kubeCli == nil {
		return nil, fmt.Errorf("kubeCli not defined")
	}
	if backup == nil {
		return nil, fmt.Errorf("backup not defined")
	}
	return &backupManager{kubeCli, backup, serviceAccount}, nil
}

func (bm *backupManager) Setup() error {
	return bm.runSidecar()
}

func (bm *backupManager) runSidecar() error {
	if err := bm.createSidecarDeployment(); err != nil {
		return fmt.Errorf("failed to create backup sidecar Deployment: %v", err)
	}
	return nil
}

func (bm *backupManager) createSidecarDeployment() error {
	d := bm.makeSidecarDeployment()
	_, err := bm.kubeCli.AppsV1beta1().Deployments(bm.backup.Namespace).Create(d)
	return err
}

func (bm *backupManager) makeSidecarDeployment() *appsv1beta1.Deployment {
	b := bm.backup
	podTemplate := k8sutil.NewBackupPodTemplate(bm.serviceAccount, b.Spec)
	k8sutil.AttachS3ToPodSpec(&podTemplate.Spec, b.Spec.S3)
	name := k8sutil.BackupSidecarName(b.Name)
	dplSel := k8sutil.LabelsForCluster(b.Spec.ClusterName)
	return k8sutil.NewBackupDeploymentManifest(name, dplSel, podTemplate, k8sutil.AsOwner(b))
}
