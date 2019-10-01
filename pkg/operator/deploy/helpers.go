package deploy

import (
	"fmt"
	"os"
	"path/***REMOVED***lepath"
	"strconv"
	"strings"

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

// InitMeteringCRDSlice initializes a slice of CRD structures, where each
// structure contains information about an individual CRD that metering manages
// and returns this slice of CRD structures.
func InitMeteringCRDSlice(manifestDir string, pathToCRDMap map[string]string) []CRD {
	var crds []CRD

	crds = append(crds, CRD{
		Name: "hivetables.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(manifestDir, pathToCRDMap["hiveTable"]),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "prestotables.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(manifestDir, pathToCRDMap["prestoTable"]),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "storagelocations.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(manifestDir, pathToCRDMap["storageLocation"]),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "reports.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(manifestDir, pathToCRDMap["report"]),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "reportqueries.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(manifestDir, pathToCRDMap["reportQuery"]),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "reportdatasources.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(manifestDir, pathToCRDMap["reportDataSource"]),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})
	crds = append(crds, CRD{
		Name: "meteringcon***REMOVED***gs.metering.openshift.io",
		Path: ***REMOVED***lepath.Join(manifestDir, pathToCRDMap["meteringCon***REMOVED***g"]),
		CRD:  new(apiextv1beta1.CustomResourceDe***REMOVED***nition),
	})

	return crds
}

func getMeteringAnsiblePath(manifestDir, platform string) (string, error) {
	if manifestDir == "" {
		return "", fmt.Errorf("Failed to set the $DEPLOY_MANIFESTS_DIR or --deploy-manifests-dir flag to a non-empty value")
	}

	deployDir, err := ***REMOVED***lepath.Abs(manifestDir)
	if err != nil {
		return "", fmt.Errorf("Failed to get the absolute path of the manifest/deploy directory %s: %v", manifestDir, err)
	}

	dirStat, err := os.Stat(manifestDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("Failed to get the stat the manifest/deploy directory %s: %v", manifestDir, err)
	}
	if !dirStat.IsDir() {
		return "", fmt.Errorf("Speci***REMOVED***ed deploy directory '%s' is not a directory", manifestDir)
	}

	var ansibleOperatorManifestDir string

	switch strings.ToLower(platform) {
	case "upstream":
		ansibleOperatorManifestDir = ***REMOVED***lepath.Join(deployDir, upstreamManifestDirname, manifestAnsibleOperator)
	case "openshift":
		ansibleOperatorManifestDir = ***REMOVED***lepath.Join(deployDir, openshiftManifestDirname, manifestAnsibleOperator)
	case "ocp-testing":
		ansibleOperatorManifestDir = ***REMOVED***lepath.Join(deployDir, ocpTestingManifestDirname, manifestAnsibleOperator)
	default:
		return "", fmt.Errorf("Failed to set $DEPLOY_PLATFORM or --platform flag to a valid value. Supported platforms: [upstream, openshift, ocp-testing]")
	}

	dirStat, err = os.Stat(ansibleOperatorManifestDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("Failed to stat the %s deploy platform directory '%s': %v", platform, ansibleOperatorManifestDir, err)
	}
	if !dirStat.IsDir() {
		return "", fmt.Errorf("Speci***REMOVED***ed %s deploy platform directory '%s' is not a directory", platform, ansibleOperatorManifestDir)
	}

	return ansibleOperatorManifestDir, nil
}

func ReadMeteringAnsibleOperatorManifests(manifestDir, platform string) (*OperatorResources, error) {
	var resources OperatorResources

	ansibleOperatorManifestDir, err := getMeteringAnsiblePath(manifestDir, platform)
	if err != nil {
		return nil, fmt.Errorf("Failed to get the path to the metering-ansible-operator directory: %v", err)
	}

	pathToCRDMap := map[string]string{
		"hiveTable":        hivetableFile,
		"prestoTable":      prestotableFile,
		"meteringCon***REMOVED***g":   meteringcon***REMOVED***gFile,
		"report":           reportFile,
		"reportDataSource": reportdatasourceFile,
		"reportQuery":      reportqueryFile,
		"storageLocation":  storagelocationFile,
	}

	resources.CRDs = InitMeteringCRDSlice(ansibleOperatorManifestDir, pathToCRDMap)

	for _, crd := range resources.CRDs {
		err := DecodeYAMLManifestToObject(crd.Path, crd.CRD)
		if err != nil {
			return nil, fmt.Errorf("Failed to decode the YAML manifest: %v", err)
		}
	}

	meteringResourceMap := map[string]struct {
		path string
		obj  interface{}
	}{
		"deployment": {
			path: meteringDeploymentFile,
			obj:  &resources.Deployment,
		},
		"serviceAccount": {
			path: meteringServiceAccountFile,
			obj:  &resources.ServiceAccount,
		},
		"roleBinding": {
			path: meteringRoleBindingFile,
			obj:  &resources.RoleBinding,
		},
		"role": {
			path: meteringRoleFile,
			obj:  &resources.Role,
		},
		"clusterRoleBinding": {
			path: meteringClusterRoleBindingFile,
			obj:  &resources.ClusterRoleBinding,
		},
		"clusterRole": {
			path: meteringClusterRoleFile,
			obj:  &resources.ClusterRole,
		},
	}

	for name, resource := range meteringResourceMap {
		path := ***REMOVED***lepath.Join(ansibleOperatorManifestDir, resource.path)

		err = DecodeYAMLManifestToObject(path, resource.obj)
		if err != nil {
			return nil, fmt.Errorf("Failed to decode the YAML manifest for the %s resource: %v", name, err)
		}
	}

	return &resources, nil
}
