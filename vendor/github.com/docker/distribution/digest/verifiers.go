package digest

import (
	"hash"
	"io"
)

// Veri***REMOVED***er presents a general veri***REMOVED***cation interface to be used with message
// digests and other byte stream veri***REMOVED***cations. Users instantiate a Veri***REMOVED***er
// from one of the various methods, write the data under test to it then check
// the result with the Veri***REMOVED***ed method.
type Veri***REMOVED***er interface {
	io.Writer

	// Veri***REMOVED***ed will return true if the content written to Veri***REMOVED***er matches
	// the digest.
	Veri***REMOVED***ed() bool
}

// NewDigestVeri***REMOVED***er returns a veri***REMOVED***er that compares the written bytes
// against a passed in digest.
func NewDigestVeri***REMOVED***er(d Digest) (Veri***REMOVED***er, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}

	return hashVeri***REMOVED***er{
		hash:   d.Algorithm().Hash(),
		digest: d,
	}, nil
}

type hashVeri***REMOVED***er struct {
	digest Digest
	hash   hash.Hash
}

func (hv hashVeri***REMOVED***er) Write(p []byte) (n int, err error) {
	return hv.hash.Write(p)
}

func (hv hashVeri***REMOVED***er) Veri***REMOVED***ed() bool {
	return hv.digest == NewDigest(hv.digest.Algorithm(), hv.hash)
}
