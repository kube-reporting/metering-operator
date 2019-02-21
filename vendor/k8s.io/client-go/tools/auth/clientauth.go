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

/*
Package auth de***REMOVED***nes a ***REMOVED***le format for holding authentication
information needed by clients of Kubernetes.  Typically,
a Kubernetes cluster will put auth info for the admin in a known
location when it is created, and will (soon) put it in a known
location within a Container's ***REMOVED***le tree for Containers that
need access to the Kubernetes API.

Having a de***REMOVED***ned format allows:
  - clients to be implemented in multiple languages
  - applications which link clients to be portable across
    clusters with different authentication styles (e.g.
    some may use SSL Client certs, others may not, etc)
  - when the format changes, applications only
    need to update this code.

The ***REMOVED***le format is json, marshalled from a struct authcfg.Info.

Clinet libraries in other languages should use the same format.

It is not intended to store general preferences, such as default
namespace, output options, etc.  CLIs (such as kubectl) and UIs should
develop their own format and may wish to inline the authcfg.Info type.

The authcfg.Info is just a ***REMOVED***le format.  It is distinct from
client.Con***REMOVED***g which holds options for creating a client.Client.
Helper functions are provided in this package to ***REMOVED***ll in a
client.Client from an authcfg.Info.

Example:

    import (
        "pkg/client"
        "pkg/client/auth"
    )

    info, err := auth.LoadFromFile(***REMOVED***lename)
    if err != nil {
      // handle error
    }
    clientCon***REMOVED***g = client.Con***REMOVED***g{}
    clientCon***REMOVED***g.Host = "example.com:4901"
    clientCon***REMOVED***g = info.MergeWithCon***REMOVED***g()
    client := client.New(clientCon***REMOVED***g)
    client.Pods(ns).List()
*/
package auth

// TODO: need a way to rotate Tokens.  Therefore, need a way for client object to be reset when the authcfg is updated.
import (
	"encoding/json"
	"io/ioutil"
	"os"

	restclient "k8s.io/client-go/rest"
)

// Info holds Kubernetes API authorization con***REMOVED***g.  It is intended
// to be read/written from a ***REMOVED***le as a JSON object.
type Info struct {
	User        string
	Password    string
	CAFile      string
	CertFile    string
	KeyFile     string
	BearerToken string
	Insecure    *bool
}

// LoadFromFile parses an Info object from a ***REMOVED***le path.
// If the ***REMOVED***le does not exist, then os.IsNotExist(err) == true
func LoadFromFile(path string) (*Info, error) {
	var info Info
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}
	return &info, err
}

// MergeWithCon***REMOVED***g returns a copy of a client.Con***REMOVED***g with values from the Info.
// The ***REMOVED***elds of client.Con***REMOVED***g with a corresponding ***REMOVED***eld in the Info are set
// with the value from the Info.
func (info Info) MergeWithCon***REMOVED***g(c restclient.Con***REMOVED***g) (restclient.Con***REMOVED***g, error) {
	var con***REMOVED***g restclient.Con***REMOVED***g = c
	con***REMOVED***g.Username = info.User
	con***REMOVED***g.Password = info.Password
	con***REMOVED***g.CAFile = info.CAFile
	con***REMOVED***g.CertFile = info.CertFile
	con***REMOVED***g.KeyFile = info.KeyFile
	con***REMOVED***g.BearerToken = info.BearerToken
	if info.Insecure != nil {
		con***REMOVED***g.Insecure = *info.Insecure
	}
	return con***REMOVED***g, nil
}

func (info Info) Complete() bool {
	return len(info.User) > 0 ||
		len(info.CertFile) > 0 ||
		len(info.BearerToken) > 0
}
