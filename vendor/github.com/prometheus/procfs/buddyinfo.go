// Copyright 2017 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this ***REMOVED***le except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the speci***REMOVED***c language governing permissions and
// limitations under the License.

package procfs

import (
	"bu***REMOVED***o"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// A BuddyInfo is the details parsed from /proc/buddyinfo.
// The data is comprised of an array of free fragments of each size.
// The sizes are 2^n*PAGE_SIZE, where n is the array index.
type BuddyInfo struct {
	Node  string
	Zone  string
	Sizes []float64
}

// NewBuddyInfo reads the buddyinfo statistics.
func NewBuddyInfo() ([]BuddyInfo, error) {
	fs, err := NewFS(DefaultMountPoint)
	if err != nil {
		return nil, err
	}

	return fs.NewBuddyInfo()
}

// NewBuddyInfo reads the buddyinfo statistics from the speci***REMOVED***ed `proc` ***REMOVED***lesystem.
func (fs FS) NewBuddyInfo() ([]BuddyInfo, error) {
	***REMOVED***le, err := os.Open(fs.Path("buddyinfo"))
	if err != nil {
		return nil, err
	}
	defer ***REMOVED***le.Close()

	return parseBuddyInfo(***REMOVED***le)
}

func parseBuddyInfo(r io.Reader) ([]BuddyInfo, error) {
	var (
		buddyInfo   = []BuddyInfo{}
		scanner     = bu***REMOVED***o.NewScanner(r)
		bucketCount = -1
	)

	for scanner.Scan() {
		var err error
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid number of ***REMOVED***elds when parsing buddyinfo")
		}

		node := strings.TrimRight(parts[1], ",")
		zone := strings.TrimRight(parts[3], ",")
		arraySize := len(parts[4:])

		if bucketCount == -1 {
			bucketCount = arraySize
		} ***REMOVED*** {
			if bucketCount != arraySize {
				return nil, fmt.Errorf("mismatch in number of buddyinfo buckets, previous count %d, new count %d", bucketCount, arraySize)
			}
		}

		sizes := make([]float64, arraySize)
		for i := 0; i < arraySize; i++ {
			sizes[i], err = strconv.ParseFloat(parts[i+4], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid value in buddyinfo: %s", err)
			}
		}

		buddyInfo = append(buddyInfo, BuddyInfo{node, zone, sizes})
	}

	return buddyInfo, scanner.Err()
}
