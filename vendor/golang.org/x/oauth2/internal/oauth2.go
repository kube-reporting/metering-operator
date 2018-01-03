// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE ***REMOVED***le.

// Package internal contains support packages for oauth2 package.
package internal

import (
	"bu***REMOVED***o"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ParseKey converts the binary contents of a private key ***REMOVED***le
// to an *rsa.PrivateKey. It detects whether the private key is in a
// PEM container or not. If so, it extracts the the private key
// from PEM container before conversion. It only supports PEM
// containers with no passphrase.
func ParseKey(key []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(key)
	if block != nil {
		key = block.Bytes
	}
	parsedKey, err := x509.ParsePKCS8PrivateKey(key)
	if err != nil {
		parsedKey, err = x509.ParsePKCS1PrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("private key should be a PEM or plain PKSC1 or PKCS8; parse error: %v", err)
		}
	}
	parsed, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is invalid")
	}
	return parsed, nil
}

func ParseINI(ini io.Reader) (map[string]map[string]string, error) {
	result := map[string]map[string]string{
		"": {}, // root section
	}
	scanner := bu***REMOVED***o.NewScanner(ini)
	currentSection := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPre***REMOVED***x(line, ";") {
			// comment.
			continue
		}
		if strings.HasPre***REMOVED***x(line, "[") && strings.HasSuf***REMOVED***x(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			result[currentSection] = map[string]string{}
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && parts[0] != "" {
			result[currentSection][strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning ini: %v", err)
	}
	return result, nil
}

func CondVal(v string) []string {
	if v == "" {
		return nil
	}
	return []string{v}
}
