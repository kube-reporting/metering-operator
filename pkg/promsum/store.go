package promsum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/***REMOVED***lepath"
	"strings"
	"time"

	"github.com/segmentio/ksuid"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

// Store manages the persistence of billing records.
type Store interface {
	// Write inserts the given records into storage.
	Write(record []BillingRecord) error

	// Read retrieves billing records within the given range. There are no ordering guarantees.
	Read(rng cb.Range, query, subject string) ([]BillingRecord, error)
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

// Write stores billing records as a ***REMOVED***le using the ***REMOVED***lename for ordering.
// Will overwrite existing entries matching range, subject, and query.
func (f FileStore) Write(records []BillingRecord) error {
	// create directory, if exists don't error
	if err := os.MkdirAll(f.directory, FileStorePerms); err != nil && !os.IsExist(err) {
		return fmt.Errorf("Could not create directory at '%s': %v", f.directory, err)
	}

	data, err := json.Marshal(records)
	if err != nil {
		return fmt.Errorf("could not record record: %v", err)
	}

	uuid, err := ksuid.NewRandom()
	if err != nil {
		return fmt.Errorf("failed to generate ***REMOVED***le uuid: %s", err)
	}

	min, max := extrema(records)
	rng := cb.Range{min, max}
	name := Name(rng, uuid.String())

	recordPath := ***REMOVED***lepath.Join(f.directory, name)
	if err = ioutil.WriteFile(recordPath, data, FileStorePerms); err != nil {
		return fmt.Errorf("failed to write billing record to '%s': %v", recordPath, err)
	}
	return nil
}

// Read retrieves all billing records for the given range, query, and subject. There are no ordering guarantees.
func (f FileStore) Read(rng cb.Range, query, subject string) (records []BillingRecord, err error) {
	err = ***REMOVED***lepath.Walk(f.directory, func(path string, ***REMOVED***le os.FileInfo, walkErr error) error {
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

		***REMOVED***leRecords, err := decodeRelevantRecords(data, rng)
		if err != nil {
			return fmt.Errorf("failed to read ***REMOVED***le '%s': %v", path, err)
		}
		records = append(records, ***REMOVED***leRecords...)
		return nil
	})

	if err != nil {
		err = fmt.Errorf("failed to read for range %v to %v: %v", rng.Start, rng.End, err)
	}
	return
}

// Name returns the name of the ***REMOVED***le for a given range.
func Name(rng cb.Range, suf***REMOVED***x string) string {
	return fmt.Sprintf("%d-%d-%x.json", rng.Start.Unix(), rng.End.Unix(), suf***REMOVED***x)
}

// PathWithinRange determines if the given ***REMOVED***lename represents a range within the one given.
func PathWithinRange(path string, rng cb.Range) (bool, error) {
	name := ***REMOVED***lepath.Base(path)
	recordRng, err := rngFromName(name)
	if err != nil {
		return false, fmt.Errorf("could not determine record range from ***REMOVED***lename for '%s': %v", path, err)
	}

	// skip if record is not in desired range
	if recordRng.Start.After(rng.End) || recordRng.End.Before(rng.Start) {
		return false, nil
	}
	return true, nil
}

// hash implements a simple hashing function for queries.
func hash(in string) (out uint64) {
	p, m := uint64(74207281), uint64(859433)
	for _, char := range in {
		out = m*out + uint64(char)
	}
	return uint64(out % p)
}

func rngFromName(name string) (cb.Range, error) {
	name = strings.TrimRight(name, ".json")
	s := strings.SplitN(name, "-", 3)
	if len(s) != 3 {
		return cb.Range{}, fmt.Errorf("'%s' does not match the format 'StartRange-EndRange.json'", name)
	}
	startStr, endStr := s[0], s[1]
	return cb.ParseUnixRange(startStr, endStr)
}

// extrema returns the min and max of a range
func extrema(records []BillingRecord) (min, max time.Time) {
	if len(records) < 1 {
		return
	}

	min = records[0].Start
	max = records[0].End
	for _, r := range records {
		if r.Start.Before(min) {
			min = r.Start
		}

		if r.End.After(max) {
			max = r.End
		}
	}
	return
}

func decodeRelevantRecords(data []byte, rng cb.Range) ([]BillingRecord, error) {
	var allRecords []BillingRecord
	if err := json.Unmarshal(data, &allRecords); err != nil {
		return nil, fmt.Errorf("could not decode record data: %v", err)
	}

	records := allRecords[:0]
	for _, r := range allRecords {
		if rng.Within(r.Start) || rng.Within(r.End) {
			records = append(records, r)
		}
	}
	return records, nil
}
