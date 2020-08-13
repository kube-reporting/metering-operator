package operator

import (
	"fmt"
	"net"
	"os"

	"k8s.io/apimachinery/pkg/util/errors"
)

// IsValidConfig checks the validity of all configuration options.
func IsValidConfig(cfg *Config) error {
	errs := []error{}

	errs = append(errs, IsValidNamespaceConfig(cfg))
	errs = append(errs, IsValidListenConfig(cfg))
	errs = append(errs, IsValidPrestoConfig(cfg))
	errs = append(errs, IsValidHiveConfig(cfg))
	errs = append(errs, IsValidKubeConfig(cfg.Kubeconfig))
	errs = append(errs, IsValidPrometheusConfig(cfg))

	if err := isValidTLSConfig(&cfg.APITLSConfig); err != nil {
		errs = append(errs, fmt.Errorf("error validating apiTLSConfig: %s", err.Error()))
	}

	if err := isValidTLSConfig(&cfg.MetricsTLSConfig); err != nil {
		errs = append(errs, fmt.Errorf("error validating metricsTLSConfig: %s", err.Error()))
	}

	if len(errs) != 0 {
		return errors.NewAggregate(errs)
	}
	return nil
}

// IsValidNamespaceConfig ensures that if you are using target namespaces the all namespace field is correct.
func IsValidNamespaceConfig(cfg *Config) error {
	if len(cfg.TargetNamespaces) > 1 && !cfg.AllNamespaces {
		return fmt.Errorf("must set allNamespaces if more than one namespace is passed to targetNamespaces")
	}
	return nil
}

// IsValidListenConfig ensures all *Listen fields are set to valid host/ports if they have a value set.
func IsValidListenConfig(cfg *Config) error {
	errs := []error{}

	errs = append(errs, isValidHostPort(cfg.APIListen, "apiListen"))

	if len(errs) > 0 {
		return errors.NewAggregate(errs)
	}
	return nil
}

// IsValidPrestoConfig ensure all Presto* fields are valid if provided.
func IsValidPrestoConfig(cfg *Config) error {
	errs := []error{}

	if !cfg.PrestoUseTLS {
		if cfg.PrestoUseClientCertAuth {
			errs = append(errs, fmt.Errorf("prestoUseClientCertAuth cannot be set to true if prestoUseTLS is false"))
		}

		if len(cfg.PrestoCAFile) > 0 {
			errs = append(errs, fmt.Errorf("prestoCAFile cannot be set if prestoUseTLS is false"))
		}
	}

	if len(cfg.PrestoCAFile) > 0 {
		if _, err := os.Stat(cfg.PrestoCAFile); err != nil {
			errs = append(errs, err)
		}
	}

	if (len(cfg.PrestoClientCertFile) > 0 && len(cfg.PrestoClientKeyFile) == 0) ||
		(len(cfg.PrestoClientKeyFile) > 0 && len(cfg.PrestoClientCertFile) == 0) {
		errs = append(errs, fmt.Errorf("prestoClientCertFile and prestoClientKeyFile must both be specified or neither specified"))
	}

	if len(errs) > 0 {
		return errors.NewAggregate(errs)
	}
	return nil
}

// IsValidHiveConfig ensure all Hive* fields are valid if provided.
func IsValidHiveConfig(cfg *Config) error {
	errs := []error{}

	if !cfg.HiveUseTLS {
		if cfg.HiveUseClientCertAuth {
			errs = append(errs, fmt.Errorf("hiveUseClientCertAuth cannot be set to true if hiveUseTLS is false"))
		}

		if len(cfg.HiveCAFile) > 0 {
			errs = append(errs, fmt.Errorf("hiveCAFile cannot be set if hiveUseTLS is false"))
		}
	}

	if len(cfg.HiveCAFile) > 0 {
		if _, err := os.Stat(cfg.HiveCAFile); err != nil {
			errs = append(errs, err)
		}
	}

	if (len(cfg.HiveClientCertFile) > 0 && len(cfg.HiveClientKeyFile) == 0) ||
		(len(cfg.HiveClientKeyFile) > 0 && len(cfg.HiveClientCertFile) == 0) {
		errs = append(errs, fmt.Errorf("hiveClientCertFile and hiveClientKeyFile must both be specified or neither specified"))
	}

	if len(errs) > 0 {
		return errors.NewAggregate(errs)
	}
	return nil
}

// IsValidKubeConfig ensures the kube config is set to a valid file if provided.
func IsValidKubeConfig(kubeconfig string) error {
	if len(kubeconfig) > 0 {
		if _, err := os.Stat(kubeconfig); err != nil {
			return err
		}
	}
	return nil
}

// IsValidPrometheusConfig ensures prometheus configuration is valid.
func IsValidPrometheusConfig(cfg *Config) error {
	errs := []error{}
	if cfg.PrometheusConfig.CAFile != "" {
		if _, err := os.Stat(cfg.PrometheusConfig.CAFile); err != nil {
			errs = append(errs, err)
		}
	}

	// PrometheusDataSourceMaxBackfillImportDuration overrides PrometheusDataSourceGlobalImportFromTime
	// don't set both.
	if cfg.PrometheusDataSourceGlobalImportFromTime != nil && cfg.PrometheusDataSourceMaxBackfillImportDuration > 0 {
		errs = append(errs, fmt.Errorf("prometheusDataSourceGlobalImportFromTime and prometheusDataSourceMaxBackfillImportDuration cannot both be set"))
	}

	if len(errs) > 0 {
		return errors.NewAggregate(errs)
	}
	return nil
}

// IsValidTLSConfig ensures the TLS config is valid.
func isValidTLSConfig(cfg *TLSConfig) error {
	if cfg.UseTLS {
		if cfg.TLSCert == "" {
			return fmt.Errorf("must set TLS certificate if TLS is enabled")
		}
		if cfg.TLSKey == "" {
			return fmt.Errorf("must set TLS private key if TLS is enabled")
		}
	}
	return nil
}

// isValidHostPort attempts to split a non empty hp into host and port, returning any errors found.
// TODO this is only validating non-empty strings.  We may want to check for empty strings an report errors.
// TODO this requires a port to be specified, is that one of our requirements?
func isValidHostPort(hp string, name string) error {
	if len(hp) > 0 {
		if _, _, err := net.SplitHostPort(hp); err != nil {
			return fmt.Errorf("invalid %s: %s", name, err.Error())
		}
	}
	return nil
}
