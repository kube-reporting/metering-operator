package promsum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/***REMOVED***lepath"
	"strings"
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
	FileStorePerms os.FileMode = 0700
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
	dir := Dir(f.directory, record.Query, record.Subject)

	// create directory, if exists don't error
	if err := os.MkdirAll(dir, FileStorePerms); err != nil && !os.IsExist(err) {
		return fmt.Errorf("Could not create directory at '%s': %v", dir, err)
	}

	data, err := json.Marshal(&record)
	if err != nil {
		return fmt.Errorf("could not record record: %v", err)
	}

	name := Name(record.Range())
	recordPath := ***REMOVED***lepath.Join(dir, name)
	if err = ioutil.WriteFile(recordPath, data, FileStorePerms); err != nil {
		return fmt.Errorf("failed to write billing record to '%s': %v", recordPath, err)
	}
	return nil
}

// Read retrieves all billing records for the given range, query, and subject. There are no ordering guarantees.
func (f FileStore) Read(rng Range, query, subject string) (records []BillingRecord, err error) {
	dir := Dir(f.directory, query, subject)

	err = ***REMOVED***lepath.Walk(dir, func(path string, ***REMOVED***le os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return err
		}

		// ignore directories
		if ***REMOVED***le.IsDir() {
			return nil
		}

		if ok, err := PathWithinRange(path, rng); err != nil {
			return fmt.Errorf("failed to determine if path '%s' is in range: %v", path, err)
		} ***REMOVED*** if !ok {
			return nil
		}

		// read record
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read ***REMOVED***le '%s': %v", path, err)
		}

		var record BillingRecord
		if err = json.Unmarshal(data, &record); err != nil {
			return fmt.Errorf("could not read record for '%s': %v", path, err)
		}
		records = append(records, record)
		return nil
	})

	if err != nil {
		err = fmt.Errorf("failed to read for range %v to %v: %v", rng.Start, rng.End, err)
	}
	return
}

// Name returns the name of the ***REMOVED***le for a given range.
func Name(rng Range) string {
	return fmt.Sprintf("%d-%d.json", rng.Start.Unix(), rng.End.Unix())
}

// Dir returns the path of the storage directory for the given query and subject
func Dir(parent, query, subject string) string {
	hashedQuery := fmt.Sprintf("%x", hash(query))
	return ***REMOVED***lepath.Join(parent, subject, hashedQuery)
}

// PathWithinRange determines if the given ***REMOVED***lename represents a range within the one given.
func PathWithinRange(path string, rng Range) (bool, error) {
	name := ***REMOVED***lepath.Base(path)
	recordRng, err := rngFromName(name)
	if err != nil {
		return false, fmt.Errorf("could not determine record range from ***REMOVED***lename for '%s': %v", path, err)
	}

	// skip if record is not in desired range
	if !rng.Within(recordRng.Start) && !rng.Within(recordRng.End) {
		return false, nil
	}
	return true, nil
}

// hash implements a simple hashing function for queries.
func hash(in string) (out uint64) {
	p, m := uint64(4423), uint64(2277)
	for _, char := range in {
		out = m*out + uint64(char)
	}
	return uint64(out % p)
}

func rngFromName(name string) (Range, error) {
	name = strings.TrimRight(name, ".json")
	s := strings.SplitN(name, "-", 2)
	if len(s) != 2 {
		return Range{}, fmt.Errorf("'%s' does not match the format 'StartRange-EndRange.json'", name)
	}
	startStr, endStr := s[0], s[1]
	return ParseUnixRange(startStr, endStr)
}
