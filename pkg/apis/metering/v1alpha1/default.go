package v1alpha1

import runtime "k8s.io/apimachinery/pkg/runtime"

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}
