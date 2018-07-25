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
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// ProcLimits represents the soft limits for each of the process's resource
// limits. For more information see getrlimit(2):
// http://man7.org/linux/man-pages/man2/getrlimit.2.html.
type ProcLimits struct {
	// CPU time limit in seconds.
	CPUTime int64
	// Maximum size of ***REMOVED***les that the process may create.
	FileSize int64
	// Maximum size of the process's data segment (initialized data,
	// uninitialized data, and heap).
	DataSize int64
	// Maximum size of the process stack in bytes.
	StackSize int64
	// Maximum size of a core ***REMOVED***le.
	CoreFileSize int64
	// Limit of the process's resident set in pages.
	ResidentSet int64
	// Maximum number of processes that can be created for the real user ID of
	// the calling process.
	Processes int64
	// Value one greater than the maximum ***REMOVED***le descriptor number that can be
	// opened by this process.
	OpenFiles int64
	// Maximum number of bytes of memory that may be locked into RAM.
	LockedMemory int64
	// Maximum size of the process's virtual memory address space in bytes.
	AddressSpace int64
	// Limit on the combined number of flock(2) locks and fcntl(2) leases that
	// this process may establish.
	FileLocks int64
	// Limit of signals that may be queued for the real user ID of the calling
	// process.
	PendingSignals int64
	// Limit on the number of bytes that can be allocated for POSIX message
	// queues for the real user ID of the calling process.
	MsqqueueSize int64
	// Limit of the nice priority set using setpriority(2) or nice(2).
	NicePriority int64
	// Limit of the real-time priority set using sched_setscheduler(2) or
	// sched_setparam(2).
	RealtimePriority int64
	// Limit (in microseconds) on the amount of CPU time that a process
	// scheduled under a real-time scheduling policy may consume without making
	// a blocking system call.
	RealtimeTimeout int64
}

const (
	limitsFields    = 3
	limitsUnlimited = "unlimited"
)

var (
	limitsDelimiter = regexp.MustCompile("  +")
)

// NewLimits returns the current soft limits of the process.
func (p Proc) NewLimits() (ProcLimits, error) {
	f, err := os.Open(p.path("limits"))
	if err != nil {
		return ProcLimits{}, err
	}
	defer f.Close()

	var (
		l = ProcLimits{}
		s = bu***REMOVED***o.NewScanner(f)
	)
	for s.Scan() {
		***REMOVED***elds := limitsDelimiter.Split(s.Text(), limitsFields)
		if len(***REMOVED***elds) != limitsFields {
			return ProcLimits{}, fmt.Errorf(
				"couldn't parse %s line %s", f.Name(), s.Text())
		}

		switch ***REMOVED***elds[0] {
		case "Max cpu time":
			l.CPUTime, err = parseInt(***REMOVED***elds[1])
		case "Max ***REMOVED***le size":
			l.FileSize, err = parseInt(***REMOVED***elds[1])
		case "Max data size":
			l.DataSize, err = parseInt(***REMOVED***elds[1])
		case "Max stack size":
			l.StackSize, err = parseInt(***REMOVED***elds[1])
		case "Max core ***REMOVED***le size":
			l.CoreFileSize, err = parseInt(***REMOVED***elds[1])
		case "Max resident set":
			l.ResidentSet, err = parseInt(***REMOVED***elds[1])
		case "Max processes":
			l.Processes, err = parseInt(***REMOVED***elds[1])
		case "Max open ***REMOVED***les":
			l.OpenFiles, err = parseInt(***REMOVED***elds[1])
		case "Max locked memory":
			l.LockedMemory, err = parseInt(***REMOVED***elds[1])
		case "Max address space":
			l.AddressSpace, err = parseInt(***REMOVED***elds[1])
		case "Max ***REMOVED***le locks":
			l.FileLocks, err = parseInt(***REMOVED***elds[1])
		case "Max pending signals":
			l.PendingSignals, err = parseInt(***REMOVED***elds[1])
		case "Max msgqueue size":
			l.MsqqueueSize, err = parseInt(***REMOVED***elds[1])
		case "Max nice priority":
			l.NicePriority, err = parseInt(***REMOVED***elds[1])
		case "Max realtime priority":
			l.RealtimePriority, err = parseInt(***REMOVED***elds[1])
		case "Max realtime timeout":
			l.RealtimeTimeout, err = parseInt(***REMOVED***elds[1])
		}
		if err != nil {
			return ProcLimits{}, err
		}
	}

	return l, s.Err()
}

func parseInt(s string) (int64, error) {
	if s == limitsUnlimited {
		return -1, nil
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("couldn't parse value %s: %s", s, err)
	}
	return i, nil
}
