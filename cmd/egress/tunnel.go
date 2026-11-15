package egress

import (
	"io"
	"net"
)

// pipe copies both ways between a hijacked client connection and the upstream
// until both directions finish.
func pipe(client net.Conn, clientIn io.Reader, up net.Conn) {
	done := make(chan struct{}, 2)
	go func() { io.Copy(up, clientIn); closeWrite(up); done <- struct{}{} }()
	go func() { io.Copy(client, up); closeWrite(client); done <- struct{}{} }()
	<-done
	<-done
}

func closeWrite(c net.Conn) {
	if cw, ok := c.(interface{ CloseWrite() error }); ok {
		cw.CloseWrite()
		return
	}
	c.Close()
}
