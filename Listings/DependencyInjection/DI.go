package main

import (
	"fmt"
	"time"
)

type Cache interface {
	Get(name string) (interface{}, bool)
	Set(name string, value interface{}) error
	ClearName(name string)
	ClearAll()
}

type MapCache map[string]interface{}

func (c MapCache) Get(name string) (res interface{}, ok bool) {
	res, ok = c[name]
	return
}
func (c MapCache) Set(name string, value interface{}) (err error) {
	c[name] = value
	return
}
func (c MapCache) ClearName(name string) {
	delete(c, name)
	return
}
func (c MapCache) ClearAll() {
	for k, _ := range c {
		delete(c, k)
	}
	return
}

type HTTPClient interface {
	SendReceive(url, method string, in interface{}) (out interface{},
		err error)
}

type EchoHTTPClient struct {
}

func (c *EchoHTTPClient) SendReceive(url, method string, in interface{}) (out interface{},
	err error) {
	out = fmt.Sprintf("SENT %s %s with %v", method, url, in)
	return
}

type Logger interface {
	Log(format string, args ...interface{})
}

type StdoutLogger struct {
}

func (l *StdoutLogger) Log(format string, args ...interface{}) {
	fmt.Printf("%s - %s\n", time.Now().Format(time.StampMilli), fmt.Sprintf(format, args...))
}

type HTTPService struct { // also a HTTPClient
	log    Logger
	client HTTPClient
	cache  Cache
	// :  other fields not using dependencies
}

func NewService(client HTTPClient, log Logger,
	cache Cache) (s *HTTPService) {
	s = &HTTPService{}
	s.log = log
	s.client = client
	s.cache = cache
	// : set other fields
	return
}

func (s *HTTPService) SendReceive(url, method string,
	in interface{}) (out interface{}, err error) {
	key := fmt.Sprintf("%s:%s", method, url)
	if xout, ok := s.cache.Get(key); ok {
		out = xout
		return
	}
	out, err = s.client.SendReceive(url, method, in)
	s.log.Log("SendReceive(%s, %s, %v)=%v", method, url, in, err)
	if err != nil {
		return
	}
	err = s.cache.Set(key, out)
	return
}

func main() {
	// :
	log := StdoutLogger{}      // concrete type
	client := EchoHTTPClient{} // concrete type
	cache := MapCache{}        // concrete type
	// :
	// create a service with all dependencies injected
	s := NewService(&client, &log, cache)
	// :
	for i:= 0; i < 5; i++ {
		if i % 3 == 0 {
			cache.ClearAll()
		}
		data, err := s.SendReceive("some URL", "GET", fmt.Sprintf("index=%d", i))
		if err != nil {
			fmt.Printf("Failed: %v\n", err)
			continue
		}
		fmt.Printf("Received: %v\n", data)
	}
	// :
}
