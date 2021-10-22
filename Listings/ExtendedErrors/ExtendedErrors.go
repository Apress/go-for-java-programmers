package main

import (
	"errors"
	"fmt"
)

type MultError []error

func (me MultError) Error() (res string) {
	res = "MultError"
	sep := " "
	for _, e := range me {
		res = fmt.Sprintf("%s%s%s", res, sep, e.Error())
		sep = "; "
	}
	return
}
func (me MultError) String() string {
	return me.Error()
}

type ErrorWithCause struct {
	Err   error
	Cause error
}

func NewError(err error) *ErrorWithCause {
	return NewErrorWithCause(err, nil)
}
func NewErrorWithCause(err error, cause error) *ErrorWithCause {
	if err == nil {
		err = errors.New("no error supplied")
	}
	return &ErrorWithCause{err, cause}
}
func (wc ErrorWithCause) Error() string {
	xerr := wc.Err
	xcause := wc.Cause
	if xcause == nil {
		xcause = errors.New("no root cause supplied")
	}
	return fmt.Sprintf("ErrorWithCause{%v %v}", xerr, xcause)
}
func (wc ErrorWithCause) String() string {
	return wc.Error()
}

type TryFunc func() error
type CatchFunc func(error) (rerr error, cerr error)
type FinallyFunc func()

type TryCatchError struct {
	tryError   error
	catchError error
}

func (tce *TryCatchError) Error() string {
	return tce.String()
}
func (tce *TryCatchError) String() string {
	return fmt.Sprintf("TryCatchError[%v %v]", tce.tryError, tce.catchError)
}
func (tce *TryCatchError) Cause() error {
	return tce.tryError
}
func (tce *TryCatchError) Catch() error {
	return tce.catchError
}

func TryFinally(t TryFunc, f FinallyFunc) (err error) {
	defer func() {
		f()
	}()
	err = t()
	if err != nil {
		err = &TryCatchError{err, nil}
	}
	return
}

func triageRecover(p interface{}, c CatchFunc) (err error) {
	if p != nil {
		var terr, cerr error
		if v, ok := p.(error); ok {
			terr = v
		}
		if xrerr, xcerr := c(terr); xrerr != nil {
			cerr = xcerr
			err = xrerr
		}
		if terr != nil || cerr != nil {
			err = &TryCatchError{terr, cerr}
		}
	}
	return err
}

func TryCatch(t TryFunc, c CatchFunc) (err error) {
	defer func() {
		if xerr := triageRecover(recover(), c); xerr != nil {
			err = xerr
		}
	}()
	err = t()
	return
}

func TryCatchFinally(t TryFunc, c CatchFunc, f FinallyFunc) (err error) {
	defer func() {
		f()
	}()
	defer func() {
		if xerr := triageRecover(recover(), c); xerr != nil {
			err = xerr
		}
	}()
	err = t()
	return
}

func main() {
	me  := MultError(make([]error,0, 10))
	for _, v := range []string{"one", "two", "three"} {
		me = append(me, errors.New(v))
	}
	fmt.Printf("MultipleError error: %s\n", me.Error())
	fmt.Printf("MultipleError value: %v\n\n", me)

	ewc := NewErrorWithCause(errors.New("error"), errors.New("cause"))
	fmt.Printf("ErrorWithCause error: %s\n", ewc.Error())
	fmt.Printf("ErrorWithCause value: %v\n\n", ewc)

	err := TryCatchFinally(func() error {
		fmt.Printf("in try\n")
		panic(errors.New("forced panic"))
	}, func(e error) (re, ce error) {
		fmt.Printf("in catch %v: %v %v\n", e, re, ce)
		return
	}, func() {
		fmt.Printf("in finally\n")
	})
	fmt.Printf("TCF returned: %v\n", err)

	err = TryFinally(func() error {
		fmt.Printf("in try\n")
		return errors.New("try error")
	}, func() {
		fmt.Printf("in finally\n")
	})
	fmt.Printf("TCF returned: %v\n", err)

	err = TryCatch(func() error {
		fmt.Printf("in try\n")
		panic(errors.New("forced panic"))
	}, func(e error) (re, ce error) {
		fmt.Printf("in catch %v: %v %v\n", e, re, ce)
		return
	})
	fmt.Printf("TCF returned: %v\n", err)

	err = TryCatch(func() error {
		fmt.Printf("in try\n")
		return nil
	}, func(e error) (re, ce error) {
		fmt.Printf("in catch %v: %v %v\n", e, re, ce)
		return
	})
	fmt.Printf("TCF returned: %v\n", err)
}