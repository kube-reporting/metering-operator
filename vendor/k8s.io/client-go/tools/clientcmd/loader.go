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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/***REMOVED***lepath"
	"reflect"
	goruntime "runtime"
	"strings"

	"github.com/imdario/mergo"
	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	restclient "k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"k8s.io/client-go/util/homedir"
)

const (
	RecommendedCon***REMOVED***gPathFlag   = "kubecon***REMOVED***g"
	RecommendedCon***REMOVED***gPathEnvVar = "KUBECONFIG"
	RecommendedHomeDir          = ".kube"
	RecommendedFileName         = "con***REMOVED***g"
	RecommendedSchemaName       = "schema"
)

var (
	RecommendedCon***REMOVED***gDir  = path.Join(homedir.HomeDir(), RecommendedHomeDir)
	RecommendedHomeFile   = path.Join(RecommendedCon***REMOVED***gDir, RecommendedFileName)
	RecommendedSchemaFile = path.Join(RecommendedCon***REMOVED***gDir, RecommendedSchemaName)
)

// currentMigrationRules returns a map that holds the history of recommended home directories used in previous versions.
// Any future changes to RecommendedHomeFile and related are expected to add a migration rule here, in order to make
// sure existing con***REMOVED***g ***REMOVED***les are migrated to their new locations properly.
func currentMigrationRules() map[string]string {
	oldRecommendedHomeFile := path.Join(os.Getenv("HOME"), "/.kube/.kubecon***REMOVED***g")
	oldRecommendedWindowsHomeFile := path.Join(os.Getenv("HOME"), RecommendedHomeDir, RecommendedFileName)

	migrationRules := map[string]string{}
	migrationRules[RecommendedHomeFile] = oldRecommendedHomeFile
	if goruntime.GOOS == "windows" {
		migrationRules[RecommendedHomeFile] = oldRecommendedWindowsHomeFile
	}
	return migrationRules
}

type ClientCon***REMOVED***gLoader interface {
	Con***REMOVED***gAccess
	// IsDefaultCon***REMOVED***g returns true if the returned con***REMOVED***g matches the defaults.
	IsDefaultCon***REMOVED***g(*restclient.Con***REMOVED***g) bool
	// Load returns the latest con***REMOVED***g
	Load() (*clientcmdapi.Con***REMOVED***g, error)
}

type Kubecon***REMOVED***gGetter func() (*clientcmdapi.Con***REMOVED***g, error)

type ClientCon***REMOVED***gGetter struct {
	kubecon***REMOVED***gGetter Kubecon***REMOVED***gGetter
}

// ClientCon***REMOVED***gGetter implements the ClientCon***REMOVED***gLoader interface.
var _ ClientCon***REMOVED***gLoader = &ClientCon***REMOVED***gGetter{}

func (g *ClientCon***REMOVED***gGetter) Load() (*clientcmdapi.Con***REMOVED***g, error) {
	return g.kubecon***REMOVED***gGetter()
}

func (g *ClientCon***REMOVED***gGetter) GetLoadingPrecedence() []string {
	return nil
}
func (g *ClientCon***REMOVED***gGetter) GetStartingCon***REMOVED***g() (*clientcmdapi.Con***REMOVED***g, error) {
	return g.kubecon***REMOVED***gGetter()
}
func (g *ClientCon***REMOVED***gGetter) GetDefaultFilename() string {
	return ""
}
func (g *ClientCon***REMOVED***gGetter) IsExplicitFile() bool {
	return false
}
func (g *ClientCon***REMOVED***gGetter) GetExplicitFile() string {
	return ""
}
func (g *ClientCon***REMOVED***gGetter) IsDefaultCon***REMOVED***g(con***REMOVED***g *restclient.Con***REMOVED***g) bool {
	return false
}

// ClientCon***REMOVED***gLoadingRules is an ExplicitPath and string slice of speci***REMOVED***c locations that are used for merging together a Con***REMOVED***g
// Callers can put the chain together however they want, but we'd recommend:
// EnvVarPathFiles if set (a list of ***REMOVED***les if set) OR the HomeDirectoryPath
// ExplicitPath is special, because if a user speci***REMOVED***cally requests a certain ***REMOVED***le be used and error is reported if this ***REMOVED***le is not present
type ClientCon***REMOVED***gLoadingRules struct {
	ExplicitPath string
	Precedence   []string

	// MigrationRules is a map of destination ***REMOVED***les to source ***REMOVED***les.  If a destination ***REMOVED***le is not present, then the source ***REMOVED***le is checked.
	// If the source ***REMOVED***le is present, then it is copied to the destination ***REMOVED***le BEFORE any further loading happens.
	MigrationRules map[string]string

	// DoNotResolvePaths indicates whether or not to resolve paths with respect to the originating ***REMOVED***les.  This is phrased as a negative so
	// that a default object that doesn't set this will usually get the behavior it wants.
	DoNotResolvePaths bool

	// DefaultClientCon***REMOVED***g is an optional ***REMOVED***eld indicating what rules to use to calculate a default con***REMOVED***guration.
	// This should match the overrides passed in to ClientCon***REMOVED***g loader.
	DefaultClientCon***REMOVED***g ClientCon***REMOVED***g
}

// ClientCon***REMOVED***gLoadingRules implements the ClientCon***REMOVED***gLoader interface.
var _ ClientCon***REMOVED***gLoader = &ClientCon***REMOVED***gLoadingRules{}

// NewDefaultClientCon***REMOVED***gLoadingRules returns a ClientCon***REMOVED***gLoadingRules object with default ***REMOVED***elds ***REMOVED***lled in.  You are not required to
// use this constructor
func NewDefaultClientCon***REMOVED***gLoadingRules() *ClientCon***REMOVED***gLoadingRules {
	chain := []string{}

	envVarFiles := os.Getenv(RecommendedCon***REMOVED***gPathEnvVar)
	if len(envVarFiles) != 0 {
		***REMOVED***leList := ***REMOVED***lepath.SplitList(envVarFiles)
		// prevent the same path load multiple times
		chain = append(chain, deduplicate(***REMOVED***leList)...)

	} ***REMOVED*** {
		chain = append(chain, RecommendedHomeFile)
	}

	return &ClientCon***REMOVED***gLoadingRules{
		Precedence:     chain,
		MigrationRules: currentMigrationRules(),
	}
}

// Load starts by running the MigrationRules and then
// takes the loading rules and returns a Con***REMOVED***g object based on following rules.
//   if the ExplicitPath, return the unmerged explicit ***REMOVED***le
//   Otherwise, return a merged con***REMOVED***g based on the Precedence slice
// A missing ExplicitPath ***REMOVED***le produces an error. Empty ***REMOVED***lenames or other missing ***REMOVED***les are ignored.
// Read errors or ***REMOVED***les with non-deserializable content produce errors.
// The ***REMOVED***rst ***REMOVED***le to set a particular map key wins and map key's value is never changed.
// BUT, if you set a struct value that is NOT contained inside of map, the value WILL be changed.
// This results in some odd looking logic to merge in one direction, merge in the other, and then merge the two.
// It also means that if two ***REMOVED***les specify a "red-user", only values from the ***REMOVED***rst ***REMOVED***le's red-user are used.  Even
// non-conflicting entries from the second ***REMOVED***le's "red-user" are discarded.
// Relative paths inside of the .kubecon***REMOVED***g ***REMOVED***les are resolved against the .kubecon***REMOVED***g ***REMOVED***le's parent folder
// and only absolute ***REMOVED***le paths are returned.
func (rules *ClientCon***REMOVED***gLoadingRules) Load() (*clientcmdapi.Con***REMOVED***g, error) {
	if err := rules.Migrate(); err != nil {
		return nil, err
	}

	errlist := []error{}

	kubeCon***REMOVED***gFiles := []string{}

	// Make sure a ***REMOVED***le we were explicitly told to use exists
	if len(rules.ExplicitPath) > 0 {
		if _, err := os.Stat(rules.ExplicitPath); os.IsNotExist(err) {
			return nil, err
		}
		kubeCon***REMOVED***gFiles = append(kubeCon***REMOVED***gFiles, rules.ExplicitPath)

	} ***REMOVED*** {
		kubeCon***REMOVED***gFiles = append(kubeCon***REMOVED***gFiles, rules.Precedence...)
	}

	kubecon***REMOVED***gs := []*clientcmdapi.Con***REMOVED***g{}
	// read and cache the con***REMOVED***g ***REMOVED***les so that we only look at them once
	for _, ***REMOVED***lename := range kubeCon***REMOVED***gFiles {
		if len(***REMOVED***lename) == 0 {
			// no work to do
			continue
		}

		con***REMOVED***g, err := LoadFromFile(***REMOVED***lename)
		if os.IsNotExist(err) {
			// skip missing ***REMOVED***les
			continue
		}
		if err != nil {
			errlist = append(errlist, fmt.Errorf("Error loading con***REMOVED***g ***REMOVED***le \"%s\": %v", ***REMOVED***lename, err))
			continue
		}

		kubecon***REMOVED***gs = append(kubecon***REMOVED***gs, con***REMOVED***g)
	}

	// ***REMOVED***rst merge all of our maps
	mapCon***REMOVED***g := clientcmdapi.NewCon***REMOVED***g()

	for _, kubecon***REMOVED***g := range kubecon***REMOVED***gs {
		mergo.MergeWithOverwrite(mapCon***REMOVED***g, kubecon***REMOVED***g)
	}

	// merge all of the struct values in the reverse order so that priority is given correctly
	// errors are not added to the list the second time
	nonMapCon***REMOVED***g := clientcmdapi.NewCon***REMOVED***g()
	for i := len(kubecon***REMOVED***gs) - 1; i >= 0; i-- {
		kubecon***REMOVED***g := kubecon***REMOVED***gs[i]
		mergo.MergeWithOverwrite(nonMapCon***REMOVED***g, kubecon***REMOVED***g)
	}

	// since values are overwritten, but maps values are not, we can merge the non-map con***REMOVED***g on top of the map con***REMOVED***g and
	// get the values we expect.
	con***REMOVED***g := clientcmdapi.NewCon***REMOVED***g()
	mergo.MergeWithOverwrite(con***REMOVED***g, mapCon***REMOVED***g)
	mergo.MergeWithOverwrite(con***REMOVED***g, nonMapCon***REMOVED***g)

	if rules.ResolvePaths() {
		if err := ResolveLocalPaths(con***REMOVED***g); err != nil {
			errlist = append(errlist, err)
		}
	}
	return con***REMOVED***g, utilerrors.NewAggregate(errlist)
}

// Migrate uses the MigrationRules map.  If a destination ***REMOVED***le is not present, then the source ***REMOVED***le is checked.
// If the source ***REMOVED***le is present, then it is copied to the destination ***REMOVED***le BEFORE any further loading happens.
func (rules *ClientCon***REMOVED***gLoadingRules) Migrate() error {
	if rules.MigrationRules == nil {
		return nil
	}

	for destination, source := range rules.MigrationRules {
		if _, err := os.Stat(destination); err == nil {
			// if the destination already exists, do nothing
			continue
		} ***REMOVED*** if os.IsPermission(err) {
			// if we can't access the ***REMOVED***le, skip it
			continue
		} ***REMOVED*** if !os.IsNotExist(err) {
			// if we had an error other than non-existence, fail
			return err
		}

		if sourceInfo, err := os.Stat(source); err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				// if the source ***REMOVED***le doesn't exist or we can't access it, there's no work to do.
				continue
			}

			// if we had an error other than non-existence, fail
			return err
		} ***REMOVED*** if sourceInfo.IsDir() {
			return fmt.Errorf("cannot migrate %v to %v because it is a directory", source, destination)
		}

		in, err := os.Open(source)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(destination)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err = io.Copy(out, in); err != nil {
			return err
		}
	}

	return nil
}

// GetLoadingPrecedence implements Con***REMOVED***gAccess
func (rules *ClientCon***REMOVED***gLoadingRules) GetLoadingPrecedence() []string {
	return rules.Precedence
}

// GetStartingCon***REMOVED***g implements Con***REMOVED***gAccess
func (rules *ClientCon***REMOVED***gLoadingRules) GetStartingCon***REMOVED***g() (*clientcmdapi.Con***REMOVED***g, error) {
	clientCon***REMOVED***g := NewNonInteractiveDeferredLoadingClientCon***REMOVED***g(rules, &Con***REMOVED***gOverrides{})
	rawCon***REMOVED***g, err := clientCon***REMOVED***g.RawCon***REMOVED***g()
	if os.IsNotExist(err) {
		return clientcmdapi.NewCon***REMOVED***g(), nil
	}
	if err != nil {
		return nil, err
	}

	return &rawCon***REMOVED***g, nil
}

// GetDefaultFilename implements Con***REMOVED***gAccess
func (rules *ClientCon***REMOVED***gLoadingRules) GetDefaultFilename() string {
	// Explicit ***REMOVED***le if we have one.
	if rules.IsExplicitFile() {
		return rules.GetExplicitFile()
	}
	// Otherwise, ***REMOVED***rst existing ***REMOVED***le from precedence.
	for _, ***REMOVED***lename := range rules.GetLoadingPrecedence() {
		if _, err := os.Stat(***REMOVED***lename); err == nil {
			return ***REMOVED***lename
		}
	}
	// If none exists, use the ***REMOVED***rst from precedence.
	if len(rules.Precedence) > 0 {
		return rules.Precedence[0]
	}
	return ""
}

// IsExplicitFile implements Con***REMOVED***gAccess
func (rules *ClientCon***REMOVED***gLoadingRules) IsExplicitFile() bool {
	return len(rules.ExplicitPath) > 0
}

// GetExplicitFile implements Con***REMOVED***gAccess
func (rules *ClientCon***REMOVED***gLoadingRules) GetExplicitFile() string {
	return rules.ExplicitPath
}

// IsDefaultCon***REMOVED***g returns true if the provided con***REMOVED***guration matches the default
func (rules *ClientCon***REMOVED***gLoadingRules) IsDefaultCon***REMOVED***g(con***REMOVED***g *restclient.Con***REMOVED***g) bool {
	if rules.DefaultClientCon***REMOVED***g == nil {
		return false
	}
	defaultCon***REMOVED***g, err := rules.DefaultClientCon***REMOVED***g.ClientCon***REMOVED***g()
	if err != nil {
		return false
	}
	return reflect.DeepEqual(con***REMOVED***g, defaultCon***REMOVED***g)
}

// LoadFromFile takes a ***REMOVED***lename and deserializes the contents into Con***REMOVED***g object
func LoadFromFile(***REMOVED***lename string) (*clientcmdapi.Con***REMOVED***g, error) {
	kubecon***REMOVED***gBytes, err := ioutil.ReadFile(***REMOVED***lename)
	if err != nil {
		return nil, err
	}
	con***REMOVED***g, err := Load(kubecon***REMOVED***gBytes)
	if err != nil {
		return nil, err
	}
	klog.V(6).Infoln("Con***REMOVED***g loaded from ***REMOVED***le", ***REMOVED***lename)

	// set LocationOfOrigin on every Cluster, User, and Context
	for key, obj := range con***REMOVED***g.AuthInfos {
		obj.LocationOfOrigin = ***REMOVED***lename
		con***REMOVED***g.AuthInfos[key] = obj
	}
	for key, obj := range con***REMOVED***g.Clusters {
		obj.LocationOfOrigin = ***REMOVED***lename
		con***REMOVED***g.Clusters[key] = obj
	}
	for key, obj := range con***REMOVED***g.Contexts {
		obj.LocationOfOrigin = ***REMOVED***lename
		con***REMOVED***g.Contexts[key] = obj
	}

	if con***REMOVED***g.AuthInfos == nil {
		con***REMOVED***g.AuthInfos = map[string]*clientcmdapi.AuthInfo{}
	}
	if con***REMOVED***g.Clusters == nil {
		con***REMOVED***g.Clusters = map[string]*clientcmdapi.Cluster{}
	}
	if con***REMOVED***g.Contexts == nil {
		con***REMOVED***g.Contexts = map[string]*clientcmdapi.Context{}
	}

	return con***REMOVED***g, nil
}

// Load takes a byte slice and deserializes the contents into Con***REMOVED***g object.
// Encapsulates deserialization without assuming the source is a ***REMOVED***le.
func Load(data []byte) (*clientcmdapi.Con***REMOVED***g, error) {
	con***REMOVED***g := clientcmdapi.NewCon***REMOVED***g()
	// if there's no data in a ***REMOVED***le, return the default object instead of failing (DecodeInto reject empty input)
	if len(data) == 0 {
		return con***REMOVED***g, nil
	}
	decoded, _, err := clientcmdlatest.Codec.Decode(data, &schema.GroupVersionKind{Version: clientcmdlatest.Version, Kind: "Con***REMOVED***g"}, con***REMOVED***g)
	if err != nil {
		return nil, err
	}
	return decoded.(*clientcmdapi.Con***REMOVED***g), nil
}

// WriteToFile serializes the con***REMOVED***g to yaml and writes it out to a ***REMOVED***le.  If not present, it creates the ***REMOVED***le with the mode 0600.  If it is present
// it stomps the contents
func WriteToFile(con***REMOVED***g clientcmdapi.Con***REMOVED***g, ***REMOVED***lename string) error {
	content, err := Write(con***REMOVED***g)
	if err != nil {
		return err
	}
	dir := ***REMOVED***lepath.Dir(***REMOVED***lename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(***REMOVED***lename, content, 0600); err != nil {
		return err
	}
	return nil
}

func lockFile(***REMOVED***lename string) error {
	// TODO: ***REMOVED***nd a way to do this with actual ***REMOVED***le locks. Will
	// probably need separate solution for windows and Linux.

	// Make sure the dir exists before we try to create a lock ***REMOVED***le.
	dir := ***REMOVED***lepath.Dir(***REMOVED***lename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(lockName(***REMOVED***lename), os.O_CREATE|os.O_EXCL, 0)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func unlockFile(***REMOVED***lename string) error {
	return os.Remove(lockName(***REMOVED***lename))
}

func lockName(***REMOVED***lename string) string {
	return ***REMOVED***lename + ".lock"
}

// Write serializes the con***REMOVED***g to yaml.
// Encapsulates serialization without assuming the destination is a ***REMOVED***le.
func Write(con***REMOVED***g clientcmdapi.Con***REMOVED***g) ([]byte, error) {
	return runtime.Encode(clientcmdlatest.Codec, &con***REMOVED***g)
}

func (rules ClientCon***REMOVED***gLoadingRules) ResolvePaths() bool {
	return !rules.DoNotResolvePaths
}

// ResolveLocalPaths resolves all relative paths in the con***REMOVED***g object with respect to the stanza's LocationOfOrigin
// this cannot be done directly inside of LoadFromFile because doing so there would make it impossible to load a ***REMOVED***le without
// modi***REMOVED***cation of its contents.
func ResolveLocalPaths(con***REMOVED***g *clientcmdapi.Con***REMOVED***g) error {
	for _, cluster := range con***REMOVED***g.Clusters {
		if len(cluster.LocationOfOrigin) == 0 {
			continue
		}
		base, err := ***REMOVED***lepath.Abs(***REMOVED***lepath.Dir(cluster.LocationOfOrigin))
		if err != nil {
			return fmt.Errorf("Could not determine the absolute path of con***REMOVED***g ***REMOVED***le %s: %v", cluster.LocationOfOrigin, err)
		}

		if err := ResolvePaths(GetClusterFileReferences(cluster), base); err != nil {
			return err
		}
	}
	for _, authInfo := range con***REMOVED***g.AuthInfos {
		if len(authInfo.LocationOfOrigin) == 0 {
			continue
		}
		base, err := ***REMOVED***lepath.Abs(***REMOVED***lepath.Dir(authInfo.LocationOfOrigin))
		if err != nil {
			return fmt.Errorf("Could not determine the absolute path of con***REMOVED***g ***REMOVED***le %s: %v", authInfo.LocationOfOrigin, err)
		}

		if err := ResolvePaths(GetAuthInfoFileReferences(authInfo), base); err != nil {
			return err
		}
	}

	return nil
}

// RelativizeClusterLocalPaths ***REMOVED***rst absolutizes the paths by calling ResolveLocalPaths.  This assumes that any NEW path is already
// absolute, but any existing path will be resolved relative to LocationOfOrigin
func RelativizeClusterLocalPaths(cluster *clientcmdapi.Cluster) error {
	if len(cluster.LocationOfOrigin) == 0 {
		return fmt.Errorf("no location of origin for %s", cluster.Server)
	}
	base, err := ***REMOVED***lepath.Abs(***REMOVED***lepath.Dir(cluster.LocationOfOrigin))
	if err != nil {
		return fmt.Errorf("could not determine the absolute path of con***REMOVED***g ***REMOVED***le %s: %v", cluster.LocationOfOrigin, err)
	}

	if err := ResolvePaths(GetClusterFileReferences(cluster), base); err != nil {
		return err
	}
	if err := RelativizePathWithNoBacksteps(GetClusterFileReferences(cluster), base); err != nil {
		return err
	}

	return nil
}

// RelativizeAuthInfoLocalPaths ***REMOVED***rst absolutizes the paths by calling ResolveLocalPaths.  This assumes that any NEW path is already
// absolute, but any existing path will be resolved relative to LocationOfOrigin
func RelativizeAuthInfoLocalPaths(authInfo *clientcmdapi.AuthInfo) error {
	if len(authInfo.LocationOfOrigin) == 0 {
		return fmt.Errorf("no location of origin for %v", authInfo)
	}
	base, err := ***REMOVED***lepath.Abs(***REMOVED***lepath.Dir(authInfo.LocationOfOrigin))
	if err != nil {
		return fmt.Errorf("could not determine the absolute path of con***REMOVED***g ***REMOVED***le %s: %v", authInfo.LocationOfOrigin, err)
	}

	if err := ResolvePaths(GetAuthInfoFileReferences(authInfo), base); err != nil {
		return err
	}
	if err := RelativizePathWithNoBacksteps(GetAuthInfoFileReferences(authInfo), base); err != nil {
		return err
	}

	return nil
}

func RelativizeCon***REMOVED***gPaths(con***REMOVED***g *clientcmdapi.Con***REMOVED***g, base string) error {
	return RelativizePathWithNoBacksteps(GetCon***REMOVED***gFileReferences(con***REMOVED***g), base)
}

func ResolveCon***REMOVED***gPaths(con***REMOVED***g *clientcmdapi.Con***REMOVED***g, base string) error {
	return ResolvePaths(GetCon***REMOVED***gFileReferences(con***REMOVED***g), base)
}

func GetCon***REMOVED***gFileReferences(con***REMOVED***g *clientcmdapi.Con***REMOVED***g) []*string {
	refs := []*string{}

	for _, cluster := range con***REMOVED***g.Clusters {
		refs = append(refs, GetClusterFileReferences(cluster)...)
	}
	for _, authInfo := range con***REMOVED***g.AuthInfos {
		refs = append(refs, GetAuthInfoFileReferences(authInfo)...)
	}

	return refs
}

func GetClusterFileReferences(cluster *clientcmdapi.Cluster) []*string {
	return []*string{&cluster.Certi***REMOVED***cateAuthority}
}

func GetAuthInfoFileReferences(authInfo *clientcmdapi.AuthInfo) []*string {
	s := []*string{&authInfo.ClientCerti***REMOVED***cate, &authInfo.ClientKey, &authInfo.TokenFile}
	// Only resolve exec command if it isn't PATH based.
	if authInfo.Exec != nil && strings.ContainsRune(authInfo.Exec.Command, ***REMOVED***lepath.Separator) {
		s = append(s, &authInfo.Exec.Command)
	}
	return s
}

// ResolvePaths updates the given refs to be absolute paths, relative to the given base directory
func ResolvePaths(refs []*string, base string) error {
	for _, ref := range refs {
		// Don't resolve empty paths
		if len(*ref) > 0 {
			// Don't resolve absolute paths
			if !***REMOVED***lepath.IsAbs(*ref) {
				*ref = ***REMOVED***lepath.Join(base, *ref)
			}
		}
	}
	return nil
}

// RelativizePathWithNoBacksteps updates the given refs to be relative paths, relative to the given base directory as long as they do not require backsteps.
// Any path requiring a backstep is left as-is as long it is absolute.  Any non-absolute path that can't be relativized produces an error
func RelativizePathWithNoBacksteps(refs []*string, base string) error {
	for _, ref := range refs {
		// Don't relativize empty paths
		if len(*ref) > 0 {
			rel, err := MakeRelative(*ref, base)
			if err != nil {
				return err
			}

			// if we have a backstep, don't mess with the path
			if strings.HasPre***REMOVED***x(rel, "../") {
				if ***REMOVED***lepath.IsAbs(*ref) {
					continue
				}

				return fmt.Errorf("%v requires backsteps and is not absolute", *ref)
			}

			*ref = rel
		}
	}
	return nil
}

func MakeRelative(path, base string) (string, error) {
	if len(path) > 0 {
		rel, err := ***REMOVED***lepath.Rel(base, path)
		if err != nil {
			return path, err
		}
		return rel, nil
	}
	return path, nil
}

// deduplicate removes any duplicated values and returns a new slice, keeping the order unchanged
func deduplicate(s []string) []string {
	encountered := map[string]bool{}
	ret := make([]string, 0)
	for i := range s {
		if encountered[s[i]] {
			continue
		}
		encountered[s[i]] = true
		ret = append(ret, s[i])
	}
	return ret
}
