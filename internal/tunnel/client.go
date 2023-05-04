package tunnel

import (
	"errors"
	"net"
	"strconv"

	"github.com/charmbracelet/log"

	"go.chensl.me/minitunnel/internal/netutil"
)

type Client struct {
	conn      net.Conn
	codec     *msgCodec
	srvAddr   string
	localAddr string
}

func NewClient(
	to string,
	localHost string,
	localPort int,
) (*Client, error) {
	srvAddr := net.JoinHostPort(to, strconv.FormatInt(_controlPort, 10))
	conn, err := net.DialTimeout("tcp", srvAddr, _networkTimeout)
	if err != nil {
		return nil, err
	}

	codec := newMsgCodec(conn)
	if err := codec.writeMsg("hello"); err != nil {
		return nil, err
	}

	msg, err := codec.readMsgTimeout()
	if err != nil {
		return nil, err
	}

	switch msg[0] {
	case "hello":
		if len(msg) != 2 {
			return nil, errors.New("invalid command")
		}
		log.Infof("mt server listening at: tcp://%v", net.JoinHostPort(to, msg[1]))
	default:
		return nil, errors.New("unexpected initial non-hello message")
	}

	return &Client{
		conn:    conn,
		codec:   codec,
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
		msg, err := c.codec.readMsg()
		if err != nil {
			return err
		}

		switch msg[0] {
		case "hello":
			log.Warn("unexpected hello")
		case "heartbeat":
		case "connection":
			if len(msg) != 2 {
				return errors.New("invalid command")
			}
			go func() {
				if err := c.handleConn(msg[1]); err != nil {
					log.Error("connection exited with error", "id", msg[1], "err", err)
				} else {
					log.Info("connection exited", "id", msg[1])
				}
			}()
		}
	}
}

func (c *Client) handleConn(id string) error {
	localConn, err := net.DialTimeout("tcp", c.localAddr, _networkTimeout)
	if err != nil {
		return err
	}

	remoteConn, err := net.DialTimeout("tcp", c.srvAddr, _networkTimeout)
	if err != nil {
		_ = localConn.Close()
		return err
	}

	if err := newMsgCodec(remoteConn).writeMsg("accept", id); err != nil {
		return err
	}

	return netutil.Proxy(localConn, remoteConn)
}
