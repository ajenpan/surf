package network

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
)

type WSServerOptions struct {
	ListenAddr       string
	HeatbeatInterval time.Duration

	OnConnPacket  FuncOnConnPacket
	OnConnStatus  FuncOnConnStatus
	OnConnAuth    FuncOnConnAuth
	OnConnAccpect func(r *http.Request) bool
}

type WSServerOption func(*WSServerOptions)

func NewWSServer(opts WSServerOptions) (*WSServer, error) {
	ret := &WSServer{
		WSServerOptions: opts,
		sockets:         make(map[string]*WSConn),
		die:             make(chan bool),
	}
	if ret.HeatbeatInterval < time.Duration(DefaultHeartbeatSec)*time.Second {
		ret.HeatbeatInterval = time.Duration(DefaultHeartbeatSec) * time.Second
	}
	h := &http.ServeMux{}
	h.HandleFunc("/", ret.ServeHTTP)
	ret.httpsvr = &http.Server{Addr: ret.ListenAddr, Handler: h}

	if ret.OnConnAccpect == nil {
		ret.upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	} else {
		ret.upgrader.CheckOrigin = ret.OnConnAccpect
	}

	return ret, nil
}

type WSServer struct {
	WSServerOptions
	mu       sync.RWMutex
	sockets  map[string]*WSConn
	die      chan bool
	httpsvr  *http.Server
	upgrader ws.Upgrader
}

func (s *WSServer) Start() error {
	ln, err := net.Listen("tcp", s.httpsvr.Addr)
	if err != nil {
		return err
	}

	go func() {
		if err := s.httpsvr.Serve(ln); err != nil {
			log.Error(err)
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

	conn := newWSConn(GenConnID(), nil, c, s.HeatbeatInterval*2)
	conn.status = Connectting

	deadline := time.Now().Add(s.HeatbeatInterval * 2)
	c.SetReadDeadline(deadline)
	c.SetWriteDeadline(deadline)

	pk, err := conn.readPacket()
	if err != nil {
		return
	}
	if pk.Meta.GetType() != PacketType_Inner || pk.Meta.GetSubFlag() != PacketInnerSubType_HandShakeStart || len(pk.GetBody()) != 0 {
		return
	}

	var us auth.User
	if s.OnConnAuth != nil {
		pk := NewHVPacket()
		pk.Meta.SetType(PacketType_Inner)
		pk.Meta.SetSubFlag(PacketInnerSubType_Cmd)
		pk.SetHead([]byte("auth"))
		if err := conn.writePacket(pk); err != nil {
			return
		}
		if pk, err := conn.readPacket(); err != nil {
			return
		} else {
			if us, err = s.OnConnAuth(pk.GetBody()); err != nil {
				pk.Meta.SetType(PacketType_Inner)
				pk.Meta.SetSubFlag(PacketInnerSubType_HandShakeFailed)
				pk.SetBody([]byte(err.Error()))
				return
			}
			conn.User = us
		}
	}

	pk.Meta.SetType(PacketType_Inner)
	pk.Meta.SetSubFlag(PacketInnerSubType_HandShakeFinish)
	pk.SetBody([]byte(conn.ConnID()))
	conn.writePacket(pk)

	// the connection is established here
	if s.OnConnStatus != nil {
		s.OnConnStatus(conn, true)
		defer s.OnConnStatus(conn, false)
	}

	conn.status = Connected

	go func() {
		defer conn.Close()
		err := conn.writeWork()
		if err != nil {
			fmt.Println(err)
		}
	}()

	go func() {
		defer conn.Close()
		err := conn.readWork()
		if err != nil {
			fmt.Println(err)
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
				if s.OnConnPacket != nil {
					s.OnConnPacket(conn, packet)
				}
			}
		}
	}
}

func (s *WSServer) Address() string {
	return s.httpsvr.Addr
}

func (s *WSServer) SocketCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sockets)
}
