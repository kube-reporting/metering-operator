package aws

import (
	"path/***REMOVED***lepath"
)

// Manifest is a representation of the ***REMOVED***le AWS provides with metadata for current usage information.
type Manifest struct {
	AssemblyID string `json:"assemblyId"`
	Account    string `json:"account"`
	Columns    []struct {
		Category string `json:"category"`
		Name     string `json:"name"`
	} `json:"columns"`
	Charset       string `json:"charset"`
	Compression   string `json:"compression"`
	ContentType   string `json:"contentType"`
	ReportID      string `json:"reportId"`
	ReportName    string `json:"reportName"`
	BillingPeriod struct {
		Begin string `json:"begin"`
		End   string `json:"end"`
	} `json:"billingPeriod"`
	Bucket                 string   `json:"bucket"`
	ReportKeys             []string `json:"reportKeys"`
	AdditionalArtifactKeys []string `json:"additionalArtifactKeys"`
}

// Paths returns the directories containing usage data. The result will be free of duplicates.
func (m Manifest) Paths() (paths []string) {
	pathMap := map[string]bool{}
	for _, key := range m.ReportKeys {
		dirPath := ***REMOVED***lepath.Dir(key)
		pathMap[dirPath] = true
	}

	for path := range pathMap {
		paths = append(paths, path)
	}
	return
}
