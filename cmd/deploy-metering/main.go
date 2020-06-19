package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	meteringclient "github.com/kube-reporting/metering-operator/pkg/generated/clientset/versioned/typed/metering/v1"
	olmclientv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1"
	olmclientv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/typed/operators/v1alpha1"
	apiextclientv1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kube-reporting/metering-operator/pkg/deploy"
)

var (
	cfg                  deploy.Config
	deployType           string
	logLevel             string
	meteringConfigCRFile string
	deployManifestsDir   string

	rootCmd = &cobra.Command{
		Use:   "deploy-metering",
		Short: "Deploying the metering operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	olmInstallCmd = &cobra.Command{
		Use:     "olm-install",
		Short:   "Install the metering operator through OLM",
		Example: "./bin/deploy-metering olm-install --channel 4.3",
		PreRun: func(cmd *cobra.Command, args []string) {
			deployType = "olm-install"
		},
		RunE: runDeployMetering,
	}

	installCmd = &cobra.Command{
		Use:     "install",
		Short:   "Install the metering operator",
		Example: "./bin/deploy-metering install --platform upstream",
		PreRun: func(cmd *cobra.Command, args []string) {
			deployType = "install"
		},
		RunE: runDeployMetering,
	}

	olmUninstallCmd = &cobra.Command{
		Use:     "olm-uninstall",
		Short:   "Uninstall the metering operator and OLM resources",
		Example: "./bin/deploy-metering olm-uninstall --delete-namespace",
		PreRun: func(cmd *cobra.Command, args []string) {
			deployType = "olm-uninstall"
		},
		RunE: runDeployMetering,
	}

	uninstallCmd = &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstall the metering operator",
		Example: "./bin/deploy-metering uninstall --delete--all",
		PreRun: func(cmd *cobra.Command, args []string) {
			deployType = "uninstall"
		},
		RunE: runDeployMetering,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfg.Namespace, "namespace", "", "The namespace to install the metering resources. This can also be specified through the METERING_NAMESPACE ENV var.")
	rootCmd.PersistentFlags().StringVar(&cfg.Platform, "platform", "openshift", "The platform to install the metering stack on. Supported options are 'openshift', 'upstream', or 'ocp-testing'. This can also be specified through the DEPLOY_PLATFORM ENV var.")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", log.InfoLevel.String(), "The logging level when deploying metering")
	rootCmd.PersistentFlags().StringVar(&meteringConfigCRFile, "meteringconfig", "", "The absolute/relative path to the MeteringConfig custom resource. This can also be specified through the METERING_CR_FILE ENV var.")
	rootCmd.PersistentFlags().StringVar(&deployManifestsDir, "deploy-manifests-dir", "manifests/deploy", "The absolute/relative path to the metering manifest directory. This can also be specified through the INSTALLER_MANIFESTS_DIR.")

	uninstallCmd.Flags().BoolVar(&cfg.DeleteCRDs, "delete-crd", false, "If true, this would delete the metering CRDs during an uninstall. This can also be specified through the METERING_DELETE_CRDS ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteCRB, "delete-crb", false, "If true, this would delete the metering cluster role bindings during an uninstall. This can also be specified through METERING_DELETE_CRB ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteNamespace, "delete-namespace", false, "If true, this would delete the namespace during an uninstall. This can also be specified through the METERING_DELETE_NAMESPACE ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeletePVCs, "delete-pvc", true, "If true, this would delete the PVCs used by metering resources during an uninstall. This can also be specified through the METERING_DELETE_PVCS ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteAll, "delete-all", false, "If true, this would delete the all metering resources during an uninstall. This can also be specified through the METERING_DELETE_ALL ENV var.")

	olmUninstallCmd.Flags().BoolVar(&cfg.DeleteCRDs, "delete-crd", false, "If true, this would delete the metering CRDs during an uninstall. This can also be specified through the METERING_DELETE_CRDS ENV var.")
	olmUninstallCmd.Flags().BoolVar(&cfg.DeleteNamespace, "delete-namespace", false, "If true, this would delete the namespace during an uninstall. This can also be specified through the METERING_DELETE_NAMESPACE ENV var.")

	installCmd.Flags().StringVar(&cfg.Repo, "repo", "", "The name of the metering-ansible-operator image repository. This can also be specified through the METERING_OPERATOR_IMAGE_REPO ENV var.")
	installCmd.Flags().StringVar(&cfg.Tag, "tag", "", "The name of the metering-ansible-operator image tag. This can also be specified through the METERING_OPERATOR_IMAGE_TAG ENV var.")
	installCmd.Flags().BoolVar(&cfg.SkipMeteringDeployment, "skip-metering-operator-deployment", false, "If true, only create the metering namespace, CRDs, and MeteringConfig resources. This can also be specified through the SKIP_METERING_OPERATOR_DEPLOY ENV var.")
	installCmd.Flags().BoolVar(&cfg.RunMeteringOperatorLocal, "run-metering-operator-local", false, "If true, skip installing the metering deployment. This can also be specified through the $SKIP_METERING_OPERATOR_DEPLOYMENT ENV var.")

	olmInstallCmd.Flags().StringVar(&cfg.SubscriptionName, "subscription", "metering-ocp", "The name of the metering subscription that gets created.")
	olmInstallCmd.Flags().StringVar(&cfg.Channel, "channel", "4.4", "The metering channel to subscribe to. Examples: 4.2, 4.3, 4.4, etc.")

	if err := initFlagsFromEnv(); err != nil {
		log.WithError(err).Fatalf("Failed to update flags from ENV vars: %v", err)
	}
}

func main() {
	rootCmd.AddCommand(installCmd, olmInstallCmd, uninstallCmd, olmUninstallCmd)

	err := rootCmd.Execute()
	if err != nil {
		log.WithError(err).Fatalf("Failed to deploy metering: %v", err)
	}
}

func runDeployMetering(cmd *cobra.Command, args []string) error {
	logger := setupLogger(logLevel)

	if meteringConfigCRFile == "" {
		return fmt.Errorf("failed to set the $METERING_CR_FILE or --meteringconfig flag")
	}

	err := deploy.DecodeYAMLManifestToObject(meteringConfigCRFile, &cfg.MeteringConfig)
	if err != nil {
		return fmt.Errorf("failed to read MeteringConfig: %v", err)
	}

	cfg.OperatorResources, err = deploy.ReadMeteringAnsibleOperatorManifests(deployManifestsDir, cfg.Platform)
	if err != nil {
		return fmt.Errorf("failed to read metering-ansible-operator manifests: %v", err)
	}

	logger.Debugf("Metering Deploy Config: %#v", cfg)

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize the kubernetes client config: %v", err)
	}

	client, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("failed to initialize the kubernetes clientset: %v", err)
	}

	apiextClient, err := apiextclientv1.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("failed to initialize the apiextensions clientset: %v", err)
	}

	meteringClient, err := meteringclient.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("failed to initialize the metering clientset: %v", err)
	}

	olmV1Client, err := olmclientv1.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("failed to initialize the v1 OLM clientset: %v", err)
	}

	olmV1Alpha1Client, err := olmclientv1alpha1.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("failed to initialize the v1alpha OLM clientset: %v", err)
	}

	deployObj, err := deploy.NewDeployer(cfg, logger, client, apiextClient, meteringClient, olmV1Client, olmV1Alpha1Client)
	if err != nil {
		return fmt.Errorf("failed to deploy metering: %v", err)
	}

	switch deployType {
	case "install":
		err := deployObj.Install()
		if err != nil {
			return fmt.Errorf("failed to install metering: %v", err)
		}
		logger.Infof("Finished installing metering")

	case "olm-install":
		err := deployObj.InstallOLM()
		if err != nil {
			return fmt.Errorf("failed to install metering through OLM: %v", err)
		}
		logger.Infof("Finished installing metering through OLM")

	case "uninstall":
		err := deployObj.Uninstall()
		if err != nil {
			return fmt.Errorf("failed to uninstall metering: %v", err)
		}
		logger.Infof("Finished uninstall metering")

	case "olm-uninstall":
		err := deployObj.UninstallOLM()
		if err != nil {
			return fmt.Errorf("failed to uninstall metering OLM resources: %v", err)
		}
		logger.Infof("Finished uninstalling metering OLM resources")

	default:
		return fmt.Errorf("invalid deployType encountered: %v", deployType)
	}

	return nil
}

func setupLogger(logLevelStr string) log.FieldLogger {
	var err error

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "01-02-2006 15:04:05",
	})

	logger := log.WithFields(log.Fields{
		"app": "deploy",
	})
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		logger.WithError(err).Fatalf("invalid log level: %s", logLevel)
	}
	logger.Infof("Setting the log level to %s", logLevel.String())
	logger.Logger.Level = logLevel

	return logger
}

func initFlagsFromEnv() error {
	flagEnvConf := []struct {
		cmd *cobra.Command
		env map[string]string
	}{
		{
			cmd: rootCmd,
			env: map[string]string{
				"METERING_NAMESPACE":        "namespace",
				"DEPLOY_PLATFORM":           "platform",
				"METERING_CR_FILE":          "meteringconfig",
				"DEPLOY_MANIFESTS_DIR":      "deploy-manifests-dir",
				"METERING_DEPLOY_LOG_LEVEL": "log-level",
			},
		},
		{
			cmd: uninstallCmd,
			env: map[string]string{
				"METERING_DELETE_CRB":       "delete-crb",
				"METERING_DELETE_CRDS":      "delete-crd",
				"METERING_DELETE_PVCS":      "delete-pvc",
				"METERING_DELETE_NAMESPACE": "delete-namespace",
				"METERING_DELETE_ALL":       "delete-all",
			},
		},
		{
			cmd: installCmd,
			env: map[string]string{
				"SKIP_METERING_OPERATOR_DEPLOYMENT": "skip-metering-operator-deployment",
				"METERING_OPERATOR_IMAGE_REPO":      "repo",
				"METERING_OPERATOR_IMAGE_TAG":       "tag",
				"DEPLOY_METERING_OPERATOR_LOCAL":    "run-metering-operator-local",
			},
		},
	}

	for _, flagConf := range flagEnvConf {
		flagSet := flagConf.cmd.Flags()

		if flagConf.cmd.HasPersistentFlags() {
			flagSet = flagConf.cmd.PersistentFlags()
		}

		if err := mapEnvVarToFlag(flagConf.env, flagSet); err != nil {
			log.WithError(err).Fatalf("Failed to update flags from ENV vars (%s): %v", flagConf.cmd.Name(), err)
		}
	}

	return nil
}

// mapEnvVarToFlag takes a mapping of ENV var names to flag names and iterates
// over that mapping attempting to set the flag value with the ENV var key name.
// see: https://github.com/spf13/viper/issues/461
func mapEnvVarToFlag(vars map[string]string, flagset *pflag.FlagSet) error {
	for env, flag := range vars {
		flagObj := flagset.Lookup(flag)
		if flagObj == nil {
			return fmt.Errorf("the %s flag doesn't exist", flag)
		}

		if val := os.Getenv(env); val != "" {
			if err := flagObj.Value.Set(val); err != nil {
				return fmt.Errorf("failed to set the %s flag: %v", flag, err)
			}
		}

	}

	return nil
}
