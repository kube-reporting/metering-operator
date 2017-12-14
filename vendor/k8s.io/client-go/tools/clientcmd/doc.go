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
Package clientcmd provides one stop shopping for building a working client from a ***REMOVED***xed con***REMOVED***g,
from a .kubecon***REMOVED***g ***REMOVED***le, from command line flags, or from any merged combination.

Sample usage from merged .kubecon***REMOVED***g ***REMOVED***les (local directory, home directory)

	loadingRules := clientcmd.NewDefaultClientCon***REMOVED***gLoadingRules()
	// if you want to change the loading rules (which ***REMOVED***les in which order), you can do so here

	con***REMOVED***gOverrides := &clientcmd.Con***REMOVED***gOverrides{}
	// if you want to change override values or bind them to flags, there are methods to help you

	kubeCon***REMOVED***g := clientcmd.NewNonInteractiveDeferredLoadingClientCon***REMOVED***g(loadingRules, con***REMOVED***gOverrides)
	con***REMOVED***g, err := kubeCon***REMOVED***g.ClientCon***REMOVED***g()
	if err != nil {
		// Do something
	}
	client, err := metav1.New(con***REMOVED***g)
	// ...
*/
package clientcmd // import "k8s.io/client-go/tools/clientcmd"
