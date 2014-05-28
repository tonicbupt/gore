package gore

import (
	"bufio"
	"net"
	"sync"
	"time"
)

const (
	connStateNotConnected = iota
	connStateConnected
	connStateReconnecting
)

// Conn holds a persistent connection to a redis server
type Conn struct {
	address        string
	tcpConn        net.Conn
	state          int
	mutex          sync.Mutex
	writeMutex     sync.Mutex
	readMutex      sync.Mutex
	rb             *bufio.Reader
	wb             *bufio.Writer
	RequestTimeout time.Duration
}

// Dial opens a TCP connection with a redis server.
func Dial(address string, timeout time.Duration) (*Conn, error) {
	conn := &Conn{
		RequestTimeout: 10 * time.Second,
	}
	err := conn.connect(address, timeout)
	return conn, err
}

// Close closes the connection with redis server
func (c *Conn) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.state == connStateNotConnected {
		return nil
	}
	c.state = connStateNotConnected
	return c.tcpConn.Close()
}

func (c *Conn) connect(address string, timeout time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.state == connStateConnected || c.state == connStateReconnecting {
		return nil
	}
	var err error
	c.address = address
	if timeout == 0 {
		c.tcpConn, err = net.Dial("tcp", address)
	} else {
		c.tcpConn, err = net.DialTimeout("tcp", address, timeout)
	}
	if err == nil {
		c.state = connStateConnected
		c.rb = bufio.NewReader(c.tcpConn)
		c.wb = bufio.NewWriter(c.tcpConn)
	}
	return err
}

func (c *Conn) reconnect() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.state == connStateReconnecting {
		return
	}
	c.tcpConn.Close()
	c.state = connStateReconnecting
	go c.doReconnect()
}

func (c *Conn) doReconnect() {
	for {
		if err := c.connect(c.address, 0); err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
}
