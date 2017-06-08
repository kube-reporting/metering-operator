package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/colinmarc/hdfs"
)

const (
	localCreateFlags = os.O_RDWR | os.O_CREATE
	localPerms       = os.ModePerm
)

// setupOutput creates a WriteCloser for either local ***REMOVED***lesystem or HDFS depending on the protocol.
func setupOutput(in string) (io.WriteCloser, error) {
	u, err := url.Parse(in)
	if err != nil {
		return nil, fmt.Errorf("a valid path with scheme (s3://, ***REMOVED***le://, or hdfs://) must be given: %v", err)
	}

	if u.Scheme == "hdfs" {
		return setupHDFS(u.Host, u.Path)
	} ***REMOVED*** if u.Scheme == "***REMOVED***le" {
		return setupFS(u.Path)
	} ***REMOVED*** if u.Scheme == "s3" {
		return setupS3(u.Path)
	} ***REMOVED*** {
		return nil, fmt.Errorf("unknown scheme '%s' given, please provide either s3://, ***REMOVED***le://, or hdfs://", u.Scheme)
	}
}

// setupHDFS creates a WriteCloser for HDFS.
func setupHDFS(host, path string) (io.WriteCloser, error) {
	client, err := hdfs.New(host)
	if err != nil {
		return nil, fmt.Errorf("unable to create client for HDFS: %v", err)
	}

	***REMOVED***le, err := client.Create(path)
	if err != nil {
		return nil, fmt.Errorf("could not create ***REMOVED***le '%s': %v", path, err)
	}

	return ***REMOVED***le, err
}

// setupFS creates a WriteCloser for the local ***REMOVED***lesystem.
func setupFS(path string) (io.WriteCloser, error) {
	***REMOVED***le, err := os.OpenFile(path, localCreateFlags, localPerms)
	if err != nil {
		return nil, fmt.Errorf("the ***REMOVED***le '%v' could not be created: %v", path, err)
	}
	return ***REMOVED***le, nil
}

// setupS3 con***REMOVED***gures writing to a temporary ***REMOVED***le and then pushing to S3 on ***REMOVED***le close
func setupS3(path string) (io.WriteCloser, error) {
	// create temporary ***REMOVED***le as buffer
	f, err := ioutil.TempFile("", "pod-data")
	if err != nil {
		return nil, fmt.Errorf("could not write object data to temp***REMOVED***le: %v", err)
	}

	// remove initial slash from name
	// ie. /bucket-name/***REMOVED***le => bucket-name/***REMOVED***le
	path = path[1:]

	// determine bucket name and object key
	slash := strings.Index(path, "/")
	bucket, key := path[0:slash], path[slash:]
	log.Printf("Writing to temp***REMOVED***le then uploading to the S3 bucket '%s' with key '%s'", bucket, key)

	return s3Writer{
		bucket: bucket,
		key:    key,
		File:   f,
	}, nil

}

type s3Writer struct {
	bucket string
	key    string
	*os.File
}

func (writer s3Writer) Close() error {
	// close writing and reset seek
	defer writer.File.Close()
	writer.File.Seek(0, 0)

	s := session.Must(session.NewSession())
	svc := s3.New(s)

	_, err := svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(writer.bucket),
		Key:    aws.String(writer.key),
		Body:   writer.File,
	})

	if err != nil {
		log.Fatalf("failed to write ***REMOVED***le to s3 bucket '%s': %v. Output has been saved to '%s'.", writer.bucket, err, writer.File.Name())
	}

	// cleanup temporary ***REMOVED***le since successful
	os.Remove(writer.File.Name())
	return nil
}
