package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

func CompressFileToNewGZIPFile(path string) (err error) {
	// dummy compression code
	fmt.Printf("Starting compression of %s...\n", path)
	start := time.Now()
	time.Sleep(time.Duration(rand.Intn(5) + 1) * time.Second)
	end := time.Now()
	fmt.Printf("Compression of %s complete in %d seconds\n", path,
		end.Sub(start) / time.Second)
	return
}

func main() {
	var wg sync.WaitGroup
	for _, arg := range os.Args[1:] { // Args[0] is program name
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			err := CompressFileToNewGZIPFile(path)
			if err != nil {
				log.Printf("File %s received error: %v\n", path, err)
				os.Exit(1)
			}
		}(arg)  // prevents duplication of arg in all goroutines
	}
	wg.Wait()
}

