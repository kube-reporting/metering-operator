// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this ***REMOVED***le except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the speci***REMOVED***c language governing permissions and
// limitations under the License.

package swag

import (
	"os"
	"path/***REMOVED***lepath"
	"runtime"
	"strings"
)

const (
	// GOPATHKey represents the env key for gopath
	GOPATHKey = "GOPATH"
)

// FindInSearchPath ***REMOVED***nds a package in a provided lists of paths
func FindInSearchPath(searchPath, pkg string) string {
	pathsList := ***REMOVED***lepath.SplitList(searchPath)
	for _, path := range pathsList {
		if evaluatedPath, err := ***REMOVED***lepath.EvalSymlinks(***REMOVED***lepath.Join(path, "src", pkg)); err == nil {
			if _, err := os.Stat(evaluatedPath); err == nil {
				return evaluatedPath
			}
		}
	}
	return ""
}

// FindInGoSearchPath ***REMOVED***nds a package in the $GOPATH:$GOROOT
func FindInGoSearchPath(pkg string) string {
	return FindInSearchPath(FullGoSearchPath(), pkg)
}

// FullGoSearchPath gets the search paths for ***REMOVED***nding packages
func FullGoSearchPath() string {
	allPaths := os.Getenv(GOPATHKey)
	if allPaths != "" {
		allPaths = strings.Join([]string{allPaths, runtime.GOROOT()}, ":")
	} ***REMOVED*** {
		allPaths = runtime.GOROOT()
	}
	return allPaths
}
