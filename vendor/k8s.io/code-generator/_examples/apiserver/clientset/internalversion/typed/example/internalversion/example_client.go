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

package internalversion

import (
	rest "k8s.io/client-go/rest"
	"k8s.io/code-generator/_examples/apiserver/clientset/internalversion/scheme"
)

type ExampleInterface interface {
	RESTClient() rest.Interface
	TestTypesGetter
}

// ExampleClient is used to interact with features provided by the example.apiserver.code-generator.k8s.io group.
type ExampleClient struct {
	restClient rest.Interface
}

func (c *ExampleClient) TestTypes(namespace string) TestTypeInterface {
	return newTestTypes(c, namespace)
}

// NewForCon***REMOVED***g creates a new ExampleClient for the given con***REMOVED***g.
func NewForCon***REMOVED***g(c *rest.Con***REMOVED***g) (*ExampleClient, error) {
	con***REMOVED***g := *c
	if err := setCon***REMOVED***gDefaults(&con***REMOVED***g); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&con***REMOVED***g)
	if err != nil {
		return nil, err
	}
	return &ExampleClient{client}, nil
}

// NewForCon***REMOVED***gOrDie creates a new ExampleClient for the given con***REMOVED***g and
// panics if there is an error in the con***REMOVED***g.
func NewForCon***REMOVED***gOrDie(c *rest.Con***REMOVED***g) *ExampleClient {
	client, err := NewForCon***REMOVED***g(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new ExampleClient for the given RESTClient.
func New(c rest.Interface) *ExampleClient {
	return &ExampleClient{c}
}

func setCon***REMOVED***gDefaults(con***REMOVED***g *rest.Con***REMOVED***g) error {
	g, err := scheme.Registry.Group("example.apiserver.code-generator.k8s.io")
	if err != nil {
		return err
	}

	con***REMOVED***g.APIPath = "/apis"
	if con***REMOVED***g.UserAgent == "" {
		con***REMOVED***g.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	if con***REMOVED***g.GroupVersion == nil || con***REMOVED***g.GroupVersion.Group != g.GroupVersion.Group {
		gv := g.GroupVersion
		con***REMOVED***g.GroupVersion = &gv
	}
	con***REMOVED***g.NegotiatedSerializer = scheme.Codecs

	if con***REMOVED***g.QPS == 0 {
		con***REMOVED***g.QPS = 5
	}
	if con***REMOVED***g.Burst == 0 {
		con***REMOVED***g.Burst = 10
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ExampleClient) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
