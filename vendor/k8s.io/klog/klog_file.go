// Go support for leveled logs, analogous to https://code.google.com/p/google-glog/
//
// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this ***REMOVED***le except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the speci***REMOVED***c language governing permissions and
// limitations under the License.

// File I/O for logs.

package klog

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/***REMOVED***lepath"
	"strings"
	"sync"
	"time"
)

// MaxSize is the maximum size of a log ***REMOVED***le in bytes.
var MaxSize uint64 = 1024 * 1024 * 1800

// logDirs lists the candidate directories for new log ***REMOVED***les.
var logDirs []string

func createLogDirs() {
	if logging.logDir != "" {
		logDirs = append(logDirs, logging.logDir)
	}
	logDirs = append(logDirs, os.TempDir())
}

var (
	pid      = os.Getpid()
	program  = ***REMOVED***lepath.Base(os.Args[0])
	host     = "unknownhost"
	userName = "unknownuser"
)

func init() {
	h, err := os.Hostname()
	if err == nil {
		host = shortHostname(h)
	}

	current, err := user.Current()
	if err == nil {
		userName = current.Username
	}

	// Sanitize userName since it may contain ***REMOVED***lepath separators on Windows.
	userName = strings.Replace(userName, `\`, "_", -1)
}

// shortHostname returns its argument, truncating at the ***REMOVED***rst period.
// For instance, given "www.google.com" it returns "www".
func shortHostname(hostname string) string {
	if i := strings.Index(hostname, "."); i >= 0 {
		return hostname[:i]
	}
	return hostname
}

// logName returns a new log ***REMOVED***le name containing tag, with start time t, and
// the name for the symlink for tag.
func logName(tag string, t time.Time) (name, link string) {
	name = fmt.Sprintf("%s.%s.%s.log.%s.%04d%02d%02d-%02d%02d%02d.%d",
		program,
		host,
		userName,
		tag,
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		pid)
	return name, program + "." + tag
}

var onceLogDirs sync.Once

// create creates a new log ***REMOVED***le and returns the ***REMOVED***le and its ***REMOVED***lename, which
// contains tag ("INFO", "FATAL", etc.) and t.  If the ***REMOVED***le is created
// successfully, create also attempts to update the symlink for that tag, ignoring
// errors.
func create(tag string, t time.Time) (f *os.File, ***REMOVED***lename string, err error) {
	if logging.logFile != "" {
		f, err := os.Create(logging.logFile)
		if err == nil {
			return f, logging.logFile, nil
		}
		return nil, "", fmt.Errorf("log: unable to create log: %v", err)
	}
	onceLogDirs.Do(createLogDirs)
	if len(logDirs) == 0 {
		return nil, "", errors.New("log: no log dirs")
	}
	name, link := logName(tag, t)
	var lastErr error
	for _, dir := range logDirs {
		fname := ***REMOVED***lepath.Join(dir, name)
		f, err := os.Create(fname)
		if err == nil {
			symlink := ***REMOVED***lepath.Join(dir, link)
			os.Remove(symlink)        // ignore err
			os.Symlink(name, symlink) // ignore err
			return f, fname, nil
		}
		lastErr = err
	}
	return nil, "", fmt.Errorf("log: cannot create log: %v", lastErr)
}
