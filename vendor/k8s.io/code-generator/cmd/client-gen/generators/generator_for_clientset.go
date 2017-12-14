/*
Copyright 2016 The Kubernetes Authors.

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

package generators

import (
	"fmt"
	"io"
	"path/***REMOVED***lepath"
	"strings"

	clientgentypes "k8s.io/code-generator/cmd/client-gen/types"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// genClientset generates a package for a clientset.
type genClientset struct {
	generator.DefaultGen
	groups             []clientgentypes.GroupVersions
	groupGoNames       map[clientgentypes.GroupVersion]string
	clientsetPackage   string
	outputPackage      string
	imports            namer.ImportTracker
	clientsetGenerated bool
}

var _ generator.Generator = &genClientset{}

func (g *genClientset) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

// We only want to call GenerateType() once.
func (g *genClientset) Filter(c *generator.Context, t *types.Type) bool {
	ret := !g.clientsetGenerated
	g.clientsetGenerated = true
	return ret
}

func (g *genClientset) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	for _, group := range g.groups {
		for _, version := range group.Versions {
			typedClientPath := ***REMOVED***lepath.Join(g.clientsetPackage, "typed", group.PackageName, version.NonEmpty())
			groupAlias := strings.ToLower(g.groupGoNames[clientgentypes.GroupVersion{group.Group, version}])
			imports = append(imports, strings.ToLower(fmt.Sprintf("%s%s \"%s\"", groupAlias, version.NonEmpty(), typedClientPath)))
		}
	}
	return
}

func (g *genClientset) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	// TODO: We actually don't need any type information to generate the clientset,
	// perhaps we can adapt the go2ild framework to this kind of usage.
	sw := generator.NewSnippetWriter(w, c, "$", "$")

	allGroups := clientgentypes.ToGroupVersionPackages(g.groups, g.groupGoNames)
	m := map[string]interface{}{
		"allGroups":                            allGroups,
		"Con***REMOVED***g":                               c.Universe.Type(types.Name{Package: "k8s.io/client-go/rest", Name: "Con***REMOVED***g"}),
		"DefaultKubernetesUserAgent":           c.Universe.Function(types.Name{Package: "k8s.io/client-go/rest", Name: "DefaultKubernetesUserAgent"}),
		"RESTClientInterface":                  c.Universe.Type(types.Name{Package: "k8s.io/client-go/rest", Name: "Interface"}),
		"DiscoveryInterface":                   c.Universe.Type(types.Name{Package: "k8s.io/client-go/discovery", Name: "DiscoveryInterface"}),
		"DiscoveryClient":                      c.Universe.Type(types.Name{Package: "k8s.io/client-go/discovery", Name: "DiscoveryClient"}),
		"NewDiscoveryClientForCon***REMOVED***g":          c.Universe.Function(types.Name{Package: "k8s.io/client-go/discovery", Name: "NewDiscoveryClientForCon***REMOVED***g"}),
		"NewDiscoveryClientForCon***REMOVED***gOrDie":     c.Universe.Function(types.Name{Package: "k8s.io/client-go/discovery", Name: "NewDiscoveryClientForCon***REMOVED***gOrDie"}),
		"NewDiscoveryClient":                   c.Universe.Function(types.Name{Package: "k8s.io/client-go/discovery", Name: "NewDiscoveryClient"}),
		"flowcontrolNewTokenBucketRateLimiter": c.Universe.Function(types.Name{Package: "k8s.io/client-go/util/flowcontrol", Name: "NewTokenBucketRateLimiter"}),
		"glogErrorf":                           c.Universe.Function(types.Name{Package: "github.com/golang/glog", Name: "Errorf"}),
	}
	sw.Do(clientsetInterface, m)
	sw.Do(clientsetTemplate, m)
	for _, g := range allGroups {
		sw.Do(clientsetInterfaceImplTemplate, g)
		// don't generated the default method if generating internalversion clientset
		if g.IsDefaultVersion && g.Version != "" {
			sw.Do(clientsetInterfaceDefaultVersionImpl, g)
		}
	}
	sw.Do(getDiscoveryTemplate, m)
	sw.Do(newClientsetForCon***REMOVED***gTemplate, m)
	sw.Do(newClientsetForCon***REMOVED***gOrDieTemplate, m)
	sw.Do(newClientsetForRESTClientTemplate, m)

	return sw.Error()
}

var clientsetInterface = `
type Interface interface {
	Discovery() $.DiscoveryInterface|raw$
    $range .allGroups$$.GroupGoName$$.Version$() $.PackageAlias$.$.GroupGoName$$.Version$Interface
	$if .IsDefaultVersion$// Deprecated: please explicitly pick a version if possible.
	$.GroupGoName$() $.PackageAlias$.$.GroupGoName$$.Version$Interface
	$end$$end$
}
`

var clientsetTemplate = `
// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*$.DiscoveryClient|raw$
    $range .allGroups$$.LowerCaseGroupGoName$$.Version$ *$.PackageAlias$.$.GroupGoName$$.Version$Client
    $end$
}
`

var clientsetInterfaceImplTemplate = `
// $.GroupGoName$$.Version$ retrieves the $.GroupGoName$$.Version$Client
func (c *Clientset) $.GroupGoName$$.Version$() $.PackageAlias$.$.GroupGoName$$.Version$Interface {
	return c.$.LowerCaseGroupGoName$$.Version$
}
`

var clientsetInterfaceDefaultVersionImpl = `
// Deprecated: $.GroupGoName$ retrieves the default version of $.GroupGoName$Client.
// Please explicitly pick a version.
func (c *Clientset) $.GroupGoName$() $.PackageAlias$.$.GroupGoName$$.Version$Interface {
	return c.$.LowerCaseGroupGoName$$.Version$
}
`

var getDiscoveryTemplate = `
// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() $.DiscoveryInterface|raw$ {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}
`

var newClientsetForCon***REMOVED***gTemplate = `
// NewForCon***REMOVED***g creates a new Clientset for the given con***REMOVED***g.
func NewForCon***REMOVED***g(c *$.Con***REMOVED***g|raw$) (*Clientset, error) {
	con***REMOVED***gShallowCopy := *c
	if con***REMOVED***gShallowCopy.RateLimiter == nil && con***REMOVED***gShallowCopy.QPS > 0 {
		con***REMOVED***gShallowCopy.RateLimiter = $.flowcontrolNewTokenBucketRateLimiter|raw$(con***REMOVED***gShallowCopy.QPS, con***REMOVED***gShallowCopy.Burst)
	}
	var cs Clientset
	var err error
$range .allGroups$    cs.$.LowerCaseGroupGoName$$.Version$, err =$.PackageAlias$.NewForCon***REMOVED***g(&con***REMOVED***gShallowCopy)
	if err!=nil {
		return nil, err
	}
$end$
	cs.DiscoveryClient, err = $.NewDiscoveryClientForCon***REMOVED***g|raw$(&con***REMOVED***gShallowCopy)
	if err!=nil {
		$.glogErrorf|raw$("failed to create the DiscoveryClient: %v", err)
		return nil, err
	}
	return &cs, nil
}
`

var newClientsetForCon***REMOVED***gOrDieTemplate = `
// NewForCon***REMOVED***gOrDie creates a new Clientset for the given con***REMOVED***g and
// panics if there is an error in the con***REMOVED***g.
func NewForCon***REMOVED***gOrDie(c *$.Con***REMOVED***g|raw$) *Clientset {
	var cs Clientset
$range .allGroups$    cs.$.LowerCaseGroupGoName$$.Version$ =$.PackageAlias$.NewForCon***REMOVED***gOrDie(c)
$end$
	cs.DiscoveryClient = $.NewDiscoveryClientForCon***REMOVED***gOrDie|raw$(c)
	return &cs
}
`

var newClientsetForRESTClientTemplate = `
// New creates a new Clientset for the given RESTClient.
func New(c $.RESTClientInterface|raw$) *Clientset {
	var cs Clientset
$range .allGroups$    cs.$.LowerCaseGroupGoName$$.Version$ =$.PackageAlias$.New(c)
$end$
	cs.DiscoveryClient = $.NewDiscoveryClient|raw$(c)
	return &cs
}
`
