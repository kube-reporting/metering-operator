package prealpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var SchemeGroupVersion = schema.GroupVersion{Group: "chargeback.coreos.com", Version: "prealpha"}

var (
	// TODO: move SchemeBuilder with zz_generated.deepcopy.go to k8s.io/api.
	// localSchemeBuilder and AddToScheme will stay in k8s.io/kubernetes.
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated ***REMOVED***les. The separation
	// makes the code compile even when the generated ***REMOVED***les are missing.
	localSchemeBuilder.Register(addKnownTypes)
}

// Resource takes an unquali***REMOVED***ed resource and returns a Group quali***REMOVED***ed GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Report{},
		&ReportList{},
		&ReportDataStore{},
		&ReportDataStoreList{},
		&ReportGenerationQuery{},
		&ReportGenerationQueryList{},
		&ReportPrometheusQuery{},
		&ReportPrometheusQueryList{},
	)

	scheme.AddKnownTypes(SchemeGroupVersion,
		&metav1.Status{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
