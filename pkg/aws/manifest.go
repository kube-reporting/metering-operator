package aws

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
