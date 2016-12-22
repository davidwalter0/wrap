package wrap

import (
	"net/http"
	"os"

	trace "github.com/davidwalter0/tracer"
	"github.com/oxtoacart/bpool"
)

var detail = false
var enable = false
var BP *bpool.SizedBufferPool
var BPSize = 32
var BPAlloc = 16384

var tracer *trace.Tracer

// turn on call trace for debug and testing
func TraceEnvConfig() bool {
	switch os.Getenv("WRAP_BUFFER_TRACE_ENABLE") {
	case "enable", "true", "1", "ok", "ack", "on":
		return EnableTrace(true)
	case "disable", "false", "0", "nak", "off":
		fallthrough
	default:
		return EnableTrace(false)
	}
}

func init() {
	tracer = trace.New()
}

func Tracer() *trace.Tracer {
	return tracer
}

func EnableTrace(e bool) bool {
	enable = e
	return e
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
