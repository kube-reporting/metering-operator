/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE ***REMOVED***le
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this ***REMOVED***le
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this ***REMOVED***le except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * speci***REMOVED***c language governing permissions and limitations
 * under the License.
 */

package thrift

import (
	"bu***REMOVED***o"
)

type TBufferedTransportFactory struct {
	size int
}

type TBufferedTransport struct {
	bu***REMOVED***o.ReadWriter
	tp TTransport
}

func (p *TBufferedTransportFactory) GetTransport(trans TTransport) (TTransport, error) {
	return NewTBufferedTransport(trans, p.size), nil
}

func NewTBufferedTransportFactory(bufferSize int) *TBufferedTransportFactory {
	return &TBufferedTransportFactory{size: bufferSize}
}

func NewTBufferedTransport(trans TTransport, bufferSize int) *TBufferedTransport {
	return &TBufferedTransport{
		ReadWriter: bu***REMOVED***o.ReadWriter{
			Reader: bu***REMOVED***o.NewReaderSize(trans, bufferSize),
			Writer: bu***REMOVED***o.NewWriterSize(trans, bufferSize),
		},
		tp: trans,
	}
}

func (p *TBufferedTransport) IsOpen() bool {
	return p.tp.IsOpen()
}

func (p *TBufferedTransport) Open() (err error) {
	return p.tp.Open()
}

func (p *TBufferedTransport) Close() (err error) {
	return p.tp.Close()
}

func (p *TBufferedTransport) Read(b []byte) (int, error) {
	n, err := p.ReadWriter.Read(b)
	if err != nil {
		p.ReadWriter.Reader.Reset(p.tp)
	}
	return n, err
}

func (p *TBufferedTransport) Write(b []byte) (int, error) {
	n, err := p.ReadWriter.Write(b)
	if err != nil {
		p.ReadWriter.Writer.Reset(p.tp)
	}
	return n, err
}

func (p *TBufferedTransport) Flush() error {
	if err := p.ReadWriter.Flush(); err != nil {
		p.ReadWriter.Writer.Reset(p.tp)
		return err
	}
	return p.tp.Flush()
}

func (p *TBufferedTransport) RemainingBytes() (num_bytes uint64) {
	return p.tp.RemainingBytes()
}
