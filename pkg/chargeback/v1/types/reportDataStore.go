package types

import (
	"encoding/json"
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	ReportDataStorePlural = "reportdatastores"
)

// +k8s:deepcopy-gen=true
type ReportDataStoreList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty"`
	Items         []*ReportDataStore `json:"items"`
}

// +k8s:deepcopy-gen=true
type ReportDataStore struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec ReportDataStoreSpec `json:"spec"`
}

// +k8s:deepcopy-gen=true
type ReportDataStoreSpec struct {
	Storage ReportDataStoreStorage `json:"storage"`
	Queries []string               `json:"queries"`
}

// +k8s:deepcopy-gen=true
type ReportDataStoreStorage struct {
	Type   string `json:"type"`
	Format string `json:"format"`
	Bucket string `json:"bucket"`
	Prefix string `json:"prefix"`
}

func ListReportDataStores(r rest.Interface, namespace string) (ReportDataStoreList, error) {
	req := r.Get().
		Namespace(namespace).
		Resource(ReportDataStorePlural).
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		fmt.Printf("list op for namespace %q got err: %v\n", namespace, err)
		return ReportDataStoreList{}, err
	}
	var dataStores ReportDataStoreList
	return dataStores, json.Unmarshal(b, &dataStores)
}

func GetReportDataStore(r rest.Interface, namespace, name string) (ReportDataStore, error) {
	req := r.Get().
		Namespace(namespace).
		Resource(ReportDataStorePlural).
		Name(name).
		FieldsSelectorParam(nil)

	b, err := req.DoRaw()
	if err != nil {
		fmt.Printf("get op for namespace %q got err: %v\n", namespace, err)
		return ReportDataStore{}, err
	}
	var dataStore ReportDataStore
	return dataStore, json.Unmarshal(b, &dataStore)
}
