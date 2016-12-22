package wrap

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func HttpScopedBufferHandler(handler http.Handler) http.Handler {
	switch strings.ToLower(os.Getenv("WRAP_BUFFER_HANDLER")) {
	case "pool", "pooledhandler", "bphandler":
		return HttpScopedBPHandlerWriter(handler)
	case "", "default", "bytes.buffer", "buffer":
		fallthrough
	default:
		return HttpScopedHandlerWriter(handler)
	}
}

// use a buffer buffer pools buffer then write/flush the buffer to the ResponseWriter.
func HttpScopedBPHandlerWriter(handler http.Handler) http.Handler {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buffer := NewBufferPoolWriter(w)
		var text string
		if enable {
			text = fmt.Sprintf("%v %p", buffer, buffer)
		}
		defer buffer.BPFlushAll()
		defer tracer.Detailed(detail).Enable(enable).ScopedTrace(text)()
		handler.ServeHTTP(buffer, r)
	})
}

// use a bytes.Buffer then write/flush the buffer to the ResponseWriter
func HttpScopedHandlerWriter(handler http.Handler) http.Handler {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buffer := NewBufferWriter(w)
		var text string
		if enable {
			text = fmt.Sprintf("%v %p", buffer, buffer)
		}
		defer buffer.FlushAll()
		defer tracer.Detailed(detail).Enable(enable).ScopedTrace(text)()
		handler.ServeHTTP(buffer, r)
	})
}

var NoOp = http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
})

// Recover recovers from any panicking goroutine
func Recover(next http.Handler) http.Handler {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
		defer func() {
			defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
			err := recover()
			if err != nil {
				fmt.Fprintf(w, "%v", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RecoverFunc(next http.HandlerFunc) http.HandlerFunc {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	return Recover(next).(http.HandlerFunc)
}

// type Chainer http.HandlerFunc
type ChainerFunc func(http.Handler) http.Handler

// Chain creates an ordered chain of handlers from an argument list
// The handlers call chain A->B->C => R(A)->R(B)->R(C)
func Chain(handlers ...http.HandlerFunc) http.Handler {
	defer tracer.Enable(enable).ScopedTrace()()
	if len(handlers) > 1 {
		next := Chain(handlers[1:]...)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer tracer.Enable(enable).ScopedTrace()()
			handlers[0].ServeHTTP(w, r)
			next.ServeHTTP(w, r)
		})
	} else if len(handlers) == 1 {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer tracer.Enable(enable).ScopedTrace()()
			handlers[0].ServeHTTP(w, r)
		})
	}
	return NoOp
}

// ChainLinkWrap wraps each handler in the argument list of handlers
// The handlers call chain A->B->C => R(A->B->C)
func ChainLinkWrap(wrapper ChainerFunc, handlers ...http.HandlerFunc) http.Handler {
	defer tracer.Enable(enable).ScopedTrace()()
	if len(handlers) > 1 {
		next := Chain(handlers[1:]...)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer tracer.Enable(enable).ScopedTrace()()
			wrapper(handlers[0]).ServeHTTP(w, r)
			wrapper(next).ServeHTTP(w, r)
		})
	} else if len(handlers) == 1 {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer tracer.Enable(enable).ScopedTrace()()
			wrapper(handlers[0]).ServeHTTP(w, r)
		})
	}
	return NoOp
}
