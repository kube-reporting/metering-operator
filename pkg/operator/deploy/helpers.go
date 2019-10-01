package deploy

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
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

// DecodeYAMLManifestToObject is a helper function that takes the path to a manifest file, e.g. the
// deployment YAML file, and opens that file using os.Open, which returns an io.Reader object that
// can be passed to the YAML/JSON decoder to build up the @resource parameter for usage in the clientsets.
func DecodeYAMLManifestToObject(path string, resource interface{}) error {
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

// InitMeteringCRDSlice initializes a slice of CRD structures, where each
// structure contains information about an individual CRD that metering manages
// and returns this slice of CRD structures.
func InitMeteringCRDSlice(manifestDir string, pathToCRDMap map[string]string) []CRD {
	var crds []CRD

	crds = append(crds, CRD{
		Name: "hivetables.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["hiveTable"]),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "prestotables.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["prestoTable"]),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "storagelocations.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["storageLocation"]),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "reports.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["report"]),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "reportqueries.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["reportQuery"]),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "reportdatasources.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["reportDataSource"]),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})
	crds = append(crds, CRD{
		Name: "meteringconfigs.metering.openshift.io",
		Path: filepath.Join(manifestDir, pathToCRDMap["meteringConfig"]),
		CRD:  new(apiextv1beta1.CustomResourceDefinition),
	})

	return crds
}

func getMeteringAnsiblePath(manifestDir, platform string) (string, error) {
	if manifestDir == "" {
		return "", fmt.Errorf("Failed to set the $DEPLOY_MANIFESTS_DIR or --deploy-manifests-dir flag to a non-empty value")
	}

	deployDir, err := filepath.Abs(manifestDir)
	if err != nil {
		return "", fmt.Errorf("Failed to get the absolute path of the manifest/deploy directory %s: %v", manifestDir, err)
	}

	dirStat, err := os.Stat(manifestDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("Failed to get the stat the manifest/deploy directory %s: %v", manifestDir, err)
	}
	if !dirStat.IsDir() {
		return "", fmt.Errorf("Specified deploy directory '%s' is not a directory", manifestDir)
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
		return "", fmt.Errorf("Failed to set $DEPLOY_PLATFORM or --platform flag to a valid value. Supported platforms: [upstream, openshift, ocp-testing]")
	}

	dirStat, err = os.Stat(ansibleOperatorManifestDir)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("Failed to stat the %s deploy platform directory '%s': %v", platform, ansibleOperatorManifestDir, err)
	}
	if !dirStat.IsDir() {
		return "", fmt.Errorf("Specified %s deploy platform directory '%s' is not a directory", platform, ansibleOperatorManifestDir)
	}

	return ansibleOperatorManifestDir, nil
}

// Resource contains information about an individual resource that metering manages
type Resource struct {
	UseAbsolutePath bool
	SkipResource    bool
	Path            string
	Resource        interface{}
}

// InitObjectFromManifest is the driver function for building up a path to the metering-ansible-operator directory
// from the base @manifestDir directory and initializing the metering resource/CRD objects from the YAML manifests.
// If @meteringConfigCRFile is empty, and the cfg.Resources.MeteringConfig is not nil, then we skip that resource
// when decoding the YAML manifests.
func InitObjectFromManifest(logger logrus.FieldLogger, cfg *Config, manifestDir, meteringConfigCRFile string) error {
	var resources MeteringResources

	ansibleOperatorManifestDir, err := getMeteringAnsiblePath(manifestDir, cfg.Platform)
	if err != nil {
		return fmt.Errorf("Failed to get the path to the metering-ansible-operator directory: %v", err)
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
			return fmt.Errorf("Failed to decode the YAML manifest: %v", err)
		}
	}

	var skipMeteringConfig bool

	// check if @cfg has a non-empty MeteringConfig obj when @meteringConfigCRFile is empty
	if meteringConfigCRFile == "" {
		// need to null check the cfg.Resources structure before null checking the MeteringConfig obj
		if cfg.Resources != nil && cfg.Resources.MeteringConfig != nil {
			resources.MeteringConfig = cfg.Resources.MeteringConfig
			skipMeteringConfig = true
		} else {
			return fmt.Errorf("Failed to set the $METERING_CR_FILE or --meteringconfig flag")
		}
	}

	meteringResourceMap := map[string]Resource{
		"meteringConfig": Resource{
			Path:            meteringConfigCRFile,
			UseAbsolutePath: true,
			SkipResource:    skipMeteringConfig,
			Resource:        &resources.MeteringConfig,
		},
		"deployment": Resource{
			Path:     meteringDeploymentFile,
			Resource: &resources.Deployment,
		},
		"serviceAccount": Resource{
			Path:     meteringServiceAccountFile,
			Resource: &resources.ServiceAccount,
		},
		"roleBinding": Resource{
			Path:     meteringRoleBindingFile,
			Resource: &resources.RoleBinding,
		},
		"role": Resource{
			Path:     meteringRoleFile,
			Resource: &resources.Role,
		},
		"clusterRoleBinding": Resource{
			Path:     meteringClusterRoleBindingFile,
			Resource: &resources.ClusterRoleBinding,
		},
		"clusterRole": Resource{
			Path:     meteringClusterRoleFile,
			Resource: &resources.ClusterRole,
		},
	}

	for name, resource := range meteringResourceMap {
		if resource.SkipResource {
			continue
		}

		path := filepath.Join(ansibleOperatorManifestDir, resource.Path)

		if resource.UseAbsolutePath {
			path = resource.Path
		}

		err = DecodeYAMLManifestToObject(path, resource.Resource)
		if err != nil {
			return fmt.Errorf("Failed to decode the YAML manifest for the %s resource: %v", name, err)
		}
	}

	cfg.Resources = &resources

	return nil
}
