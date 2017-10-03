package main

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/coreos-inc/kube-chargeback/pkg/promsum"
)

// setupStore configures a store using the given URL.
func setupStore(u url.URL) (promsum.Store, error) {
	if u.Scheme == "file" {
		return setupFS(u.Path)
	} else if u.Scheme == "s3" {
		return setupS3(path.Join(u.Host, u.Path))
	} else {
		return nil, fmt.Errorf("unknown scheme '%s' given, please provide either s3:// or file://", u.Scheme)
	}
}

// setupFS creates a WriteCloser for the local filesystem.
func setupFS(path string) (promsum.Store, error) {
	file, err := promsum.NewFileStore(path)
	if err != nil {
		return nil, fmt.Errorf("store for path '%v' could not be created: %v", path, err)
	}
	return file, nil
}

// setupS3 configures writing to a temporary file and then pushing to S3 on file close
func setupS3(path string) (promsum.Store, error) {
	// determine bucket name and object key
	slash := strings.Index(path, "/")
	bucket, key := path[0:slash], path[slash:]
	log.Printf("Uploading to the S3 bucket '%s' with key '%s'", bucket, key)

	return promsum.NewS3Store(bucket, key), nil
}
