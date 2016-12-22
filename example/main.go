package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	trace "github.com/davidwalter0/tracer"
	"github.com/davidwalter0/wrap"
)

var detail = false
var enable = false
var tracer = trace.New()

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

	http.HandleFunc("/", handler.ServeHTTP)
	http.HandleFunc("/panic", handlerPanic.ServeHTTP)
	http.HandleFunc("/panicky", handlerPanicky.ServeHTTP)
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		fmt.Println(err)
	}

}
