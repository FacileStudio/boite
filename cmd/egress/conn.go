package egress

import (
	"errors"
	"net"
	"sync"
)

var errServed = errors.New("connection finished")

type oneShot struct {
	once sync.Once
	conn chan net.Conn
	done chan struct{}
}

func newOneShot(c net.Conn) *oneShot {
	o := &oneShot{conn: make(chan net.Conn, 1), done: make(chan struct{})}
	o.conn <- &notifyConn{Conn: c, done: o.done}
	return o
}

func (o *oneShot) Accept() (net.Conn, error) {
	select {
	case c := <-o.conn:
		return c, nil
	case <-o.done:
		return nil, errServed
	}
}

func (o *oneShot) Close() error   { return nil }
func (o *oneShot) Addr() net.Addr { return pipeAddr{} }

type notifyConn struct {
	net.Conn
	done chan struct{}
	once sync.Once
}

func (c *notifyConn) Close() error {
	c.once.Do(func() { close(c.done) })
	return c.Conn.Close()
}
