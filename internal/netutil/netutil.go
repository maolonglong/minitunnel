package netutil

import (
	"io"
	"net"
)

func Proxy(c1, c2 net.Conn) {
	defer c1.Close()
	defer c2.Close()

	errc := make(chan error, 1)

	go func() {
		_, err := io.Copy(c1, c2)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(c2, c1)
		errc <- err
	}()

	<-errc
}
