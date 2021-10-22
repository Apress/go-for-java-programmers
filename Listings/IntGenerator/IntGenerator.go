/*
 Copyright (c) DellEMC 2022. All rights reserved.

 This code is owned by DellEMC.  Any copying, extension or reproduction without DellEMC's explicit permission is not authorized.
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func generateIntValues(ctx context.Context, values chan<- int) {
loop:
	for {
		v, err := genIntValue()
		if err != nil {
			fmt.Printf("genIntValue error: %v\n", err)
			close(values)
			break
		}
		select {
		case values <- v: // output value
			fmt.Printf("generateIntValues sent: %v\n", v)
		case <-ctx.Done():
			break loop // done when something received
		}
	}
}
func genIntValue() (v int, err error) {
	test := rand.Intn(20) % 5
	if test == 0 {
		err = errors.New(fmt.Sprintf("fake some error"))
		return
	}
	v = rand.Intn(100)
	fmt.Printf("genIntValue next: %d\n", v)
	return
}

func main() {
	values := make(chan int, 10)
	ctx, cf := context.WithTimeout(context.Background(), 5*time.Second)
	go generateIntValues(ctx, values)
	for v := range values { // get all generated
		fmt.Printf("generateIntValues received: %d\n", v)
	}
	cf()
	fmt.Printf("generateIntValues done\n")

}
