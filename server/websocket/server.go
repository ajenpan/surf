package websocket

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gobwas/ws"
)

func NewServer(opts ServerOptions) *HttpServer {
	ret := &HttpServer{
		opts: opts,
		die:  make(chan bool),
		httpsvr: &http.Server{
			Addr: opts.Address,
		},
	}

	if ret.opts.NewIDFunc == nil {
		ret.opts.NewIDFunc = nextID
	}
	return ret
}

type ServerOptions struct {
	Address          string
	HeatbeatInterval time.Duration
	OnMessage        OnMessageFunc
	OnConn           OnConnStatFunc
	NewIDFunc        NewIDFunc
}

type HttpServer struct {
	opts ServerOptions

	die     chan bool
	httpsvr *http.Server
}

func (s *HttpServer) Start() error {
	s.httpsvr.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			log.Println(err)
			return
		}

		socket := NewSocket(s.opts.NewIDFunc(), conn)
		defer socket.Close()

		if s.opts.OnConn != nil {
			s.opts.OnConn(socket, Connected)
			defer s.opts.OnConn(socket, Disconnected)
		}

		go socket.writeWork()

		for {
			p := &Packet{}
			err := socket.readPacket(p)
			if err != nil {
				break
			}

			if s.opts.OnMessage != nil {
				s.opts.OnMessage(socket, p)
			}

		}
	})

	err := s.httpsvr.ListenAndServe()
	return err
}

func (s *HttpServer) Stop() error {
	select {
	case <-s.die:
	default:
		close(s.die)
		s.httpsvr.Shutdown(context.Background())
	}
	return nil
}
