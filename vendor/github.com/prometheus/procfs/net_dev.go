// Copyright 2018 The Prometheus Authors
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
	"errors"
	"os"
	"sort"
	"strconv"
	"strings"
)

// NetDevLine is single line parsed from /proc/net/dev or /proc/[pid]/net/dev.
type NetDevLine struct {
	Name         string `json:"name"`          // The name of the interface.
	RxBytes      uint64 `json:"rx_bytes"`      // Cumulative count of bytes received.
	RxPackets    uint64 `json:"rx_packets"`    // Cumulative count of packets received.
	RxErrors     uint64 `json:"rx_errors"`     // Cumulative count of receive errors encountered.
	RxDropped    uint64 `json:"rx_dropped"`    // Cumulative count of packets dropped while receiving.
	RxFIFO       uint64 `json:"rx_***REMOVED***fo"`       // Cumulative count of FIFO buffer errors.
	RxFrame      uint64 `json:"rx_frame"`      // Cumulative count of packet framing errors.
	RxCompressed uint64 `json:"rx_compressed"` // Cumulative count of compressed packets received by the device driver.
	RxMulticast  uint64 `json:"rx_multicast"`  // Cumulative count of multicast frames received by the device driver.
	TxBytes      uint64 `json:"tx_bytes"`      // Cumulative count of bytes transmitted.
	TxPackets    uint64 `json:"tx_packets"`    // Cumulative count of packets transmitted.
	TxErrors     uint64 `json:"tx_errors"`     // Cumulative count of transmit errors encountered.
	TxDropped    uint64 `json:"tx_dropped"`    // Cumulative count of packets dropped while transmitting.
	TxFIFO       uint64 `json:"tx_***REMOVED***fo"`       // Cumulative count of FIFO buffer errors.
	TxCollisions uint64 `json:"tx_collisions"` // Cumulative count of collisions detected on the interface.
	TxCarrier    uint64 `json:"tx_carrier"`    // Cumulative count of carrier losses detected by the device driver.
	TxCompressed uint64 `json:"tx_compressed"` // Cumulative count of compressed packets transmitted by the device driver.
}

// NetDev is parsed from /proc/net/dev or /proc/[pid]/net/dev. The map keys
// are interface names.
type NetDev map[string]NetDevLine

// NewNetDev returns kernel/system statistics read from /proc/net/dev.
func NewNetDev() (NetDev, error) {
	fs, err := NewFS(DefaultMountPoint)
	if err != nil {
		return nil, err
	}

	return fs.NewNetDev()
}

// NewNetDev returns kernel/system statistics read from /proc/net/dev.
func (fs FS) NewNetDev() (NetDev, error) {
	return newNetDev(fs.Path("net/dev"))
}

// NewNetDev returns kernel/system statistics read from /proc/[pid]/net/dev.
func (p Proc) NewNetDev() (NetDev, error) {
	return newNetDev(p.path("net/dev"))
}

// newNetDev creates a new NetDev from the contents of the given ***REMOVED***le.
func newNetDev(***REMOVED***le string) (NetDev, error) {
	f, err := os.Open(***REMOVED***le)
	if err != nil {
		return NetDev{}, err
	}
	defer f.Close()

	nd := NetDev{}
	s := bu***REMOVED***o.NewScanner(f)
	for n := 0; s.Scan(); n++ {
		// Skip the 2 header lines.
		if n < 2 {
			continue
		}

		line, err := nd.parseLine(s.Text())
		if err != nil {
			return nd, err
		}

		nd[line.Name] = *line
	}

	return nd, s.Err()
}

// parseLine parses a single line from the /proc/net/dev ***REMOVED***le. Header lines
// must be ***REMOVED***ltered prior to calling this method.
func (nd NetDev) parseLine(rawLine string) (*NetDevLine, error) {
	parts := strings.SplitN(rawLine, ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid net/dev line, missing colon")
	}
	***REMOVED***elds := strings.Fields(strings.TrimSpace(parts[1]))

	var err error
	line := &NetDevLine{}

	// Interface Name
	line.Name = strings.TrimSpace(parts[0])
	if line.Name == "" {
		return nil, errors.New("invalid net/dev line, empty interface name")
	}

	// RX
	line.RxBytes, err = strconv.ParseUint(***REMOVED***elds[0], 10, 64)
	if err != nil {
		return nil, err
	}
	line.RxPackets, err = strconv.ParseUint(***REMOVED***elds[1], 10, 64)
	if err != nil {
		return nil, err
	}
	line.RxErrors, err = strconv.ParseUint(***REMOVED***elds[2], 10, 64)
	if err != nil {
		return nil, err
	}
	line.RxDropped, err = strconv.ParseUint(***REMOVED***elds[3], 10, 64)
	if err != nil {
		return nil, err
	}
	line.RxFIFO, err = strconv.ParseUint(***REMOVED***elds[4], 10, 64)
	if err != nil {
		return nil, err
	}
	line.RxFrame, err = strconv.ParseUint(***REMOVED***elds[5], 10, 64)
	if err != nil {
		return nil, err
	}
	line.RxCompressed, err = strconv.ParseUint(***REMOVED***elds[6], 10, 64)
	if err != nil {
		return nil, err
	}
	line.RxMulticast, err = strconv.ParseUint(***REMOVED***elds[7], 10, 64)
	if err != nil {
		return nil, err
	}

	// TX
	line.TxBytes, err = strconv.ParseUint(***REMOVED***elds[8], 10, 64)
	if err != nil {
		return nil, err
	}
	line.TxPackets, err = strconv.ParseUint(***REMOVED***elds[9], 10, 64)
	if err != nil {
		return nil, err
	}
	line.TxErrors, err = strconv.ParseUint(***REMOVED***elds[10], 10, 64)
	if err != nil {
		return nil, err
	}
	line.TxDropped, err = strconv.ParseUint(***REMOVED***elds[11], 10, 64)
	if err != nil {
		return nil, err
	}
	line.TxFIFO, err = strconv.ParseUint(***REMOVED***elds[12], 10, 64)
	if err != nil {
		return nil, err
	}
	line.TxCollisions, err = strconv.ParseUint(***REMOVED***elds[13], 10, 64)
	if err != nil {
		return nil, err
	}
	line.TxCarrier, err = strconv.ParseUint(***REMOVED***elds[14], 10, 64)
	if err != nil {
		return nil, err
	}
	line.TxCompressed, err = strconv.ParseUint(***REMOVED***elds[15], 10, 64)
	if err != nil {
		return nil, err
	}

	return line, nil
}

// Total aggregates the values across interfaces and returns a new NetDevLine.
// The Name ***REMOVED***eld will be a sorted comma separated list of interface names.
func (nd NetDev) Total() NetDevLine {
	total := NetDevLine{}

	names := make([]string, 0, len(nd))
	for _, ifc := range nd {
		names = append(names, ifc.Name)
		total.RxBytes += ifc.RxBytes
		total.RxPackets += ifc.RxPackets
		total.RxPackets += ifc.RxPackets
		total.RxErrors += ifc.RxErrors
		total.RxDropped += ifc.RxDropped
		total.RxFIFO += ifc.RxFIFO
		total.RxFrame += ifc.RxFrame
		total.RxCompressed += ifc.RxCompressed
		total.RxMulticast += ifc.RxMulticast
		total.TxBytes += ifc.TxBytes
		total.TxPackets += ifc.TxPackets
		total.TxErrors += ifc.TxErrors
		total.TxDropped += ifc.TxDropped
		total.TxFIFO += ifc.TxFIFO
		total.TxCollisions += ifc.TxCollisions
		total.TxCarrier += ifc.TxCarrier
		total.TxCompressed += ifc.TxCompressed
	}
	sort.Strings(names)
	total.Name = strings.Join(names, ", ")

	return total
}
