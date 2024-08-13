package network

import (
	"net"
	"sync"
	"time"

	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/log"
)

type TcpServerOptions struct {
	ListenAddr       string
	HeatbeatInterval time.Duration

	OnConnPacket  FuncOnConnPacket
	OnConnEnable  FuncOnConnEnable
	OnConnAuth    FuncOnConnAuth
	OnConnAccpect func(net.Conn) bool
}

type TcpServerOption func(*TcpServerOptions)

func NewTcpServer(opts TcpServerOptions) (*TcpServer, error) {
	ret := &TcpServer{
		opts:    opts,
		sockets: make(map[string]*TcpConn),
		die:     make(chan bool),
	}
	if ret.opts.HeatbeatInterval < time.Duration(DefaultMinTimeoutSec)*time.Second {
		ret.opts.HeatbeatInterval = time.Duration(DefaultTimeoutSec) * time.Second
	}

	listener, err := net.Listen("tcp", opts.ListenAddr)
	if err != nil {
		return nil, err
	}
	ret.listener = listener
	return ret, nil
}

type TcpServer struct {
	opts     TcpServerOptions
	mu       sync.RWMutex
	sockets  map[string]*TcpConn
	die      chan bool
	listener net.Listener
}

func (s *TcpServer) Stop() error {
	select {
	case <-s.die:
		return nil
	default:
		close(s.die)
	}
	s.listener.Close()
	return nil
}

func (s *TcpServer) Start() error {
	go func() {
		var tempDelay time.Duration = 0
		for {
			select {
			case <-s.die:
				return
			default:
				conn, err := s.listener.Accept()
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						if tempDelay == 0 {
							tempDelay = 5 * time.Millisecond
						} else {
							tempDelay *= 2
						}
						if max := 1 * time.Second; tempDelay > max {
							tempDelay = max
						}
						time.Sleep(tempDelay)
						continue
					}
					log.Error(err)
					return
				}
				tempDelay = 0
				go s.onAccept(conn)
			}
		}
	}()
	return nil
}

func (s *TcpServer) onAccept(c net.Conn) {
	defer c.Close()

	if s.opts.OnConnAccpect != nil {
		if !s.opts.OnConnAccpect(c) {
			return
		}
	}

	conn, err := s.handshake(c)
	if err != nil {
		return
	}

	conn.status = Connected

	// the connection is established here
	go func() {
		defer conn.Close()
		conn.writeWork()
	}()

	go func() {
		defer conn.Close()
		conn.readWork()
	}()

	if s.opts.OnConnEnable != nil {
		s.opts.OnConnEnable(conn, true)
		defer s.opts.OnConnEnable(conn, false)
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

			if packet.Meta.GetType() == (PacketType_Inner) {
				switch packet.Meta.GetSubFlag() {
				case uint8(PacketType_Inner_Heartbeat):
					conn.Send(packet)
				}
			} else {
				s.opts.OnConnPacket(conn, packet)
			}

		}
	}
}

func (s *TcpServer) handshake(conn net.Conn) (*TcpConn, error) {
	deadline := time.Now().Add(s.opts.HeatbeatInterval)
	conn.SetReadDeadline(deadline)
	conn.SetWriteDeadline(deadline)
	pk := NewHVPacket()
	_, err := pk.ReadFrom(conn)
	if err != nil {
		return nil, err
	}

	if pk.Meta.GetType() != PacketType_Inner || pk.Meta.GetSubFlag() != PacketType_Inner_HandShake || len(pk.GetBody()) != 0 {
		return nil, ErrInvalidPacket
	}

	var us auth.User
	if s.opts.OnConnAuth != nil {
		pk.Meta.SetSubFlag(PacketType_Inner_Cmd)
		pk.SetBody([]byte("auth"))
		if _, err = pk.WriteTo(conn); err != nil {
			return nil, err
		}
		if _, err = pk.ReadFrom(conn); err != nil {
			return nil, err
		}

		if pk.Meta.GetType() != PacketType_Inner || pk.Meta.GetSubFlag() != PacketType_Inner_CmdResult {
			return nil, ErrInvalidPacket
		}

		if us, err = s.opts.OnConnAuth(pk.GetBody()); err != nil {
			return nil, err
		}
	}

	socketid := GenConnID()

	pk.Meta.SetType(PacketType_Inner)
	pk.Meta.SetSubFlag(PacketType_Inner_HandShakeFinish)
	pk.SetBody([]byte(socketid))
	if _, err := pk.WriteTo(conn); err != nil {
		return nil, err
	}

	socket := &TcpConn{
		User:     us,
		id:       socketid,
		conn:     conn,
		timeOut:  s.opts.HeatbeatInterval,
		chClosed: make(chan struct{}),
		status:   Disconnected,
		chWrite:  make(chan *HVPacket, 10),
		chRead:   make(chan *HVPacket, 10),
	}

	return socket, nil
}

func (s *TcpServer) Address() net.Addr {
	return s.listener.Addr()
}

func (s *TcpServer) SocketCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sockets)
}
