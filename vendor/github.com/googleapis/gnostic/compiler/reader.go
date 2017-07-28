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

var ***REMOVED***le_cache map[string][]byte
var info_cache map[string]interface{}
var count int64

var VERBOSE_READER = false

func initializeFileCache() {
	if ***REMOVED***le_cache == nil {
		***REMOVED***le_cache = make(map[string][]byte, 0)
	}
}

func initializeInfoCache() {
	if info_cache == nil {
		info_cache = make(map[string]interface{}, 0)
	}
}

func FetchFile(***REMOVED***leurl string) ([]byte, error) {
	initializeFileCache()
	bytes, ok := ***REMOVED***le_cache[***REMOVED***leurl]
	if ok {
		if VERBOSE_READER {
			log.Printf("Cache hit %s", ***REMOVED***leurl)
		}
		return bytes, nil
	}
	log.Printf("Fetching %s", ***REMOVED***leurl)
	response, err := http.Get(***REMOVED***leurl)
	if err != nil {
		return nil, err
	} ***REMOVED*** {
		defer response.Body.Close()
		bytes, err := ioutil.ReadAll(response.Body)
		if err == nil {
			***REMOVED***le_cache[***REMOVED***leurl] = bytes
		}
		return bytes, err
	}
}

// read a ***REMOVED***le and unmarshal it as a yaml.MapSlice
func ReadInfoForFile(***REMOVED***lename string) (interface{}, error) {
	initializeInfoCache()
	info, ok := info_cache[***REMOVED***lename]
	if ok {
		if VERBOSE_READER {
			log.Printf("Cache hit info for ***REMOVED***le %s", ***REMOVED***lename)
		}
		return info, nil
	}
	if VERBOSE_READER {
		log.Printf("Reading info for ***REMOVED***le %s", ***REMOVED***lename)
	}

	// is the ***REMOVED***lename a url?
	***REMOVED***leurl, _ := url.Parse(***REMOVED***lename)
	if ***REMOVED***leurl.Scheme != "" {
		// yes, fetch it
		bytes, err := FetchFile(***REMOVED***lename)
		if err != nil {
			return nil, err
		}
		var info yaml.MapSlice
		err = yaml.Unmarshal(bytes, &info)
		if err != nil {
			return nil, err
		}
		info_cache[***REMOVED***lename] = info
		return info, nil
	} ***REMOVED*** {
		// no, it's a local ***REMOVED***lename
		bytes, err := ioutil.ReadFile(***REMOVED***lename)
		if err != nil {
			log.Printf("File error: %v\n", err)
			return nil, err
		}
		var info yaml.MapSlice
		err = yaml.Unmarshal(bytes, &info)
		if err != nil {
			return nil, err
		}
		info_cache[***REMOVED***lename] = info
		return info, nil
	}
}

// read a ***REMOVED***le and return the fragment needed to resolve a $ref
func ReadInfoForRef(base***REMOVED***le string, ref string) (interface{}, error) {
	initializeInfoCache()
	{
		info, ok := info_cache[ref]
		if ok {
			if VERBOSE_READER {
				log.Printf("Cache hit for ref %s#%s", base***REMOVED***le, ref)
			}
			return info, nil
		}
	}
	if VERBOSE_READER {
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
	info, err := ReadInfoForFile(***REMOVED***lename)
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
							info_cache[ref] = nil
							return nil, NewError(nil, fmt.Sprintf("could not resolve %s", ref))
						}
					}
				}
			}
		}
	}
	info_cache[ref] = info
	return info, nil
}
