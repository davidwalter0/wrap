package wrap

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func header() http.Header {
	var h http.Header = make(http.Header)
	h["Content-Type"] = []string{"text/plain; charset=utf-8"}
	return http.Header(h)
}

func body() *bytes.Buffer {
	var b bytes.Buffer
	b.Write([]byte("Body Text"))
	return &b
}

func Response() *httptest.ResponseRecorder {
	var response httptest.ResponseRecorder = httptest.ResponseRecorder{
		Code:      200,
		HeaderMap: header(),
		Body:      body(),
		Flushed:   false,
	}
	return &response
}

func failure() *bytes.Buffer {
	var b bytes.Buffer
	b.Write([]byte("Body Text:Failure"))
	return &b
}

func FailureResponse() *httptest.ResponseRecorder {
	var response httptest.ResponseRecorder = httptest.ResponseRecorder{
		Code:      200,
		HeaderMap: header(),
		Body:      failure(),
		Flushed:   false,
	}
	return &response
}

func EmptyResponse() *httptest.ResponseRecorder {
	var response httptest.ResponseRecorder = httptest.ResponseRecorder{
		Code:      200,
		HeaderMap: http.Header{},
		Body:      &bytes.Buffer{},
		Flushed:   false,
	}
	return &response
}

func compare(lhs, rhs *httptest.ResponseRecorder) bool {
	var rc bool
	lv, lok := lhs.HeaderMap["Content-Type"]
	rv, rok := rhs.HeaderMap["Content-Type"]
	if lok && rok {
		rc = lhs.Code == rhs.Code &&
			reflect.DeepEqual(lhs.HeaderMap["Content-Type"], rhs.HeaderMap["Content-Type"]) &&
			lhs.Body.String() == rhs.Body.String()
	} else {
		rc = lhs.Code == rhs.Code &&
			lv == nil && rv == nil &&
			lhs.Body.String() == rhs.Body.String()
	}
	return rc
}

func x(w http.ResponseWriter, r *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
}

var X = http.HandlerFunc(x)

func y(w http.ResponseWriter, r *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
}

var Y = http.HandlerFunc(y)

func z(w http.ResponseWriter, r *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
}

var Z = http.HandlerFunc(z)

func a(w http.ResponseWriter, r *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
	w.Write(body().Bytes())
}

var A = http.HandlerFunc(a)

func b(w http.ResponseWriter, r *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
}

var B = http.HandlerFunc(b)

func failer(w http.ResponseWriter, r *http.Request) {
	defer tracer.Enable(enable).ScopedTrace()()
	panic(":Failure")
}

func Test_BufferWriteChain(t *testing.T) {
	handler := HttpScopedHandlerWriter(Chain(x, y, z, a, RecoverFunc(failer)))
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		fmt.Println("Creating 'GET /' request failed!")
		os.Exit(0)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !compare(rec, FailureResponse()) {
		t.Fail()
	}
}

func Test_BufferBPWriteChain1(t *testing.T) {
	handler := HttpScopedBPHandlerWriter(ChainLinkWrap(Recover, x, y, z, a))
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		fmt.Println("Creating 'GET /' request failed!")
		os.Exit(0)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !compare(rec, Response()) {
		t.Fail()
	}
}

func Test_BufferBPWriteChain2(t *testing.T) {
	handler := HttpScopedBPHandlerWriter(ChainLinkWrap(Recover, x, y, z, a))
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		fmt.Println("Creating 'GET /' request failed!")
		os.Exit(0)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !compare(rec, Response()) {
		t.Fail()
	}
}

func Test_BufferWriteChain1(t *testing.T) {
	handler := HttpScopedHandlerWriter(ChainLinkWrap(Recover, x, y, z, a))
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		fmt.Println("Creating 'GET /' request failed!")
		os.Exit(0)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !compare(rec, Response()) {
		t.Fail()
	}
}

func Test_BufferWriteChain2(t *testing.T) {
	handler := HttpScopedHandlerWriter(ChainLinkWrap(Recover, x, y, z, a))
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		fmt.Println("Creating 'GET /' request failed!")
		os.Exit(0)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !compare(rec, Response()) {
		t.Fail()
	}
}

func Test_Chain(t *testing.T) {
	handler := ChainLinkWrap(Recover, x, y, z, a)
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		fmt.Println("Creating 'GET /' request failed!")
		os.Exit(0)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !compare(rec, Response()) {
		t.Fail()
	}
}

func Test_EmptyChain(t *testing.T) {
	handler := Chain()
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		fmt.Println("Creating 'GET /' request failed!")
		os.Exit(0)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !compare(rec, EmptyResponse()) {
		t.Fail()
	}
}

const (
	N int = 8192
)

func UnBufferedWriter(w http.ResponseWriter) {
	const n int = N
	for i := 0; i < n; i++ {
		w.Write([]byte(fmt.Sprintf("%256d", i)))
	}
}

func Filler() []byte {
	const n int = N
	var b string
	for i := 0; i < n; i++ {
		if len(b) == 0 {
			b = fmt.Sprintf("%256d", i)
		} else {
			b += fmt.Sprintf(", %8d", i)
		}
	}
	return []byte(b)
}

var filler []byte = Filler()

func BufferFillHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(filler)
}

func UnBufferedFillHandler(w http.ResponseWriter, r *http.Request) {
	UnBufferedWriter(w)
}

func Benchmark_Buffer_BP_Multi_Chain_Create_A(b *testing.B) {
	enable = false
	detail = false
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	iterations := 100
	for i := 0; i < iterations; i++ {
		handler := HttpScopedBPHandlerWriter(Chain(x, y, z, UnBufferedFillHandler))
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Println("Creating 'GET /' request failed!")
			os.Exit(0)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func Benchmark_Buffer_Multi_Chain_Create_B(b *testing.B) {
	enable = false
	detail = false
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	iterations := 100
	for i := 0; i < iterations; i++ {
		handler := HttpScopedHandlerWriter(Chain(x, y, z, UnBufferedFillHandler))
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Println("Creating 'GET /' request failed!")
			os.Exit(0)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func Benchmark_UnBuffered_Multi_Chain_Create_C(b *testing.B) {
	enable = false
	detail = false
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	iterations := 100
	handler := Chain(x, y, z, UnBufferedFillHandler)
	for i := 0; i < iterations; i++ {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Println("Creating 'GET /' request failed!")
			os.Exit(0)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func Benchmark_Buffer_BP_Make_One_Chain_D(b *testing.B) {
	enable = false
	detail = false
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	iterations := 100
	handler := HttpScopedBPHandlerWriter(Chain(x, y, z, BufferFillHandler))
	for i := 0; i < iterations; i++ {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Println("Creating 'GET /' request failed!")
			os.Exit(0)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func Benchmark_Buffer_Make_One_Chain_E(b *testing.B) {
	enable = false
	detail = false
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	iterations := 100
	handler := HttpScopedHandlerWriter(Chain(x, y, z, BufferFillHandler))
	for i := 0; i < iterations; i++ {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Println("Creating 'GET /' request failed!")
			os.Exit(0)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func Benchmark_UnBuffered_Make_One_Chain_F(b *testing.B) {
	enable = false
	detail = false
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	iterations := 100
	handler := Chain(x, y, z, UnBufferedFillHandler)
	for i := 0; i < iterations; i++ {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Println("Creating 'GET /' request failed!")
			os.Exit(0)
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func Benchmark_Buffer_BP_G(b *testing.B) {
	enable = false
	detail = false
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	iterations := 100
	for i := 0; i < iterations; i++ {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Println("Creating 'GET /' request failed!")
			os.Exit(0)
		}
		rec := httptest.NewRecorder()
		X.ServeHTTP(rec, req)
		Y.ServeHTTP(rec, req)
		Z.ServeHTTP(rec, req)
		BufferFillHandler(rec, req)
	}
}

func Benchmark_Buffer_H(b *testing.B) {
	enable = false
	detail = false
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	iterations := 100

	for i := 0; i < iterations; i++ {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Println("Creating 'GET /' request failed!")
			os.Exit(0)
		}
		rec := httptest.NewRecorder()
		X.ServeHTTP(rec, req)
		Y.ServeHTTP(rec, req)
		Z.ServeHTTP(rec, req)
		BufferFillHandler(rec, req)
	}
}

func Benchmark_UnBuffered_I(b *testing.B) {
	enable = false
	detail = false
	defer tracer.Detailed(detail).Enable(enable).ScopedTrace()()
	iterations := 100
	for i := 0; i < iterations; i++ {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			fmt.Println("Creating 'GET /' request failed!")
			os.Exit(0)
		}
		rec := httptest.NewRecorder()
		X.ServeHTTP(rec, req)
		Y.ServeHTTP(rec, req)
		Z.ServeHTTP(rec, req)
		UnBufferedFillHandler(rec, req)
	}
}
