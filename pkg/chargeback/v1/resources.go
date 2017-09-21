package v1

import (
	"fmt"

	extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Resources = []*extensions.CustomResourceDefinition{
//ReportResource,
}

var ReportResource = &extensions.CustomResourceDefinition{
	ObjectMeta: metav1.ObjectMeta{
		Name:      fmt.Sprintf("%s.%s", ReportPlural, Group),
		Namespace: "team-chargeback",
	},
	Spec: extensions.CustomResourceDefinitionSpec{
		Group:   Group,
		Version: Version,
		Names: extensions.CustomResourceDefinitionNames{
			Plural: ReportPlural,
			Kind:   ReportKind,
		},
		Scope: extensions.NamespaceScoped,
	},
}
