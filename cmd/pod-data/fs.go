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

// setupOutput creates a WriteCloser for either local filesystem or HDFS depending on the protocol.
func setupOutput(in string) (io.WriteCloser, error) {
	u, err := url.Parse(in)
	if err != nil {
		return nil, fmt.Errorf("a valid path with scheme (s3://, file://, or hdfs://) must be given: %v", err)
	}

	if u.Scheme == "hdfs" {
		return setupHDFS(u.Host, u.Path)
	} else if u.Scheme == "file" {
		return setupFS(u.Path)
	} else if u.Scheme == "s3" {
		return setupS3(u.Path)
	} else {
		return nil, fmt.Errorf("unknown scheme '%s' given, please provide either s3://, file://, or hdfs://", u.Scheme)
	}
}

// setupHDFS creates a WriteCloser for HDFS.
func setupHDFS(host, path string) (io.WriteCloser, error) {
	client, err := hdfs.New(host)
	if err != nil {
		return nil, fmt.Errorf("unable to create client for HDFS: %v", err)
	}

	file, err := client.Create(path)
	if err != nil {
		return nil, fmt.Errorf("could not create file '%s': %v", path, err)
	}

	return file, err
}

// setupFS creates a WriteCloser for the local filesystem.
func setupFS(path string) (io.WriteCloser, error) {
	file, err := os.OpenFile(path, localCreateFlags, localPerms)
	if err != nil {
		return nil, fmt.Errorf("the file '%v' could not be created: %v", path, err)
	}
	return file, nil
}

// setupS3 configures writing to a temporary file and then pushing to S3 on file close
func setupS3(path string) (io.WriteCloser, error) {
	// create temporary file as buffer
	f, err := ioutil.TempFile("", "pod-data")
	if err != nil {
		return nil, fmt.Errorf("could not write object data to tempfile: %v", err)
	}

	// remove initial slash from name
	// ie. /bucket-name/file => bucket-name/file
	path = path[1:]

	// determine bucket name and object key
	slash := strings.Index(path, "/")
	bucket, key := path[0:slash], path[slash:]
	log.Printf("Writing to tempfile then uploading to the S3 bucket '%s' with key '%s'", bucket, key)

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
		log.Fatalf("failed to write file to s3 bucket '%s': %v. Output has been saved to '%s'.", writer.bucket, err, writer.File.Name())
	}

	// cleanup temporary file since successful
	os.Remove(writer.File.Name())
	return nil
}
