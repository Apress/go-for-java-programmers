package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func greet(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, fmt.Sprintf("Method %s not supported", req.Method), 405)
		return
	}
	var name string
	if err := req.ParseForm(); err == nil {
		name = strings.TrimSpace(req.FormValue("name"))
	}
	if len(name) == 0 {
		name = "World"
	}
	w.Header().Add(http.CanonicalHeaderKey("content-type"),
		"text/plain")
	io.WriteString(w, fmt.Sprintf("Hello %s!\n", name))
}

func now(w http.ResponseWriter, req *http.Request) {
	// request checks like in greet
	w.Header().Add(http.CanonicalHeaderKey("content-type"),
		"text/plain")
	io.WriteString(w, fmt.Sprintf("%s", time.Now()))
}

func main() {
	fs := http.FileServer(http.Dir(`/temp`))

	http.HandleFunc("/greet", greet)
	http.HandleFunc("/now", now)
	http.Handle( "/static/", http.StripPrefix( "/static", fs ) )
	log.Fatal(http.ListenAndServe(":8088", nil))
}
