package server

import (
	"crypto/rsa"
	"fmt"
	"net"

	"github.com/ajenpan/surf/auth"
	"github.com/ajenpan/surf/log"
	"github.com/ajenpan/surf/server/tcp"

	"github.com/ajenpan/surf/server/marshal"
)

type TcpServerOptions struct {
	ListenAddr       string
	AuthPublicKey    *rsa.PublicKey
	OnSessionMessage FuncOnSessionMessage
	OnSessionStatus  FuncOnSessionStatus
	Marshal          marshal.Marshaler
}

func NewTcpServer(opts *TcpServerOptions) (*TcpServer, error) {
	ret := &TcpServer{
		opts:       opts,
		listenAddr: opts.ListenAddr,
	}
	if opts.Marshal == nil {
		opts.Marshal = &marshal.ProtoMarshaler{}
	}
	tcpopt := tcp.ServerOptions{
		Address:         opts.ListenAddr,
		OnSocketMessage: ret.OnMessage,
		OnSocketConn:    ret.OnConn,
		OnSocketDisconn: ret.OnDisconn,
		OnAccpect: func(conn net.Conn) bool {
			log.Debugf("OnAccpectConn remote:%s, local:%s", conn.RemoteAddr(), conn.LocalAddr())
			return true
		},
		NewIDFunc: NewSessionID,
	}

	if opts.AuthPublicKey != nil {
		tcpopt.AuthFunc = auth.RsaTokenAuth(opts.AuthPublicKey)
	}

	imp, err := tcp.NewServer(tcpopt)
	if err != nil {
		return nil, err
	}
	ret.imp = imp
	return ret, nil
}

type TcpServer struct {
	imp *tcp.Server

	opts       *TcpServerOptions
	listenAddr string
}

type TcpSession struct {
	*tcp.Socket
}

var tcpSessionKey = &struct{}{}

func (s *TcpSession) Send(p *Message) error {
	return s.Socket.Send(p)
}

func (s *TcpSession) SessionType() string {
	return "tcp-session"
}

func (s *TcpServer) Start() error {
	return s.imp.Start()
}

func (s *TcpServer) Stop() error {
	return s.imp.Stop()
}

func loadTcpSession(socket *tcp.Socket) *TcpSession {
	v, ok := socket.Meta.Load(tcpSessionKey)
	if !ok {
		return nil
	}
	return v.(*TcpSession)
}

func (s *TcpServer) OnMessage(socket *tcp.Socket, p tcp.Packet) {
	if p.PacketType() != PacketBinaryRouteType {
		fmt.Println("unknow packet type:", p.PacketType())
		return
	}

	sess := loadTcpSession(socket)
	if sess == nil {
		return
	}

	if s.opts.OnSessionMessage != nil {
		msg, ok := p.(*Message)
		if ok {
			s.opts.OnSessionMessage(sess, msg)
		}
	}
}

func (s *TcpServer) OnConn(socket *tcp.Socket) {
	sess := &TcpSession{
		Socket: socket,
	}
	socket.Meta.Store(tcpSessionKey, sess)

	if s.opts.OnSessionStatus != nil {
		s.opts.OnSessionStatus(sess, true)
	}
}

func (s *TcpServer) OnDisconn(socket *tcp.Socket, err error) {
	sess := loadTcpSession(socket)
	socket.Meta.Delete(tcpSessionKey)
	if sess != nil && s.opts.OnSessionStatus != nil {
		s.opts.OnSessionStatus(sess, false)
	}
	log.Errorf("OnDisconn:%v %v", socket.SessionID(), err)
}
