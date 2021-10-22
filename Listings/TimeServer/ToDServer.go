package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"
)

var functionMap = template.FuncMap{
	"formatDate": formatDate,
	"formatTime": formatTime,
}
var parsedTodTemplate *template.Template

func loadTemplate(path string) {
	parsedTodTemplate = template.New("tod")
	parsedTodTemplate.Funcs(functionMap)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panicf("failed reading template %s: %v", path, err)
	}
	if _, err := parsedTodTemplate.Parse(string(data)); err != nil {
		log.Panicf("failed parsing template %s: %v", path, err)
	}
}
func formatDate(dt time.Time) string {
	return dt.Format("Mon Jan _2 2006")
}
func formatTime(dt time.Time) string {
	return dt.Format("15:04:05 MST")
}

type TODData struct {
	TOD time.Time
}

func processTODRequest(w http.ResponseWriter, req *http.Request) {
	var data = &TODData{time.Now()}
	parsedTodTemplate.Execute(w, data) // assume cannot fail
}

var serverPort = 8085

func timeServer() {
	loadTemplate( `C:\Temp\tod.tmpl`)
	http.HandleFunc("/tod", processTODRequest)
	spec := fmt.Sprintf(":%d", serverPort)
	if err := http.ListenAndServe(spec, nil); err != nil {
		log.Fatalf("failed to start server on port %s: %v", spec, err)
	}
	log.Println("server exited")
}

func main() {
	timeServer()
}

