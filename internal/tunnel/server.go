package tunnel

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"

	"go.chensl.me/minitunnel/internal/netutil"
)

const _controlPort = 6101

type Server struct {
	conns sync.Map
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Run() error {
	ln, err := net.Listen("tcp", ":"+strconv.FormatInt(_controlPort, 10))
	if err != nil {
		return err
	}

	log.Info("server listening", "addr", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		logger := log.With("client_addr", conn.RemoteAddr())
		logger.Info("new client")
		go func() {
			if err := s.handleConn(conn, logger); err != nil {
				logger.Error("client exited with error", "err", err)
			} else {
				logger.Info("client exited")
			}
		}()
	}
}

func (s *Server) handleConn(conn net.Conn, logger *log.Logger) error {
	defer conn.Close()
	codec := newMsgCodec(conn)

	msg, err := codec.readMsgTimeout()
	if err != nil {
		return err
	}
	switch msg[0] {
	case "hello":
		addr, _ := net.ResolveTCPAddr("tcp", ":0")
		ln, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}

		_, port, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			return err
		}

		if err := codec.writeMsg("hello", port); err != nil {
			return err
		}

		for {
			if err := codec.writeMsg("heartbeat"); err != nil {
				return nil
			}

			_ = ln.SetDeadline(time.Now().Add(500 * time.Millisecond))
			conn2, err := ln.Accept()
			if err != nil {
				var operr *net.OpError
				if errors.As(err, &operr) && operr.Timeout() {
					continue
				}
				return err
			}

			id := uuid.NewString()
			s.conns.Store(id, conn2)

			time.AfterFunc(10*time.Second, func() {
				conn2, ok := s.conns.LoadAndDelete(id)
				if ok {
					_ = conn2.(net.Conn).Close()
					logger.Warn("removed stale connection", "id", id)
				}
			})

			if err := codec.writeMsg("connection", id); err != nil {
				return err
			}
		}

	case "accept":
		if len(msg) != 2 {
			return errors.New("invalid command")
		}
		logger.Info("forwarding connection", "id", msg[1])
		v, ok := s.conns.LoadAndDelete(msg[1])
		if !ok {
			logger.Warn("missing connection", "id", msg[1])
			return nil
		}
		conn2 := v.(net.Conn)
		netutil.Proxy(conn, conn2)
		return nil

	default:
		return errors.New("invalid command")
	}
}
