package tcp

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var socketIdx uint64

func nextID() string {
	idx := atomic.AddUint64(&socketIdx, 1)
	if idx == 0 {
		idx = atomic.AddUint64(&socketIdx, 1)
	}
	return fmt.Sprintf("tcp_%v", idx)
}

type ServerOptions struct {
	Address          string
	HeatbeatInterval time.Duration
	OnMessage        OnMessageFunc
	OnConn           OnConnStatFunc
	NewIDFunc        NewIDFunc
	AuthTokenChecker func(string) (*UserInfo, error)
}

type ServerOption func(*ServerOptions)

func NewServer(opts ServerOptions) (*Server, error) {
	ret := &Server{
		opts:    opts,
		sockets: make(map[string]*Socket),
		die:     make(chan bool),
	}
	listener, err := net.Listen("tcp", opts.Address)
	if err != nil {
		return nil, err
	}

	ret.listener = listener

	if ret.opts.OnMessage == nil {
		ret.opts.OnMessage = func(s *Socket, p *THVPacket) {}
	}
	if ret.opts.OnConn == nil {
		ret.opts.OnConn = func(s *Socket, enable bool) {}
	}
	if ret.opts.HeatbeatInterval == 0 {
		ret.opts.HeatbeatInterval = time.Duration(DefaultTimeoutSec) * time.Second
	}
	if ret.opts.NewIDFunc == nil {
		ret.opts.NewIDFunc = nextID
	}
	return ret, nil
}

type Server struct {
	opts     ServerOptions
	mu       sync.RWMutex
	sockets  map[string]*Socket
	die      chan bool
	wgConns  sync.WaitGroup
	listener net.Listener
}

func (s *Server) Stop() error {
	select {
	case <-s.die:
	default:
		close(s.die)
	}
	s.wgConns.Wait()
	s.listener.Close()
	return nil
}

func (s *Server) Start() error {
	var tempDelay time.Duration = 0
	for {
		select {
		case <-s.die:
			return nil
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
				return err
			}
			tempDelay = 0

			socket := NewSocket(conn, SocketOptions{
				ID:      s.opts.NewIDFunc(),
				Timeout: s.opts.HeatbeatInterval,
			})

			go s.onAccept(socket)
		}
	}
}

func (n *Server) onAccept(socket *Socket) {
	n.wgConns.Add(1)
	defer n.wgConns.Done()
	defer socket.Close()

	//read ack
	ack := NewEmptyTHVPacket()
	if err := socket.readPacket(ack); err != nil {
		return
	}

	if ack.GetType() != PacketTypeAck {
		return
	}

	// set socket id
	ack.SetBody([]byte(socket.ID()))
	if err := socket.writePacket(ack); err != nil {
		return
	}

	// auth packet
	ack.Reset()
	if err := socket.readPacket(ack); err != nil || ack.GetType() != PacketTypeAuth {
		return
	}

	// auth token
	if n.opts.AuthTokenChecker != nil {
		var err error
		if socket.UserInfo, err = n.opts.AuthTokenChecker(string(ack.GetBody())); err != nil {
			ack.Body = []uint8(err.Error())
			socket.writePacket(ack)
			return
		}
	}

	ack.SetBody([]byte("ok"))
	if err := socket.writePacket(ack); err != nil {
		return
	}

	// the connection is established
	go socket.writeWork()
	n.storeSocket(socket)
	defer n.removeSocket(socket)

	if n.opts.OnConn != nil {
		n.opts.OnConn(socket, true)
		defer n.opts.OnConn(socket, false)
	}

	var socketErr error = nil

	for {
		p := NewEmptyTHVPacket()
		socketErr = socket.readPacket(p)
		if socketErr != nil {
			break
		}

		typ := p.GetType()
		if typ <= PacketTypeInnerEndAt_ {
			switch typ {
			case PacketTypeHeartbeat:
				fallthrough
			case PacketTypeEcho:
				socket.SendPacket(p)
			}
		} else {
			if n.opts.OnMessage != nil {
				n.opts.OnMessage(socket, p)
			}
		}
	}
}

func (s *Server) Address() net.Addr {
	return s.listener.Addr()
}

func (s *Server) GetSocket(id string) *Socket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ret, ok := s.sockets[id]
	if ok {
		return ret
	}
	return nil
}

func (s *Server) SocketCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sockets)
}

func (s *Server) storeSocket(conn *Socket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sockets[conn.ID()] = conn
}

func (s *Server) removeSocket(conn *Socket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sockets, conn.ID())
}
