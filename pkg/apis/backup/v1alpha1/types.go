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

	// BackupIntervalInSecond specifies the interval between two backups.
	BackupIntervalInSecond int `json:"backupIntervalInSecond"`
}

type EtcdBackupStatus struct {
	// Initialized indicates if the Vault service is initialized.
	Initialized bool `json:"initialized"`
}
