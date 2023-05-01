package netutil

import (
	"errors"
	"io"
	"net"
	"syscall"

	"github.com/oklog/run"
)

func Proxy(c1 net.Conn, c2 net.Conn) error {
	var g run.Group

	g.Add(func() error {
		_, err := io.Copy(c1, c2)
		if err != nil && !isConnClosed(err) {
			return err
		}
		_ = c1.Close()
		return nil
	}, func(_ error) {
		_ = c2.Close()
	})

	g.Add(func() error {
		_, err := io.Copy(c2, c1)
		if err != nil && !isConnClosed(err) {
			return err
		}
		_ = c2.Close()
		return nil
	}, func(_ error) {
		_ = c1.Close()
	})

	return g.Run()
}

func isConnClosed(err error) bool {
	switch {
	case
		errors.Is(err, net.ErrClosed),
		errors.Is(err, io.EOF),
		errors.Is(err, syscall.EPIPE):
		return true
	default:
		return false
	}
}
