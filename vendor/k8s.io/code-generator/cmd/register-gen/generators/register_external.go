/*
Copyright 2018 The Kubernetes Authors.

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
	"io"
	"sort"

	clientgentypes "k8s.io/code-generator/cmd/client-gen/types"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

type registerExternalGenerator struct {
	generator.DefaultGen
	outputPackage   string
	gv              clientgentypes.GroupVersion
	typesToGenerate []*types.Type
	imports         namer.ImportTracker
}

var _ generator.Generator = &registerExternalGenerator{}

func (g *registerExternalGenerator) Filter(_ *generator.Context, _ *types.Type) bool {
	return false
}

func (g *registerExternalGenerator) Imports(c *generator.Context) (imports []string) {
	return g.imports.ImportLines()
}

func (g *registerExternalGenerator) Namers(_ *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *registerExternalGenerator) Finalize(context *generator.Context, w io.Writer) error {
	typesToGenerateOnlyNames := make([]string, len(g.typesToGenerate))
	for index, typeToGenerate := range g.typesToGenerate {
		typesToGenerateOnlyNames[index] = typeToGenerate.Name.Name
	}

	// sort the list of types to register, so that the generator produces stable output
	sort.Strings(typesToGenerateOnlyNames)

	sw := generator.NewSnippetWriter(w, context, "$", "$")
	m := map[string]interface{}{
		"groupName":         g.gv.Group,
		"version":           g.gv.Version,
		"types":             typesToGenerateOnlyNames,
		"addToGroupVersion": context.Universe.Function(types.Name{Package: "k8s.io/apimachinery/pkg/apis/meta/v1", Name: "AddToGroupVersion"}),
		"groupVersion":      context.Universe.Type(types.Name{Package: "k8s.io/apimachinery/pkg/apis/meta/v1", Name: "GroupVersion"}),
	}
	sw.Do(registerExternalTypesTemplate, m)
	return sw.Error()
}

var registerExternalTypesTemplate = `
// GroupName speci***REMOVED***es the group name used to register the objects.
const GroupName = "$.groupName$"

// GroupVersion speci***REMOVED***es the group and the version used to register the objects.
var GroupVersion = $.groupVersion|raw${Group: GroupName, Version: "$.version$"}

// SchemeGroupVersion is group version used to register these objects
// Deprecated: use GroupVersion instead.
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "$.version$"}

// Resource takes an unquali***REMOVED***ed resource and returns a Group quali***REMOVED***ed GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// localSchemeBuilder and AddToScheme will stay in k8s.io/kubernetes.
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
    // Depreciated: use Install instead
	AddToScheme        = localSchemeBuilder.AddToScheme
	Install            = localSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated ***REMOVED***les. The separation
	// makes the code compile even when the generated ***REMOVED***les are missing.
	localSchemeBuilder.Register(addKnownTypes)
}

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
    $range .types -$
        &$.${},
    $end$
	)
    // AddToGroupVersion allows the serialization of client types like ListOptions.
	$.addToGroupVersion|raw$(scheme, SchemeGroupVersion)
	return nil
}
`
