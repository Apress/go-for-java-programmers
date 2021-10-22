package main

import (
	"log"
	"math/rand"
	"net/http"
)

var messages = []string{
	"Now is the time for all good Devops to come the aid of their servers.",
	"Alas poor Altair 8800; I knew it well!",
	"In the beginning there was ARPA Net and its domain was limited.",
	// assume many more
	"A blog a day helps keep the hacker away.",
}

func sendRandomMessage(w http.ResponseWriter, req *http.Request) {
	w.Header().Add(http.CanonicalHeaderKey("content-type"),
		"text/plain")
	w.Write([]byte(messages[rand.Intn(len(messages))]))
}

var spec = ":8080" // means localhost:8080
func main() {
	http.HandleFunc("/message", sendRandomMessage)
	if err := http.ListenAndServe(spec, nil); err != nil {
		log.Fatalf("Failed to start server on %s: %v", spec, err)
	}
}
