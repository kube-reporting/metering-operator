package e2e

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	metering "github.com/operator-framework/operator-metering/pkg/apis/metering/v1"
	"github.com/operator-framework/operator-metering/test/deployframework"
	"github.com/operator-framework/operator-metering/test/reportingframework"
	"github.com/operator-framework/operator-metering/test/testhelpers"
)

var (
	df *deployframework.DeployFramework

	kubeConfig string
	logLevel   string

	deployManifestsDir         string
	meteringOperatorImageRepo  string
	meteringOperatorImageTag   string
	reportingOperatorImageRepo string
	reportingOperatorImageTag  string
	namespacePrefix            string
	cleanupScriptPath          string
	testOutputPath             string

	defaultTargetPods          = 7
	kubeNamespaceCharLimit     = 63
	meteringconfigMetadataName = "operator-metering"
)

func init() {
	runAWSBillingTests = os.Getenv("ENABLE_AWS_BILLING_TESTS") == "true"

	meteringOperatorImageRepo = os.Getenv("METERING_OPERATOR_IMAGE_REPO")
	meteringOperatorImageTag = os.Getenv("METERING_OPERATOR_IMAGE_TAG")
	reportingOperatorImageRepo = os.Getenv("REPORTING_OPERATOR_IMAGE_REPO")
	reportingOperatorImageTag = os.Getenv("REPORTING_OPERATOR_IMAGE_TAG")
}

func TestMain(m *testing.M) {
	flag.StringVar(&kubeConfig, "kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	flag.StringVar(&logLevel, "log-level", logrus.DebugLevel.String(), "The log level")

	flag.StringVar(&deployManifestsDir, "deploy-manifests-dir", "../../manifests/deploy", "The absolute/relative path to the metering manifest directory.")
	flag.BoolVar(&runAWSBillingTests, "run-aws-billing-tests", runAWSBillingTests, "")

	flag.StringVar(&meteringOperatorImageRepo, "metering-operator-image-repo", meteringOperatorImageRepo, "")
	flag.StringVar(&meteringOperatorImageTag, "metering-operator-image-tag", meteringOperatorImageTag, "")
	flag.StringVar(&reportingOperatorImageRepo, "reporting-operator-image-repo", reportingOperatorImageRepo, "")
	flag.StringVar(&reportingOperatorImageTag, "reporting-operator-image-tag", reportingOperatorImageTag, "")

	flag.StringVar(&namespacePrefix, "namespace-prefix", "", "The namespace prefix to install the metering resources.")
	flag.StringVar(&cleanupScriptPath, "cleanup-script-path", "../../hack/run-test-cleanup.sh", "The absolute/relative path to the testing cleanup hack script.")
	flag.StringVar(&testOutputPath, "test-output-path", "", "The absolute/relative path that you want to store test logs within.")
	flag.Parse()

	logger := testhelpers.SetupLogger(logLevel)

	var err error
	if df, err = deployframework.New(logger, namespacePrefix, deployManifestsDir, kubeConfig); err != nil {
		logger.Fatalf("Failed to create a new deploy framework: %v", err)
	}

	os.Exit(m.Run())
}

type InstallTestCase struct {
	Name     string
	TestFunc func(t *testing.T, testReportingFramework *reportingframework.ReportingFramework)
}

func TestManualMeteringInstall(t *testing.T) {
	testInstallConfigs := []struct {
		TargetPods                int
		Skip                      bool
		Name                      string
		MeteringOperatorImageRepo string
		MeteringOperatorImageTag  string
		MeteringConfigSpec        metering.MeteringConfigSpec
		InstallSubTest            InstallTestCase
	}{
		{
			Name:                      "HDFSInstallTestReportingProducesData",
			TargetPods:                defaultTargetPods,
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			Skip:                      false,
			InstallSubTest: InstallTestCase{
				Name:     "testReportingProducesData",
				TestFunc: testReportingProducesData,
			},
			MeteringConfigSpec: metering.MeteringConfigSpec{
				LogHelmTemplate: testhelpers.PtrToBool(true),
				UnsupportedFeatures: &metering.UnsupportedFeaturesConfig{
					EnableHDFS: testhelpers.PtrToBool(true),
				},
				Storage: &metering.StorageConfig{
					Type: "hive",
					Hive: &metering.HiveStorageConfig{
						Type: "hdfs",
						Hdfs: &metering.HiveHDFSConfig{
							Namenode: "hdfs-namenode-0.hdfs-namenode:9820",
						},
					},
				},
				ReportingOperator: &metering.ReportingOperator{
					Spec: &metering.ReportingOperatorSpec{
						Resources: &v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("1"),
								v1.ResourceMemory: resource.MustParse("250Mi"),
							},
						},
						Image: &metering.ImageConfig{},
						Config: &metering.ReportingOperatorConfig{
							LogLevel: "debug",
							Prometheus: &metering.ReportingOperatorPrometheusConfig{
								MetricsImporter: &metering.ReportingOperatorPrometheusMetricsImporterConfig{
									Config: &metering.ReportingOperatorPrometheusMetricsImporterConfigSpec{
										ChunkSize:                 &meta.Duration{Duration: 5 * time.Minute},
										PollInterval:              &meta.Duration{Duration: 30 * time.Second},
										StepSize:                  &meta.Duration{Duration: 1 * time.Minute},
										MaxImportBackfillDuration: &meta.Duration{Duration: 15 * time.Minute},
										MaxQueryRangeDuration:     &meta.Duration{Duration: 5 * time.Minute},
									},
								},
							},
						},
					},
				},
				Presto: &metering.Presto{
					Spec: &metering.PrestoSpec{
						Coordinator: &metering.PrestoCoordinatorSpec{
							Resources: &v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
				Hive: &metering.Hive{
					Spec: &metering.HiveSpec{
						Metastore: &metering.HiveMetastoreSpec{
							Resources: &v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("650Mi"),
								},
							},
							Storage: &metering.HiveMetastoreStorageConfig{
								Size: "5Gi",
							},
						},
						Server: &metering.HiveServerSpec{
							Resources: &v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("650Mi"),
								},
							},
						},
					},
				},
				Hadoop: &metering.Hadoop{
					Spec: &metering.HadoopSpec{
						HDFS: &metering.HadoopHDFS{
							Enabled: testhelpers.PtrToBool(true),
							Datanode: &metering.HadoopHDFSDatanodeSpec{
								Resources: &v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("500Mi"),
									},
								},
								Storage: &metering.HadoopHDFSStorageConfig{
									Size: "5Gi",
								},
							},
							Namenode: &metering.HadoopHDFSNamenodeSpec{
								Resources: &v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("500Mi"),
									},
								},
								Storage: &metering.HadoopHDFSStorageConfig{
									Size: "5Gi",
								},
							},
						},
					},
				},
			},
		},
		{
			Name:                      "HDFSInstallTestReportingProducesCorrectDataForInput",
			TargetPods:                defaultTargetPods,
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
			Skip:                      false,
			InstallSubTest: InstallTestCase{
				Name:     "testReportingProducesCorrectDataForInput",
				TestFunc: testReportingProducesCorrectDataForInput,
			},
			MeteringConfigSpec: metering.MeteringConfigSpec{
				LogHelmTemplate: testhelpers.PtrToBool(true),
				UnsupportedFeatures: &metering.UnsupportedFeaturesConfig{
					EnableHDFS: testhelpers.PtrToBool(true),
				},
				Storage: &metering.StorageConfig{
					Type: "hive",
					Hive: &metering.HiveStorageConfig{
						Type: "hdfs",
						Hdfs: &metering.HiveHDFSConfig{
							Namenode: "hdfs-namenode-0.hdfs-namenode:9820",
						},
					},
				},
				ReportingOperator: &metering.ReportingOperator{
					Spec: &metering.ReportingOperatorSpec{
						Resources: &v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    resource.MustParse("1"),
								v1.ResourceMemory: resource.MustParse("250Mi"),
							},
						},
						Image: &metering.ImageConfig{},
						Config: &metering.ReportingOperatorConfig{
							LogLevel: "debug",
							Prometheus: &metering.ReportingOperatorPrometheusConfig{
								MetricsImporter: &metering.ReportingOperatorPrometheusMetricsImporterConfig{
									Enabled: testhelpers.PtrToBool(false),
								},
							},
						},
					},
				},
				Presto: &metering.Presto{
					Spec: &metering.PrestoSpec{
						Coordinator: &metering.PrestoCoordinatorSpec{
							Resources: &v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
				Hive: &metering.Hive{
					Spec: &metering.HiveSpec{
						Metastore: &metering.HiveMetastoreSpec{
							Resources: &v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("650Mi"),
								},
							},
							Storage: &metering.HiveMetastoreStorageConfig{
								Size: "5Gi",
							},
						},
						Server: &metering.HiveServerSpec{
							Resources: &v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("650Mi"),
								},
							},
						},
					},
				},
				Hadoop: &metering.Hadoop{
					Spec: &metering.HadoopSpec{
						HDFS: &metering.HadoopHDFS{
							Enabled: testhelpers.PtrToBool(true),
							Datanode: &metering.HadoopHDFSDatanodeSpec{
								Resources: &v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("500Mi"),
									},
								},
								Storage: &metering.HadoopHDFSStorageConfig{
									Size: "5Gi",
								},
							},
							Namenode: &metering.HadoopHDFSNamenodeSpec{
								Resources: &v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("500Mi"),
									},
								},
								Storage: &metering.HadoopHDFSStorageConfig{
									Size: "5Gi",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testInstallConfigs {
		t := t
		testCase := testCase

		if !testCase.Skip {
			t.Run(testCase.Name, func(t *testing.T) {
				testInstall(
					t,
					testCase.Name,
					namespacePrefix,
					testCase.MeteringOperatorImageRepo,
					testCase.MeteringOperatorImageTag,
					testOutputPath,
					cleanupScriptPath,
					testCase.TargetPods,
					testCase.InstallSubTest,
					testCase.MeteringConfigSpec,
				)
			})
		}
	}
}

func testInstall(
	t *testing.T,
	testCaseName,
	namespacePrefix,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	testOuputPath,
	cleanupScriptPath string,
	testTargetPods int,
	testInstallFunction InstallTestCase,
	testMeteringConfigSpec metering.MeteringConfigSpec,
) {
	// create a directory used to store the @testCaseName container and resource logs
	testCaseOutputBaseDir := filepath.Join(testOuputPath, testCaseName)
	err := os.Mkdir(testCaseOutputBaseDir, 0777)
	assert.NoError(t, err, "creating the test case output directory should produce no error")

	rand.Seed(time.Now().UnixNano())

	// randomize the namespace to avoid existing namespaces
	testFuncNamespace := fmt.Sprintf("%s-%s-%d", namespacePrefix, strings.ToLower(testCaseName), rand.Intn(50))
	if len(testFuncNamespace) > kubeNamespaceCharLimit {
		df.Logger.Infof("The test function namespace exceeded the %d kube namespace character limit, retrying without the test case name", kubeNamespaceCharLimit)
		testFuncNamespace = fmt.Sprintf("%s-%d", namespacePrefix, rand.Intn(50))

		// if the length of the truncated namespace is still too long, fail the test and continue onto the next test function iteration
		if len(testFuncNamespace) > kubeNamespaceCharLimit {
			require.Fail(t, "the length of the test function namespace exceeds the kube namespace limit")
		}
	}

	deployerCtx, err := df.NewDeployerCtx(
		testFuncNamespace,
		meteringOperatorImageRepo,
		meteringOperatorImageTag,
		reportingOperatorImageRepo,
		reportingOperatorImageTag,
		testMeteringConfigSpec,
		testCaseOutputBaseDir,
		testTargetPods,
	)
	require.NoError(t, err, "creating a new deployer context should produce no error")

	rf, err := deployerCtx.Setup()
	assert.Nil(t, err, "expected there would be no error installing and setting up the metering stack")

	if rf != nil {
		t.Run(testInstallFunction.Name, func(t *testing.T) {
			testInstallFunction.TestFunc(t, rf)
		})
	}

	err = deployerCtx.Teardown(cleanupScriptPath)
	assert.NoError(t, err, "capturing logs and uninstalling metering should produce no error")
}
