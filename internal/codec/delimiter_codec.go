package codec

import (
	"bufio"
	"net"
	"sync"

	"github.com/mailru/easyjson"

	"go.chensl.me/minitunnel/internal/msg"
)

const _delim byte = 0

type DelimiterCodec struct {
	closed bool
	conn   net.Conn
	r      *bufio.Reader
	w      *bufio.Writer
	mu     sync.Mutex
}

func NewDelimiterCodec(conn net.Conn) Codec {
	return &DelimiterCodec{
		closed: false,
		conn:   conn,
		r:      bufio.NewReader(conn),
		w:      bufio.NewWriter(conn),
	}
}

func (c *DelimiterCodec) ReadCommand() (*msg.Command, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	b, err := c.r.ReadBytes(_delim)
	if err != nil {
		return nil, err
	}

	if b[len(b)-1] == _delim {
		b = b[:len(b)-1]
	}

	var cmd msg.Command
	err = easyjson.Unmarshal(b, &cmd)
	if err != nil {
		return nil, err
	}

	return &cmd, err
}

func (c *DelimiterCodec) WriteCommand(cmd *msg.Command) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	b, err := easyjson.Marshal(cmd)
	if err != nil {
		return err
	}

	_, err = c.w.Write(b)
	if err != nil {
		return err
	}

	err = c.w.WriteByte(_delim)
	if err != nil {
		return err
	}

	return c.w.Flush()
}

func (c *DelimiterCodec) Conn() net.Conn {
	return c.conn
}

func (c *DelimiterCodec) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}
