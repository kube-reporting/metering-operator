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
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

// IPVSStats holds IPVS statistics, as exposed by the kernel in `/proc/net/ip_vs_stats`.
type IPVSStats struct {
	// Total count of connections.
	Connections uint64
	// Total incoming packages processed.
	IncomingPackets uint64
	// Total outgoing packages processed.
	OutgoingPackets uint64
	// Total incoming traf***REMOVED***c.
	IncomingBytes uint64
	// Total outgoing traf***REMOVED***c.
	OutgoingBytes uint64
}

// IPVSBackendStatus holds current metrics of one virtual / real address pair.
type IPVSBackendStatus struct {
	// The local (virtual) IP address.
	LocalAddress net.IP
	// The remote (real) IP address.
	RemoteAddress net.IP
	// The local (virtual) port.
	LocalPort uint16
	// The remote (real) port.
	RemotePort uint16
	// The local ***REMOVED***rewall mark
	LocalMark string
	// The transport protocol (TCP, UDP).
	Proto string
	// The current number of active connections for this virtual/real address pair.
	ActiveConn uint64
	// The current number of inactive connections for this virtual/real address pair.
	InactConn uint64
	// The current weight of this virtual/real address pair.
	Weight uint64
}

// NewIPVSStats reads the IPVS statistics.
func NewIPVSStats() (IPVSStats, error) {
	fs, err := NewFS(DefaultMountPoint)
	if err != nil {
		return IPVSStats{}, err
	}

	return fs.NewIPVSStats()
}

// NewIPVSStats reads the IPVS statistics from the speci***REMOVED***ed `proc` ***REMOVED***lesystem.
func (fs FS) NewIPVSStats() (IPVSStats, error) {
	***REMOVED***le, err := os.Open(fs.Path("net/ip_vs_stats"))
	if err != nil {
		return IPVSStats{}, err
	}
	defer ***REMOVED***le.Close()

	return parseIPVSStats(***REMOVED***le)
}

// parseIPVSStats performs the actual parsing of `ip_vs_stats`.
func parseIPVSStats(***REMOVED***le io.Reader) (IPVSStats, error) {
	var (
		statContent []byte
		statLines   []string
		statFields  []string
		stats       IPVSStats
	)

	statContent, err := ioutil.ReadAll(***REMOVED***le)
	if err != nil {
		return IPVSStats{}, err
	}

	statLines = strings.SplitN(string(statContent), "\n", 4)
	if len(statLines) != 4 {
		return IPVSStats{}, errors.New("ip_vs_stats corrupt: too short")
	}

	statFields = strings.Fields(statLines[2])
	if len(statFields) != 5 {
		return IPVSStats{}, errors.New("ip_vs_stats corrupt: unexpected number of ***REMOVED***elds")
	}

	stats.Connections, err = strconv.ParseUint(statFields[0], 16, 64)
	if err != nil {
		return IPVSStats{}, err
	}
	stats.IncomingPackets, err = strconv.ParseUint(statFields[1], 16, 64)
	if err != nil {
		return IPVSStats{}, err
	}
	stats.OutgoingPackets, err = strconv.ParseUint(statFields[2], 16, 64)
	if err != nil {
		return IPVSStats{}, err
	}
	stats.IncomingBytes, err = strconv.ParseUint(statFields[3], 16, 64)
	if err != nil {
		return IPVSStats{}, err
	}
	stats.OutgoingBytes, err = strconv.ParseUint(statFields[4], 16, 64)
	if err != nil {
		return IPVSStats{}, err
	}

	return stats, nil
}

// NewIPVSBackendStatus reads and returns the status of all (virtual,real) server pairs.
func NewIPVSBackendStatus() ([]IPVSBackendStatus, error) {
	fs, err := NewFS(DefaultMountPoint)
	if err != nil {
		return []IPVSBackendStatus{}, err
	}

	return fs.NewIPVSBackendStatus()
}

// NewIPVSBackendStatus reads and returns the status of all (virtual,real) server pairs from the speci***REMOVED***ed `proc` ***REMOVED***lesystem.
func (fs FS) NewIPVSBackendStatus() ([]IPVSBackendStatus, error) {
	***REMOVED***le, err := os.Open(fs.Path("net/ip_vs"))
	if err != nil {
		return nil, err
	}
	defer ***REMOVED***le.Close()

	return parseIPVSBackendStatus(***REMOVED***le)
}

func parseIPVSBackendStatus(***REMOVED***le io.Reader) ([]IPVSBackendStatus, error) {
	var (
		status       []IPVSBackendStatus
		scanner      = bu***REMOVED***o.NewScanner(***REMOVED***le)
		proto        string
		localMark    string
		localAddress net.IP
		localPort    uint16
		err          error
	)

	for scanner.Scan() {
		***REMOVED***elds := strings.Fields(scanner.Text())
		if len(***REMOVED***elds) == 0 {
			continue
		}
		switch {
		case ***REMOVED***elds[0] == "IP" || ***REMOVED***elds[0] == "Prot" || ***REMOVED***elds[1] == "RemoteAddress:Port":
			continue
		case ***REMOVED***elds[0] == "TCP" || ***REMOVED***elds[0] == "UDP":
			if len(***REMOVED***elds) < 2 {
				continue
			}
			proto = ***REMOVED***elds[0]
			localMark = ""
			localAddress, localPort, err = parseIPPort(***REMOVED***elds[1])
			if err != nil {
				return nil, err
			}
		case ***REMOVED***elds[0] == "FWM":
			if len(***REMOVED***elds) < 2 {
				continue
			}
			proto = ***REMOVED***elds[0]
			localMark = ***REMOVED***elds[1]
			localAddress = nil
			localPort = 0
		case ***REMOVED***elds[0] == "->":
			if len(***REMOVED***elds) < 6 {
				continue
			}
			remoteAddress, remotePort, err := parseIPPort(***REMOVED***elds[1])
			if err != nil {
				return nil, err
			}
			weight, err := strconv.ParseUint(***REMOVED***elds[3], 10, 64)
			if err != nil {
				return nil, err
			}
			activeConn, err := strconv.ParseUint(***REMOVED***elds[4], 10, 64)
			if err != nil {
				return nil, err
			}
			inactConn, err := strconv.ParseUint(***REMOVED***elds[5], 10, 64)
			if err != nil {
				return nil, err
			}
			status = append(status, IPVSBackendStatus{
				LocalAddress:  localAddress,
				LocalPort:     localPort,
				LocalMark:     localMark,
				RemoteAddress: remoteAddress,
				RemotePort:    remotePort,
				Proto:         proto,
				Weight:        weight,
				ActiveConn:    activeConn,
				InactConn:     inactConn,
			})
		}
	}
	return status, nil
}

func parseIPPort(s string) (net.IP, uint16, error) {
	var (
		ip  net.IP
		err error
	)

	switch len(s) {
	case 13:
		ip, err = hex.DecodeString(s[0:8])
		if err != nil {
			return nil, 0, err
		}
	case 46:
		ip = net.ParseIP(s[1:40])
		if ip == nil {
			return nil, 0, fmt.Errorf("invalid IPv6 address: %s", s[1:40])
		}
	default:
		return nil, 0, fmt.Errorf("unexpected IP:Port: %s", s)
	}

	portString := s[len(s)-4:]
	if len(portString) != 4 {
		return nil, 0, fmt.Errorf("unexpected port string format: %s", portString)
	}
	port, err := strconv.ParseUint(portString, 16, 16)
	if err != nil {
		return nil, 0, err
	}

	return ip, uint16(port), nil
}
