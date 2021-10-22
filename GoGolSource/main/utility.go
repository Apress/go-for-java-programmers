package main

import (
	"bytes"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const NanosPerMs = 1_000_000
const FilePrefix = "file:" // local (vs. HTTP) file

func LoadImage(url string) (img image.Image, kind string, err error) {
	switch {
	case strings.HasPrefix(url, FilePrefix):
		url = url[len(FilePrefix):]
		var b []byte
		b, err = ioutil.ReadFile(url) // read image from file
		if err != nil {
			return
		}
		r := bytes.NewReader(b)
		img, kind, err = image.Decode(r)
		if err != nil {
			return
		}
	default:
		var resp *http.Response
		resp, err = http.Get(url) // get image from network
		if err != nil {
			return
		}
		img, kind, err = image.Decode(resp.Body)
		resp.Body.Close() // error ignored
		if err != nil {
			return
		}
	}
	return
}

// Fail if passed an error.
func fatalIfError(v ...interface{}) {
	if v != nil && len(v) > 0 {
		if err, ok := v[len(v)-1].(error); ok && err != nil {
			log.Fatalf("unexpected error: %v\n", err)
		}
	}
}
