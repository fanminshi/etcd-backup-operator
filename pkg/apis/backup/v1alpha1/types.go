package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EtcdBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []EtcdBackup `json:"items"`
}

type EtcdBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              EtcdBackupSpec   `json:"spec"`
	Status            EtcdBackupStatus `json:"status,omitempty"`
}

type EtcdBackupSpec struct {
	// clusterName is the etcd cluster name.
	ClusterName string `json:"clusterName,omitempty"`

	StorageType string `json:"storageType"`

	StorageSource `json:",inline"`
}

type EtcdBackupStatus struct {
	// Succeeded indicates if the backup is Succeeded.
	Succeeded bool `json:"succeeded"`
	// Reason indicates reason for any backup failure.
	Reason string `json:"Reason"`
}
