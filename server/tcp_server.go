package server

import (
	"crypto/rsa"

	"github.com/ajenpan/surf/server/tcp"

	"google.golang.org/protobuf/proto"
)

type TcpServerOptions struct {
	ListenAddr       string
	AuthPublicKey    *rsa.PublicKey
	OnSessionMessage FuncOnSessionMessage
	OnSessionStatus  FuncOnSessionStatus
}

func NewTcpServer(opts *TcpServerOptions) (*TcpServer, error) {
	ret := &TcpServer{
		opts:       opts,
		listenAddr: opts.ListenAddr,
	}
	tcpopt := tcp.ServerOptions{
		Address:   opts.ListenAddr,
		OnMessage: ret.OnTcpMessage,
		OnConn:    ret.OnTcpConn,
		NewIDFunc: NewSessionID,
	}

	if opts.AuthPublicKey != nil {
		tcpopt.AuthFunc = func(b []byte) (*tcp.UserInfo, error) {
			uid, uname, role, err := VerifyToken(opts.AuthPublicKey, string(b))
			if err != nil {
				return nil, err
			}
			return &tcp.UserInfo{
				UId:   uid,
				UName: uname,
				Role:  role,
			}, nil
		}
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

func (s *TcpSession) Send(msg *Message) error {
	p, err := s.msg2pkg(msg)
	if err != nil {
		return err
	}
	return s.Socket.SendPacket(p)
}

func (s *TcpSession) SessionType() string {
	return "tcp-session"
}

func (s *TcpSession) msg2pkg(p *Message) (*tcp.THVPacket, error) {
	head, err := proto.Marshal(p.Head)
	if err != nil {
		return nil, err
	}
	return tcp.NewTHVPacket(head, p.Body), nil
}

func (s *TcpSession) pkg2msg(p *tcp.THVPacket) (*Message, error) {
	msg := NewMessage()
	msg.Body = p.GetBody()
	err := proto.Unmarshal(p.GetHead(), msg.Head)
	if err != nil {
		return nil, err
	}
	return msg, nil
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

func (s *TcpServer) OnTcpMessage(socket *tcp.Socket, p *tcp.THVPacket) {
	sess := loadTcpSession(socket)
	if sess == nil {
		return
	}
	msg, err := sess.pkg2msg(p)
	if err != nil {
		return
	}

	if s.opts.OnSessionMessage != nil {
		s.opts.OnSessionMessage(sess, msg)
	}
}

func (s *TcpServer) OnTcpConn(socket *tcp.Socket, valid bool) {
	var sess *TcpSession
	if valid {
		sess = &TcpSession{
			Socket: socket,
		}
		socket.Meta.Store(tcpSessionKey, sess)
	} else {
		sess = loadTcpSession(socket)
		socket.Meta.Delete(tcpSessionKey)
	}

	if sess != nil && s.opts.OnSessionStatus != nil {
		s.opts.OnSessionStatus(sess, valid)
	}
}
