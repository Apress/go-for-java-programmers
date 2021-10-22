package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func CountWordsInFile(path string) (counts map[string]int, err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return
	}
	defer f.Close()
	counts, err = scan(f)
	return
}

func scan(r io.Reader) (counts map[string]int, err error) {
	counts = make(map[string]int)
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanWords) // make into words
	for s.Scan() {           // true while words left
		lcw := strings.ToLower(s.Text()) // get last scanned word
		counts[lcw] = counts[lcw] + 1 // missing is zero value
	}
	err = s.Err() // notice any error
	return
}


func main() {
	path := `/temp/words.txt` // point to a real file
	counts, err := CountWordsInFile(path)
	if err != nil {
		fmt.Printf("Count failed: %v\n", err)
		return
	}
	fmt.Printf("Counts for %q:\n", path)
	for k, v := range counts {
		fmt.Printf("  %-20s = %v\n", k, v)
	}
}
