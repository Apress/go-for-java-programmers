package main

import (
	"fmt"
	"time"
)

type RequestFunc func(arg interface{}) interface{}

type GenericRequest struct {
	Handler     RequestFunc
	Args        interface{}
	Result      chan interface{}
}

func NewGenericRequest(h RequestFunc, a interface{},
	r chan interface{}) (gr * GenericRequest) {
	gr = &GenericRequest{h, a, r}
	if gr.Result == nil {
		gr.Result = make(chan interface{})
	}
	return
}

func HandleGenericRequests(requests chan *GenericRequest) {
	for req := range requests {
		req := req
		go func() {
			req.Result <- req.Handler(req.Args)
		}()
	}
}

var Requests = make(chan *GenericRequest, 100)

func sumFloat(arg interface{}) interface{} {
	var res float64
	values, ok := arg.([]float64)
	if ok {
		for _, v := range values {
			res += v
		}
	}
	return res
}

func main() {
	reqs := make([]*GenericRequest, 0, 10)
	reqs = append(reqs, NewGenericRequest(sumFloat, []float64{1, 2, 3}, nil))
	reqs = append(reqs, NewGenericRequest(sumFloat, []float64{5, 6, 7}, nil))
	reqs = append(reqs, NewGenericRequest(sumFloat, []float64{7, 8, 9}, nil))
	for _, r := range reqs {
		// accepts < 100  requests without blocking
		Requests <- r
	}
	go HandleGenericRequests(Requests)

	time.Sleep(5 * time.Second)  // simulate doing other work

	for i, r := range reqs {
		fmt.Printf("sum %d: %v\n", i+1, <-r.Result) // wait for each to finish
	}
	close(Requests)
}
