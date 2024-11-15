package core

type ServerType = uint16

const (
	ServerType_Client ServerType = 0
	ServerType_Gate   ServerType = 100
	ServerType_Battle ServerType = 101
)
