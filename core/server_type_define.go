package core

type ServerType = uint16

const (
	ServerType_User   ServerType = 0
	ServerType_Core   ServerType = 100
	ServerType_Gate   ServerType = 101
	ServerType_Battle ServerType = 102
)
