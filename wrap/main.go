package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	trace "github.com/davidwalter0/tracer"
	"github.com/davidwalter0/wrap"
)

var detail = false
var enable = false
var tracer = trace.New()

type logWriter struct {
}

type Format string
type Formats []Format

func (format Format) Split(on string) (formats Formats) {
	for {
		where := strings.Index(string(format), on)
		switch where {
		case -1:
			return append(formats, format)
		default:
			formats = append(formats, format[0:where])
			format = format[len(on)+where:]
		}
	}
	return formats
}

func (format Format) SplitRecurse(on string) (formats Formats) {
	where := strings.Index(string(format), on)
	switch where {
	case -1:
		return Formats{format}
	default:
		return append(Formats{format[0:where]}, format[len(on)+where:].SplitRecurse(on)...)
	}
}

func (f Formats) Join(with string) (s Format) {
	fmt.Println("Join", f, with)
	for _, t := range f {
		if len(s) == 0 {
			s += Format(t)
		} else {
			s += Format(with) + Format(t)
		}
	}
	return
}

// using varargs to pass option zero or one args, only use the first
// if specified
func (f Formats) String(args ...string) string {
	var with string = " "
	if len(args) > 0 {
		with = args[0]
	}
	return string(f.Join(with))
}

func (f Format) String() string {
	return string(f)
}

var format Format = Format(fmt.Sprintf("%s", time.Now().UTC().Format("2006.01.02.15.04.05.000.-0700.MST")))

var TimeForm string = "2006.01.02.15.04.05.000.-0700.MST"

func header() http.Header {
	var h http.Header = make(http.Header)
	h["Content-Type"] = []string{"text/plain; charset=utf-8"}
	return http.Header(h)
}

func body() *bytes.Buffer {
	var b bytes.Buffer
	b.Write([]byte(" [A: Body Text] "))
	return &b
}

func b(w http.ResponseWriter, r *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
	w.Write([]byte(" [B: This is more text] "))
}

var B = http.HandlerFunc(b)

func a(w http.ResponseWriter, r *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
	w.Write(body().Bytes())
}

var A = http.HandlerFunc(a)

func panicky(w http.ResponseWriter, r *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
	panic(string([]byte(" going... <" + string(body().Bytes()) + " > going...")))
	w.Write(body().Bytes())
}

var Panicky = http.HandlerFunc(panicky)

func (writer logWriter) Write(bytes []byte) (int, error) {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	return fmt.Printf(time.Now().UTC().Format(TimeForm) + " [DEBUG] TEST TEXT> \n")
}

func Time(w http.ResponseWriter, r *http.Request) {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	log.Printf(time.Now().UTC().Format(TimeForm) + " [DEBUG] TEST TEXT > \n")
	fmt.Fprintf(w, time.Now().UTC().Format(TimeForm)+" [DEBUG] TEST TEXT > ")
}

func Recover(next http.Handler) http.Handler {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
		defer func() {
			defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
			err := recover()
			if err != nil {
				fmt.Fprintf(w, " >>>%v<<< ", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RecoverFunc(next http.HandlerFunc) http.HandlerFunc {
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	return Recover(next).(http.HandlerFunc)
}

func main() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	port := os.Getenv("PORT")
	host := os.Getenv("HOST")

	if len(port) == 0 {
		fmt.Println("PORT not set, using 8080")
		port = "8080"
	} else {
		fmt.Println("PORT=" + port)
	}

	if len(host) == 0 {
		fmt.Println("HOST not set, default bind all")
		host = "0.0.0.0"
	} else {
		fmt.Println("HOST=" + host)
	}
	listen := host + ":" + port

	fmt.Println("PORT on which  " + ":" + port)
	fmt.Println("HOST interface " + ":" + host)

	fmt.Println("listening on " + listen)

	handler := wrap.Chain(Time, A, B, A)
	// Panic and fail to complete the writes
	handlerPanic := wrap.Chain(Time, A, B, A, RecoverFunc(Panicky))
	// Panic / recover and write
	handlerPanicky := wrap.ChainLinkWrap(Recover, Time, A, Panicky, B, A)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { handler.ServeHTTP(w, r) })
	http.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) { handlerPanic.ServeHTTP(w, r) })
	http.HandleFunc("/panicky", func(w http.ResponseWriter, r *http.Request) { handlerPanicky.ServeHTTP(w, r) })
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		fmt.Println(err)
	}

}
