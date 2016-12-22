/*
This source file was originally licensed under the following:

The MIT License (MIT)

Copyright (c) 2014 go-on webframework for golang

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

*/

package wrap

import (
	"bytes"
	"net/http"
)

type Buffer interface {
	Context(ctxPtr interface{}) bool
	SetContext(ctxPtr interface{})
	Header() http.Header
	WriteHeader(i int)
	Write(b []byte) (int, error)
	Reset()
	FlushAll()
	Body() []byte
	BodyString() string
	HasChanged() bool
	IsOk() bool
	FlushCode()
	FlushHeaders()
}

// Buffer is a ResponseWriter wrapper that may be used as buffer.
type BufferWriter struct {

	// ResponseWriter is the underlying response writer that is wrapped
	// by BufferWriter
	http.ResponseWriter

	// BufferWriter is the underlying io.Writer that buffers the response
	// body
	Buffer *bytes.Buffer

	// Code is the cached status code
	Code int

	// changed tracks modifications to ResponseWriter and reads from the
	// header - tracked as changes
	changed bool

	// header is the cached header
	header http.Header
}

// NewBufferWriter returns a BufferWriter wrapping the given response
// writer.
func NewBufferWriter(w http.ResponseWriter) (bf *BufferWriter) {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	bf = &BufferWriter{}
	bf.ResponseWriter = w
	bf.header = make(http.Header)
	bf.Buffer = &bytes.Buffer{}
	return
}

// Header returns the cached http.Header and tracks this call as
// change
func (bf *BufferWriter) Header() http.Header {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	bf.changed = true
	return bf.header
}

// WriteHeader writes the cached status code and tracks this call as
// change
func (bf *BufferWriter) WriteHeader(i int) {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	bf.changed = true
	bf.Code = i
}

// Write writes to the underlying buffer and tracks this call as
// change
func (bf *BufferWriter) Write(b []byte) (int, error) {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	bf.changed = true
	return bf.Buffer.Write(b)
}

// Reset set the BufferWriter to the defaults
func (bf *BufferWriter) Reset() {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	bf.Buffer.Reset()
	bf.Code = 0
	bf.changed = false
	bf.header = make(http.Header)
}

// FlushAll flushes headers, status code and body to the underlying
// ResponseWriter, if something changed
func (bf *BufferWriter) FlushAll() {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	if bf.HasChanged() {
		bf.FlushHeaders()
		bf.FlushCode()
		bf.ResponseWriter.Write(bf.Buffer.Bytes())
	}
}

// Body returns the bytes of the underlying buffer (that is meant to
// be the body of the response)
func (bf *BufferWriter) Body() []byte {
	return bf.Buffer.Bytes()
}

// BodyString returns the string of the underlying buffer (that is
// meant to be the body of the response)
func (bf *BufferWriter) BodyString() string {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	return bf.Buffer.String()
}

// HasChanged returns true if Header, WriteHeader or Write has been
// called
func (bf *BufferWriter) HasChanged() bool {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	return bf.changed
}

// IsOk returns true if the cached status code is not set or in the
// 2xx range.
func (bf *BufferWriter) IsOk() bool {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	if bf.Code == 0 {
		return true
	}
	if bf.Code >= 200 && bf.Code < 300 {
		return true
	}
	return false
}

// FlushCode flushes the status code to the underlying responsewriter
// if it was set.
func (bf *BufferWriter) FlushCode() {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	if bf.Code != 0 {
		bf.ResponseWriter.WriteHeader(bf.Code)
	}
}

// FlushHeaders adds the headers to the underlying ResponseWriter,
// removing them from BufferWriter.
func (bf *BufferWriter) FlushHeaders() {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	header := bf.ResponseWriter.Header()
	for k, v := range bf.header {
		header.Del(k)
		for _, val := range v {
			header.Add(k, val)
		}
	}
}
