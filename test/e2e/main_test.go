package e2e

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
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

	runAWSBillingTests         bool
	meteringOperatorImageRepo  string
	meteringOperatorImageTag   string
	reportingOperatorImageRepo string
	reportingOperatorImageTag  string

	defaultTargetPods = 7
)

func init() {
	runAWSBillingTests = os.Getenv("ENABLE_AWS_BILLING_TESTS") == "true"
	meteringOperatorImageRepo = os.Getenv("METERING_OPERATOR_IMAGE_REPO")
	meteringOperatorImageTag = os.Getenv("METERING_OPERATOR_IMAGE_TAG")
	reportingOperatorImageRepo = os.Getenv("REPORTING_OPERATOR_IMAGE_REPO")
	reportingOperatorImageTag = os.Getenv("REPORTING_OPERATOR_IMAGE_TAG")
}

func TestMain(m *testing.M) {
	kubeConfigFlag := flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
	nsPrefixFlag := flag.String("namespace-prefix", "", "The namespace prefix to install the metering resources.")
	manifestDirFlag := flag.String("deploy-manifests-dir", "../../manifests/deploy", "The absolute/relative path to the metering manifest directory.")
	cleanupScriptPathFlag := flag.String("cleanup-script-path", "../../hack/run-test-cleanup.sh", "The absolute/relative path to the testing cleanup hack script.")
	testOutputPathFlag := flag.String("test-output-path", "", "The absolute/relative path that you want to store test logs within.")
	logLevelFlag := flag.String("log-level", logrus.DebugLevel.String(), "The log level")
	flag.Parse()

	logger := testhelpers.SetupLogger(*logLevelFlag)

	var err error
	if df, err = deployframework.New(
		logger,
		*nsPrefixFlag,
		*manifestDirFlag,
		*kubeConfigFlag,
		*cleanupScriptPathFlag,
		*testOutputPathFlag,
		reportingOperatorImageRepo,
		reportingOperatorImageTag,
	); err != nil {
		logger.Fatalf("Failed to create a new deploy framework: %v", err)
	}

	os.Exit(m.Run())
}

func TestInstallMeteringAndReportingProducesData(t *testing.T) {
	testInstallConfigs := []struct {
		TargetPods                int
		Name                      string
		MeteringOperatorImageRepo string
		MeteringOperatorImageTag  string
		MeteringConfigSpec        metering.MeteringConfigSpec
	}{
		{
			Name:                      "HDFSInstall",
			TargetPods:                defaultTargetPods,
			MeteringOperatorImageRepo: meteringOperatorImageRepo,
			MeteringOperatorImageTag:  meteringOperatorImageTag,
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
						Image: &metering.ImageConfig{
							Repository: reportingOperatorImageRepo,
							Tag:        reportingOperatorImageTag,
						},
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
	}

	for _, testCase := range testInstallConfigs {
		t := t
		testCase := testCase

		t.Run(testCase.Name, func(t *testing.T) {
			testInstall(t, testCase.MeteringOperatorImageRepo, testCase.MeteringOperatorImageTag, testCase.Name, testCase.TargetPods, testCase.MeteringConfigSpec)
		})
	}
}

func testInstall(
	t *testing.T,
	meteringOperatorImageRepo,
	meteringOperatorImageTag,
	testName string,
	targetPods int,
	spec metering.MeteringConfigSpec,
) {
	// create a directory used to store the @testName container and resource logs
	testOutputDir := filepath.Join(df.LoggingPath, testName)
	err := os.Mkdir(testOutputDir, 0777)
	require.NoError(t, err, "creating the test case output directory should produce no error")

	// randomize the namespace to avoid existing namespaces
	rand.Seed(time.Now().UnixNano())
	namespace := df.NamespacePrefix + "-" + strconv.Itoa(rand.Intn(50))

	deployerCtx, err := df.NewDeployerCtx(meteringOperatorImageRepo, meteringOperatorImageTag, namespace, testOutputDir, targetPods, spec)
	require.NoError(t, err, "creating a new deployer context should produce no error")

	rf, err := deployerCtx.Setup()
	assert.NoError(t, err, "deploying metering should produce no error")

	defer func() {
		err := deployerCtx.Teardown(df.CleanupScriptPath)
		assert.NoError(t, err, "capturing logs and uninstalling metering should produce no error")
	}()

	// note: in order to respect the defer closure and cleanup responsibly, we need
	// to avoid returning errors or ending the function early as testInstall could
	// be run in a goroutine - so check if the cfg object is nil to avoid a panic
	if rf != nil {
		// note: we need to run the testReportingProducesData as a group in order
		// to avoid preemptively running the defer closure early and uninstalling
		// the metering resources during the middle of the test execution
		// for more information, see:
		// https://github.com/golang/go/issues/22993
		// https://github.com/golang/go/issues/31651
		t.Run("testReportingProducesData", func(t *testing.T) {
			testReportingProducesData(t, rf)
		})
	}
}

func testReportingProducesData(t *testing.T, testReportingFramework *reportingframework.ReportingFramework) {
	// cron schedule to run every minute
	cronSchedule := &metering.ReportSchedule{
		Period: metering.ReportPeriodCron,
		Cron: &metering.ReportScheduleCron{
			Expression: fmt.Sprintf("*/1 * * * *"),
		},
	}

	queries := []struct {
		queryName   string
		skip        bool
		nonParallel bool
	}{
		{queryName: "namespace-cpu-request"},
		{queryName: "namespace-cpu-usage"},
		{queryName: "namespace-memory-request"},
		{queryName: "namespace-persistentvolumeclaim-request"},
		{queryName: "namespace-persistentvolumeclaim-usage"},
		{queryName: "namespace-memory-usage"},
		{queryName: "persistentvolumeclaim-usage"},
		{queryName: "persistentvolumeclaim-capacity"},
		{queryName: "persistentvolumeclaim-request"},
		{queryName: "pod-cpu-request"},
		{queryName: "pod-cpu-usage"},
		{queryName: "pod-memory-request"},
		{queryName: "pod-memory-usage"},
		{queryName: "node-cpu-utilization"},
		{queryName: "node-memory-utilization"},
		{queryName: "cluster-persistentvolumeclaim-request"},
		{queryName: "cluster-cpu-capacity"},
		{queryName: "cluster-memory-capacity"},
		{queryName: "cluster-cpu-usage"},
		{queryName: "cluster-memory-usage"},
		{queryName: "cluster-cpu-utilization"},
		{queryName: "cluster-memory-utilization"},
		{queryName: "namespace-memory-utilization"},
		{queryName: "namespace-cpu-utilization"},
		{queryName: "pod-cpu-request-aws", skip: !runAWSBillingTests, nonParallel: true},
		{queryName: "pod-memory-request-aws", skip: !runAWSBillingTests, nonParallel: true},
		{queryName: "aws-ec2-cluster-cost", skip: !runAWSBillingTests, nonParallel: true},
	}

	var reportsProduceDataTestCases []reportProducesDataTestCase

	for _, query := range queries {
		reportcronTestCase := reportProducesDataTestCase{
			name:          query.queryName + "-cron",
			queryName:     query.queryName,
			schedule:      cronSchedule,
			newReportFunc: testReportingFramework.NewSimpleReport,
			skip:          query.skip,
			parallel:      !query.nonParallel,
		}
		reportRunOnceTestCase := reportProducesDataTestCase{
			name:          query.queryName + "-runonce",
			queryName:     query.queryName,
			schedule:      nil, // runOnce
			newReportFunc: testReportingFramework.NewSimpleReport,
			skip:          query.skip,
			parallel:      !query.nonParallel,
		}

		reportsProduceDataTestCases = append(reportsProduceDataTestCases, reportcronTestCase, reportRunOnceTestCase)
	}

	testReportsProduceData(t, testReportingFramework, reportsProduceDataTestCases)
}
