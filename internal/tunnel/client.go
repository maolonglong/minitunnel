package tunnel

import (
	"errors"
	"net"
	"strconv"

	"github.com/charmbracelet/log"

	"go.chensl.me/minitunnel/internal/codec"
	"go.chensl.me/minitunnel/internal/msg"
	"go.chensl.me/minitunnel/internal/netutil"
)

type Client struct {
	conn      codec.Codec
	srvAddr   string
	localAddr string
}

func NewClient(
	to string,
	localHost string,
	localPort int,
) (*Client, error) {
	srvAddr := net.JoinHostPort(to, strconv.FormatInt(_controlPort, 10))
	raw, err := net.Dial("tcp", srvAddr)
	if err != nil {
		return nil, err
	}

	conn := codec.NewDelimiterCodec(raw)
	if err := conn.WriteCommand(msg.C("hello")); err != nil {
		return nil, err
	}

	cmd, err := conn.ReadCommand()
	if err != nil {
		return nil, err
	}

	switch cmd.Name {
	case "hello":
		if len(cmd.Args) != 1 {
			return nil, errors.New("invalid command")
		}
		log.Infof("listening at: tcp://%v", net.JoinHostPort(to, cmd.Args[0]))
	default:
		return nil, errors.New("unexpected initial non-hello message")
	}

	return &Client{
		conn:    conn,
		srvAddr: srvAddr,
		localAddr: net.JoinHostPort(
			localHost,
			strconv.FormatInt(int64(localPort), 10),
		),
	}, nil
}

func (c *Client) Run() error {
	defer c.conn.Close()
	for {
		cmd, err := c.conn.ReadCommand()
		if err != nil {
			return err
		}

		switch cmd.Name {
		case "hello":
			log.Warn("unexpected hello")
		case "heartbeat":
		case "connection":
			if len(cmd.Args) != 1 {
				return errors.New("invalid command")
			}
			go func() {
				if err := c.handleConn(cmd.Args[0]); err != nil {
					log.Error(
						"connection exited with error",
						"id",
						cmd.Args[0],
						"err",
						err,
					)
				} else {
					log.Info("connection exited", "id", cmd.Args[0])
				}
			}()
		}
	}
}

func (c *Client) handleConn(id string) error {
	localConn, err := net.Dial("tcp", c.localAddr)
	if err != nil {
		return err
	}

	remoteConn, err := net.Dial("tcp", c.srvAddr)
	if err != nil {
		return err
	}

	if err := codec.NewDelimiterCodec(remoteConn).WriteCommand(msg.C("accept", id)); err != nil {
		return err
	}

	return netutil.Proxy(localConn, remoteConn)
}
