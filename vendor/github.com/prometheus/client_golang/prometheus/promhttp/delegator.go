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

package promhttp

import (
	"bu***REMOVED***o"
	"io"
	"net"
	"net/http"
)

const (
	closeNoti***REMOVED***er = 1 << iota
	flusher
	hijacker
	readerFrom
	pusher
)

type delegator interface {
	http.ResponseWriter

	Status() int
	Written() int64
}

type responseWriterDelegator struct {
	http.ResponseWriter

	handler, method    string
	status             int
	written            int64
	wroteHeader        bool
	observeWriteHeader func(int)
}

func (r *responseWriterDelegator) Status() int {
	return r.status
}

func (r *responseWriterDelegator) Written() int64 {
	return r.written
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
	if r.observeWriteHeader != nil {
		r.observeWriteHeader(code)
	}
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

type closeNoti***REMOVED***erDelegator struct{ *responseWriterDelegator }
type flusherDelegator struct{ *responseWriterDelegator }
type hijackerDelegator struct{ *responseWriterDelegator }
type readerFromDelegator struct{ *responseWriterDelegator }

func (d *closeNoti***REMOVED***erDelegator) CloseNotify() <-chan bool {
	return d.ResponseWriter.(http.CloseNoti***REMOVED***er).CloseNotify()
}
func (d *flusherDelegator) Flush() {
	d.ResponseWriter.(http.Flusher).Flush()
}
func (d *hijackerDelegator) Hijack() (net.Conn, *bu***REMOVED***o.ReadWriter, error) {
	return d.ResponseWriter.(http.Hijacker).Hijack()
}
func (d *readerFromDelegator) ReadFrom(re io.Reader) (int64, error) {
	if !d.wroteHeader {
		d.WriteHeader(http.StatusOK)
	}
	n, err := d.ResponseWriter.(io.ReaderFrom).ReadFrom(re)
	d.written += n
	return n, err
}

var pickDelegator = make([]func(*responseWriterDelegator) delegator, 32)

func init() {
	// TODO(beorn7): Code generation would help here.
	pickDelegator[0] = func(d *responseWriterDelegator) delegator { // 0
		return d
	}
	pickDelegator[closeNoti***REMOVED***er] = func(d *responseWriterDelegator) delegator { // 1
		return closeNoti***REMOVED***erDelegator{d}
	}
	pickDelegator[flusher] = func(d *responseWriterDelegator) delegator { // 2
		return flusherDelegator{d}
	}
	pickDelegator[flusher+closeNoti***REMOVED***er] = func(d *responseWriterDelegator) delegator { // 3
		return struct {
			*responseWriterDelegator
			http.Flusher
			http.CloseNoti***REMOVED***er
		}{d, &flusherDelegator{d}, &closeNoti***REMOVED***erDelegator{d}}
	}
	pickDelegator[hijacker] = func(d *responseWriterDelegator) delegator { // 4
		return hijackerDelegator{d}
	}
	pickDelegator[hijacker+closeNoti***REMOVED***er] = func(d *responseWriterDelegator) delegator { // 5
		return struct {
			*responseWriterDelegator
			http.Hijacker
			http.CloseNoti***REMOVED***er
		}{d, &hijackerDelegator{d}, &closeNoti***REMOVED***erDelegator{d}}
	}
	pickDelegator[hijacker+flusher] = func(d *responseWriterDelegator) delegator { // 6
		return struct {
			*responseWriterDelegator
			http.Hijacker
			http.Flusher
		}{d, &hijackerDelegator{d}, &flusherDelegator{d}}
	}
	pickDelegator[hijacker+flusher+closeNoti***REMOVED***er] = func(d *responseWriterDelegator) delegator { // 7
		return struct {
			*responseWriterDelegator
			http.Hijacker
			http.Flusher
			http.CloseNoti***REMOVED***er
		}{d, &hijackerDelegator{d}, &flusherDelegator{d}, &closeNoti***REMOVED***erDelegator{d}}
	}
	pickDelegator[readerFrom] = func(d *responseWriterDelegator) delegator { // 8
		return readerFromDelegator{d}
	}
	pickDelegator[readerFrom+closeNoti***REMOVED***er] = func(d *responseWriterDelegator) delegator { // 9
		return struct {
			*responseWriterDelegator
			io.ReaderFrom
			http.CloseNoti***REMOVED***er
		}{d, &readerFromDelegator{d}, &closeNoti***REMOVED***erDelegator{d}}
	}
	pickDelegator[readerFrom+flusher] = func(d *responseWriterDelegator) delegator { // 10
		return struct {
			*responseWriterDelegator
			io.ReaderFrom
			http.Flusher
		}{d, &readerFromDelegator{d}, &flusherDelegator{d}}
	}
	pickDelegator[readerFrom+flusher+closeNoti***REMOVED***er] = func(d *responseWriterDelegator) delegator { // 11
		return struct {
			*responseWriterDelegator
			io.ReaderFrom
			http.Flusher
			http.CloseNoti***REMOVED***er
		}{d, &readerFromDelegator{d}, &flusherDelegator{d}, &closeNoti***REMOVED***erDelegator{d}}
	}
	pickDelegator[readerFrom+hijacker] = func(d *responseWriterDelegator) delegator { // 12
		return struct {
			*responseWriterDelegator
			io.ReaderFrom
			http.Hijacker
		}{d, &readerFromDelegator{d}, &hijackerDelegator{d}}
	}
	pickDelegator[readerFrom+hijacker+closeNoti***REMOVED***er] = func(d *responseWriterDelegator) delegator { // 13
		return struct {
			*responseWriterDelegator
			io.ReaderFrom
			http.Hijacker
			http.CloseNoti***REMOVED***er
		}{d, &readerFromDelegator{d}, &hijackerDelegator{d}, &closeNoti***REMOVED***erDelegator{d}}
	}
	pickDelegator[readerFrom+hijacker+flusher] = func(d *responseWriterDelegator) delegator { // 14
		return struct {
			*responseWriterDelegator
			io.ReaderFrom
			http.Hijacker
			http.Flusher
		}{d, &readerFromDelegator{d}, &hijackerDelegator{d}, &flusherDelegator{d}}
	}
	pickDelegator[readerFrom+hijacker+flusher+closeNoti***REMOVED***er] = func(d *responseWriterDelegator) delegator { // 15
		return struct {
			*responseWriterDelegator
			io.ReaderFrom
			http.Hijacker
			http.Flusher
			http.CloseNoti***REMOVED***er
		}{d, &readerFromDelegator{d}, &hijackerDelegator{d}, &flusherDelegator{d}, &closeNoti***REMOVED***erDelegator{d}}
	}
}
