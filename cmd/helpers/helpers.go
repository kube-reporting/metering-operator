package helpers

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// SetupLogger is responsible for building up a basic logrus FieldLogger
// instance with a specific log level and fields configuration.
func SetupLogger(logLevelStr string, useDefaultFormatter bool, fields log.Fields) log.FieldLogger {
	var err error

	logger := log.WithFields(fields)
	logLevel, err := log.ParseLevel(logLevelStr)
	if err != nil {
		logger.WithError(err).Fatalf("invalid log level: %s", logLevel)
	}
	logger.Infof("Setting the log level to %s", logLevel.String())
	logger.Logger.Level = logLevel

	if useDefaultFormatter {
		logger.Logger.SetFormatter(&log.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "01-02-2006 15:04:05",
		})
	}

	return logger
}

// MapEnvVarToFlag takes a mapping of ENV var names to flag names and iterates
// over that mapping attempting to set the flag value with the ENV var key name.
// see: https://github.com/spf13/viper/issues/461
func MapEnvVarToFlag(vars map[string]string, flagset *pflag.FlagSet) error {
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

// SetFlagsFromEnv parses all registered flags in the given flagset,
// and if they are not already set it attempts to set their values from
// environment variables. Environment variables take the name of the flag but
// are UPPERCASE, and any dashes are replaced by underscores. Environment
// variables additionally are prefixed by the given string followed by
// and underscore. For example, if prefix=PREFIX: some-flag => PREFIX_SOME_FLAG
func SetFlagsFromEnv(fs *pflag.FlagSet, prefix string) (err error) {
	alreadySet := make(map[string]bool)
	fs.Visit(func(f *pflag.Flag) {
		alreadySet[f.Name] = true
	})
	fs.VisitAll(func(f *pflag.Flag) {
		if !alreadySet[f.Name] {
			key := prefix + "_" + strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
			val := os.Getenv(key)
			if val != "" {
				if serr := fs.Set(f.Name, val); serr != nil {
					err = fmt.Errorf("invalid value %q for %s: %v", val, key, serr)
				}
			}
		}
	})
	return err
}
