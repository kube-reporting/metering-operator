package chargeback

import (
	"fmt"

	extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Resources = []*extensions.CustomResourceDe***REMOVED***nition{
	ReportResource,
}

var ReportResource = &extensions.CustomResourceDe***REMOVED***nition{
	ObjectMeta: metav1.ObjectMeta{
		Name: fmt.Sprintf("%s.%s", ReportPlural, Group),
	},
	Spec: extensions.CustomResourceDe***REMOVED***nitionSpec{
		Group:   Group,
		Version: Version,
		Names: extensions.CustomResourceDe***REMOVED***nitionNames{
			Plural: ReportPlural,
			Kind:   ReportKind,
		},
		Scope: extensions.ClusterScoped,
	},
}
