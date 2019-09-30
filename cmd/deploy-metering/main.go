package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/operator-framework/operator-metering/pkg/operator/deploy"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	cfg                  deploy.Con***REMOVED***g
	deployType           string
	logLevel             string
	meteringCon***REMOVED***gCRFile string
	deployManifestsDir   string

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
	rootCmd.PersistentFlags().StringVar(&cfg.Namespace, "namespace", "", "The namespace to install the metering resources. This can also be speci***REMOVED***ed through the METERING_NAMESPACE ENV var.")
	rootCmd.PersistentFlags().StringVar(&cfg.Platform, "platform", "openshift", "The platform to install the metering stack on. Supported options are 'openshift', 'upstream', or 'ocp-testing'. This can also be speci***REMOVED***ed through the DEPLOY_PLATFORM ENV var.")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", log.InfoLevel.String(), "The logging level when deploying metering")
	rootCmd.PersistentFlags().StringVar(&meteringCon***REMOVED***gCRFile, "meteringcon***REMOVED***g", "", "The absolute/relative path to the MeteringCon***REMOVED***g custom resource. This can also be speci***REMOVED***ed through the METERING_CR_FILE ENV var.")
	rootCmd.PersistentFlags().StringVar(&deployManifestsDir, "deploy-manifests-dir", "manifests/deploy", "The absolute/relative path to the metering manifest directory. This can also be speci***REMOVED***ed through the INSTALLER_MANIFESTS_DIR.")

	uninstallCmd.Flags().BoolVar(&cfg.DeleteCRDs, "delete-crd", false, "If true, this would delete the metering CRDs during an uninstall. This can also be speci***REMOVED***ed through the METERING_DELETE_CRDS ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteCRB, "delete-crb", false, "If true, this would delete the metering cluster role bindings during an uninstall. This can also be speci***REMOVED***ed through METERING_DELETE_CRB ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteNamespace, "delete-namespace", false, "If true, this would delete the namespace during an uninstall. This can also be speci***REMOVED***ed through the METERING_DELETE_NAMESPACE ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeletePVCs, "delete-pvc", true, "If true, this would delete the PVCs used by metering resources during an uninstall. This can also be speci***REMOVED***ed through the METERING_DELETE_PVCS ENV var.")
	uninstallCmd.Flags().BoolVar(&cfg.DeleteAll, "delete-all", false, "If true, this would delete the all metering resources during an uninstall. This can also be speci***REMOVED***ed through the METERING_DELETE_ALL ENV var.")

	installCmd.Flags().StringVar(&cfg.Repo, "repo", "", "The name of the metering-ansible-operator image repository. This can also be speci***REMOVED***ed through the METERING_OPERATOR_IMAGE_REPO ENV var.")
	installCmd.Flags().StringVar(&cfg.Tag, "tag", "", "The name of the metering-ansible-operator image tag. This can also be speci***REMOVED***ed through the METERING_OPERATOR_IMAGE_TAG ENV var.")
	installCmd.Flags().BoolVar(&cfg.SkipMeteringDeployment, "skip-metering-operator-deployment", false, "If true, only create the metering namespace, CRDs, and MeteringCon***REMOVED***g resources. This can also be speci***REMOVED***ed through the SKIP_METERING_OPERATOR_DEPLOY ENV var.")

	if err := initFlagsFromEnv(); err != nil {
		log.WithError(err).Fatalf("Failed to update flags from ENV vars: %v", err)
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
	logger := setupLogger(logLevel)

	if meteringCon***REMOVED***gCRFile == "" {
		return fmt.Errorf("Failed to set the $METERING_CR_FILE or --meteringcon***REMOVED***g flag")
	}

	err := deploy.InitObjectFromManifest(deployManifestsDir, meteringCon***REMOVED***gCRFile, &cfg)
	if err != nil {
		return fmt.Errorf("Failed to initialize metering resource objects from YAML manifests: %v", err)
	}

	logger.Debugf("Metering Deploy Con***REMOVED***g: %#v", cfg)

	deployObj, err := deploy.NewDeployer(cfg, logger)
	if err != nil {
		return fmt.Errorf("Failed to deploy metering: %v", err)
	}

	if deployType == "install" {
		err := deployObj.Install()
		if err != nil {
			return fmt.Errorf("Failed to install metering: %v", err)
		}
		logger.Infof("Finished installing metering")
	} ***REMOVED*** if deployType == "uninstall" {
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
				"METERING_CR_FILE":          "meteringcon***REMOVED***g",
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
