/*
Copyright 2017 The Kubernetes Authors.

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

package util

import (
	gobuild "go/build"
	"path"
	"path/***REMOVED***lepath"
	"reflect"
	"strings"
)

type empty struct{}

// CurrentPackage returns the go package of the current directory, or "" if it cannot
// be derived from the GOPATH.
func CurrentPackage() string {
	for _, root := range gobuild.Default.SrcDirs() {
		if pkg, ok := hasSubdir(root, "."); ok {
			return pkg
		}
	}
	return ""
}

func hasSubdir(root, dir string) (rel string, ok bool) {
	// ensure a tailing separator to properly compare on word-boundaries
	const sep = string(***REMOVED***lepath.Separator)
	root = ***REMOVED***lepath.Clean(root)
	if !strings.HasSuf***REMOVED***x(root, sep) {
		root += sep
	}

	// check whether root dir starts with root
	dir = ***REMOVED***lepath.Clean(dir)
	if !strings.HasPre***REMOVED***x(dir, root) {
		return "", false
	}

	// cut off root
	return ***REMOVED***lepath.ToSlash(dir[len(root):]), true
}

// BoilerplatePath uses the boilerplate in code-generator by calculating the relative path to it.
func BoilerplatePath() string {
	return path.Join(reflect.TypeOf(empty{}).PkgPath(), "/../../hack/boilerplate.go.txt")
}
