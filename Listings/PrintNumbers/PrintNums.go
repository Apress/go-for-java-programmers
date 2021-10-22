package main

import (
	"fmt"
	"math/rand"
	"time"
)

func printNum(id string, count int) {
	for i := 0; i < count; i++ {
		fmt.Printf("%s: %d\n", id, i)
		delay := time.Duration(rand.Intn(10)) * time.Millisecond
		time.Sleep(delay)  // delay a bit
	}
}

func main() {
	printNum("one", 5)
	printNum("two", 5)
	printNum("main", 5)
	fmt.Println("---------------------------------")
	go printNum("one", 5)
	go printNum("two", 5)
	printNum("main", 5)
}

