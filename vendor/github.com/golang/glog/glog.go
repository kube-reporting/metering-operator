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

// Package glog implements logging analogous to the Google-internal C++ INFO/ERROR/V setup.
// It provides functions Info, Warning, Error, Fatal, plus formatting variants such as
// Infof. It also provides V-style logging controlled by the -v and -vmodule=***REMOVED***le=2 flags.
//
// Basic examples:
//
//	glog.Info("Prepare to repel boarders")
//
//	glog.Fatalf("Initialization failed: %s", err)
//
// See the documentation for the V function for an explanation of these examples:
//
//	if glog.V(2) {
//		glog.Info("Starting transaction...")
//	}
//
//	glog.V(2).Infoln("Processed", nItems, "elements")
//
// Log output is buffered and written periodically using Flush. Programs
// should call Flush before exiting to guarantee all log output is written.
//
// By default, all log statements write to ***REMOVED***les in a temporary directory.
// This package provides several flags that modify this behavior.
// As a result, flag.Parse must be called before any logging is done.
//
//	-logtostderr=false
//		Logs are written to standard error instead of to ***REMOVED***les.
//	-alsologtostderr=false
//		Logs are written to standard error as well as to ***REMOVED***les.
//	-stderrthreshold=ERROR
//		Log events at or above this severity are logged to standard
//		error as well as to ***REMOVED***les.
//	-log_dir=""
//		Log ***REMOVED***les will be written to this directory instead of the
//		default temporary directory.
//
//	Other flags provide aids to debugging.
//
//	-log_backtrace_at=""
//		When set to a ***REMOVED***le and line number holding a logging statement,
//		such as
//			-log_backtrace_at=gopherflakes.go:234
//		a stack trace will be written to the Info log whenever execution
//		hits that statement. (Unlike with -vmodule, the ".go" must be
//		present.)
//	-v=0
//		Enable V-leveled logging at the speci***REMOVED***ed level.
//	-vmodule=""
//		The syntax of the argument is a comma-separated list of pattern=N,
//		where pattern is a literal ***REMOVED***le name (minus the ".go" suf***REMOVED***x) or
//		"glob" pattern and N is a V level. For instance,
//			-vmodule=gopher*=3
//		sets the V level to 3 in all Go ***REMOVED***les whose names begin "gopher".
//
package glog

import (
	"bu***REMOVED***o"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	stdLog "log"
	"os"
	"path/***REMOVED***lepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// severity identi***REMOVED***es the sort of log: info, warning etc. It also implements
// the flag.Value interface. The -stderrthreshold flag is of type severity and
// should be modi***REMOVED***ed only through the flag.Value interface. The values match
// the corresponding constants in C++.
type severity int32 // sync/atomic int32

// These constants identify the log levels in order of increasing severity.
// A message written to a high-severity log ***REMOVED***le is also written to each
// lower-severity log ***REMOVED***le.
const (
	infoLog severity = iota
	warningLog
	errorLog
	fatalLog
	numSeverity = 4
)

const severityChar = "IWEF"

var severityName = []string{
	infoLog:    "INFO",
	warningLog: "WARNING",
	errorLog:   "ERROR",
	fatalLog:   "FATAL",
}

// get returns the value of the severity.
func (s *severity) get() severity {
	return severity(atomic.LoadInt32((*int32)(s)))
}

// set sets the value of the severity.
func (s *severity) set(val severity) {
	atomic.StoreInt32((*int32)(s), int32(val))
}

// String is part of the flag.Value interface.
func (s *severity) String() string {
	return strconv.FormatInt(int64(*s), 10)
}

// Get is part of the flag.Value interface.
func (s *severity) Get() interface{} {
	return *s
}

// Set is part of the flag.Value interface.
func (s *severity) Set(value string) error {
	var threshold severity
	// Is it a known name?
	if v, ok := severityByName(value); ok {
		threshold = v
	} ***REMOVED*** {
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		threshold = severity(v)
	}
	logging.stderrThreshold.set(threshold)
	return nil
}

func severityByName(s string) (severity, bool) {
	s = strings.ToUpper(s)
	for i, name := range severityName {
		if name == s {
			return severity(i), true
		}
	}
	return 0, false
}

// OutputStats tracks the number of output lines and bytes written.
type OutputStats struct {
	lines int64
	bytes int64
}

// Lines returns the number of lines written.
func (s *OutputStats) Lines() int64 {
	return atomic.LoadInt64(&s.lines)
}

// Bytes returns the number of bytes written.
func (s *OutputStats) Bytes() int64 {
	return atomic.LoadInt64(&s.bytes)
}

// Stats tracks the number of lines of output and number of bytes
// per severity level. Values must be read with atomic.LoadInt64.
var Stats struct {
	Info, Warning, Error OutputStats
}

var severityStats = [numSeverity]*OutputStats{
	infoLog:    &Stats.Info,
	warningLog: &Stats.Warning,
	errorLog:   &Stats.Error,
}

// Level is exported because it appears in the arguments to V and is
// the type of the v flag, which can be set programmatically.
// It's a distinct type because we want to discriminate it from logType.
// Variables of type level are only changed under logging.mu.
// The -v flag is read only with atomic ops, so the state of the logging
// module is consistent.

// Level is treated as a sync/atomic int32.

// Level speci***REMOVED***es a level of verbosity for V logs. *Level implements
// flag.Value; the -v flag is of type Level and should be modi***REMOVED***ed
// only through the flag.Value interface.
type Level int32

// get returns the value of the Level.
func (l *Level) get() Level {
	return Level(atomic.LoadInt32((*int32)(l)))
}

// set sets the value of the Level.
func (l *Level) set(val Level) {
	atomic.StoreInt32((*int32)(l), int32(val))
}

// String is part of the flag.Value interface.
func (l *Level) String() string {
	return strconv.FormatInt(int64(*l), 10)
}

// Get is part of the flag.Value interface.
func (l *Level) Get() interface{} {
	return *l
}

// Set is part of the flag.Value interface.
func (l *Level) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	logging.mu.Lock()
	defer logging.mu.Unlock()
	logging.setVState(Level(v), logging.vmodule.***REMOVED***lter, false)
	return nil
}

// moduleSpec represents the setting of the -vmodule flag.
type moduleSpec struct {
	***REMOVED***lter []modulePat
}

// modulePat contains a ***REMOVED***lter for the -vmodule flag.
// It holds a verbosity level and a ***REMOVED***le pattern to match.
type modulePat struct {
	pattern string
	literal bool // The pattern is a literal string
	level   Level
}

// match reports whether the ***REMOVED***le matches the pattern. It uses a string
// comparison if the pattern contains no metacharacters.
func (m *modulePat) match(***REMOVED***le string) bool {
	if m.literal {
		return ***REMOVED***le == m.pattern
	}
	match, _ := ***REMOVED***lepath.Match(m.pattern, ***REMOVED***le)
	return match
}

func (m *moduleSpec) String() string {
	// Lock because the type is not atomic. TODO: clean this up.
	logging.mu.Lock()
	defer logging.mu.Unlock()
	var b bytes.Buffer
	for i, f := range m.***REMOVED***lter {
		if i > 0 {
			b.WriteRune(',')
		}
		fmt.Fprintf(&b, "%s=%d", f.pattern, f.level)
	}
	return b.String()
}

// Get is part of the (Go 1.2)  flag.Getter interface. It always returns nil for this flag type since the
// struct is not exported.
func (m *moduleSpec) Get() interface{} {
	return nil
}

var errVmoduleSyntax = errors.New("syntax error: expect comma-separated list of ***REMOVED***lename=N")

// Syntax: -vmodule=recordio=2,***REMOVED***le=1,gfs*=3
func (m *moduleSpec) Set(value string) error {
	var ***REMOVED***lter []modulePat
	for _, pat := range strings.Split(value, ",") {
		if len(pat) == 0 {
			// Empty strings such as from a trailing comma can be ignored.
			continue
		}
		patLev := strings.Split(pat, "=")
		if len(patLev) != 2 || len(patLev[0]) == 0 || len(patLev[1]) == 0 {
			return errVmoduleSyntax
		}
		pattern := patLev[0]
		v, err := strconv.Atoi(patLev[1])
		if err != nil {
			return errors.New("syntax error: expect comma-separated list of ***REMOVED***lename=N")
		}
		if v < 0 {
			return errors.New("negative value for vmodule level")
		}
		if v == 0 {
			continue // Ignore. It's harmless but no point in paying the overhead.
		}
		// TODO: check syntax of ***REMOVED***lter?
		***REMOVED***lter = append(***REMOVED***lter, modulePat{pattern, isLiteral(pattern), Level(v)})
	}
	logging.mu.Lock()
	defer logging.mu.Unlock()
	logging.setVState(logging.verbosity, ***REMOVED***lter, true)
	return nil
}

// isLiteral reports whether the pattern is a literal string, that is, has no metacharacters
// that require ***REMOVED***lepath.Match to be called to match the pattern.
func isLiteral(pattern string) bool {
	return !strings.ContainsAny(pattern, `\*?[]`)
}

// traceLocation represents the setting of the -log_backtrace_at flag.
type traceLocation struct {
	***REMOVED***le string
	line int
}

// isSet reports whether the trace location has been speci***REMOVED***ed.
// logging.mu is held.
func (t *traceLocation) isSet() bool {
	return t.line > 0
}

// match reports whether the speci***REMOVED***ed ***REMOVED***le and line matches the trace location.
// The argument ***REMOVED***le name is the full path, not the basename speci***REMOVED***ed in the flag.
// logging.mu is held.
func (t *traceLocation) match(***REMOVED***le string, line int) bool {
	if t.line != line {
		return false
	}
	if i := strings.LastIndex(***REMOVED***le, "/"); i >= 0 {
		***REMOVED***le = ***REMOVED***le[i+1:]
	}
	return t.***REMOVED***le == ***REMOVED***le
}

func (t *traceLocation) String() string {
	// Lock because the type is not atomic. TODO: clean this up.
	logging.mu.Lock()
	defer logging.mu.Unlock()
	return fmt.Sprintf("%s:%d", t.***REMOVED***le, t.line)
}

// Get is part of the (Go 1.2) flag.Getter interface. It always returns nil for this flag type since the
// struct is not exported
func (t *traceLocation) Get() interface{} {
	return nil
}

var errTraceSyntax = errors.New("syntax error: expect ***REMOVED***le.go:234")

// Syntax: -log_backtrace_at=gopherflakes.go:234
// Note that unlike vmodule the ***REMOVED***le extension is included here.
func (t *traceLocation) Set(value string) error {
	if value == "" {
		// Unset.
		t.line = 0
		t.***REMOVED***le = ""
	}
	***REMOVED***elds := strings.Split(value, ":")
	if len(***REMOVED***elds) != 2 {
		return errTraceSyntax
	}
	***REMOVED***le, line := ***REMOVED***elds[0], ***REMOVED***elds[1]
	if !strings.Contains(***REMOVED***le, ".") {
		return errTraceSyntax
	}
	v, err := strconv.Atoi(line)
	if err != nil {
		return errTraceSyntax
	}
	if v <= 0 {
		return errors.New("negative or zero value for level")
	}
	logging.mu.Lock()
	defer logging.mu.Unlock()
	t.line = v
	t.***REMOVED***le = ***REMOVED***le
	return nil
}

// flushSyncWriter is the interface satis***REMOVED***ed by logging destinations.
type flushSyncWriter interface {
	Flush() error
	Sync() error
	io.Writer
}

func init() {
	flag.BoolVar(&logging.toStderr, "logtostderr", false, "log to standard error instead of ***REMOVED***les")
	flag.BoolVar(&logging.alsoToStderr, "alsologtostderr", false, "log to standard error as well as ***REMOVED***les")
	flag.Var(&logging.verbosity, "v", "log level for V logs")
	flag.Var(&logging.stderrThreshold, "stderrthreshold", "logs at or above this threshold go to stderr")
	flag.Var(&logging.vmodule, "vmodule", "comma-separated list of pattern=N settings for ***REMOVED***le-***REMOVED***ltered logging")
	flag.Var(&logging.traceLocation, "log_backtrace_at", "when logging hits line ***REMOVED***le:N, emit a stack trace")

	// Default stderrThreshold is ERROR.
	logging.stderrThreshold = errorLog

	logging.setVState(0, nil, false)
	go logging.flushDaemon()
}

// Flush flushes all pending log I/O.
func Flush() {
	logging.lockAndFlushAll()
}

// loggingT collects all the global state of the logging setup.
type loggingT struct {
	// Boolean flags. Not handled atomically because the flag.Value interface
	// does not let us avoid the =true, and that shorthand is necessary for
	// compatibility. TODO: does this matter enough to ***REMOVED***x? Seems unlikely.
	toStderr     bool // The -logtostderr flag.
	alsoToStderr bool // The -alsologtostderr flag.

	// Level flag. Handled atomically.
	stderrThreshold severity // The -stderrthreshold flag.

	// freeList is a list of byte buffers, maintained under freeListMu.
	freeList *buffer
	// freeListMu maintains the free list. It is separate from the main mutex
	// so buffers can be grabbed and printed to without holding the main lock,
	// for better parallelization.
	freeListMu sync.Mutex

	// mu protects the remaining elements of this structure and is
	// used to synchronize logging.
	mu sync.Mutex
	// ***REMOVED***le holds writer for each of the log types.
	***REMOVED***le [numSeverity]flushSyncWriter
	// pcs is used in V to avoid an allocation when computing the caller's PC.
	pcs [1]uintptr
	// vmap is a cache of the V Level for each V() call site, identi***REMOVED***ed by PC.
	// It is wiped whenever the vmodule flag changes state.
	vmap map[uintptr]Level
	// ***REMOVED***lterLength stores the length of the vmodule ***REMOVED***lter chain. If greater
	// than zero, it means vmodule is enabled. It may be read safely
	// using sync.LoadInt32, but is only modi***REMOVED***ed under mu.
	***REMOVED***lterLength int32
	// traceLocation is the state of the -log_backtrace_at flag.
	traceLocation traceLocation
	// These flags are modi***REMOVED***ed only under lock, although verbosity may be fetched
	// safely using atomic.LoadInt32.
	vmodule   moduleSpec // The state of the -vmodule flag.
	verbosity Level      // V logging level, the value of the -v flag/
}

// buffer holds a byte Buffer for reuse. The zero value is ready for use.
type buffer struct {
	bytes.Buffer
	tmp  [64]byte // temporary byte array for creating headers.
	next *buffer
}

var logging loggingT

// setVState sets a consistent state for V logging.
// l.mu is held.
func (l *loggingT) setVState(verbosity Level, ***REMOVED***lter []modulePat, setFilter bool) {
	// Turn verbosity off so V will not ***REMOVED***re while we are in transition.
	logging.verbosity.set(0)
	// Ditto for ***REMOVED***lter length.
	atomic.StoreInt32(&logging.***REMOVED***lterLength, 0)

	// Set the new ***REMOVED***lters and wipe the pc->Level map if the ***REMOVED***lter has changed.
	if setFilter {
		logging.vmodule.***REMOVED***lter = ***REMOVED***lter
		logging.vmap = make(map[uintptr]Level)
	}

	// Things are consistent now, so enable ***REMOVED***ltering and verbosity.
	// They are enabled in order opposite to that in V.
	atomic.StoreInt32(&logging.***REMOVED***lterLength, int32(len(***REMOVED***lter)))
	logging.verbosity.set(verbosity)
}

// getBuffer returns a new, ready-to-use buffer.
func (l *loggingT) getBuffer() *buffer {
	l.freeListMu.Lock()
	b := l.freeList
	if b != nil {
		l.freeList = b.next
	}
	l.freeListMu.Unlock()
	if b == nil {
		b = new(buffer)
	} ***REMOVED*** {
		b.next = nil
		b.Reset()
	}
	return b
}

// putBuffer returns a buffer to the free list.
func (l *loggingT) putBuffer(b *buffer) {
	if b.Len() >= 256 {
		// Let big buffers die a natural death.
		return
	}
	l.freeListMu.Lock()
	b.next = l.freeList
	l.freeList = b
	l.freeListMu.Unlock()
}

var timeNow = time.Now // Stubbed out for testing.

/*
header formats a log header as de***REMOVED***ned by the C++ implementation.
It returns a buffer containing the formatted header and the user's ***REMOVED***le and line number.
The depth speci***REMOVED***es how many stack frames above lives the source line to be identi***REMOVED***ed in the log message.

Log lines have this form:
	Lmmdd hh:mm:ss.uuuuuu threadid ***REMOVED***le:line] msg...
where the ***REMOVED***elds are de***REMOVED***ned as follows:
	L                A single character, representing the log level (eg 'I' for INFO)
	mm               The month (zero padded; ie May is '05')
	dd               The day (zero padded)
	hh:mm:ss.uuuuuu  Time in hours, minutes and fractional seconds
	threadid         The space-padded thread ID as returned by GetTID()
	***REMOVED***le             The ***REMOVED***le name
	line             The line number
	msg              The user-supplied message
*/
func (l *loggingT) header(s severity, depth int) (*buffer, string, int) {
	_, ***REMOVED***le, line, ok := runtime.Caller(3 + depth)
	if !ok {
		***REMOVED***le = "???"
		line = 1
	} ***REMOVED*** {
		slash := strings.LastIndex(***REMOVED***le, "/")
		if slash >= 0 {
			***REMOVED***le = ***REMOVED***le[slash+1:]
		}
	}
	return l.formatHeader(s, ***REMOVED***le, line), ***REMOVED***le, line
}

// formatHeader formats a log header using the provided ***REMOVED***le name and line number.
func (l *loggingT) formatHeader(s severity, ***REMOVED***le string, line int) *buffer {
	now := timeNow()
	if line < 0 {
		line = 0 // not a real line number, but acceptable to someDigits
	}
	if s > fatalLog {
		s = infoLog // for safety.
	}
	buf := l.getBuffer()

	// Avoid Fprintf, for speed. The format is so simple that we can do it quickly by hand.
	// It's worth about 3X. Fprintf is hard.
	_, month, day := now.Date()
	hour, minute, second := now.Clock()
	// Lmmdd hh:mm:ss.uuuuuu threadid ***REMOVED***le:line]
	buf.tmp[0] = severityChar[s]
	buf.twoDigits(1, int(month))
	buf.twoDigits(3, day)
	buf.tmp[5] = ' '
	buf.twoDigits(6, hour)
	buf.tmp[8] = ':'
	buf.twoDigits(9, minute)
	buf.tmp[11] = ':'
	buf.twoDigits(12, second)
	buf.tmp[14] = '.'
	buf.nDigits(6, 15, now.Nanosecond()/1000, '0')
	buf.tmp[21] = ' '
	buf.nDigits(7, 22, pid, ' ') // TODO: should be TID
	buf.tmp[29] = ' '
	buf.Write(buf.tmp[:30])
	buf.WriteString(***REMOVED***le)
	buf.tmp[0] = ':'
	n := buf.someDigits(1, line)
	buf.tmp[n+1] = ']'
	buf.tmp[n+2] = ' '
	buf.Write(buf.tmp[:n+3])
	return buf
}

// Some custom tiny helper functions to print the log header ef***REMOVED***ciently.

const digits = "0123456789"

// twoDigits formats a zero-pre***REMOVED***xed two-digit integer at buf.tmp[i].
func (buf *buffer) twoDigits(i, d int) {
	buf.tmp[i+1] = digits[d%10]
	d /= 10
	buf.tmp[i] = digits[d%10]
}

// nDigits formats an n-digit integer at buf.tmp[i],
// padding with pad on the left.
// It assumes d >= 0.
func (buf *buffer) nDigits(n, i, d int, pad byte) {
	j := n - 1
	for ; j >= 0 && d > 0; j-- {
		buf.tmp[i+j] = digits[d%10]
		d /= 10
	}
	for ; j >= 0; j-- {
		buf.tmp[i+j] = pad
	}
}

// someDigits formats a zero-pre***REMOVED***xed variable-width integer at buf.tmp[i].
func (buf *buffer) someDigits(i, d int) int {
	// Print into the top, then copy down. We know there's space for at least
	// a 10-digit number.
	j := len(buf.tmp)
	for {
		j--
		buf.tmp[j] = digits[d%10]
		d /= 10
		if d == 0 {
			break
		}
	}
	return copy(buf.tmp[i:], buf.tmp[j:])
}

func (l *loggingT) println(s severity, args ...interface{}) {
	buf, ***REMOVED***le, line := l.header(s, 0)
	fmt.Fprintln(buf, args...)
	l.output(s, buf, ***REMOVED***le, line, false)
}

func (l *loggingT) print(s severity, args ...interface{}) {
	l.printDepth(s, 1, args...)
}

func (l *loggingT) printDepth(s severity, depth int, args ...interface{}) {
	buf, ***REMOVED***le, line := l.header(s, depth)
	fmt.Fprint(buf, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, buf, ***REMOVED***le, line, false)
}

func (l *loggingT) printf(s severity, format string, args ...interface{}) {
	buf, ***REMOVED***le, line := l.header(s, 0)
	fmt.Fprintf(buf, format, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, buf, ***REMOVED***le, line, false)
}

// printWithFileLine behaves like print but uses the provided ***REMOVED***le and line number.  If
// alsoLogToStderr is true, the log message always appears on standard error; it
// will also appear in the log ***REMOVED***le unless --logtostderr is set.
func (l *loggingT) printWithFileLine(s severity, ***REMOVED***le string, line int, alsoToStderr bool, args ...interface{}) {
	buf := l.formatHeader(s, ***REMOVED***le, line)
	fmt.Fprint(buf, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	l.output(s, buf, ***REMOVED***le, line, alsoToStderr)
}

// output writes the data to the log ***REMOVED***les and releases the buffer.
func (l *loggingT) output(s severity, buf *buffer, ***REMOVED***le string, line int, alsoToStderr bool) {
	l.mu.Lock()
	if l.traceLocation.isSet() {
		if l.traceLocation.match(***REMOVED***le, line) {
			buf.Write(stacks(false))
		}
	}
	data := buf.Bytes()
	if l.toStderr {
		os.Stderr.Write(data)
	} ***REMOVED*** {
		if alsoToStderr || l.alsoToStderr || s >= l.stderrThreshold.get() {
			os.Stderr.Write(data)
		}
		if l.***REMOVED***le[s] == nil {
			if err := l.createFiles(s); err != nil {
				os.Stderr.Write(data) // Make sure the message appears somewhere.
				l.exit(err)
			}
		}
		switch s {
		case fatalLog:
			l.***REMOVED***le[fatalLog].Write(data)
			fallthrough
		case errorLog:
			l.***REMOVED***le[errorLog].Write(data)
			fallthrough
		case warningLog:
			l.***REMOVED***le[warningLog].Write(data)
			fallthrough
		case infoLog:
			l.***REMOVED***le[infoLog].Write(data)
		}
	}
	if s == fatalLog {
		// If we got here via Exit rather than Fatal, print no stacks.
		if atomic.LoadUint32(&fatalNoStacks) > 0 {
			l.mu.Unlock()
			timeoutFlush(10 * time.Second)
			os.Exit(1)
		}
		// Dump all goroutine stacks before exiting.
		// First, make sure we see the trace for the current goroutine on standard error.
		// If -logtostderr has been speci***REMOVED***ed, the loop below will do that anyway
		// as the ***REMOVED***rst stack in the full dump.
		if !l.toStderr {
			os.Stderr.Write(stacks(false))
		}
		// Write the stack trace for all goroutines to the ***REMOVED***les.
		trace := stacks(true)
		logExitFunc = func(error) {} // If we get a write error, we'll still exit below.
		for log := fatalLog; log >= infoLog; log-- {
			if f := l.***REMOVED***le[log]; f != nil { // Can be nil if -logtostderr is set.
				f.Write(trace)
			}
		}
		l.mu.Unlock()
		timeoutFlush(10 * time.Second)
		os.Exit(255) // C++ uses -1, which is silly because it's anded with 255 anyway.
	}
	l.putBuffer(buf)
	l.mu.Unlock()
	if stats := severityStats[s]; stats != nil {
		atomic.AddInt64(&stats.lines, 1)
		atomic.AddInt64(&stats.bytes, int64(len(data)))
	}
}

// timeoutFlush calls Flush and returns when it completes or after timeout
// elapses, whichever happens ***REMOVED***rst.  This is needed because the hooks invoked
// by Flush may deadlock when glog.Fatal is called from a hook that holds
// a lock.
func timeoutFlush(timeout time.Duration) {
	done := make(chan bool, 1)
	go func() {
		Flush() // calls logging.lockAndFlushAll()
		done <- true
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		fmt.Fprintln(os.Stderr, "glog: Flush took longer than", timeout)
	}
}

// stacks is a wrapper for runtime.Stack that attempts to recover the data for all goroutines.
func stacks(all bool) []byte {
	// We don't know how big the traces are, so grow a few times if they don't ***REMOVED***t. Start large, though.
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}

// logExitFunc provides a simple mechanism to override the default behavior
// of exiting on error. Used in testing and to guarantee we reach a required exit
// for fatal logs. Instead, exit could be a function rather than a method but that
// would make its use clumsier.
var logExitFunc func(error)

// exit is called if there is trouble creating or writing log ***REMOVED***les.
// It flushes the logs and exits the program; there's no point in hanging around.
// l.mu is held.
func (l *loggingT) exit(err error) {
	fmt.Fprintf(os.Stderr, "log: exiting because of error: %s\n", err)
	// If logExitFunc is set, we do that instead of exiting.
	if logExitFunc != nil {
		logExitFunc(err)
		return
	}
	l.flushAll()
	os.Exit(2)
}

// syncBuffer joins a bu***REMOVED***o.Writer to its underlying ***REMOVED***le, providing access to the
// ***REMOVED***le's Sync method and providing a wrapper for the Write method that provides log
// ***REMOVED***le rotation. There are conflicting methods, so the ***REMOVED***le cannot be embedded.
// l.mu is held for all its methods.
type syncBuffer struct {
	logger *loggingT
	*bu***REMOVED***o.Writer
	***REMOVED***le   *os.File
	sev    severity
	nbytes uint64 // The number of bytes written to this ***REMOVED***le
}

func (sb *syncBuffer) Sync() error {
	return sb.***REMOVED***le.Sync()
}

func (sb *syncBuffer) Write(p []byte) (n int, err error) {
	if sb.nbytes+uint64(len(p)) >= MaxSize {
		if err := sb.rotateFile(time.Now()); err != nil {
			sb.logger.exit(err)
		}
	}
	n, err = sb.Writer.Write(p)
	sb.nbytes += uint64(n)
	if err != nil {
		sb.logger.exit(err)
	}
	return
}

// rotateFile closes the syncBuffer's ***REMOVED***le and starts a new one.
func (sb *syncBuffer) rotateFile(now time.Time) error {
	if sb.***REMOVED***le != nil {
		sb.Flush()
		sb.***REMOVED***le.Close()
	}
	var err error
	sb.***REMOVED***le, _, err = create(severityName[sb.sev], now)
	sb.nbytes = 0
	if err != nil {
		return err
	}

	sb.Writer = bu***REMOVED***o.NewWriterSize(sb.***REMOVED***le, bufferSize)

	// Write header.
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Log ***REMOVED***le created at: %s\n", now.Format("2006/01/02 15:04:05"))
	fmt.Fprintf(&buf, "Running on machine: %s\n", host)
	fmt.Fprintf(&buf, "Binary: Built with %s %s for %s/%s\n", runtime.Compiler, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&buf, "Log line format: [IWEF]mmdd hh:mm:ss.uuuuuu threadid ***REMOVED***le:line] msg\n")
	n, err := sb.***REMOVED***le.Write(buf.Bytes())
	sb.nbytes += uint64(n)
	return err
}

// bufferSize sizes the buffer associated with each log ***REMOVED***le. It's large
// so that log records can accumulate without the logging thread blocking
// on disk I/O. The flushDaemon will block instead.
const bufferSize = 256 * 1024

// createFiles creates all the log ***REMOVED***les for severity from sev down to infoLog.
// l.mu is held.
func (l *loggingT) createFiles(sev severity) error {
	now := time.Now()
	// Files are created in decreasing severity order, so as soon as we ***REMOVED***nd one
	// has already been created, we can stop.
	for s := sev; s >= infoLog && l.***REMOVED***le[s] == nil; s-- {
		sb := &syncBuffer{
			logger: l,
			sev:    s,
		}
		if err := sb.rotateFile(now); err != nil {
			return err
		}
		l.***REMOVED***le[s] = sb
	}
	return nil
}

const flushInterval = 30 * time.Second

// flushDaemon periodically flushes the log ***REMOVED***le buffers.
func (l *loggingT) flushDaemon() {
	for _ = range time.NewTicker(flushInterval).C {
		l.lockAndFlushAll()
	}
}

// lockAndFlushAll is like flushAll but locks l.mu ***REMOVED***rst.
func (l *loggingT) lockAndFlushAll() {
	l.mu.Lock()
	l.flushAll()
	l.mu.Unlock()
}

// flushAll flushes all the logs and attempts to "sync" their data to disk.
// l.mu is held.
func (l *loggingT) flushAll() {
	// Flush from fatal down, in case there's trouble flushing.
	for s := fatalLog; s >= infoLog; s-- {
		***REMOVED***le := l.***REMOVED***le[s]
		if ***REMOVED***le != nil {
			***REMOVED***le.Flush() // ignore error
			***REMOVED***le.Sync()  // ignore error
		}
	}
}

// CopyStandardLogTo arranges for messages written to the Go "log" package's
// default logs to also appear in the Google logs for the named and lower
// severities.  Subsequent changes to the standard log's default output location
// or format may break this behavior.
//
// Valid names are "INFO", "WARNING", "ERROR", and "FATAL".  If the name is not
// recognized, CopyStandardLogTo panics.
func CopyStandardLogTo(name string) {
	sev, ok := severityByName(name)
	if !ok {
		panic(fmt.Sprintf("log.CopyStandardLogTo(%q): unrecognized severity name", name))
	}
	// Set a log format that captures the user's ***REMOVED***le and line:
	//   d.go:23: message
	stdLog.SetFlags(stdLog.Lshort***REMOVED***le)
	stdLog.SetOutput(logBridge(sev))
}

// logBridge provides the Write method that enables CopyStandardLogTo to connect
// Go's standard logs to the logs provided by this package.
type logBridge severity

// Write parses the standard logging line and passes its components to the
// logger for severity(lb).
func (lb logBridge) Write(b []byte) (n int, err error) {
	var (
		***REMOVED***le = "???"
		line = 1
		text string
	)
	// Split "d.go:23: message" into "d.go", "23", and "message".
	if parts := bytes.SplitN(b, []byte{':'}, 3); len(parts) != 3 || len(parts[0]) < 1 || len(parts[2]) < 1 {
		text = fmt.Sprintf("bad log format: %s", b)
	} ***REMOVED*** {
		***REMOVED***le = string(parts[0])
		text = string(parts[2][1:]) // skip leading space
		line, err = strconv.Atoi(string(parts[1]))
		if err != nil {
			text = fmt.Sprintf("bad line number: %s", b)
			line = 1
		}
	}
	// printWithFileLine with alsoToStderr=true, so standard log messages
	// always appear on standard error.
	logging.printWithFileLine(severity(lb), ***REMOVED***le, line, true, text)
	return len(b), nil
}

// setV computes and remembers the V level for a given PC
// when vmodule is enabled.
// File pattern matching takes the basename of the ***REMOVED***le, stripped
// of its .go suf***REMOVED***x, and uses ***REMOVED***lepath.Match, which is a little more
// general than the *? matching used in C++.
// l.mu is held.
func (l *loggingT) setV(pc uintptr) Level {
	fn := runtime.FuncForPC(pc)
	***REMOVED***le, _ := fn.FileLine(pc)
	// The ***REMOVED***le is something like /a/b/c/d.go. We want just the d.
	if strings.HasSuf***REMOVED***x(***REMOVED***le, ".go") {
		***REMOVED***le = ***REMOVED***le[:len(***REMOVED***le)-3]
	}
	if slash := strings.LastIndex(***REMOVED***le, "/"); slash >= 0 {
		***REMOVED***le = ***REMOVED***le[slash+1:]
	}
	for _, ***REMOVED***lter := range l.vmodule.***REMOVED***lter {
		if ***REMOVED***lter.match(***REMOVED***le) {
			l.vmap[pc] = ***REMOVED***lter.level
			return ***REMOVED***lter.level
		}
	}
	l.vmap[pc] = 0
	return 0
}

// Verbose is a boolean type that implements Infof (like Printf) etc.
// See the documentation of V for more information.
type Verbose bool

// V reports whether verbosity at the call site is at least the requested level.
// The returned value is a boolean of type Verbose, which implements Info, Infoln
// and Infof. These methods will write to the Info log if called.
// Thus, one may write either
//	if glog.V(2) { glog.Info("log this") }
// or
//	glog.V(2).Info("log this")
// The second form is shorter but the ***REMOVED***rst is cheaper if logging is off because it does
// not evaluate its arguments.
//
// Whether an individual call to V generates a log record depends on the setting of
// the -v and --vmodule flags; both are off by default. If the level in the call to
// V is at least the value of -v, or of -vmodule for the source ***REMOVED***le containing the
// call, the V call will log.
func V(level Level) Verbose {
	// This function tries hard to be cheap unless there's work to do.
	// The fast path is two atomic loads and compares.

	// Here is a cheap but safe test to see if V logging is enabled globally.
	if logging.verbosity.get() >= level {
		return Verbose(true)
	}

	// It's off globally but it vmodule may still be set.
	// Here is another cheap but safe test to see if vmodule is enabled.
	if atomic.LoadInt32(&logging.***REMOVED***lterLength) > 0 {
		// Now we need a proper lock to use the logging structure. The pcs ***REMOVED***eld
		// is shared so we must lock before accessing it. This is fairly expensive,
		// but if V logging is enabled we're slow anyway.
		logging.mu.Lock()
		defer logging.mu.Unlock()
		if runtime.Callers(2, logging.pcs[:]) == 0 {
			return Verbose(false)
		}
		v, ok := logging.vmap[logging.pcs[0]]
		if !ok {
			v = logging.setV(logging.pcs[0])
		}
		return Verbose(v >= level)
	}
	return Verbose(false)
}

// Info is equivalent to the global Info function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) Info(args ...interface{}) {
	if v {
		logging.print(infoLog, args...)
	}
}

// Infoln is equivalent to the global Infoln function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) Infoln(args ...interface{}) {
	if v {
		logging.println(infoLog, args...)
	}
}

// Infof is equivalent to the global Infof function, guarded by the value of v.
// See the documentation of V for usage.
func (v Verbose) Infof(format string, args ...interface{}) {
	if v {
		logging.printf(infoLog, format, args...)
	}
}

// Info logs to the INFO log.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Info(args ...interface{}) {
	logging.print(infoLog, args...)
}

// InfoDepth acts as Info but uses depth to determine which call frame to log.
// InfoDepth(0, "msg") is the same as Info("msg").
func InfoDepth(depth int, args ...interface{}) {
	logging.printDepth(infoLog, depth, args...)
}

// Infoln logs to the INFO log.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func Infoln(args ...interface{}) {
	logging.println(infoLog, args...)
}

// Infof logs to the INFO log.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Infof(format string, args ...interface{}) {
	logging.printf(infoLog, format, args...)
}

// Warning logs to the WARNING and INFO logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Warning(args ...interface{}) {
	logging.print(warningLog, args...)
}

// WarningDepth acts as Warning but uses depth to determine which call frame to log.
// WarningDepth(0, "msg") is the same as Warning("msg").
func WarningDepth(depth int, args ...interface{}) {
	logging.printDepth(warningLog, depth, args...)
}

// Warningln logs to the WARNING and INFO logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func Warningln(args ...interface{}) {
	logging.println(warningLog, args...)
}

// Warningf logs to the WARNING and INFO logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Warningf(format string, args ...interface{}) {
	logging.printf(warningLog, format, args...)
}

// Error logs to the ERROR, WARNING, and INFO logs.
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Error(args ...interface{}) {
	logging.print(errorLog, args...)
}

// ErrorDepth acts as Error but uses depth to determine which call frame to log.
// ErrorDepth(0, "msg") is the same as Error("msg").
func ErrorDepth(depth int, args ...interface{}) {
	logging.printDepth(errorLog, depth, args...)
}

// Errorln logs to the ERROR, WARNING, and INFO logs.
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func Errorln(args ...interface{}) {
	logging.println(errorLog, args...)
}

// Errorf logs to the ERROR, WARNING, and INFO logs.
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Errorf(format string, args ...interface{}) {
	logging.printf(errorLog, format, args...)
}

// Fatal logs to the FATAL, ERROR, WARNING, and INFO logs,
// including a stack trace of all running goroutines, then calls os.Exit(255).
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Fatal(args ...interface{}) {
	logging.print(fatalLog, args...)
}

// FatalDepth acts as Fatal but uses depth to determine which call frame to log.
// FatalDepth(0, "msg") is the same as Fatal("msg").
func FatalDepth(depth int, args ...interface{}) {
	logging.printDepth(fatalLog, depth, args...)
}

// Fatalln logs to the FATAL, ERROR, WARNING, and INFO logs,
// including a stack trace of all running goroutines, then calls os.Exit(255).
// Arguments are handled in the manner of fmt.Println; a newline is appended if missing.
func Fatalln(args ...interface{}) {
	logging.println(fatalLog, args...)
}

// Fatalf logs to the FATAL, ERROR, WARNING, and INFO logs,
// including a stack trace of all running goroutines, then calls os.Exit(255).
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Fatalf(format string, args ...interface{}) {
	logging.printf(fatalLog, format, args...)
}

// fatalNoStacks is non-zero if we are to exit without dumping goroutine stacks.
// It allows Exit and relatives to use the Fatal logs.
var fatalNoStacks uint32

// Exit logs to the FATAL, ERROR, WARNING, and INFO logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Print; a newline is appended if missing.
func Exit(args ...interface{}) {
	atomic.StoreUint32(&fatalNoStacks, 1)
	logging.print(fatalLog, args...)
}

// ExitDepth acts as Exit but uses depth to determine which call frame to log.
// ExitDepth(0, "msg") is the same as Exit("msg").
func ExitDepth(depth int, args ...interface{}) {
	atomic.StoreUint32(&fatalNoStacks, 1)
	logging.printDepth(fatalLog, depth, args...)
}

// Exitln logs to the FATAL, ERROR, WARNING, and INFO logs, then calls os.Exit(1).
func Exitln(args ...interface{}) {
	atomic.StoreUint32(&fatalNoStacks, 1)
	logging.println(fatalLog, args...)
}

// Exitf logs to the FATAL, ERROR, WARNING, and INFO logs, then calls os.Exit(1).
// Arguments are handled in the manner of fmt.Printf; a newline is appended if missing.
func Exitf(format string, args ...interface{}) {
	atomic.StoreUint32(&fatalNoStacks, 1)
	logging.printf(fatalLog, format, args...)
}
