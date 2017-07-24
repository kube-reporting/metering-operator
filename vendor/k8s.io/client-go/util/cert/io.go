/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this ***REMOVED***le except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the speci***REMOVED***c language governing permissions and
limitations under the License.
*/

package cert

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/***REMOVED***lepath"
)

// CanReadCertAndKey returns true if the certi***REMOVED***cate and key ***REMOVED***les already exists,
// otherwise returns false. If lost one of cert and key, returns error.
func CanReadCertAndKey(certPath, keyPath string) (bool, error) {
	certReadable := canReadFile(certPath)
	keyReadable := canReadFile(keyPath)

	if certReadable == false && keyReadable == false {
		return false, nil
	}

	if certReadable == false {
		return false, fmt.Errorf("error reading %s, certi***REMOVED***cate and key must be supplied as a pair", certPath)
	}

	if keyReadable == false {
		return false, fmt.Errorf("error reading %s, certi***REMOVED***cate and key must be supplied as a pair", keyPath)
	}

	return true, nil
}

// If the ***REMOVED***le represented by path exists and
// readable, returns true otherwise returns false.
func canReadFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}

	defer f.Close()

	return true
}

// WriteCert writes the pem-encoded certi***REMOVED***cate data to certPath.
// The certi***REMOVED***cate ***REMOVED***le will be created with ***REMOVED***le mode 0644.
// If the certi***REMOVED***cate ***REMOVED***le already exists, it will be overwritten.
// The parent directory of the certPath will be created as needed with ***REMOVED***le mode 0755.
func WriteCert(certPath string, data []byte) error {
	if err := os.MkdirAll(***REMOVED***lepath.Dir(certPath), os.FileMode(0755)); err != nil {
		return err
	}
	if err := ioutil.WriteFile(certPath, data, os.FileMode(0644)); err != nil {
		return err
	}
	return nil
}

// WriteKey writes the pem-encoded key data to keyPath.
// The key ***REMOVED***le will be created with ***REMOVED***le mode 0600.
// If the key ***REMOVED***le already exists, it will be overwritten.
// The parent directory of the keyPath will be created as needed with ***REMOVED***le mode 0755.
func WriteKey(keyPath string, data []byte) error {
	if err := os.MkdirAll(***REMOVED***lepath.Dir(keyPath), os.FileMode(0755)); err != nil {
		return err
	}
	if err := ioutil.WriteFile(keyPath, data, os.FileMode(0600)); err != nil {
		return err
	}
	return nil
}

// LoadOrGenerateKeyFile looks for a key in the ***REMOVED***le at the given path. If it
// can't ***REMOVED***nd one, it will generate a new key and store it there.
func LoadOrGenerateKeyFile(keyPath string) (data []byte, wasGenerated bool, err error) {
	loadedData, err := ioutil.ReadFile(keyPath)
	if err == nil {
		return loadedData, false, err
	}
	if !os.IsNotExist(err) {
		return nil, false, fmt.Errorf("error loading key from %s: %v", keyPath, err)
	}

	generatedData, err := MakeEllipticPrivateKeyPEM()
	if err != nil {
		return nil, false, fmt.Errorf("error generating key: %v", err)
	}
	if err := WriteKey(keyPath, generatedData); err != nil {
		return nil, false, fmt.Errorf("error writing key to %s: %v", keyPath, err)
	}
	return generatedData, true, nil
}

// NewPool returns an x509.CertPool containing the certi***REMOVED***cates in the given PEM-encoded ***REMOVED***le.
// Returns an error if the ***REMOVED***le could not be read, a certi***REMOVED***cate could not be parsed, or if the ***REMOVED***le does not contain any certi***REMOVED***cates
func NewPool(***REMOVED***lename string) (*x509.CertPool, error) {
	certs, err := CertsFromFile(***REMOVED***lename)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pool.AddCert(cert)
	}
	return pool, nil
}

// CertsFromFile returns the x509.Certi***REMOVED***cates contained in the given PEM-encoded ***REMOVED***le.
// Returns an error if the ***REMOVED***le could not be read, a certi***REMOVED***cate could not be parsed, or if the ***REMOVED***le does not contain any certi***REMOVED***cates
func CertsFromFile(***REMOVED***le string) ([]*x509.Certi***REMOVED***cate, error) {
	pemBlock, err := ioutil.ReadFile(***REMOVED***le)
	if err != nil {
		return nil, err
	}
	certs, err := ParseCertsPEM(pemBlock)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", ***REMOVED***le, err)
	}
	return certs, nil
}

// PrivateKeyFromFile returns the private key in rsa.PrivateKey or ecdsa.PrivateKey format from a given PEM-encoded ***REMOVED***le.
// Returns an error if the ***REMOVED***le could not be read or if the private key could not be parsed.
func PrivateKeyFromFile(***REMOVED***le string) (interface{}, error) {
	pemBlock, err := ioutil.ReadFile(***REMOVED***le)
	if err != nil {
		return nil, err
	}
	key, err := ParsePrivateKeyPEM(pemBlock)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", ***REMOVED***le, err)
	}
	return key, nil
}
