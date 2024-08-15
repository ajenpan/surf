package server

type Server interface {
	ServerName() string
	ServerType() uint16
}
