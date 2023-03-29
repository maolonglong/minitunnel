package codec

import (
	"io"
	"net"

	"go.chensl.me/minitunnel/internal/msg"
)

type Codec interface {
	io.Closer

	ReadCommand() (*msg.Command, error)
	WriteCommand(cmd *msg.Command) error
	Conn() net.Conn
}
