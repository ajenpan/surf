package server

type Agent interface {
	OnSessionStatus(Session, bool)
	OnSessionMessage(Session, *Message)
}
