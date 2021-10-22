package main

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"
)
const factorialnput = 100
const factorialExpect = "93326215443944152681699238856266700490715968264381621468592963895217599993229915608941463976156518286253697920827223758251185210916864000000000000000000000000"


// test the functions; happy case only

func TestEchoInt(t *testing.T) {
	//fmt.Println("in TestEchoInt")
	expect := 10
	got := EchoInt(expect)
	if got != expect {
		reportNoMatch(t, got, expect)
	}
}

func TestSumInt(t *testing.T) {
	//fmt.Println("in TestSumInt")
	expect := 10
	got := SumInt(expect, expect)
	if got != expect+expect {
		reportNoMatch(t, got, expect+expect)
	}
}

func TestEchoFloat(t *testing.T) {
	//fmt.Println("in TestEchoFloat")
	expect := 10.0
	got := EchoFloat(expect)
	if got != expect {
		reportNoMatch(t, got, expect)
	}
}

func TestSumFloat(t *testing.T) {
	//fmt.Println("in TestSumFloat")
	expect := 10.0
	got := SumFloat(expect, expect)
	if got != expect+expect {
		reportNoMatch(t, got, expect+expect)
	}
}

func TestEchoString(t *testing.T) {
	fmt.Println("in TestEchoString")
	expect := "hello"
	got := EchoString(expect)
	if got != expect {
		reportNoMatch(t, got, expect)
	}
}

func TestSumString(t *testing.T) {
	//fmt.Println("in TestSumString")
	expect := "hello"
	got := SumString(expect, expect)
	if got != expect+expect {
		reportNoMatch(t, got, expect+expect)
	}
}

func TestFactorialIterate(t *testing.T) {
	//fmt.Println("in TestFactorialIterate")
	expect := big.NewInt(0)
	expect.SetString(factorialExpect, 10)
	got, err := FactorialIterate(factorialnput)
	if err != nil {
		reportFail(t, err)
	}
	if expect.Cmp(got) != 0 {
		reportNoMatch(t, got, expect)
	}
}

func TestFactorialRecurse(t *testing.T) {
	//fmt.Println("in TestFactorialRecurse")
	expect := big.NewInt(0)
	expect.SetString(factorialExpect, 10)
	got, err := FactorialRecurse(factorialnput)
	if err != nil {
		reportFail(t, err)
	}
	if expect.Cmp(got) != 0 {
		reportNoMatch(t, got, expect)
	}
}

// benchmarks

func BenchmarkFacInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FactorialIterate(factorialnput)
	}
}

func BenchmarkFacRec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FactorialRecurse(factorialnput)
	}
}

// helpers

func reportNoMatch(t *testing.T, got interface{}, expect interface{}) {
	t.Error(fmt.Sprintf("got(%v) != expect(%v)", got, expect))
}

func reportFail(t *testing.T, err error) {
	t.Error(fmt.Sprintf("failure: %v", err))
}

var start time.Time

// do any test setup
func setup() {
	// do any setup here
	fmt.Printf("starting tests...\n")
	start = time.Now()
}

// do any test cleanup
func teardown() {
	end := time.Now()
	// do any cleanup here
	fmt.Printf("tests complete in %dms\n", end.Sub(start)/time.Millisecond)
}

// runs test with setup and cleanup
func TestMain(m *testing.M) {
	setup()
	rc := m.Run()
	teardown()
	os.Exit(rc)
}

