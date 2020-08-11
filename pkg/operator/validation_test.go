package operator

import (
	"strings"
	"testing"
	"time"
)

func TestIsValidConfig(t *testing.T) {
	tests := map[string]struct {
		makeCfg     func() *Config
		expectedErr string
	}{
		"valid config": {
			makeCfg: func() *Config { return validConfig() },
		},
		"invalid namespace config": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.AllNamespaces = false
				cfg.TargetNamespaces = []string{"foo", "bar"}
				return cfg
			},
			expectedErr: "must set allNamespaces if more than one namespace is passed to targetNamespaces",
		},
		"listen config - valid": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.APIListen = "foo:8080"
				return cfg
			},
		},
		"listen config - invalid API Listen": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.APIListen = "foo"
				return cfg
			},
			expectedErr: "invalid apiListen",
		},
		"presto config - valid": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.PrestoHost = "foo:8080"
				return cfg
			},
		},
		"presto config - invalid use tls with client cert auth": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.PrestoUseTLS = false
				cfg.PrestoUseClientCertAuth = true
				return cfg
			},
			expectedErr: "prestoUseClientCertAuth cannot be set to true if prestoUseTLS is false",
		},
		"presto config - invalid use tls with CA file": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.PrestoUseTLS = false
				cfg.PrestoCAFile = "/tmp" // is this going to fail on macs?  Could probably be done more reliably
				return cfg
			},
			expectedErr: "prestoCAFile cannot be set if prestoUseTLS is false",
		},
		"presto config - invalid use tls with CA file that doesn't exist": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.PrestoUseTLS = true
				cfg.PrestoCAFile = "/garbageFile"
				return cfg
			},
			expectedErr: "no such file or directory",
		},
		"presto config - valid client cert/key config": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.PrestoClientCertFile = "/foo"
				cfg.PrestoClientKeyFile = "/foo"
				return cfg
			},
		},
		"presto config - invalid client cert/key config - no cert": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.PrestoClientKeyFile = "/foo"
				return cfg
			},
			expectedErr: "prestoClientCertFile and prestoClientKeyFile must both be specified or neither specified",
		},
		"presto config - invalid client cert/key config - no key": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.PrestoClientCertFile = "/foo"
				return cfg
			},
			expectedErr: "prestoClientCertFile and prestoClientKeyFile must both be specified or neither specified",
		},
		"hive config - valid": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.HiveHost = "foo:8080"
				return cfg
			},
		},
		"hive config - invalid use tls with client cert auth": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.HiveUseTLS = false
				cfg.HiveUseClientCertAuth = true
				return cfg
			},
			expectedErr: "hiveUseClientCertAuth cannot be set to true if hiveUseTLS is false",
		},
		"hive config - invalid use tls with CA file": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.HiveUseTLS = false
				cfg.HiveCAFile = "/tmp" // is this going to fail on macs?  Could probably be done more reliably
				return cfg
			},
			expectedErr: "hiveCAFile cannot be set if hiveUseTLS is false",
		},
		"hive config - invalid use tls with CA file that doesn't exist": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.HiveUseTLS = true
				cfg.HiveCAFile = "/garbageFile"
				return cfg
			},
			expectedErr: "no such file or directory",
		},
		"hive config - valid client cert/key config": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.HiveClientCertFile = "/foo"
				cfg.HiveClientKeyFile = "/foo"
				return cfg
			},
		},
		"hive config - invalid client cert/key config - no cert": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.HiveClientKeyFile = "/foo"
				return cfg
			},
			expectedErr: "hiveClientCertFile and hiveClientKeyFile must both be specified or neither specified",
		},
		"hive config - invalid client cert/key config - no key": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.HiveClientCertFile = "/foo"
				return cfg
			},
			expectedErr: "hiveClientCertFile and hiveClientKeyFile must both be specified or neither specified",
		},
		"kube config - valid": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.Kubeconfig = "/tmp" // will this fail on macs?
				return cfg
			},
		},
		"kube config - invalid": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.Kubeconfig = "/garbageFile"
				return cfg
			},
			expectedErr: "no such file or directory",
		},
		"prometheus config - invalid CA File": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.PrometheusConfig.CAFile = "/garbageFile"
				return cfg
			},
			expectedErr: "no such file or directory",
		},
		"prometheus config - invalid import fields": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.PrometheusDataSourceMaxBackfillImportDuration = time.Second
				now := time.Now()
				cfg.PrometheusDataSourceGlobalImportFromTime = &now
				return cfg
			},
			expectedErr: "prometheusDataSourceGlobalImportFromTime and prometheusDataSourceMaxBackfillImportDuration cannot both be set",
		},
		"api tls config - valid": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.APITLSConfig = TLSConfig{
					UseTLS:  true,
					TLSCert: "/foo",
					TLSKey:  "/foo",
				}
				return cfg
			},
		},
		"api tls config - missing cert": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.APITLSConfig = TLSConfig{
					UseTLS: true,
					TLSKey: "/foo",
				}
				return cfg
			},
			expectedErr: "must set TLS certificate if TLS is enabled",
		},
		"api tls config - missing key": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.APITLSConfig = TLSConfig{
					UseTLS:  true,
					TLSCert: "/foo",
				}
				return cfg
			},
			expectedErr: "must set TLS private key if TLS is enabled",
		},
		"metrics tls config - valid": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.MetricsTLSConfig = TLSConfig{
					UseTLS:  true,
					TLSCert: "/foo",
					TLSKey:  "/foo",
				}
				return cfg
			},
		},
		"metrics tls config - missing cert": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.MetricsTLSConfig = TLSConfig{
					UseTLS: true,
					TLSKey: "/foo",
				}
				return cfg
			},
			expectedErr: "must set TLS certificate if TLS is enabled",
		},
		"metrics tls config - missing key": {
			makeCfg: func() *Config {
				cfg := validConfig()
				cfg.MetricsTLSConfig = TLSConfig{
					UseTLS:  true,
					TLSCert: "/foo",
				}
				return cfg
			},
			expectedErr: "must set TLS private key if TLS is enabled",
		},
	}

	for name, test := range tests {
		err := IsValidConfig(test.makeCfg())
		if len(test.expectedErr) == 0 && err == nil {
			continue //good test
		}
		if len(test.expectedErr) == 0 && err != nil {
			t.Errorf("%s expected no error but received %s", name, err.Error())
			continue
		}
		if len(test.expectedErr) > 0 && err == nil {
			t.Errorf("%s expected an error but didn't receive one", name)
			continue
		}
		// expected error, got error, check that the error was what we wanted to avoid false passes
		if !strings.Contains(err.Error(), test.expectedErr) {
			t.Errorf("%s did not find the expected error string in %s", name, err.Error())
		}
	}
}

func validConfig() *Config {
	return &Config{
		AllNamespaces: true,
	}
}
