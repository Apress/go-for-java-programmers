package main

import (
	"fmt"
)

var count int
var done = make(chan bool, 100)

func sayDone(index int) {
	done <- true
	fmt.Printf("go %d done\n", index)
}

func waitUntilAllDone(done chan bool, count int) {
	for count > 0 {
		if <-done {
			count--
		}
	}
}

func main() {
	fmt.Println("Started")
	for i := 0; i < 5; i++ {
		count++
		go func(index int) {
			defer sayDone(index)
			fmt.Printf("go %d running\n", index)
		}(i)
	}

	waitUntilAllDone(done, count)
	fmt.Println("Done")
}

