package deploy

import (
	"fmt"
	"os"
	"path/filepath"
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

// decodeYAMLManifestToObject is a helper function that takes the path to a manifest file, e.g. the
// deployment YAML file, and opens that file using os.Open, which returns an io.Reader object that
// can be passed to the YAML/JSON decoder to build up the @resource parameter for usage in the clientsets.
func decodeYAMLManifestToObject(path string, resource interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Failed to open %s, got: %v", path, err)
	}

	err = yaml.NewYAMLOrJSONDecoder(file, 100).Decode(&resource)
	if err != nil {
		return fmt.Errorf("invalid manifest passed, wasn't able to decode the YAML file, got: %v", err)
	}

	return nil
}

// initMeteringCRDSlice initializes a slice of CRD structures, where each
// structure contains information about an individual CRD that metering manages
// and assigns the deploy CRD field to this slice of structures
func (deploy *Deployer) initMeteringCRDSlice() {
	var crds []CRD

	crds = append(crds, CRD{
		Name: "hivetables.metering.openshift.io",
		Path: filepath.Join(deploy.config.ManifestLocation, hivetableFile),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "prestotables.metering.openshift.io",
		Path: filepath.Join(deploy.config.ManifestLocation, prestotableFile),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "storagelocations.metering.openshift.io",
		Path: filepath.Join(deploy.config.ManifestLocation, storagelocationFile),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "reports.metering.openshift.io",
		Path: filepath.Join(deploy.config.ManifestLocation, reportFile),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "reportqueries.metering.openshift.io",
		Path: filepath.Join(deploy.config.ManifestLocation, reportqueryFile),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "reportdatasources.metering.openshift.io",
		Path: filepath.Join(deploy.config.ManifestLocation, reportdatasourceFile),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "meteringconfigs.metering.openshift.io",
		Path: filepath.Join(deploy.config.ManifestLocation, meteringconfigFile),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})

	deploy.crds = crds
}
