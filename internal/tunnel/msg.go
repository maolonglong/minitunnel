package tunnel

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"time"
)

var _zero time.Time

const _networkTimeout = 3 * time.Second

type msgCodec struct {
	inner net.Conn
	rw    *bufio.ReadWriter
}

func newMsgCodec(conn net.Conn) *msgCodec {
	return &msgCodec{
		inner: conn,
		rw: bufio.NewReadWriter(
			bufio.NewReader(conn),
			bufio.NewWriter(conn),
		),
	}
}

func (c *msgCodec) readMsg() ([]string, error) {
	return c._readMsg(_zero)
}

func (c *msgCodec) readMsgTimeout() ([]string, error) {
	return c._readMsg(time.Now().Add(_networkTimeout))
}

func (c *msgCodec) _readMsg(deadline time.Time) ([]string, error) {
	_ = c.inner.SetReadDeadline(deadline)

	l, err := c.rw.ReadByte()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, l)
	_, err = io.ReadFull(c.rw, buf)
	if err != nil {
		return nil, err
	}

	var msg []string
	if err := json.Unmarshal(buf, &msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (c msgCodec) writeMsg(msg ...string) error {
	b, err := json.Marshal(&msg)
	if err != nil {
		return err
	}

	if err := c.rw.WriteByte(byte(len(b))); err != nil {
		return err
	}

	_, err = c.rw.Write(b)
	if err != nil {
		return err
	}

	return c.rw.Flush()
}
