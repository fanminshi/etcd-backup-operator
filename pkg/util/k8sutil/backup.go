package k8sutil

import (
	"encoding/json"
	"fmt"

	api "github.com/coreos/etcd-backup-operator/pkg/apis/backup/v1alpha1"
	"github.com/coreos/etcd-backup-operator/pkg/util/constants"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	BackupImage = "gcr.io/coreos-k8s-scale-testing/etcd-backup-operator:fanmin"
	BackupSpec  = "BACKUP_SPEC"
)

func NewBackupPodTemplate(account string, bs api.EtcdBackupSpec) v1.PodTemplateSpec {
	b, err := json.Marshal(bs)
	if err != nil {
		panic("unexpected json error " + err.Error())
	}

	ps := v1.PodSpec{
		ServiceAccountName: account,
		Containers: []v1.Container{
			{
				Name:  "backup",
				Image: BackupImage,
				Command: []string{
					"/usr/local/bin/etcd-backup",
					"--etcd-cluster=" + bs.ClusterName,
				},
				Env: []v1.EnvVar{{
					Name:      constants.EnvOperatorPodNamespace,
					ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "metadata.namespace"}},
				}, {
					Name:  BackupSpec,
					Value: string(b),
				}},
			},
		},
	}

	pl := v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:   bs.ClusterName,
			Labels: BackupSidecarLabels(bs.ClusterName),
		},
		Spec: ps,
	}

	return pl
}

const (
	awsCredentialDir          = "/root/.aws/"
	awsSecretVolName          = "secret-aws"
	AWSS3Bucket               = "AWS_S3_BUCKET"
	BackupPodSelectorAppField = "etcd_backup_tool"
)

func AttachS3ToPodSpec(ps *v1.PodSpec, ss *api.S3Source) {
	ps.Containers[0].VolumeMounts = append(ps.Containers[0].VolumeMounts, v1.VolumeMount{
		Name:      awsSecretVolName,
		MountPath: awsCredentialDir,
	})
	ps.Volumes = append(ps.Volumes, v1.Volume{
		Name: awsSecretVolName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: ss.AWSSecret,
			},
		},
	})
	ps.Containers[0].Env = append(ps.Containers[0].Env, v1.EnvVar{
		Name:  AWSS3Bucket,
		Value: ss.S3Bucket,
	})
}

func BackupSidecarName(name string) string {
	return fmt.Sprintf("%s-backup-sidecar", name)
}

func BackupSidecarLabels(clusterName string) map[string]string {
	return map[string]string{
		"app":          BackupPodSelectorAppField,
		"etcd_cluster": clusterName,
	}
}

func NewBackupDeploymentManifest(name string, dplSel map[string]string, pl v1.PodTemplateSpec, owner metav1.OwnerReference) *appsv1beta1.Deployment {
	d := &appsv1beta1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: dplSel,
		},
		Spec: appsv1beta1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: pl.ObjectMeta.Labels},
			Template: pl,
			Strategy: appsv1beta1.DeploymentStrategy{
				Type: appsv1beta1.RecreateDeploymentStrategyType,
			},
		},
	}
	AddOwnerRefToObject(d.GetObjectMeta(), owner)
	return d
}
