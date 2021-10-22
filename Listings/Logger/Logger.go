package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Logger interface {
	Error(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Info(format string, args ...interface{})
	Debug(format string, args ...interface{})
	Trace(format string, args ...interface{})
}

type DefaultLoggerImpl struct {
	logger *log.Logger
}

func (l *DefaultLoggerImpl) output(level, format string, args ...interface{}) {
	l.logger.Printf(fmt.Sprintf("%s %s %s\n", getCallerDetails(2, "-"), level, fmt.Sprintf(format, args...)))
}
func (l *DefaultLoggerImpl) Error(format string, args ...interface{}) {
	l.output("ERROR", format, args...)
}
func (l *DefaultLoggerImpl) Warn(format string, args ...interface{}) {
	l.output("WARN ", format, args...)
}
func (l *DefaultLoggerImpl) Info(format string, args ...interface{}) {
	l.output("INFO ", format, args...)
}
func (l *DefaultLoggerImpl) Debug(format string, args ...interface{}) {
	l.output("DEBUG", format, args...)
}
func (l *DefaultLoggerImpl) Trace(format string, args ...interface{}) {
	l.output("TRACE", format, args...)
}

var DefaultLogger *DefaultLoggerImpl

func init() {
	DefaultLogger = &DefaultLoggerImpl{}
	DefaultLogger.logger = log.New(os.Stdout, "GoBook ", log.LstdFlags|log.Lmicroseconds|log.LUTC)
}

// get details about the caller.
func getCallerDetails(level int, lead string) string {
	level++
	if pc, file, line, ok := runtime.Caller(level); ok {
		file = getName(file)
		goId := getGID()
		xlineCount := atomic.AddUint64(&lineCount, 1)
		lead = fmt.Sprintf("%7d go%-5d %08X %-40v@%4v", xlineCount, goId, pc, file, line)
	}
	return lead
}

var lineCount uint64

// Get the current goroutine id.
func getGID() (n uint64) {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ = strconv.ParseUint(string(b), 10, 64)
	return
}

//  Get the file name part.
func getName(file string) string {
	posn := strings.Index(file, src)
	if posn >= 0 {
		file = file[posn+len(src):]
		if strings.HasSuffix(file, goExtension) {
			file = file[0 : len(file)-len(goExtension)]
		}
	}
	return file
}

const src = "/src/"
const goExtension = ".go"

func main() {
	DefaultLogger.Trace("Hello %s!", "World")
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			DefaultLogger.Info("Hello from goroutine %d!", id)
			time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
			DefaultLogger.Info("Goodbye from goroutine %d!", id)
		}(i)
	}
	wg.Wait()
	DefaultLogger.Trace("Goodbye %s!", "World")
}
