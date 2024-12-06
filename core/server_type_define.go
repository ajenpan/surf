package core

type ServerType = uint16

const (
	ServerType_Client ServerType = 0
	ServerType_Core   ServerType = 100
	ServerType_Gate   ServerType = 101
	ServerType_Lobby  ServerType = 103
	ServerType_UAuth  ServerType = 104
)
