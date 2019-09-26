package deploy

import (
	"fmt"
	"os"
	"path/***REMOVED***lepath"
	"strconv"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// getBoolEnv is a helper function that queries the users' shell environment for the @env parameter,
// and if that environment variable has been set, attempts to parse the string returned from
// os.Getenv into a boolean value, or returns an error. If the @env environment variable is not set
// then we return the value stored in the @defaultVar parameter.
func getBoolEnv(env string, defaultVal bool) (bool, error) {
	key := os.Getenv(env)
	if key == "" {
		return defaultVal, nil
	}

	val, err := strconv.ParseBool(key)
	if err != nil {
		return false, fmt.Errorf("Failed to convert the %s env variable into a boolean: %v", env, err)
	}

	return val, nil
}

// DecodeYAMLManifestToObject is a helper function that takes the path to a manifest ***REMOVED***le, e.g. the
// deployment YAML ***REMOVED***le, and opens that ***REMOVED***le using os.Open, which returns an io.Reader object that
// can be passed to the YAML/JSON decoder to build up the @resource parameter for usage in the clientsets.
func DecodeYAMLManifestToObject(path string, resource interface{}) error {
	***REMOVED***le, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Failed to open %s, got: %v", path, err)
	}

	err = yaml.NewYAMLOrJSONDecoder(***REMOVED***le, 100).Decode(&resource)
	if err != nil {
		return fmt.Errorf("invalid manifest passed, wasn't able to decode the YAML ***REMOVED***le, got: %v", err)
	}

	return nil
}

// initMeteringCRDSlice initializes a slice of CRD structures, where each
// structure contains information about an individual CRD that metering manages
// and assigns the deploy CRD ***REMOVED***eld to this slice of structures
func (deploy *Deployer) initMeteringCRDSlice() {
	var crds []CRD

	crds = append(crds, CRD{
		Name: "hivetables.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, hivetableFile),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "prestotables.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, prestotableFile),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "storagelocations.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, storagelocationFile),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "reports.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, reportFile),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "reportqueries.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, reportqueryFile),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "reportdatasources.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, reportdatasourceFile),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "meteringcon***REMOVED***gs.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(deploy.ansibleOperatorManifestsLocation, meteringcon***REMOVED***gFile),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})

	deploy.crds = crds
}
