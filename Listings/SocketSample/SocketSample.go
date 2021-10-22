package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var xlog = log.New(os.Stderr, "", log.Ltime+log.Lmicroseconds)

// Get the current goroutine id.
func getGID() (n uint64) {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ = strconv.ParseUint(string(b), 10, 64)
	return
}

func SocketClientGo(wg *sync.WaitGroup) {
	defer wg.Done()
	gid := getGID()
	cmds := []string{TODCommand, SayingCommand}
	max := 10

	var xwg sync.WaitGroup
	for i := 0; i < max; i++ {
		xwg.Add(1)
		go func(index, max int) {
			defer xwg.Done()
			time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
			sc := newSocketClient("127.0.0.1", 8080)
			xlog.Printf("%5d SocketClientGo request %d of %d\n", gid, index, max)
			resp, err := sc.GetCmd(cmds[rand.Intn(len(cmds))])
			if err != nil {
				xlog.Printf("%5d SocketClientGo failed: %v\n", gid, err)
				return
			}
			xlog.Printf("%5d SocketClientGo response: %s\n", gid, resp)
		}(i+1, max)
	}
	xwg.Wait()
}

// allowed commands
const (
	TODCommand    = "TOD"
	SayingCommand = "Saying"
)

var delim = byte('~')

// some saying to return
var sayings = make([]string, 0, 100)

func init() {
	sayings = append(sayings,
		`Now is the time...`,
		`I'm busy.`,
		`I pity the fool that tries to stop me!`,
		`Out wit; Out play; Out last!`,
		`It's beginning to look like TBD!`,
		)
}

// a Server
type SocketServer struct {
	Accepting bool
}

func NewSocketServer() (ss *SocketServer) {
	ss = &SocketServer{}
	ss.Accepting = true
	return
}

// Accept connection until told to stop.
func (ss *SocketServer) AcceptConnections(port int) (err error) {
	gid := getGID()
	xlog.Printf("%5d accept listening on port: %d\n", gid, port)
	listen, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return
	}
	for ss.Accepting {
		conn, err := listen.Accept()
		if err != nil {
			xlog.Printf("%5d accept failed: %v\n", gid, err)
			continue
		}
		xlog.Printf("%5d accepted connection: %#v\n", gid, conn)
		go ss.handleConnectionGo(conn)
	}
	return
}

var nesting int32

// Process each connection.
// Only one command per connection.
func (ss *SocketServer) handleConnectionGo(c net.Conn) {
	defer c.Close()
	nest := atomic.AddInt32(&nesting, 1)
	defer func() {
		atomic.AddInt32(&nesting, -1)
	}()
	gid := getGID()
	data := make([]byte, 0, 1000)
	err := readData(c, &data, delim, cap(data))
	if err != nil {
		xlog.Printf("%5d handleConnection failed: %v\n", gid, err)
		return
	}
	cmd := string(data)
	xlog.Printf("%5d handleConnection request: %s, nest: %d, conn: %#v\n", gid, cmd, nest, c)
	if strings.HasSuffix(cmd, string(delim)) {
		cmd = cmd[0 : len(cmd)-1]
	}
	xlog.Printf("%5d received command: %s\n", gid, cmd)
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond) // make request take a while
	var out string
	switch cmd {
	case SayingCommand:
		out = sayings[rand.Intn(len(sayings))]
	case TODCommand:
		out = fmt.Sprintf("%s", time.Now())
	default:
		xlog.Printf("%5d handleConnection unknown request: %s\n", gid, cmd)
		out = "bad command: " + cmd
	}
	_, err = writeData(c, []byte(out+string(delim)))
	if err != nil {
		xlog.Printf("%5d %s failed: %v\n", gid, cmd, err)
	}

}

// a Client
type SocketClient struct {
	Address    string
	Port       int
	Connection net.Conn
}

func newSocketClient(address string, port int) (sc *SocketClient) {
	sc = &SocketClient{}
	sc.Address = address
	sc.Port = port
	return
}
func (sc *SocketClient) Connect() (err error) {
	gid := getGID()
	xlog.Printf("%5d attempting connection: %s:%d\n", gid, sc.Address, sc.Port)
	sc.Connection, err = net.Dial("tcp", fmt.Sprintf("%s:%d", sc.Address, sc.Port))
	if err != nil {
		return
	}
	xlog.Printf("%5d made connection: %#v\n", gid, sc.Connection)
	return
}
func (sc *SocketClient) SendCommand(cmd string) (err error) {
	gid := getGID()
	c, err := sc.Connection.Write([]byte(cmd + string(delim)))
	if err != nil {
		return
	}
	xlog.Printf("%5d sent command: %s, count=%d\n", gid, cmd, c)
	return
}
func (sc *SocketClient) ReadResponse(data *[]byte, max int) (err error) {
	err = readData(sc.Connection, data, delim, 1000)
	return
}

// send command and get response.
func (sc *SocketClient) GetCmd(cmd string) (tod string, err error) {
	err = sc.Connect()
	if err != nil {
		return
	}
	defer sc.Connection.Close()
	err = sc.SendCommand(cmd)
	data := make([]byte, 0, 1000)
	err = readData(sc.Connection, &data, delim, cap(data))
	if err != nil {
		return
	}
	tod = string(data)
	return
}

func readData(c net.Conn, data *[]byte, delim byte, max int) (err error) {
	for {
		xb := make([]byte, 1, 1)
		c, xerr := c.Read(xb)
		if xerr != nil {
			err = xerr
			return
		}
		if c > 0 {
			if len(*data) > max {
				break
			}
			b := xb[0]
			*data = append(*data, b)
			if b == delim {
				break
			}
		}
	}
	return
}
func writeData(c net.Conn, data []byte) (count int, err error) {
	count, err = c.Write(data)
	return
}

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go SocketClientGo(&wg)
	ss := NewSocketServer()
	go func() {
		gid := getGID()
		err := ss.AcceptConnections(8080)
		if err != nil {
			xlog.Printf("%5d testSocketServer accept failed: %v\n", gid, err)
			return
		}
	}()
	wg.Wait()
	ss.Accepting = false
}
