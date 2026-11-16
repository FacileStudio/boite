package qemu

import (
	"io"
	"net"
	"time"
)

// sshBannerReady reports whether addr answers with an SSH banner. A bare TCP
// connect is not enough: passt and slirp accept the forwarded port before the
// guest's sshd listens, then reset the connection.
func sshBannerReady(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 4)
	_, err = io.ReadFull(conn, buf)
	return err == nil && string(buf) == "SSH-"
}
