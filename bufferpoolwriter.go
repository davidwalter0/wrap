package wrap

import (
	"github.com/oxtoacart/bpool"
	"net/http"
)

var detail = false
var enable = false
var BP *bpool.SizedBufferPool
var BPSize = 32
var BPAlloc = 16384

func EnableTrace(e bool) {
	enable = e
}

func BufferPool() *bpool.SizedBufferPool {
	if BP == nil {
		BP = bpool.NewSizedBufferPool(BPSize, BPAlloc)
	}
	return BP
}

// NewBufferPoolWriter returns a BufferWriter wrapping the given
// response writer.
func NewBufferPoolWriter(w http.ResponseWriter) (bf *BufferWriter) {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	bf = &BufferWriter{}
	bf.ResponseWriter = w
	bf.header = make(http.Header)
	bf.Buffer = BufferPool().Get()
	return
}

// FlushAll flushes headers, status code and body to the underlying
// ResponseWriter, if something changed
func (bf *BufferWriter) BPFlushAll() {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	if bf.HasChanged() {
		bf.FlushHeaders()
		bf.FlushCode()
		bf.ResponseWriter.Write(bf.Buffer.Bytes())
		BufferPool().Put(bf.Buffer)
	}
}
