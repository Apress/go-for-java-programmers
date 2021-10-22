package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func LogWrapper(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		method, path := req.Method, req.URL
		fmt.Printf("entered handler for %s %s\n", method, path)
		f(w, req)
		fmt.Printf("exited handler for %s %s\n", method, path)
	}
}
func ElapsedTimeWrapper(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		method, path := req.Method, req.URL
		start := time.Now().UnixNano()
		f(w, req)
		fmt.Printf("elapsed time for %s %s: %dns\n",
			method, path, time.Now().UnixNano() - start)
	}
}

var spec = ":8086"  // localhost

func main() {
	// regular HTTP request handler
	handler := func(w http.ResponseWriter, req *http.Request) {
		fmt.Printf("in handler %v %v\n", req.Method, req.URL)
		time.Sleep(1 * time.Second)
		w.Write([]byte(fmt.Sprintf("In handler for %s %s", req.Method, req.URL)))
	}
	// advised handler
	http.HandleFunc("/test", LogWrapper(ElapsedTimeWrapper(handler)))
	if err := http.ListenAndServe(spec, nil); err != nil {
		log.Fatalf("Failed to start server on %s: %v", spec, err)
	}
}