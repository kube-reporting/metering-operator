package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	meteringclientv1 "github.com/operator-framework/operator-metering/pkg/generated/clientset/versioned/typed/metering/v1"
	"github.com/operator-framework/operator-metering/pkg/operator/deploy"
	apiextclientv1beta1 "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	cfg        deploy.Config
	deployType string
	logLevel   string

	rootCmd = &cobra.Command{
		Use:   "deploy-metering",
		Short: "Deploying the metering operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	installCmd = &cobra.Command{
		Use:     "install",
		Short:   "Install the metering operator",
		Example: "operator-metering install --platform upstream",
		PreRun: func(cmd *cobra.Command, args []string) {
			deployType = "install"
		},
		RunE: runDeployMetering,
	}

	uninstallCmd = &cobra.Command{
		Use:     "uninstall",
		Short:   "Uninstall the metering operator",
		Example: "operator-metering uninstall --delete--all",
		PreRun: func(cmd *cobra.Command, args []string) {
			deployType = "uninstall"
		},
		RunE: runDeployMetering,
	}
)

func init() {
	uninstallCmd.Flags().StringVar(&cfg.Namespace, "namespace", "", "The namespace to install the metering resources. This can also be specified through the METERING_NAMESPACE ENV var.")
	uninstallCmd.Flags().StringVar(&cfg.Platform, "platform", "openshift", "The platform to install the metering stack on. Supported options are 'openshift', 'upstream', or 'ocp-testing'. This can also be specified through the DEPLOY_PLATFORM ENV var.")
	uninstallCmd.Flags().StringVar(&cfg.MeteringCR, "meteringconfig", "", "The absolute/relative path to the MeteringConfig custom resource. This can also be specified through the METERING_CR_FILE ENV var")
	uninstallCmd.Flags().StringVar(&cfg.ManifestLocation, "manifest-dir", "", "The absolute/relative path to the metering manfiest directory. This can also be specified through the INSTALLER_MANIFESTS_DIR")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteCRDs, "delete-crd", false, "If true, this would delete the metering CRDs during an uninstall. This can also be specified through the METERING_DELETE_CRDS ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteCRB, "delete-crb", false, "If true, this would delete the metering cluster role bindings during an uninstall. This can also be specified through METERING_DELETE_CRB ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteNamespace, "delete-namespace", false, "If true, this would delete the namespace during an uninstall. This can also be specified through the METERING_DELETE_NAMESPACE ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeletePVCs, "delete-pvc", true, "If true, this would delete the PVCs used by metering resources during an uninstall. This can also be specified through the METERING_DELETE_PVCS ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteAll, "delete-all", false, "If true, this would delete the all metering resources during an uninstall. This can also be specified through the METERING_DELETE_ALL ENV var.")

	installCmd.Flags().StringVar(&cfg.Namespace, "namespace", "", "The namespace to install the metering resources. This can also be specified through the METERING_NAMESPACE ENV var.")
	installCmd.Flags().StringVar(&cfg.Platform, "platform", "openshift", "The platform to install the metering stack on. Supported options are 'openshift', 'upstream', or 'ocp-testing'. This can also be specified through the DEPLOY_PLATFORM ENV var.")
	installCmd.Flags().StringVar(&cfg.MeteringCR, "meteringconfig", "", "The absolute/relative path to the MeteringConfig custom resource. This can also be specified through the METERING_CR_FILE ENV var")
	installCmd.Flags().StringVar(&cfg.DeployManifestsDirectory, "deploy-manifests-dir", "manifests/deploy", "The absolute/relative path to the metering manfiest directory. This can also be specified through the INSTALLER_MANIFESTS_DIR")
	installCmd.Flags().StringVar(&cfg.Repo, "repo", "", "The name of the metering-ansible-operator image repository. This can also be specified through the METERING_OPERATOR_IMAGE_REPO ENV var")
	installCmd.Flags().StringVar(&cfg.Tag, "tag", "", "The name of the metering-ansible-operator image tag. This can also be specified through the METERING_OPERATOR_IMAGE_TAG ENV var")
	installCmd.Flags().BoolVar(&cfg.SkipMeteringDeployment, "skip-metering-operator-deployment", false, "If true, only create the metering namespace, CRDs, and MeteringConfig resources. This can also be specified through the SKIP_METERING_OPERATOR_DEPLOY ENV var.")

	installEnv := map[string]string{
		"METERING_NAMESPACE":                "namespace",
		"DEPLOY_PLATFORM":                   "platform",
		"METERING_CR_FILE":                  "meteringconfig",
		"SKIP_METERING_OPERATOR_DEPLOYMENT": "skip-metering-operator-deployment",
		"DEPLOY_MANIFESTS_DIR":              "deploy-manifests-dir",
		"METERING_OPERATOR_IMAGE_REPO":      "repo",
		"METERING_OPERATOR_IMAGE_TAG":       "tag",
	}

	err := mapEnvVarToFlag(installEnv, installCmd.Flags())
	if err != nil {
		log.WithError(err).Fatalf("Failed to update flags from ENV vars: %v", err)
	}

	uninstallEnv := map[string]string{
		"METERING_NAMESPACE":        "namespace",
		"DEPLOY_PLATFORM":           "platform",
		"METERING_CR_FILE":          "meteringconfig",
		"DEPLOY_MANIFESTS_DIR":      "deploy-manifests-dir",
		"METERING_DELETE_CRB":       "delete-crb",
		"METERING_DELETE_CRDS":      "delete-crd",
		"METERING_DELETE_PVCS":      "delete-pvc",
		"METERING_DELETE_NAMESPACE": "delete-namespace",
		"METERING_DELETE_ALL":       "delete-all",
	}

	err = mapEnvVarToFlag(uninstallEnv, uninstallCmd.Flags())
	if err != nil {
		log.WithError(err).Fatalf("Failed to update flags from ENV vars (uninstall): %v", err)
	}
}

func main() {
	rootCmd.AddCommand(installCmd, uninstallCmd)

	err := rootCmd.Execute()
	if err != nil {
		log.WithError(err).Fatalf("Failed to deploy metering: %v", err)
	}
}

func runDeployMetering(cmd *cobra.Command, args []string) error {
	var err error

	logger := setupLogger(logLevel)

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return fmt.Errorf("Failed to initialize the kubernetes client config: %v", err)
	}

	client, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("Failed to initialize the kubernetes clientset: %v", err)
	}

	apiextClient, err := apiextclientv1beta1.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("Failed to initialize the apiextensions clientset: %v", err)
	}

	meteringClient, err := meteringclientv1.NewForConfig(restconfig)
	if err != nil {
		return fmt.Errorf("Failed to initialize the metering clientset: %v", err)
	}

	logger.Debugf("Metering Deploy Config: %#v", cfg)

	deployObj, err := deploy.NewDeployer(cfg, client, apiextClient, meteringClient, logger)
	if err != nil {
		return fmt.Errorf("Failed to deploy metering: %v", err)
	}

	if deployType == "install" {
		err := deployObj.Install()
		if err != nil {
			return fmt.Errorf("Failed to install metering: %v", err)
		}
		logger.Infof("Finished installing metering")
	} else if deployType == "uninstall" {
		err := deployObj.Uninstall()
		if err != nil {
			return fmt.Errorf("Failed to uninstall metering: %v", err)
		}
		logger.Infof("Finished uninstall metering")
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

// mapEnvVarToFlag takes a mapping of ENV var names to flag names and iterates
// over that mapping attempting to set the flag value with the ENV var key name.
// see: https://github.com/spf13/viper/issues/461
func mapEnvVarToFlag(vars map[string]string, flagset *pflag.FlagSet) error {
	for env, flag := range vars {
		flagObj := flagset.Lookup(flag)
		if flagObj == nil {
			return fmt.Errorf("The %s flag doesn't exist", flag)
		}

		if val := os.Getenv(env); val != "" {
			if err := flagObj.Value.Set(val); err != nil {
				return fmt.Errorf("Failed to set the %s flag: %v", flag, err)
			}
		}

	}

	return nil
}
