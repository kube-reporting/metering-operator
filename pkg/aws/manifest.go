package aws

import (
	"path/filepath"
	"time"
)

// Manifest is a representation of the file AWS provides with metadata for current usage information.
type Manifest struct {
	AssemblyID             string        `json:"assemblyId"`
	Account                string        `json:"account"`
	Columns                []Column      `json:"columns"`
	Charset                string        `json:"charset"`
	Compression            string        `json:"compression"`
	ContentType            string        `json:"contentType"`
	ReportID               string        `json:"reportId"`
	ReportName             string        `json:"reportName"`
	BillingPeriod          BillingPeriod `json:"billingPeriod"`
	Bucket                 string        `json:"bucket"`
	ReportKeys             []string      `json:"reportKeys"`
	AdditionalArtifactKeys []string      `json:"additionalArtifactKeys"`
}

type BillingPeriod struct {
	Start Time `json:"start"`
	End   Time `json:"end"`
}

// Column is a description of a field from a AWS usage report manifest file.
type Column struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

// Paths returns the directories containing usage data. The result will be free of duplicates.
func (m Manifest) Paths() (paths []string) {
	pathMap := map[string]struct{}{}
	for _, key := range m.ReportKeys {
		dirPath := filepath.Dir(key)
		pathMap[dirPath] = struct{}{}
	}

	for path := range pathMap {
		paths = append(paths, path)
	}

	return
}

type Time struct {
	time.Time
}

const manifestTime = "20060102T000000.000Z"

func (t *Time) UnmarshalJSON(b []byte) error {
	// b contains quotes around the timestamp
	tt, err := time.Parse(manifestTime, string(b[1:len(b)-1]))
	if err == nil {
		*t = Time{tt}
	}
	return err
}

func (t *Time) String() string {
	return t.Format(manifestTime)
}
