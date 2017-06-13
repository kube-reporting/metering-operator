package promsum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/***REMOVED***lepath"
)

// Store manages the persistence of billing records.
type Store interface {
	// Write inserts the given record into storage.
	Write(record BillingRecord) error

	// Read retrieves billing records within the given range. There are no ordering guarantees.
	Read(rng Range, query, subject string) ([]BillingRecord, error)
}

var (
	// FileStorePerms are the permissions ***REMOVED***les and directories storing billing records are created with.
	FileStorePerms os.FileMode = 0644
)

// NewFileStore creates a store which writes records to the given path.
func NewFileStore(dir string) (FileStore, error) {
	dir = ***REMOVED***lepath.Clean(dir)
	if ***REMOVED***le, err := os.Stat(dir); err != nil {
		// don't throw error if just doesn't exist
		if !os.IsNotExist(err) {
			return FileStore{}, fmt.Errorf("could not access path '%s': %v", dir, err)
		}

		if err = os.MkdirAll(dir, FileStorePerms); err != nil {
			return FileStore{}, fmt.Errorf("could not create directory '%s': %v", dir, err)
		}
	} ***REMOVED*** if !***REMOVED***le.IsDir() {
		return FileStore{}, fmt.Errorf("the path '%s' is a ***REMOVED***le", dir)
	}

	return FileStore{
		directory: dir,
	}, nil
}

// FileStore is a simple implementation of Store which writes ***REMOVED***les to disk.
type FileStore struct {
	directory string
}

// FileStore must implement the Store interface
var _ Store = FileStore{}

// Write stores a billing record as a ***REMOVED***le using the ***REMOVED***lename for ordering.
// Will overwrite existing entries matching range, subject, and query.
func (f FileStore) Write(record BillingRecord) error {
	data, err := json.Marshal(&record)
	if err != nil {
		return fmt.Errorf("could not record record: %v", err)
	}

	recordPath := f.Path(record)
	if err = ioutil.WriteFile(recordPath, data, FileStorePerms); err != nil {
		return fmt.Errorf("failed to write billing record to '%s': %v", recordPath, err)
	}
	return nil
}

func (f FileStore) Read(rng Range, query, subject string) ([]BillingRecord, error) {
	return nil, nil
}

// Path returns the path where the given billing record is stored.
func (f FileStore) Path(record BillingRecord) string {
	return fmt.Sprintf("%s/%s/%x/%d-%d.json", f.directory, record.Subject, record.Query,
		record.Start.Unix(), record.End.Unix())
}
