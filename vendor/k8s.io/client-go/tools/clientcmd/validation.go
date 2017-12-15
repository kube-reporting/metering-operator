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
	"fmt"
	"os"
	"reflect"
	"strings"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	ErrNoContext   = errors.New("no context chosen")
	ErrEmptyCon***REMOVED***g = errors.New("no con***REMOVED***guration has been provided")
	// message is for consistency with old behavior
	ErrEmptyCluster = errors.New("cluster has no server de***REMOVED***ned")
)

type errContextNotFound struct {
	ContextName string
}

func (e *errContextNotFound) Error() string {
	return fmt.Sprintf("context was not found for speci***REMOVED***ed context: %v", e.ContextName)
}

// IsContextNotFound returns a boolean indicating whether the error is known to
// report that a context was not found
func IsContextNotFound(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*errContextNotFound); ok || err == ErrNoContext {
		return true
	}
	return strings.Contains(err.Error(), "context was not found for speci***REMOVED***ed context")
}

// IsEmptyCon***REMOVED***g returns true if the provided error indicates the provided con***REMOVED***guration
// is empty.
func IsEmptyCon***REMOVED***g(err error) bool {
	switch t := err.(type) {
	case errCon***REMOVED***gurationInvalid:
		return len(t) == 1 && t[0] == ErrEmptyCon***REMOVED***g
	}
	return err == ErrEmptyCon***REMOVED***g
}

// errCon***REMOVED***gurationInvalid is a set of errors indicating the con***REMOVED***guration is invalid.
type errCon***REMOVED***gurationInvalid []error

// errCon***REMOVED***gurationInvalid implements error and Aggregate
var _ error = errCon***REMOVED***gurationInvalid{}
var _ utilerrors.Aggregate = errCon***REMOVED***gurationInvalid{}

func newErrCon***REMOVED***gurationInvalid(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	default:
		return errCon***REMOVED***gurationInvalid(errs)
	}
}

// Error implements the error interface
func (e errCon***REMOVED***gurationInvalid) Error() string {
	return fmt.Sprintf("invalid con***REMOVED***guration: %v", utilerrors.NewAggregate(e).Error())
}

// Errors implements the AggregateError interface
func (e errCon***REMOVED***gurationInvalid) Errors() []error {
	return e
}

// IsCon***REMOVED***gurationInvalid returns true if the provided error indicates the con***REMOVED***guration is invalid.
func IsCon***REMOVED***gurationInvalid(err error) bool {
	switch err.(type) {
	case *errContextNotFound, errCon***REMOVED***gurationInvalid:
		return true
	}
	return IsContextNotFound(err)
}

// Validate checks for errors in the Con***REMOVED***g.  It does not return early so that it can ***REMOVED***nd as many errors as possible.
func Validate(con***REMOVED***g clientcmdapi.Con***REMOVED***g) error {
	validationErrors := make([]error, 0)

	if clientcmdapi.IsCon***REMOVED***gEmpty(&con***REMOVED***g) {
		return newErrCon***REMOVED***gurationInvalid([]error{ErrEmptyCon***REMOVED***g})
	}

	if len(con***REMOVED***g.CurrentContext) != 0 {
		if _, exists := con***REMOVED***g.Contexts[con***REMOVED***g.CurrentContext]; !exists {
			validationErrors = append(validationErrors, &errContextNotFound{con***REMOVED***g.CurrentContext})
		}
	}

	for contextName, context := range con***REMOVED***g.Contexts {
		validationErrors = append(validationErrors, validateContext(contextName, *context, con***REMOVED***g)...)
	}

	for authInfoName, authInfo := range con***REMOVED***g.AuthInfos {
		validationErrors = append(validationErrors, validateAuthInfo(authInfoName, *authInfo)...)
	}

	for clusterName, clusterInfo := range con***REMOVED***g.Clusters {
		validationErrors = append(validationErrors, validateClusterInfo(clusterName, *clusterInfo)...)
	}

	return newErrCon***REMOVED***gurationInvalid(validationErrors)
}

// Con***REMOVED***rmUsable looks a particular context and determines if that particular part of the con***REMOVED***g is useable.  There might still be errors in the con***REMOVED***g,
// but no errors in the sections requested or referenced.  It does not return early so that it can ***REMOVED***nd as many errors as possible.
func Con***REMOVED***rmUsable(con***REMOVED***g clientcmdapi.Con***REMOVED***g, passedContextName string) error {
	validationErrors := make([]error, 0)

	if clientcmdapi.IsCon***REMOVED***gEmpty(&con***REMOVED***g) {
		return newErrCon***REMOVED***gurationInvalid([]error{ErrEmptyCon***REMOVED***g})
	}

	var contextName string
	if len(passedContextName) != 0 {
		contextName = passedContextName
	} ***REMOVED*** {
		contextName = con***REMOVED***g.CurrentContext
	}

	if len(contextName) == 0 {
		return ErrNoContext
	}

	context, exists := con***REMOVED***g.Contexts[contextName]
	if !exists {
		validationErrors = append(validationErrors, &errContextNotFound{contextName})
	}

	if exists {
		validationErrors = append(validationErrors, validateContext(contextName, *context, con***REMOVED***g)...)
		validationErrors = append(validationErrors, validateAuthInfo(context.AuthInfo, *con***REMOVED***g.AuthInfos[context.AuthInfo])...)
		validationErrors = append(validationErrors, validateClusterInfo(context.Cluster, *con***REMOVED***g.Clusters[context.Cluster])...)
	}

	return newErrCon***REMOVED***gurationInvalid(validationErrors)
}

// validateClusterInfo looks for conflicts and errors in the cluster info
func validateClusterInfo(clusterName string, clusterInfo clientcmdapi.Cluster) []error {
	validationErrors := make([]error, 0)

	emptyCluster := clientcmdapi.NewCluster()
	if reflect.DeepEqual(*emptyCluster, clusterInfo) {
		return []error{ErrEmptyCluster}
	}

	if len(clusterInfo.Server) == 0 {
		if len(clusterName) == 0 {
			validationErrors = append(validationErrors, fmt.Errorf("default cluster has no server de***REMOVED***ned"))
		} ***REMOVED*** {
			validationErrors = append(validationErrors, fmt.Errorf("no server found for cluster %q", clusterName))
		}
	}
	// Make sure CA data and CA ***REMOVED***le aren't both speci***REMOVED***ed
	if len(clusterInfo.Certi***REMOVED***cateAuthority) != 0 && len(clusterInfo.Certi***REMOVED***cateAuthorityData) != 0 {
		validationErrors = append(validationErrors, fmt.Errorf("certi***REMOVED***cate-authority-data and certi***REMOVED***cate-authority are both speci***REMOVED***ed for %v. certi***REMOVED***cate-authority-data will override.", clusterName))
	}
	if len(clusterInfo.Certi***REMOVED***cateAuthority) != 0 {
		clientCertCA, err := os.Open(clusterInfo.Certi***REMOVED***cateAuthority)
		defer clientCertCA.Close()
		if err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("unable to read certi***REMOVED***cate-authority %v for %v due to %v", clusterInfo.Certi***REMOVED***cateAuthority, clusterName, err))
		}
	}

	return validationErrors
}

// validateAuthInfo looks for conflicts and errors in the auth info
func validateAuthInfo(authInfoName string, authInfo clientcmdapi.AuthInfo) []error {
	validationErrors := make([]error, 0)

	usingAuthPath := false
	methods := make([]string, 0, 3)
	if len(authInfo.Token) != 0 {
		methods = append(methods, "token")
	}
	if len(authInfo.Username) != 0 || len(authInfo.Password) != 0 {
		methods = append(methods, "basicAuth")
	}

	if len(authInfo.ClientCerti***REMOVED***cate) != 0 || len(authInfo.ClientCerti***REMOVED***cateData) != 0 {
		// Make sure cert data and ***REMOVED***le aren't both speci***REMOVED***ed
		if len(authInfo.ClientCerti***REMOVED***cate) != 0 && len(authInfo.ClientCerti***REMOVED***cateData) != 0 {
			validationErrors = append(validationErrors, fmt.Errorf("client-cert-data and client-cert are both speci***REMOVED***ed for %v. client-cert-data will override.", authInfoName))
		}
		// Make sure key data and ***REMOVED***le aren't both speci***REMOVED***ed
		if len(authInfo.ClientKey) != 0 && len(authInfo.ClientKeyData) != 0 {
			validationErrors = append(validationErrors, fmt.Errorf("client-key-data and client-key are both speci***REMOVED***ed for %v; client-key-data will override", authInfoName))
		}
		// Make sure a key is speci***REMOVED***ed
		if len(authInfo.ClientKey) == 0 && len(authInfo.ClientKeyData) == 0 {
			validationErrors = append(validationErrors, fmt.Errorf("client-key-data or client-key must be speci***REMOVED***ed for %v to use the clientCert authentication method.", authInfoName))
		}

		if len(authInfo.ClientCerti***REMOVED***cate) != 0 {
			clientCertFile, err := os.Open(authInfo.ClientCerti***REMOVED***cate)
			defer clientCertFile.Close()
			if err != nil {
				validationErrors = append(validationErrors, fmt.Errorf("unable to read client-cert %v for %v due to %v", authInfo.ClientCerti***REMOVED***cate, authInfoName, err))
			}
		}
		if len(authInfo.ClientKey) != 0 {
			clientKeyFile, err := os.Open(authInfo.ClientKey)
			defer clientKeyFile.Close()
			if err != nil {
				validationErrors = append(validationErrors, fmt.Errorf("unable to read client-key %v for %v due to %v", authInfo.ClientKey, authInfoName, err))
			}
		}
	}

	// authPath also provides information for the client to identify the server, so allow multiple auth methods in that case
	if (len(methods) > 1) && (!usingAuthPath) {
		validationErrors = append(validationErrors, fmt.Errorf("more than one authentication method found for %v; found %v, only one is allowed", authInfoName, methods))
	}

	// ImpersonateGroups or ImpersonateUserExtra should be requested with a user
	if (len(authInfo.ImpersonateGroups) > 0 || len(authInfo.ImpersonateUserExtra) > 0) && (len(authInfo.Impersonate) == 0) {
		validationErrors = append(validationErrors, fmt.Errorf("requesting groups or user-extra for %v without impersonating a user", authInfoName))
	}
	return validationErrors
}

// validateContext looks for errors in the context.  It is not transitive, so errors in the reference authInfo or cluster con***REMOVED***gs are not included in this return
func validateContext(contextName string, context clientcmdapi.Context, con***REMOVED***g clientcmdapi.Con***REMOVED***g) []error {
	validationErrors := make([]error, 0)

	if len(context.AuthInfo) == 0 {
		validationErrors = append(validationErrors, fmt.Errorf("user was not speci***REMOVED***ed for context %q", contextName))
	} ***REMOVED*** if _, exists := con***REMOVED***g.AuthInfos[context.AuthInfo]; !exists {
		validationErrors = append(validationErrors, fmt.Errorf("user %q was not found for context %q", context.AuthInfo, contextName))
	}

	if len(context.Cluster) == 0 {
		validationErrors = append(validationErrors, fmt.Errorf("cluster was not speci***REMOVED***ed for context %q", contextName))
	} ***REMOVED*** if _, exists := con***REMOVED***g.Clusters[context.Cluster]; !exists {
		validationErrors = append(validationErrors, fmt.Errorf("cluster %q was not found for context %q", context.Cluster, contextName))
	}

	if len(context.Namespace) != 0 {
		if len(validation.IsDNS1123Label(context.Namespace)) != 0 {
			validationErrors = append(validationErrors, fmt.Errorf("namespace %q for context %q does not conform to the kubernetes DNS_LABEL rules", context.Namespace, contextName))
		}
	}

	return validationErrors
}
