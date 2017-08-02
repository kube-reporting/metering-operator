package v1

import (
	"fmt"

	extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Group   = "cron.coreos.com"
	Version = "v1"
)

var (
	// Resources are the CRDs deployed for Cron.
	Resources = []*extensions.CustomResourceDe***REMOVED***nition{
		ReportResource,
	}
)

var ReportResource = &extensions.CustomResourceDe***REMOVED***nition{
	ObjectMeta: metav1.ObjectMeta{
		Name: fmt.Sprintf("%s.%s", CronPlural, Group),
	},
	Spec: extensions.CustomResourceDe***REMOVED***nitionSpec{
		Group:   Group,
		Version: Version,
		Names: extensions.CustomResourceDe***REMOVED***nitionNames{
			Plural: CronPlural,
			Kind:   CronKind,
		},
		Scope: extensions.NamespaceScoped,
	},
}
