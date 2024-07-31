package network

import (
	"net/http"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"

	"github.com/ajenpan/surf/core/auth"
)

type WSServerOptions struct {
	ListenAddr       string
	HeatbeatInterval time.Duration

	OnConnPacket  FuncOnConnPacket
	OnConnEnable  FuncOnConnEnable
	OnConnAuth    FuncOnConnAuth
	OnConnAccpect func(r *http.Request) bool
}

type WSServerOption func(*WSServerOptions)

func NewWSServer(opts WSServerOptions) *WSServer {
	ret := &WSServer{
		WSServerOptions: opts,
		sockets:         make(map[string]*WSConn),
		die:             make(chan bool),
	}
	if ret.HeatbeatInterval < time.Duration(DefaultMinTimeoutSec)*time.Second {
		ret.HeatbeatInterval = time.Duration(DefaultTimeoutSec) * time.Second
	}
	h := &http.ServeMux{}
	h.HandleFunc("/", ret.ServeHTTP)
	ret.listener = &http.Server{Addr: ret.ListenAddr, Handler: h}

	if ret.OnConnAccpect == nil {
		ret.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	} else {
		ret.upgrader.CheckOrigin = ret.OnConnAccpect
	}

	return ret
}

type WSServer struct {
	WSServerOptions
	mu       sync.RWMutex
	sockets  map[string]*WSConn
	die      chan bool
	listener *http.Server

	upgrader ws.Upgrader
}

func (s *WSServer) Start() error {
	go s.listener.ListenAndServe()
	return nil
}

func (s *WSServer) Stop() error {
	select {
	case <-s.die:
		return nil
	default:
		close(s.die)
	}
	s.listener.Close()
	return nil
}

func (s *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	conn := &WSConn{
		timeOut:  s.HeatbeatInterval,
		imp:      c,
		status:   Connectting,
		id:       GenConnID(),
		chClosed: make(chan struct{}),
		chWrite:  make(chan *HVPacket, 10),
		chRead:   make(chan *HVPacket, 10),
	}

	deadline := time.Now().Add(s.HeatbeatInterval * 2)
	c.SetReadDeadline(deadline)
	c.SetWriteDeadline(deadline)

	pk, err := conn.readPacket()
	if err != nil {
		return
	}
	if pk.GetFlag() != hvpFlagInit || pk.GetSubFlag() != hvpSubFlagHandShake || len(pk.GetBody()) != 0 {
		return
	}

	var us auth.User
	if s.OnConnAuth != nil {
		pk := NewHVPacket()
		pk.SetFlag(hvpFlagInit)
		pk.SetSubFlag(hvpSubFlagCmd)
		pk.SetBody([]byte("auth"))
		if err := conn.writePacket(pk); err != nil {
			return
		}
		if pk, err := conn.readPacket(); err != nil {
			return
		} else {
			if us, err = s.OnConnAuth(pk.GetBody()); err != nil {
				return
			}
			conn.User = us
		}
	}

	pk.SetFlag(hvpFlagInit)
	pk.SetSubFlag(hvpSubFlagHandShakeFinish)
	pk.SetBody([]byte(conn.ConnID()))
	conn.writePacket(pk)

	// the connection is established here
	go func() {
		defer conn.Close()
		conn.writeWork()
	}()

	go func() {
		defer conn.Close()
		conn.readWork()
	}()

	if s.OnConnEnable != nil {
		s.OnConnEnable(conn, true)
		defer s.OnConnEnable(conn, false)
	}

	for {
		select {
		case <-conn.chClosed:
			return
		case <-s.die:
			conn.Close()
			return
		case packet, ok := <-conn.chRead:
			if !ok {
				return
			}

			if packet.GetFlag() == hvpFlagInit {
				switch packet.GetSubFlag() {
				case hvpSubFlagHeartbeat:
					conn.Send(packet)

				}
			} else {
				if s.OnConnPacket != nil {
					s.OnConnPacket(conn, packet)
				}
			}
		}
	}
}

func (s *WSServer) Address() string {
	return s.listener.Addr
}

func (s *WSServer) SocketCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sockets)
}
