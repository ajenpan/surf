package core

import (
	"github.com/ajenpan/surf/core/auth"
	"github.com/ajenpan/surf/core/network"
)

type Context interface {
	Response(msg interface{}, err error)
	SendAsync(msg interface{}) error
	Caller() auth.User
}

type context struct {
	Conn network.Conn
	Core *Surf
}

func (ctx *context) Response(msg interface{}, err error) {

}

func (ctx *context) SendAsync(msg interface{}) error {
	return nil
}

func (ctx *context) Caller() auth.User {
	return nil
}
