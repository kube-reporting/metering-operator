package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// DecodeYAMLManifestToObject is a helper function that takes the path to a manifest file, e.g. the
// deployment YAML file, and opens that file using os.Open, which returns an io.Reader object that
// can be passed to the YAML/JSON decoder to build up the @resource parameter for usage in the clientsets.
func DecodeYAMLManifestToObject(path string, resource interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s, got: %v", path, err)
	}

	err = yaml.NewYAMLOrJSONDecoder(file, 100).Decode(&resource)
	if err != nil {
		return fmt.Errorf("invalid manifest passed, wasn't able to decode the YAML file, got: %v", err)
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
		Path: filepath.Join(manifestDir, pathToCRDMap["hiveTable"]),
		CRD:  new(apiextv1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "prestotables.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["prestoTable"]),
		CRD:  new(apiextv1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "storagelocations.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["storageLocation"]),
		CRD:  new(apiextv1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "reports.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["report"]),
		CRD:  new(apiextv1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "reportqueries.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["reportQuery"]),
		CRD:  new(apiextv1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "reportdatasources.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["reportDataSource"]),
		CRD:  new(apiextv1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "meteringconfigs.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["meteringConfig"]),
		CRD:  new(apiextv1.CustomResourceDefinition),
	})

	return crds
}

func getMeteringAnsiblePath(manifestDir, platform string) (string, error) {
	if manifestDir == "" {
		return "", fmt.Errorf("failed to set the $DEPLOY_MANIFESTS_DIR or --deploy-manifests-dir flag to a non-empty value")
	}

	deployDir, err := filepath.Abs(manifestDir)
	if err != nil {
		return "", fmt.Errorf("failed to get the absolute path of the manifest/deploy directory %s: %v", manifestDir, err)
	}

	dirStat, err := os.Stat(manifestDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("failed to get the stat the manifest/deploy directory %s: %v", manifestDir, err)
	}
	if !dirStat.IsDir() {
		return "", fmt.Errorf("specified deploy directory '%s' is not a directory", manifestDir)
	}

	var ansibleOperatorManifestDir string

	switch strings.ToLower(platform) {
	case "upstream":
		ansibleOperatorManifestDir = filepath.Join(deployDir, upstreamManifestDirname, manifestAnsibleOperator)
	case "openshift":
		ansibleOperatorManifestDir = filepath.Join(deployDir, openshiftManifestDirname, manifestAnsibleOperator)
	case "ocp-testing":
		ansibleOperatorManifestDir = filepath.Join(deployDir, ocpTestingManifestDirname, manifestAnsibleOperator)
	default:
		return "", fmt.Errorf("failed to set $DEPLOY_PLATFORM or --platform flag to a valid value. Supported platforms: [upstream, openshift, ocp-testing]")
	}

	dirStat, err = os.Stat(ansibleOperatorManifestDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("failed to stat the %s deploy platform directory '%s': %v", platform, ansibleOperatorManifestDir, err)
	}
	if !dirStat.IsDir() {
		return "", fmt.Errorf("specified %s deploy platform directory '%s' is not a directory", platform, ansibleOperatorManifestDir)
	}

	return ansibleOperatorManifestDir, nil
}

func ReadMeteringAnsibleOperatorManifests(manifestDir, platform string) (*OperatorResources, error) {
	var resources OperatorResources

	ansibleOperatorManifestDir, err := getMeteringAnsiblePath(manifestDir, platform)
	if err != nil {
		return nil, fmt.Errorf("failed to get the path to the metering-ansible-operator directory: %v", err)
	}

	pathToCRDMap := map[string]string{
		"hiveTable":        hivetableFile,
		"prestoTable":      prestotableFile,
		"meteringConfig":   meteringconfigFile,
		"report":           reportFile,
		"reportDataSource": reportdatasourceFile,
		"reportQuery":      reportqueryFile,
		"storageLocation":  storagelocationFile,
	}

	resources.CRDs = InitMeteringCRDSlice(ansibleOperatorManifestDir, pathToCRDMap)

	for _, crd := range resources.CRDs {
		err := DecodeYAMLManifestToObject(crd.Path, crd.CRD)
		if err != nil {
			return nil, fmt.Errorf("failed to decode the YAML manifest: %v", err)
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
		path := filepath.Join(ansibleOperatorManifestDir, resource.path)

		err = DecodeYAMLManifestToObject(path, resource.obj)
		if err != nil {
			return nil, fmt.Errorf("failed to decode the YAML manifest for the %s resource: %v", name, err)
		}
	}

	return &resources, nil
}
