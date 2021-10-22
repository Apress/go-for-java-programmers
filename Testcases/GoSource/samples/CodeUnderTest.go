package main

import (
	"errors"
	"math/big"
	"math/rand"
	"time"
)

// a set of functions to be tested

// Echo my input

func EchoInt(in int) (out int) {
	randomSleep(50 * time.Millisecond)
	out = in
	return
}

func EchoFloat(in float64) (out float64) {
	randomSleep(50 * time.Millisecond)
	out = in
	return
}

func EchoString(in string) (out string) {
	randomSleep(50 * time.Millisecond)
	out = in
	return
}

// Sum my inputs


func SumInt(in1, in2 int) (out int) {
	randomSleep(50 * time.Millisecond)
	out = in1 + in2
	return
}

func SumFloat(in1, in2 float64) (out float64) {
	randomSleep(5)
	out = in1 + in2
	return
}

func SumString(in1, in2 string) (out string) {
	randomSleep(50 * time.Millisecond)
	out = in1 + in2
	return
}

// Factorial computation: factorial(n):
// n < 0 - undefined
// n == 0 - 1
// n > 0 - n * factorial(n-1)

var ErrInvalidInput = errors.New("invalid input")

// Factorial via iteration
func FactorialIterate(n int64) (res *big.Int, err error) {
	if n < 0 {
		err = ErrInvalidInput
		return
	}
	res = big.NewInt(1)
	if n == 0 {
		return
	}
	for  i := int64(1); i <= n; i++ {
		term := big.NewInt(i)
		//res.Mul(res, big.NewInt(i))
		res = term.Mul(term, res)
	}
	return
}

// Factorial via recursion
func FactorialRecurse(n int64) (res *big.Int, err error) {
	if n < 0 {
		err = ErrInvalidInput
		return
	}
	res = big.NewInt(1)
	if n == 0 {
		return
	}
	term := big.NewInt(n)
	facm1, err := FactorialRecurse(n - 1)
	if err != nil {
		return
	}
	res = term.Mul(term, facm1)
	return
}

// a helper

func randomSleep(dur time.Duration ) {
	time.Sleep(time.Duration((1 + rand.Intn(int(dur)))))
}
