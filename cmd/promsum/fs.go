package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/coreos-inc/kube-chargeback/pkg/promsum"
)

// setupStore configures a store using the given URL.
func setupStore(u url.URL) (promsum.Store, error) {
	if u.Scheme == "file" {
		return setupFS(u)
	} else if u.Scheme == "s3" {
		return setupS3(u)
	} else {
		return nil, fmt.Errorf("unknown scheme '%s' given, please provide either s3:// or file://", u.Scheme)
	}
}

// setupFS creates a WriteCloser for the local filesystem.
func setupFS(u url.URL) (promsum.Store, error) {
	file, err := promsum.NewFileStore(u.Path)
	if err != nil {
		return nil, fmt.Errorf("store for path '%v' could not be created: %v", u.Path, err)
	}
	return file, nil
}

// setupS3 configures writing to a temporary file and then pushing to S3 on file close
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
