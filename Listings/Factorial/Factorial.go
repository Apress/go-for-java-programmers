package main

import (
	"errors"
	"fmt"
	"math/big"
)

var ErrBadArgument = errors.New("invalid argument")

var maxInput = 1_000 // limit result and time

func factorial(n int) (res *big.Int, err error) {
	if n < 0 || n > maxInput {
		err = ErrBadArgument
		return // or raise panic
	}
	res = big.NewInt(1)
	for i := 2; i <= n; i++ {
		res = res.Mul(res, big.NewInt(int64(i)))
	}
	return
}

func main() {
	fact, _ := factorial(100)
	fmt.Println("Factorial(100):", fact)
}
