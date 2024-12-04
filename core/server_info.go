package core

import "github.com/ajenpan/surf/core/utils/calltable"

type Server interface {
	ServerType() uint16
	ServerName() string

	OnInit(*Surf) error
	OnReady()
	OnStop() error
}

type ServerInfo struct {
	Server    Server
	Calltable *calltable.CallTable
}
