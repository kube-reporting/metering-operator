package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/coreos-inc/kube-chargeback/pkg/promsum"
)

// setupStore con***REMOVED***gures a store using the given URL.
func setupStore(u url.URL) (promsum.Store, error) {
	if u.Scheme == "***REMOVED***le" {
		return setupFS(u)
	} ***REMOVED*** if u.Scheme == "s3" {
		return setupS3(u)
	} ***REMOVED*** {
		return nil, fmt.Errorf("unknown scheme '%s' given, please provide either s3:// or ***REMOVED***le://", u.Scheme)
	}
}

// setupFS creates a WriteCloser for the local ***REMOVED***lesystem.
func setupFS(u url.URL) (promsum.Store, error) {
	***REMOVED***le, err := promsum.NewFileStore(u.Path)
	if err != nil {
		return nil, fmt.Errorf("store for path '%v' could not be created: %v", u.Path, err)
	}
	return ***REMOVED***le, nil
}

// setupS3 con***REMOVED***gures writing to a temporary ***REMOVED***le and then pushing to S3 on ***REMOVED***le close
func setupS3(u url.URL) (promsum.Store, error) {
	// determine bucket name and object key
	splitPath := strings.SplitN(u.Path, "/", 2)
	var key string
	bucket := splitPath[0]
	if len(splitPath) > 1 {
		key = splitPath[1]
	}

	log.Printf("Uploading to the S3 bucket '%s' with key '%s'", bucket, key)
	return promsum.NewS3Store(bucket, key), nil
}
