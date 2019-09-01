package azure

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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const (
	// EnvironmentFilepathName captures the name of the environment variable containing the path to the ***REMOVED***le
	// to be used while populating the Azure Environment.
	EnvironmentFilepathName = "AZURE_ENVIRONMENT_FILEPATH"

	// NotAvailable is used for endpoints and resource IDs that are not available for a given cloud.
	NotAvailable = "N/A"
)

var environments = map[string]Environment{
	"AZURECHINACLOUD":        ChinaCloud,
	"AZUREGERMANCLOUD":       GermanCloud,
	"AZUREPUBLICCLOUD":       PublicCloud,
	"AZUREUSGOVERNMENTCLOUD": USGovernmentCloud,
}

// ResourceIdenti***REMOVED***er contains a set of Azure resource IDs.
type ResourceIdenti***REMOVED***er struct {
	Graph               string `json:"graph"`
	KeyVault            string `json:"keyVault"`
	Datalake            string `json:"datalake"`
	Batch               string `json:"batch"`
	OperationalInsights string `json:"operationalInsights"`
	Storage             string `json:"storage"`
}

// Environment represents a set of endpoints for each of Azure's Clouds.
type Environment struct {
	Name                         string             `json:"name"`
	ManagementPortalURL          string             `json:"managementPortalURL"`
	PublishSettingsURL           string             `json:"publishSettingsURL"`
	ServiceManagementEndpoint    string             `json:"serviceManagementEndpoint"`
	ResourceManagerEndpoint      string             `json:"resourceManagerEndpoint"`
	ActiveDirectoryEndpoint      string             `json:"activeDirectoryEndpoint"`
	GalleryEndpoint              string             `json:"galleryEndpoint"`
	KeyVaultEndpoint             string             `json:"keyVaultEndpoint"`
	GraphEndpoint                string             `json:"graphEndpoint"`
	ServiceBusEndpoint           string             `json:"serviceBusEndpoint"`
	BatchManagementEndpoint      string             `json:"batchManagementEndpoint"`
	StorageEndpointSuf***REMOVED***x        string             `json:"storageEndpointSuf***REMOVED***x"`
	SQLDatabaseDNSSuf***REMOVED***x         string             `json:"sqlDatabaseDNSSuf***REMOVED***x"`
	Traf***REMOVED***cManagerDNSSuf***REMOVED***x      string             `json:"traf***REMOVED***cManagerDNSSuf***REMOVED***x"`
	KeyVaultDNSSuf***REMOVED***x            string             `json:"keyVaultDNSSuf***REMOVED***x"`
	ServiceBusEndpointSuf***REMOVED***x     string             `json:"serviceBusEndpointSuf***REMOVED***x"`
	ServiceManagementVMDNSSuf***REMOVED***x string             `json:"serviceManagementVMDNSSuf***REMOVED***x"`
	ResourceManagerVMDNSSuf***REMOVED***x   string             `json:"resourceManagerVMDNSSuf***REMOVED***x"`
	ContainerRegistryDNSSuf***REMOVED***x   string             `json:"containerRegistryDNSSuf***REMOVED***x"`
	CosmosDBDNSSuf***REMOVED***x            string             `json:"cosmosDBDNSSuf***REMOVED***x"`
	TokenAudience                string             `json:"tokenAudience"`
	ResourceIdenti***REMOVED***ers          ResourceIdenti***REMOVED***er `json:"resourceIdenti***REMOVED***ers"`
}

var (
	// PublicCloud is the default public Azure cloud environment
	PublicCloud = Environment{
		Name:                         "AzurePublicCloud",
		ManagementPortalURL:          "https://manage.windowsazure.com/",
		PublishSettingsURL:           "https://manage.windowsazure.com/publishsettings/index",
		ServiceManagementEndpoint:    "https://management.core.windows.net/",
		ResourceManagerEndpoint:      "https://management.azure.com/",
		ActiveDirectoryEndpoint:      "https://login.microsoftonline.com/",
		GalleryEndpoint:              "https://gallery.azure.com/",
		KeyVaultEndpoint:             "https://vault.azure.net/",
		GraphEndpoint:                "https://graph.windows.net/",
		ServiceBusEndpoint:           "https://servicebus.windows.net/",
		BatchManagementEndpoint:      "https://batch.core.windows.net/",
		StorageEndpointSuf***REMOVED***x:        "core.windows.net",
		SQLDatabaseDNSSuf***REMOVED***x:         "database.windows.net",
		Traf***REMOVED***cManagerDNSSuf***REMOVED***x:      "traf***REMOVED***cmanager.net",
		KeyVaultDNSSuf***REMOVED***x:            "vault.azure.net",
		ServiceBusEndpointSuf***REMOVED***x:     "servicebus.windows.net",
		ServiceManagementVMDNSSuf***REMOVED***x: "cloudapp.net",
		ResourceManagerVMDNSSuf***REMOVED***x:   "cloudapp.azure.com",
		ContainerRegistryDNSSuf***REMOVED***x:   "azurecr.io",
		CosmosDBDNSSuf***REMOVED***x:            "documents.azure.com",
		TokenAudience:                "https://management.azure.com/",
		ResourceIdenti***REMOVED***ers: ResourceIdenti***REMOVED***er{
			Graph:               "https://graph.windows.net/",
			KeyVault:            "https://vault.azure.net",
			Datalake:            "https://datalake.azure.net/",
			Batch:               "https://batch.core.windows.net/",
			OperationalInsights: "https://api.loganalytics.io",
			Storage:             "https://storage.azure.com/",
		},
	}

	// USGovernmentCloud is the cloud environment for the US Government
	USGovernmentCloud = Environment{
		Name:                         "AzureUSGovernmentCloud",
		ManagementPortalURL:          "https://manage.windowsazure.us/",
		PublishSettingsURL:           "https://manage.windowsazure.us/publishsettings/index",
		ServiceManagementEndpoint:    "https://management.core.usgovcloudapi.net/",
		ResourceManagerEndpoint:      "https://management.usgovcloudapi.net/",
		ActiveDirectoryEndpoint:      "https://login.microsoftonline.us/",
		GalleryEndpoint:              "https://gallery.usgovcloudapi.net/",
		KeyVaultEndpoint:             "https://vault.usgovcloudapi.net/",
		GraphEndpoint:                "https://graph.windows.net/",
		ServiceBusEndpoint:           "https://servicebus.usgovcloudapi.net/",
		BatchManagementEndpoint:      "https://batch.core.usgovcloudapi.net/",
		StorageEndpointSuf***REMOVED***x:        "core.usgovcloudapi.net",
		SQLDatabaseDNSSuf***REMOVED***x:         "database.usgovcloudapi.net",
		Traf***REMOVED***cManagerDNSSuf***REMOVED***x:      "usgovtraf***REMOVED***cmanager.net",
		KeyVaultDNSSuf***REMOVED***x:            "vault.usgovcloudapi.net",
		ServiceBusEndpointSuf***REMOVED***x:     "servicebus.usgovcloudapi.net",
		ServiceManagementVMDNSSuf***REMOVED***x: "usgovcloudapp.net",
		ResourceManagerVMDNSSuf***REMOVED***x:   "cloudapp.windowsazure.us",
		ContainerRegistryDNSSuf***REMOVED***x:   "azurecr.us",
		CosmosDBDNSSuf***REMOVED***x:            "documents.azure.us",
		TokenAudience:                "https://management.usgovcloudapi.net/",
		ResourceIdenti***REMOVED***ers: ResourceIdenti***REMOVED***er{
			Graph:               "https://graph.windows.net/",
			KeyVault:            "https://vault.usgovcloudapi.net",
			Datalake:            NotAvailable,
			Batch:               "https://batch.core.usgovcloudapi.net/",
			OperationalInsights: "https://api.loganalytics.us",
			Storage:             "https://storage.azure.com/",
		},
	}

	// ChinaCloud is the cloud environment operated in China
	ChinaCloud = Environment{
		Name:                         "AzureChinaCloud",
		ManagementPortalURL:          "https://manage.chinacloudapi.com/",
		PublishSettingsURL:           "https://manage.chinacloudapi.com/publishsettings/index",
		ServiceManagementEndpoint:    "https://management.core.chinacloudapi.cn/",
		ResourceManagerEndpoint:      "https://management.chinacloudapi.cn/",
		ActiveDirectoryEndpoint:      "https://login.chinacloudapi.cn/",
		GalleryEndpoint:              "https://gallery.chinacloudapi.cn/",
		KeyVaultEndpoint:             "https://vault.azure.cn/",
		GraphEndpoint:                "https://graph.chinacloudapi.cn/",
		ServiceBusEndpoint:           "https://servicebus.chinacloudapi.cn/",
		BatchManagementEndpoint:      "https://batch.chinacloudapi.cn/",
		StorageEndpointSuf***REMOVED***x:        "core.chinacloudapi.cn",
		SQLDatabaseDNSSuf***REMOVED***x:         "database.chinacloudapi.cn",
		Traf***REMOVED***cManagerDNSSuf***REMOVED***x:      "traf***REMOVED***cmanager.cn",
		KeyVaultDNSSuf***REMOVED***x:            "vault.azure.cn",
		ServiceBusEndpointSuf***REMOVED***x:     "servicebus.chinacloudapi.cn",
		ServiceManagementVMDNSSuf***REMOVED***x: "chinacloudapp.cn",
		ResourceManagerVMDNSSuf***REMOVED***x:   "cloudapp.azure.cn",
		ContainerRegistryDNSSuf***REMOVED***x:   "azurecr.cn",
		CosmosDBDNSSuf***REMOVED***x:            "documents.azure.cn",
		TokenAudience:                "https://management.chinacloudapi.cn/",
		ResourceIdenti***REMOVED***ers: ResourceIdenti***REMOVED***er{
			Graph:               "https://graph.chinacloudapi.cn/",
			KeyVault:            "https://vault.azure.cn",
			Datalake:            NotAvailable,
			Batch:               "https://batch.chinacloudapi.cn/",
			OperationalInsights: NotAvailable,
			Storage:             "https://storage.azure.com/",
		},
	}

	// GermanCloud is the cloud environment operated in Germany
	GermanCloud = Environment{
		Name:                         "AzureGermanCloud",
		ManagementPortalURL:          "http://portal.microsoftazure.de/",
		PublishSettingsURL:           "https://manage.microsoftazure.de/publishsettings/index",
		ServiceManagementEndpoint:    "https://management.core.cloudapi.de/",
		ResourceManagerEndpoint:      "https://management.microsoftazure.de/",
		ActiveDirectoryEndpoint:      "https://login.microsoftonline.de/",
		GalleryEndpoint:              "https://gallery.cloudapi.de/",
		KeyVaultEndpoint:             "https://vault.microsoftazure.de/",
		GraphEndpoint:                "https://graph.cloudapi.de/",
		ServiceBusEndpoint:           "https://servicebus.cloudapi.de/",
		BatchManagementEndpoint:      "https://batch.cloudapi.de/",
		StorageEndpointSuf***REMOVED***x:        "core.cloudapi.de",
		SQLDatabaseDNSSuf***REMOVED***x:         "database.cloudapi.de",
		Traf***REMOVED***cManagerDNSSuf***REMOVED***x:      "azuretraf***REMOVED***cmanager.de",
		KeyVaultDNSSuf***REMOVED***x:            "vault.microsoftazure.de",
		ServiceBusEndpointSuf***REMOVED***x:     "servicebus.cloudapi.de",
		ServiceManagementVMDNSSuf***REMOVED***x: "azurecloudapp.de",
		ResourceManagerVMDNSSuf***REMOVED***x:   "cloudapp.microsoftazure.de",
		ContainerRegistryDNSSuf***REMOVED***x:   NotAvailable,
		CosmosDBDNSSuf***REMOVED***x:            "documents.microsoftazure.de",
		TokenAudience:                "https://management.microsoftazure.de/",
		ResourceIdenti***REMOVED***ers: ResourceIdenti***REMOVED***er{
			Graph:               "https://graph.cloudapi.de/",
			KeyVault:            "https://vault.microsoftazure.de",
			Datalake:            NotAvailable,
			Batch:               "https://batch.cloudapi.de/",
			OperationalInsights: NotAvailable,
			Storage:             "https://storage.azure.com/",
		},
	}
)

// EnvironmentFromName returns an Environment based on the common name speci***REMOVED***ed.
func EnvironmentFromName(name string) (Environment, error) {
	// IMPORTANT
	// As per @radhikagupta5:
	// This is technical debt, fundamentally here because Kubernetes is not currently accepting
	// contributions to the providers. Once that is an option, the provider should be updated to
	// directly call `EnvironmentFromFile`. Until then, we rely on dispatching Azure Stack environment creation
	// from this method based on the name that is provided to us.
	if strings.EqualFold(name, "AZURESTACKCLOUD") {
		return EnvironmentFromFile(os.Getenv(EnvironmentFilepathName))
	}

	name = strings.ToUpper(name)
	env, ok := environments[name]
	if !ok {
		return env, fmt.Errorf("autorest/azure: There is no cloud environment matching the name %q", name)
	}

	return env, nil
}

// EnvironmentFromFile loads an Environment from a con***REMOVED***guration ***REMOVED***le available on disk.
// This function is particularly useful in the Hybrid Cloud model, where one must de***REMOVED***ne their own
// endpoints.
func EnvironmentFromFile(location string) (unmarshaled Environment, err error) {
	***REMOVED***leContents, err := ioutil.ReadFile(location)
	if err != nil {
		return
	}

	err = json.Unmarshal(***REMOVED***leContents, &unmarshaled)

	return
}
