package k8sutil

import (
	"fmt"
	"strings"

	api "github.com/coreos/etcd-backup-operator/pkg/apis/backup/v1alpha1"

	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// CascadeDeleteBackground returns a background delete policy option which causes the garbage collector to delete the dependents in the background.
func CascadeDeleteBackground() *metav1.DeleteOptions {
	return &metav1.DeleteOptions{
		PropagationPolicy: func() *metav1.DeletionPropagation {
			background := metav1.DeletePropagationBackground
			return &background
		}(),
	}
}

// PodDNSName constructs the dns name on which a pod can be addressed
func PodDNSName(p v1.Pod) string {
	podIP := strings.Replace(p.Status.PodIP, ".", "-", -1)
	return fmt.Sprintf("%s.%s.pod", podIP, p.Namespace)
}

// AddOwnerRefToObject appends the desired OwnerReference to the object
func AddOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}

// AsOwner returns an owner reference set as the vault cluster CR
func AsOwner(eb *api.EtcdBackup) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.SchemeGroupVersion.String(),
		Kind:       api.CRDResourceKind,
		Name:       eb.Name,
		UID:        eb.UID,
		Controller: &trueVar,
	}
}

func LabelsForCluster(clusterName string) map[string]string {
	return map[string]string{
		"etcd_cluster": clusterName,
		"app":          "etcd",
	}
}

func IsKubernetesResourceNotFoundError(err error) bool {
	return apierrors.IsNotFound(err)
}

// We are using internal api types for cluster related.
func ClusterListOpt(clusterName string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(LabelsForCluster(clusterName)).String(),
	}
}
