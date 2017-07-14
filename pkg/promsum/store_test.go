package promsum

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	cb "github.com/coreos-inc/kube-chargeback/pkg/chargeback"
)

func TestFileStoreReadWrite(t *testing.T) {
	name, err := ioutil.TempDir("", "promsum-store-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(name)

	store, err := NewFileStore(name)
	if err != nil {
		panic(err)
	}

	testStoreReadWrite(t, store)
}

func testStoreReadWrite(t *testing.T, s Store) {
	subject, query := "test-subject", "test-query"

	all := cb.Range{
		Start: time.Unix(1, 0),
		End:   time.Unix(4000, 0),
	}

	if read, err := s.Read(all, query, subject); err != nil {
		t.Error("Could not perform read: ", err)
	} ***REMOVED*** if len(read) != 0 {
		t.Error("No records should have been returned, found ", len(read))
	}

	records := []BillingRecord{
		{
			Start:   time.Unix(5, 0),
			End:     time.Unix(10, 0),
			Subject: subject,
			Query:   query,
		},
		{
			Start:   time.Unix(10, 0),
			End:     time.Unix(15, 0),
			Subject: subject,
			Query:   query,
		},
		{
			Start:   time.Unix(20, 0),
			End:     time.Unix(30, 0),
			Subject: subject,
			Query:   query,
		},
		{
			Start:   time.Unix(30, 0),
			End:     time.Unix(45, 0),
			Subject: subject,
			Query:   query,
		},
		{
			Start:   time.Unix(60, 0),
			End:     time.Unix(90, 0),
			Subject: subject,
			Query:   query,
		},
	}

	for _, r := range records {
		if err := s.Write(r); err != nil {
			t.Error("could not write record to store: ", err)
		}
	}

	if read, err := s.Read(all, query, subject); err != nil {
		t.Error("Could not perform read: ", err)
	} ***REMOVED*** if len(read) != len(records) {
		t.Error("Should have retrieved 5 records, found ", len(read))
	}

	some := cb.Range{
		Start: time.Unix(12, 0),
		End:   time.Unix(40, 0),
	}
	if read, err := s.Read(some, query, subject); err != nil {
		t.Error("Could not perform read: ", err)
	} ***REMOVED*** if len(read) != 3 {
		t.Error("Should have retrieved 3 records, found ", len(read))
	}
}
