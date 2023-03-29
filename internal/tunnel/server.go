package tunnel

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/oklog/run"

	"go.chensl.me/minitunnel/internal/codec"
	"go.chensl.me/minitunnel/internal/msg"
	"go.chensl.me/minitunnel/internal/netutil"
)

const _controlPort = 6101

type Server struct {
	heartbeatInterval time.Duration
	conns             sync.Map
}

func NewServer(controlPort int, heartbeatInterval time.Duration) *Server {
	return &Server{
		heartbeatInterval: heartbeatInterval,
	}
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
		go func() {
			if err := s.handleConn(conn); err != nil {
				log.Error("connection exited with error", "err", err)
			} else {
				log.Info("connection exited")
			}
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) error {
	c := codec.NewDelimiterCodec(conn)
	defer c.Close()

	cmd, err := c.ReadCommand()
	if err != nil {
		return err
	}
	switch cmd.Name {
	case "hello":
		log.Info("new client")

		addr, _ := net.ResolveTCPAddr("tcp", ":0")
		ln, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}

		_, port, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			return err
		}

		if err := c.WriteCommand(msg.C("hello", port)); err != nil {
			return err
		}

		var g run.Group

		exitCh := make(chan struct{})
		g.Add(func() error {
			timer := time.NewTicker(s.heartbeatInterval)
			for {
				select {
				case <-timer.C:
					if err := c.WriteCommand(msg.C("heartbeat")); err != nil {
						return err
					}
				case <-exitCh:
					timer.Stop()
					return nil
				}
			}
		}, func(_ error) {
			close(exitCh)
		})

		exitCh2 := make(chan struct{})
		g.Add(func() error {
			for {
				select {
				case <-exitCh2:
					return nil
				default:
				}
				_ = ln.SetDeadline(time.Now().Add(3 * time.Second))
				conn2, err := ln.Accept()
				if err != nil {
					if operr, ok := err.(*net.OpError); ok && operr.Timeout() {
						continue
					}
					return err
				}

				id := uuid.NewString()
				s.conns.Store(id, conn2)

				time.AfterFunc(10*time.Second, func() {
					_, ok := s.conns.LoadAndDelete(id)
					if ok {
						log.Warn("removed stale connection", "id", id)
					}
				})

				if err := c.WriteCommand(msg.C("connection", id)); err != nil {
					return err
				}
			}
		}, func(_ error) {
			close(exitCh2)
		})

		return g.Run()

	case "accept":
		if len(cmd.Args) != 1 {
			return errors.New("invalid command")
		}
		log.Info("forwarding connection", "id", cmd.Args[0])
		v, ok := s.conns.LoadAndDelete(cmd.Args[0])
		if !ok {
			log.Warn("missing connection", "id", cmd.Args[0])
			return nil
		}
		conn2 := v.(net.Conn)
		return netutil.Proxy(conn, conn2)

	default:
		return errors.New("invalid command")
	}
}
