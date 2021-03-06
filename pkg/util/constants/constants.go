package constants

import "time"

const (
	DefaultDialTimeout      = 5 * time.Second
	DefaultRequestTimeout   = 5 * time.Second
	DefaultSnapshotTimeout  = 1 * time.Minute
	DefaultSnapshotInterval = 1800 * time.Second

	DefaultBackupPodHTTPPort = 19999

	OperatorRoot   = "/var/tmp/etcd-operator"
	BackupMountDir = "/var/etcd-backup"

	PVProvisionerGCEPD  = "kubernetes.io/gce-pd"
	PVProvisionerAWSEBS = "kubernetes.io/aws-ebs"
	PVProvisionerNone   = "none"

	EnvOperatorPodName      = "MY_POD_NAME"
	EnvOperatorPodNamespace = "MY_POD_NAMESPACE"
)
