// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this ***REMOVED***le except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the speci***REMOVED***c language governing permissions and
// limitations under the License.

package compiler

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/***REMOVED***lepath"
	"strings"
)

var ***REMOVED***leCache map[string][]byte
var infoCache map[string]interface{}
var count int64

var verboseReader = false

func initializeFileCache() {
	if ***REMOVED***leCache == nil {
		***REMOVED***leCache = make(map[string][]byte, 0)
	}
}

func initializeInfoCache() {
	if infoCache == nil {
		infoCache = make(map[string]interface{}, 0)
	}
}

// FetchFile gets a speci***REMOVED***ed ***REMOVED***le from the local ***REMOVED***lesystem or a remote location.
func FetchFile(***REMOVED***leurl string) ([]byte, error) {
	initializeFileCache()
	bytes, ok := ***REMOVED***leCache[***REMOVED***leurl]
	if ok {
		if verboseReader {
			log.Printf("Cache hit %s", ***REMOVED***leurl)
		}
		return bytes, nil
	}
	log.Printf("Fetching %s", ***REMOVED***leurl)
	response, err := http.Get(***REMOVED***leurl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	bytes, err = ioutil.ReadAll(response.Body)
	if err == nil {
		***REMOVED***leCache[***REMOVED***leurl] = bytes
	}
	return bytes, err
}

// ReadBytesForFile reads the bytes of a ***REMOVED***le.
func ReadBytesForFile(***REMOVED***lename string) ([]byte, error) {
	// is the ***REMOVED***lename a url?
	***REMOVED***leurl, _ := url.Parse(***REMOVED***lename)
	if ***REMOVED***leurl.Scheme != "" {
		// yes, fetch it
		bytes, err := FetchFile(***REMOVED***lename)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	}
	// no, it's a local ***REMOVED***lename
	bytes, err := ioutil.ReadFile(***REMOVED***lename)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// ReadInfoFromBytes unmarshals a ***REMOVED***le as a yaml.MapSlice.
func ReadInfoFromBytes(***REMOVED***lename string, bytes []byte) (interface{}, error) {
	initializeInfoCache()
	cachedInfo, ok := infoCache[***REMOVED***lename]
	if ok {
		if verboseReader {
			log.Printf("Cache hit info for ***REMOVED***le %s", ***REMOVED***lename)
		}
		return cachedInfo, nil
	}
	if verboseReader {
		log.Printf("Reading info for ***REMOVED***le %s", ***REMOVED***lename)
	}
	var info yaml.MapSlice
	err := yaml.Unmarshal(bytes, &info)
	if err != nil {
		return nil, err
	}
	infoCache[***REMOVED***lename] = info
	return info, nil
}

// ReadInfoForRef reads a ***REMOVED***le and return the fragment needed to resolve a $ref.
func ReadInfoForRef(base***REMOVED***le string, ref string) (interface{}, error) {
	initializeInfoCache()
	{
		info, ok := infoCache[ref]
		if ok {
			if verboseReader {
				log.Printf("Cache hit for ref %s#%s", base***REMOVED***le, ref)
			}
			return info, nil
		}
	}
	if verboseReader {
		log.Printf("Reading info for ref %s#%s", base***REMOVED***le, ref)
	}
	count = count + 1
	basedir, _ := ***REMOVED***lepath.Split(base***REMOVED***le)
	parts := strings.Split(ref, "#")
	var ***REMOVED***lename string
	if parts[0] != "" {
		***REMOVED***lename = basedir + parts[0]
	} ***REMOVED*** {
		***REMOVED***lename = base***REMOVED***le
	}
	bytes, err := ReadBytesForFile(***REMOVED***lename)
	if err != nil {
		return nil, err
	}
	info, err := ReadInfoFromBytes(***REMOVED***lename, bytes)
	if err != nil {
		log.Printf("File error: %v\n", err)
	} ***REMOVED*** {
		if len(parts) > 1 {
			path := strings.Split(parts[1], "/")
			for i, key := range path {
				if i > 0 {
					m, ok := info.(yaml.MapSlice)
					if ok {
						found := false
						for _, section := range m {
							if section.Key == key {
								info = section.Value
								found = true
							}
						}
						if !found {
							infoCache[ref] = nil
							return nil, NewError(nil, fmt.Sprintf("could not resolve %s", ref))
						}
					}
				}
			}
		}
	}
	infoCache[ref] = info
	return info, nil
}
