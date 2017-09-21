package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	CRDResourceKind   = "EtcdBackup"
	CRDResourcePlural = "etcdbackups"
	groupName         = "etcd.database.coreos.com"
)

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme

	// SchemeGroupVersion is the group version used to register these objects.
	SchemeGroupVersion   = schema.GroupVersion{Group: groupName, Version: "v1alpha1"}
	CRDName              = CRDResourcePlural + "." + groupName
	CRDResourceShortName = []string{"eb"}
)

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&EtcdBackup{},
		&EtcdBackupList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
