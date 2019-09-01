package adal

// Copyright 2017 Microsoft Corporation
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this ***REMOVED***le except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the speci***REMOVED***c language governing permissions and
//  limitations under the License.

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/***REMOVED***lepath"
)

// LoadToken restores a Token object from a ***REMOVED***le located at 'path'.
func LoadToken(path string) (*Token, error) {
	***REMOVED***le, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ***REMOVED***le (%s) while loading token: %v", path, err)
	}
	defer ***REMOVED***le.Close()

	var token Token

	dec := json.NewDecoder(***REMOVED***le)
	if err = dec.Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode contents of ***REMOVED***le (%s) into Token representation: %v", path, err)
	}
	return &token, nil
}

// SaveToken persists an oauth token at the given location on disk.
// It moves the new ***REMOVED***le into place so it can safely be used to replace an existing ***REMOVED***le
// that maybe accessed by multiple processes.
func SaveToken(path string, mode os.FileMode, token Token) error {
	dir := ***REMOVED***lepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory (%s) to store token in: %v", dir, err)
	}

	newFile, err := ioutil.TempFile(dir, "token")
	if err != nil {
		return fmt.Errorf("failed to create the temp ***REMOVED***le to write the token: %v", err)
	}
	tempPath := newFile.Name()

	if err := json.NewEncoder(newFile).Encode(token); err != nil {
		return fmt.Errorf("failed to encode token to ***REMOVED***le (%s) while saving token: %v", tempPath, err)
	}
	if err := newFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp ***REMOVED***le %s: %v", tempPath, err)
	}

	// Atomic replace to avoid multi-writer ***REMOVED***le corruptions
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to move temporary token to desired output location. src=%s dst=%s: %v", tempPath, path, err)
	}
	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("failed to chmod the token ***REMOVED***le %s: %v", path, err)
	}
	return nil
}
