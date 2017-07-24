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

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	legacyAPIPath  = "/api"
	defaultAPIPath = "/apis"
)

// TODO: Is this obsoleted by the discovery client?

// ServerAPIVersions returns the GroupVersions supported by the API server.
// It creates a RESTClient based on the passed in con***REMOVED***g, but it doesn't rely
// on the Version and Codec of the con***REMOVED***g, because it uses AbsPath and
// takes the raw response.
func ServerAPIVersions(c *Con***REMOVED***g) (groupVersions []string, err error) {
	transport, err := TransportFor(c)
	if err != nil {
		return nil, err
	}
	client := http.Client{Transport: transport}

	con***REMOVED***gCopy := *c
	con***REMOVED***gCopy.GroupVersion = nil
	con***REMOVED***gCopy.APIPath = ""
	baseURL, _, err := defaultServerUrlFor(&con***REMOVED***gCopy)
	if err != nil {
		return nil, err
	}
	// Get the groupVersions exposed at /api
	originalPath := baseURL.Path
	baseURL.Path = path.Join(originalPath, legacyAPIPath)
	resp, err := client.Get(baseURL.String())
	if err != nil {
		return nil, err
	}
	var v metav1.APIVersions
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return nil, fmt.Errorf("unexpected error: %v", err)
	}

	groupVersions = append(groupVersions, v.Versions...)
	// Get the groupVersions exposed at /apis
	baseURL.Path = path.Join(originalPath, defaultAPIPath)
	resp2, err := client.Get(baseURL.String())
	if err != nil {
		return nil, err
	}
	var apiGroupList metav1.APIGroupList
	defer resp2.Body.Close()
	err = json.NewDecoder(resp2.Body).Decode(&apiGroupList)
	if err != nil {
		return nil, fmt.Errorf("unexpected error: %v", err)
	}

	for _, g := range apiGroupList.Groups {
		for _, gv := range g.Versions {
			groupVersions = append(groupVersions, gv.GroupVersion)
		}
	}

	return groupVersions, nil
}
