/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this ***REMOVED***le except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the speci***REMOVED***c language governing permissions and
limitations under the License.
*/

package clientcmd

import (
	"errors"
	"os"
	"path"
	"path/***REMOVED***lepath"
	"reflect"
	"sort"

	"github.com/golang/glog"

	restclient "k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Con***REMOVED***gAccess is used by subcommands and methods in this package to load and modify the appropriate con***REMOVED***g ***REMOVED***les
type Con***REMOVED***gAccess interface {
	// GetLoadingPrecedence returns the slice of ***REMOVED***les that should be used for loading and inspecting the con***REMOVED***g
	GetLoadingPrecedence() []string
	// GetStartingCon***REMOVED***g returns the con***REMOVED***g that subcommands should being operating against.  It may or may not be merged depending on loading rules
	GetStartingCon***REMOVED***g() (*clientcmdapi.Con***REMOVED***g, error)
	// GetDefaultFilename returns the name of the ***REMOVED***le you should write into (create if necessary), if you're trying to create a new stanza as opposed to updating an existing one.
	GetDefaultFilename() string
	// IsExplicitFile indicates whether or not this command is interested in exactly one ***REMOVED***le.  This implementation only ever does that  via a flag, but implementations that handle local, global, and flags may have more
	IsExplicitFile() bool
	// GetExplicitFile returns the particular ***REMOVED***le this command is operating against.  This implementation only ever has one, but implementations that handle local, global, and flags may have more
	GetExplicitFile() string
}

type PathOptions struct {
	// GlobalFile is the full path to the ***REMOVED***le to load as the global (***REMOVED***nal) option
	GlobalFile string
	// EnvVar is the env var name that points to the list of kubecon***REMOVED***g ***REMOVED***les to load
	EnvVar string
	// ExplicitFileFlag is the name of the flag to use for prompting for the kubecon***REMOVED***g ***REMOVED***le
	ExplicitFileFlag string

	// GlobalFileSubpath is an optional value used for displaying help
	GlobalFileSubpath string

	LoadingRules *ClientCon***REMOVED***gLoadingRules
}

func (o *PathOptions) GetEnvVarFiles() []string {
	if len(o.EnvVar) == 0 {
		return []string{}
	}

	envVarValue := os.Getenv(o.EnvVar)
	if len(envVarValue) == 0 {
		return []string{}
	}

	return ***REMOVED***lepath.SplitList(envVarValue)
}

func (o *PathOptions) GetLoadingPrecedence() []string {
	if envVarFiles := o.GetEnvVarFiles(); len(envVarFiles) > 0 {
		return envVarFiles
	}

	return []string{o.GlobalFile}
}

func (o *PathOptions) GetStartingCon***REMOVED***g() (*clientcmdapi.Con***REMOVED***g, error) {
	// don't mutate the original
	loadingRules := *o.LoadingRules
	loadingRules.Precedence = o.GetLoadingPrecedence()

	clientCon***REMOVED***g := NewNonInteractiveDeferredLoadingClientCon***REMOVED***g(&loadingRules, &Con***REMOVED***gOverrides{})
	rawCon***REMOVED***g, err := clientCon***REMOVED***g.RawCon***REMOVED***g()
	if os.IsNotExist(err) {
		return clientcmdapi.NewCon***REMOVED***g(), nil
	}
	if err != nil {
		return nil, err
	}

	return &rawCon***REMOVED***g, nil
}

func (o *PathOptions) GetDefaultFilename() string {
	if o.IsExplicitFile() {
		return o.GetExplicitFile()
	}

	if envVarFiles := o.GetEnvVarFiles(); len(envVarFiles) > 0 {
		if len(envVarFiles) == 1 {
			return envVarFiles[0]
		}

		// if any of the envvar ***REMOVED***les already exists, return it
		for _, envVarFile := range envVarFiles {
			if _, err := os.Stat(envVarFile); err == nil {
				return envVarFile
			}
		}

		// otherwise, return the last one in the list
		return envVarFiles[len(envVarFiles)-1]
	}

	return o.GlobalFile
}

func (o *PathOptions) IsExplicitFile() bool {
	if len(o.LoadingRules.ExplicitPath) > 0 {
		return true
	}

	return false
}

func (o *PathOptions) GetExplicitFile() string {
	return o.LoadingRules.ExplicitPath
}

func NewDefaultPathOptions() *PathOptions {
	ret := &PathOptions{
		GlobalFile:       RecommendedHomeFile,
		EnvVar:           RecommendedCon***REMOVED***gPathEnvVar,
		ExplicitFileFlag: RecommendedCon***REMOVED***gPathFlag,

		GlobalFileSubpath: path.Join(RecommendedHomeDir, RecommendedFileName),

		LoadingRules: NewDefaultClientCon***REMOVED***gLoadingRules(),
	}
	ret.LoadingRules.DoNotResolvePaths = true

	return ret
}

// ModifyCon***REMOVED***g takes a Con***REMOVED***g object, iterates through Clusters, AuthInfos, and Contexts, uses the LocationOfOrigin if speci***REMOVED***ed or
// uses the default destination ***REMOVED***le to write the results into.  This results in multiple ***REMOVED***le reads, but it's very easy to follow.
// Preferences and CurrentContext should always be set in the default destination ***REMOVED***le.  Since we can't distinguish between empty and missing values
// (no nil strings), we're forced have separate handling for them.  In the kubecon***REMOVED***g cases, newCon***REMOVED***g should have at most one difference,
// that means that this code will only write into a single ***REMOVED***le.  If you want to relativizePaths, you must provide a fully quali***REMOVED***ed path in any
// modi***REMOVED***ed element.
func ModifyCon***REMOVED***g(con***REMOVED***gAccess Con***REMOVED***gAccess, newCon***REMOVED***g clientcmdapi.Con***REMOVED***g, relativizePaths bool) error {
	possibleSources := con***REMOVED***gAccess.GetLoadingPrecedence()
	// sort the possible kubecon***REMOVED***g ***REMOVED***les so we always "lock" in the same order
	// to avoid deadlock (note: this can fail w/ symlinks, but... come on).
	sort.Strings(possibleSources)
	for _, ***REMOVED***lename := range possibleSources {
		if err := lockFile(***REMOVED***lename); err != nil {
			return err
		}
		defer unlockFile(***REMOVED***lename)
	}

	startingCon***REMOVED***g, err := con***REMOVED***gAccess.GetStartingCon***REMOVED***g()
	if err != nil {
		return err
	}

	// We need to ***REMOVED***nd all differences, locate their original ***REMOVED***les, read a partial con***REMOVED***g to modify only that stanza and write out the ***REMOVED***le.
	// Special case the test for current context and preferences since those always write to the default ***REMOVED***le.
	if reflect.DeepEqual(*startingCon***REMOVED***g, newCon***REMOVED***g) {
		// nothing to do
		return nil
	}

	if startingCon***REMOVED***g.CurrentContext != newCon***REMOVED***g.CurrentContext {
		if err := writeCurrentContext(con***REMOVED***gAccess, newCon***REMOVED***g.CurrentContext); err != nil {
			return err
		}
	}

	if !reflect.DeepEqual(startingCon***REMOVED***g.Preferences, newCon***REMOVED***g.Preferences) {
		if err := writePreferences(con***REMOVED***gAccess, newCon***REMOVED***g.Preferences); err != nil {
			return err
		}
	}

	// Search every cluster, authInfo, and context.  First from new to old for differences, then from old to new for deletions
	for key, cluster := range newCon***REMOVED***g.Clusters {
		startingCluster, exists := startingCon***REMOVED***g.Clusters[key]
		if !reflect.DeepEqual(cluster, startingCluster) || !exists {
			destinationFile := cluster.LocationOfOrigin
			if len(destinationFile) == 0 {
				destinationFile = con***REMOVED***gAccess.GetDefaultFilename()
			}

			con***REMOVED***gToWrite, err := getCon***REMOVED***gFromFile(destinationFile)
			if err != nil {
				return err
			}
			t := *cluster

			con***REMOVED***gToWrite.Clusters[key] = &t
			con***REMOVED***gToWrite.Clusters[key].LocationOfOrigin = destinationFile
			if relativizePaths {
				if err := RelativizeClusterLocalPaths(con***REMOVED***gToWrite.Clusters[key]); err != nil {
					return err
				}
			}

			if err := WriteToFile(*con***REMOVED***gToWrite, destinationFile); err != nil {
				return err
			}
		}
	}

	for key, context := range newCon***REMOVED***g.Contexts {
		startingContext, exists := startingCon***REMOVED***g.Contexts[key]
		if !reflect.DeepEqual(context, startingContext) || !exists {
			destinationFile := context.LocationOfOrigin
			if len(destinationFile) == 0 {
				destinationFile = con***REMOVED***gAccess.GetDefaultFilename()
			}

			con***REMOVED***gToWrite, err := getCon***REMOVED***gFromFile(destinationFile)
			if err != nil {
				return err
			}
			con***REMOVED***gToWrite.Contexts[key] = context

			if err := WriteToFile(*con***REMOVED***gToWrite, destinationFile); err != nil {
				return err
			}
		}
	}

	for key, authInfo := range newCon***REMOVED***g.AuthInfos {
		startingAuthInfo, exists := startingCon***REMOVED***g.AuthInfos[key]
		if !reflect.DeepEqual(authInfo, startingAuthInfo) || !exists {
			destinationFile := authInfo.LocationOfOrigin
			if len(destinationFile) == 0 {
				destinationFile = con***REMOVED***gAccess.GetDefaultFilename()
			}

			con***REMOVED***gToWrite, err := getCon***REMOVED***gFromFile(destinationFile)
			if err != nil {
				return err
			}
			t := *authInfo
			con***REMOVED***gToWrite.AuthInfos[key] = &t
			con***REMOVED***gToWrite.AuthInfos[key].LocationOfOrigin = destinationFile
			if relativizePaths {
				if err := RelativizeAuthInfoLocalPaths(con***REMOVED***gToWrite.AuthInfos[key]); err != nil {
					return err
				}
			}

			if err := WriteToFile(*con***REMOVED***gToWrite, destinationFile); err != nil {
				return err
			}
		}
	}

	for key, cluster := range startingCon***REMOVED***g.Clusters {
		if _, exists := newCon***REMOVED***g.Clusters[key]; !exists {
			destinationFile := cluster.LocationOfOrigin
			if len(destinationFile) == 0 {
				destinationFile = con***REMOVED***gAccess.GetDefaultFilename()
			}

			con***REMOVED***gToWrite, err := getCon***REMOVED***gFromFile(destinationFile)
			if err != nil {
				return err
			}
			delete(con***REMOVED***gToWrite.Clusters, key)

			if err := WriteToFile(*con***REMOVED***gToWrite, destinationFile); err != nil {
				return err
			}
		}
	}

	for key, context := range startingCon***REMOVED***g.Contexts {
		if _, exists := newCon***REMOVED***g.Contexts[key]; !exists {
			destinationFile := context.LocationOfOrigin
			if len(destinationFile) == 0 {
				destinationFile = con***REMOVED***gAccess.GetDefaultFilename()
			}

			con***REMOVED***gToWrite, err := getCon***REMOVED***gFromFile(destinationFile)
			if err != nil {
				return err
			}
			delete(con***REMOVED***gToWrite.Contexts, key)

			if err := WriteToFile(*con***REMOVED***gToWrite, destinationFile); err != nil {
				return err
			}
		}
	}

	for key, authInfo := range startingCon***REMOVED***g.AuthInfos {
		if _, exists := newCon***REMOVED***g.AuthInfos[key]; !exists {
			destinationFile := authInfo.LocationOfOrigin
			if len(destinationFile) == 0 {
				destinationFile = con***REMOVED***gAccess.GetDefaultFilename()
			}

			con***REMOVED***gToWrite, err := getCon***REMOVED***gFromFile(destinationFile)
			if err != nil {
				return err
			}
			delete(con***REMOVED***gToWrite.AuthInfos, key)

			if err := WriteToFile(*con***REMOVED***gToWrite, destinationFile); err != nil {
				return err
			}
		}
	}

	return nil
}

func PersisterForUser(con***REMOVED***gAccess Con***REMOVED***gAccess, user string) restclient.AuthProviderCon***REMOVED***gPersister {
	return &persister{con***REMOVED***gAccess, user}
}

type persister struct {
	con***REMOVED***gAccess Con***REMOVED***gAccess
	user         string
}

func (p *persister) Persist(con***REMOVED***g map[string]string) error {
	newCon***REMOVED***g, err := p.con***REMOVED***gAccess.GetStartingCon***REMOVED***g()
	if err != nil {
		return err
	}
	authInfo, ok := newCon***REMOVED***g.AuthInfos[p.user]
	if ok && authInfo.AuthProvider != nil {
		authInfo.AuthProvider.Con***REMOVED***g = con***REMOVED***g
		ModifyCon***REMOVED***g(p.con***REMOVED***gAccess, *newCon***REMOVED***g, false)
	}
	return nil
}

// writeCurrentContext takes three possible paths.
// If newCurrentContext is the same as the startingCon***REMOVED***g's current context, then we exit.
// If newCurrentContext has a value, then that value is written into the default destination ***REMOVED***le.
// If newCurrentContext is empty, then we ***REMOVED***nd the con***REMOVED***g ***REMOVED***le that is setting the CurrentContext and clear the value from that ***REMOVED***le
func writeCurrentContext(con***REMOVED***gAccess Con***REMOVED***gAccess, newCurrentContext string) error {
	if startingCon***REMOVED***g, err := con***REMOVED***gAccess.GetStartingCon***REMOVED***g(); err != nil {
		return err
	} ***REMOVED*** if startingCon***REMOVED***g.CurrentContext == newCurrentContext {
		return nil
	}

	if con***REMOVED***gAccess.IsExplicitFile() {
		***REMOVED***le := con***REMOVED***gAccess.GetExplicitFile()
		currCon***REMOVED***g, err := getCon***REMOVED***gFromFile(***REMOVED***le)
		if err != nil {
			return err
		}
		currCon***REMOVED***g.CurrentContext = newCurrentContext
		if err := WriteToFile(*currCon***REMOVED***g, ***REMOVED***le); err != nil {
			return err
		}

		return nil
	}

	if len(newCurrentContext) > 0 {
		destinationFile := con***REMOVED***gAccess.GetDefaultFilename()
		con***REMOVED***g, err := getCon***REMOVED***gFromFile(destinationFile)
		if err != nil {
			return err
		}
		con***REMOVED***g.CurrentContext = newCurrentContext

		if err := WriteToFile(*con***REMOVED***g, destinationFile); err != nil {
			return err
		}

		return nil
	}

	// we're supposed to be clearing the current context.  We need to ***REMOVED***nd the ***REMOVED***rst spot in the chain that is setting it and clear it
	for _, ***REMOVED***le := range con***REMOVED***gAccess.GetLoadingPrecedence() {
		if _, err := os.Stat(***REMOVED***le); err == nil {
			currCon***REMOVED***g, err := getCon***REMOVED***gFromFile(***REMOVED***le)
			if err != nil {
				return err
			}

			if len(currCon***REMOVED***g.CurrentContext) > 0 {
				currCon***REMOVED***g.CurrentContext = newCurrentContext
				if err := WriteToFile(*currCon***REMOVED***g, ***REMOVED***le); err != nil {
					return err
				}

				return nil
			}
		}
	}

	return errors.New("no con***REMOVED***g found to write context")
}

func writePreferences(con***REMOVED***gAccess Con***REMOVED***gAccess, newPrefs clientcmdapi.Preferences) error {
	if startingCon***REMOVED***g, err := con***REMOVED***gAccess.GetStartingCon***REMOVED***g(); err != nil {
		return err
	} ***REMOVED*** if reflect.DeepEqual(startingCon***REMOVED***g.Preferences, newPrefs) {
		return nil
	}

	if con***REMOVED***gAccess.IsExplicitFile() {
		***REMOVED***le := con***REMOVED***gAccess.GetExplicitFile()
		currCon***REMOVED***g, err := getCon***REMOVED***gFromFile(***REMOVED***le)
		if err != nil {
			return err
		}
		currCon***REMOVED***g.Preferences = newPrefs
		if err := WriteToFile(*currCon***REMOVED***g, ***REMOVED***le); err != nil {
			return err
		}

		return nil
	}

	for _, ***REMOVED***le := range con***REMOVED***gAccess.GetLoadingPrecedence() {
		currCon***REMOVED***g, err := getCon***REMOVED***gFromFile(***REMOVED***le)
		if err != nil {
			return err
		}

		if !reflect.DeepEqual(currCon***REMOVED***g.Preferences, newPrefs) {
			currCon***REMOVED***g.Preferences = newPrefs
			if err := WriteToFile(*currCon***REMOVED***g, ***REMOVED***le); err != nil {
				return err
			}

			return nil
		}
	}

	return errors.New("no con***REMOVED***g found to write preferences")
}

// getCon***REMOVED***gFromFile tries to read a kubecon***REMOVED***g ***REMOVED***le and if it can't, returns an error.  One exception, missing ***REMOVED***les result in empty con***REMOVED***gs, not an error.
func getCon***REMOVED***gFromFile(***REMOVED***lename string) (*clientcmdapi.Con***REMOVED***g, error) {
	con***REMOVED***g, err := LoadFromFile(***REMOVED***lename)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if con***REMOVED***g == nil {
		con***REMOVED***g = clientcmdapi.NewCon***REMOVED***g()
	}
	return con***REMOVED***g, nil
}

// GetCon***REMOVED***gFromFileOrDie tries to read a kubecon***REMOVED***g ***REMOVED***le and if it can't, it calls exit.  One exception, missing ***REMOVED***les result in empty con***REMOVED***gs, not an exit
func GetCon***REMOVED***gFromFileOrDie(***REMOVED***lename string) *clientcmdapi.Con***REMOVED***g {
	con***REMOVED***g, err := getCon***REMOVED***gFromFile(***REMOVED***lename)
	if err != nil {
		glog.FatalDepth(1, err)
	}

	return con***REMOVED***g
}
