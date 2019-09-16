package adal

// Copyright 2017 Microsoft Corporation
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this ***REMOVED***le except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the speci***REMOVED***c language governing permissions and
//  limitations under the License.

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	activeDirectoryEndpointTemplate = "%s/oauth2/%s%s"
)

// OAuthCon***REMOVED***g represents the endpoints needed
// in OAuth operations
type OAuthCon***REMOVED***g struct {
	AuthorityEndpoint  url.URL `json:"authorityEndpoint"`
	AuthorizeEndpoint  url.URL `json:"authorizeEndpoint"`
	TokenEndpoint      url.URL `json:"tokenEndpoint"`
	DeviceCodeEndpoint url.URL `json:"deviceCodeEndpoint"`
}

// IsZero returns true if the OAuthCon***REMOVED***g object is zero-initialized.
func (oac OAuthCon***REMOVED***g) IsZero() bool {
	return oac == OAuthCon***REMOVED***g{}
}

func validateStringParam(param, name string) error {
	if len(param) == 0 {
		return fmt.Errorf("parameter '" + name + "' cannot be empty")
	}
	return nil
}

// NewOAuthCon***REMOVED***g returns an OAuthCon***REMOVED***g with tenant speci***REMOVED***c urls
func NewOAuthCon***REMOVED***g(activeDirectoryEndpoint, tenantID string) (*OAuthCon***REMOVED***g, error) {
	apiVer := "1.0"
	return NewOAuthCon***REMOVED***gWithAPIVersion(activeDirectoryEndpoint, tenantID, &apiVer)
}

// NewOAuthCon***REMOVED***gWithAPIVersion returns an OAuthCon***REMOVED***g with tenant speci***REMOVED***c urls.
// If apiVersion is not nil the "api-version" query parameter will be appended to the endpoint URLs with the speci***REMOVED***ed value.
func NewOAuthCon***REMOVED***gWithAPIVersion(activeDirectoryEndpoint, tenantID string, apiVersion *string) (*OAuthCon***REMOVED***g, error) {
	if err := validateStringParam(activeDirectoryEndpoint, "activeDirectoryEndpoint"); err != nil {
		return nil, err
	}
	api := ""
	// it's legal for tenantID to be empty so don't validate it
	if apiVersion != nil {
		if err := validateStringParam(*apiVersion, "apiVersion"); err != nil {
			return nil, err
		}
		api = fmt.Sprintf("?api-version=%s", *apiVersion)
	}
	u, err := url.Parse(activeDirectoryEndpoint)
	if err != nil {
		return nil, err
	}
	authorityURL, err := u.Parse(tenantID)
	if err != nil {
		return nil, err
	}
	authorizeURL, err := u.Parse(fmt.Sprintf(activeDirectoryEndpointTemplate, tenantID, "authorize", api))
	if err != nil {
		return nil, err
	}
	tokenURL, err := u.Parse(fmt.Sprintf(activeDirectoryEndpointTemplate, tenantID, "token", api))
	if err != nil {
		return nil, err
	}
	deviceCodeURL, err := u.Parse(fmt.Sprintf(activeDirectoryEndpointTemplate, tenantID, "devicecode", api))
	if err != nil {
		return nil, err
	}

	return &OAuthCon***REMOVED***g{
		AuthorityEndpoint:  *authorityURL,
		AuthorizeEndpoint:  *authorizeURL,
		TokenEndpoint:      *tokenURL,
		DeviceCodeEndpoint: *deviceCodeURL,
	}, nil
}

// MultiTenantOAuthCon***REMOVED***g provides endpoints for primary and aulixiary tenant IDs.
type MultiTenantOAuthCon***REMOVED***g interface {
	PrimaryTenant() *OAuthCon***REMOVED***g
	AuxiliaryTenants() []*OAuthCon***REMOVED***g
}

// OAuthOptions contains optional OAuthCon***REMOVED***g creation arguments.
type OAuthOptions struct {
	APIVersion string
}

func (c OAuthOptions) apiVersion() string {
	if c.APIVersion != "" {
		return fmt.Sprintf("?api-version=%s", c.APIVersion)
	}
	return "1.0"
}

// NewMultiTenantOAuthCon***REMOVED***g creates an object that support multitenant OAuth con***REMOVED***guration.
// See https://docs.microsoft.com/en-us/azure/azure-resource-manager/authenticate-multi-tenant for more information.
func NewMultiTenantOAuthCon***REMOVED***g(activeDirectoryEndpoint, primaryTenantID string, auxiliaryTenantIDs []string, options OAuthOptions) (MultiTenantOAuthCon***REMOVED***g, error) {
	if len(auxiliaryTenantIDs) == 0 || len(auxiliaryTenantIDs) > 3 {
		return nil, errors.New("must specify one to three auxiliary tenants")
	}
	mtCfg := multiTenantOAuthCon***REMOVED***g{
		cfgs: make([]*OAuthCon***REMOVED***g, len(auxiliaryTenantIDs)+1),
	}
	apiVer := options.apiVersion()
	pri, err := NewOAuthCon***REMOVED***gWithAPIVersion(activeDirectoryEndpoint, primaryTenantID, &apiVer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OAuthCon***REMOVED***g for primary tenant: %v", err)
	}
	mtCfg.cfgs[0] = pri
	for i := range auxiliaryTenantIDs {
		aux, err := NewOAuthCon***REMOVED***g(activeDirectoryEndpoint, auxiliaryTenantIDs[i])
		if err != nil {
			return nil, fmt.Errorf("failed to create OAuthCon***REMOVED***g for tenant '%s': %v", auxiliaryTenantIDs[i], err)
		}
		mtCfg.cfgs[i+1] = aux
	}
	return mtCfg, nil
}

type multiTenantOAuthCon***REMOVED***g struct {
	// ***REMOVED***rst con***REMOVED***g in the slice is the primary tenant
	cfgs []*OAuthCon***REMOVED***g
}

func (m multiTenantOAuthCon***REMOVED***g) PrimaryTenant() *OAuthCon***REMOVED***g {
	return m.cfgs[0]
}

func (m multiTenantOAuthCon***REMOVED***g) AuxiliaryTenants() []*OAuthCon***REMOVED***g {
	return m.cfgs[1:]
}
