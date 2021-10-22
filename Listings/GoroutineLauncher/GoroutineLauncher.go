package main

import (
	"errors"
	"fmt"
	"time"
)
var NoError = errors.New("no error")  // special error

func GoroutineLauncher(gr func(), c *(chan error)) {
	go func(){
		defer func(){
			if p := recover(); p != nil {
				if c != nil {
					// ensure we send an error
					if err, ok := p.(error); ok {
						*c <- err
						return
					}
					*c <- errors.New(fmt.Sprintf("%v", p))
				}
				return
			}
			if c != nil {
				*c <- NoError  // could also send nil and test for it
			}
		}()
		gr()
	}()
}

var N = 5

func main() {
	var errchan = make(chan error, N)  // N >= 1 based on max active goroutines
	// :
	GoroutineLauncher (func(){
		time.Sleep(2 * time.Second)  // simulate complex work
		panic("panic happened!")
	}, &errchan)
	// :
	time.Sleep(5 * time.Second)        // simulate other work
	// :
	err := <- errchan  // wait for result
	if err != NoError {
		fmt.Printf("got %q" , err.Error())
	}
}
