package egress

import (
	"net"
	"os"
)

type pipeAddr struct{}

func (pipeAddr) Network() string { return "stdio" }
func (pipeAddr) String() string  { return "stdio" }

type pipeConn struct {
	*os.File
	w *os.File
}

func (p *pipeConn) Write(b []byte) (int, error) { return p.w.Write(b) }
func (p *pipeConn) Close() error                { p.File.Close(); return p.w.Close() }
func (p *pipeConn) LocalAddr() net.Addr         { return pipeAddr{} }
func (p *pipeConn) RemoteAddr() net.Addr        { return pipeAddr{} }
