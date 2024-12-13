package network

import (
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
)

type WSServerOptions struct {
	ListenAddr       string
	HeatbeatInterval time.Duration

	OnConnPacket  FuncOnConnPacket
	OnConnEnable  FuncOnConnEnable
	OnConnAuth    FuncOnConnAuth
	OnConnAccpect func(r *http.Request) bool
	Log           *slog.Logger
}

type WSServerOption func(*WSServerOptions)

func NewWSServer(opts WSServerOptions) (*WSServer, error) {
	ret := &WSServer{
		opts:    opts,
		sockets: make(map[string]*WSConn),
		die:     make(chan bool),
	}
	if ret.opts.HeatbeatInterval < time.Duration(DefaultHeartbeatSec)*time.Second {
		ret.opts.HeatbeatInterval = time.Duration(DefaultHeartbeatSec) * time.Second
	}
	h := &http.ServeMux{}
	h.HandleFunc("/", ret.ServeHTTP)
	ln, err := net.Listen("tcp", ret.opts.ListenAddr)
	if err != nil {
		return nil, err
	}
	ret.listener = ln
	ret.addr = ln.Addr()
	ret.httpsvr = &http.Server{Addr: ret.addr.String(), Handler: h}
	if ret.opts.OnConnAccpect == nil {
		ret.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	} else {
		ret.upgrader.CheckOrigin = ret.opts.OnConnAccpect
	}
	if ret.opts.Log == nil {
		ret.opts.Log = slog.Default().With("module", "ws")
	}
	return ret, nil
}

type WSServer struct {
	opts     WSServerOptions
	mu       sync.RWMutex
	sockets  map[string]*WSConn
	die      chan bool
	httpsvr  *http.Server
	upgrader ws.Upgrader

	listener net.Listener
	addr     net.Addr
}

func (s *WSServer) log() *slog.Logger {
	return s.opts.Log
}

func (s *WSServer) Start() error {
	go func() {
		if err := s.httpsvr.Serve(s.listener); err != nil {
			s.log().Error("Serve err", "err", err)
		}
	}()
	return nil
}

func (s *WSServer) Stop() error {
	select {
	case <-s.die:
		return nil
	default:
		close(s.die)
	}
	s.httpsvr.Close()
	return nil
}

func (s *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	conn := newWSConn(GenConnID(), nil, c, s.opts.HeatbeatInterval*2)
	conn.status = Connectting

	deadline := time.Now().Add(s.opts.HeatbeatInterval * 2)
	c.SetReadDeadline(deadline)
	c.SetWriteDeadline(deadline)

	pk, err := conn.ReadPacket()
	if err != nil {
		return
	}

	if pk.Meta.GetType() != PacketType_Inner || pk.Meta.GetSubFlag() != PacketInnerSubType_HandShakeStart || len(pk.GetBody()) != 0 {
		return
	}

	var uInfo User

	if s.opts.OnConnAuth != nil {
		pk := NewHVPacket()
		pk.Meta.SetType(PacketType_Inner)
		pk.Meta.SetSubFlag(PacketInnerSubType_Cmd)
		pk.SetHead([]byte("auth"))
		if err := conn.WritePacket(pk); err != nil {
			return
		}

		pk, err = conn.ReadPacket()
		if err != nil {
			return
		}
		if uInfo, err = s.opts.OnConnAuth(conn, pk.GetBody()); err != nil {
			pk.Meta.SetType(PacketType_Inner)
			pk.Meta.SetSubFlag(PacketInnerSubType_HandShakeFailed)
			pk.SetHead([]byte("auth"))
			pk.SetBody([]byte(err.Error()))
			conn.WritePacket(pk)
			time.Sleep(1 * time.Second)
			return
		}
		conn.userInfo.fromUser(uInfo)
	}

	pk.Meta.SetType(PacketType_Inner)
	pk.Meta.SetSubFlag(PacketInnerSubType_HandShakeFinish)
	pk.SetBody([]byte(conn.ConnId()))
	conn.WritePacket(pk)

	// the connection is established here
	if s.opts.OnConnEnable != nil {
		s.opts.OnConnEnable(conn, true)
		defer s.opts.OnConnEnable(conn, false)
	}

	conn.status = Connected

	go func() {
		defer conn.Close()
		err := conn.writeWork()
		if err != nil {
			s.log().Error("writeWork err", "err", err)
		}
	}()

	go func() {
		defer conn.Close()
		err := conn.readWork()
		if err != nil {
			if err != io.EOF {
				s.log().Error("readWork err", "err", err)
			}
		}
	}()

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
			if packet.Meta.GetType() == PacketType_Inner {
				switch packet.Meta.GetSubFlag() {
				case PacketInnerSubType_Heartbeat:
					body := make([]byte, 8)
					binary.LittleEndian.PutUint64(body, uint64(time.Now().UnixMilli()))
					packet.SetBody(body)
					conn.Send(packet)
				}
			} else {
				if s.opts.OnConnPacket != nil {
					s.opts.OnConnPacket(conn, packet)
				}
			}
		}
	}
}

func (s *WSServer) Address() string {
	return s.addr.String()
}

func (s *WSServer) SocketCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sockets)
}
