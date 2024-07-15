package core

import (
	"github.com/ajenpan/surf/core/auth"
)

type Context struct {
}

func (ctx *Context) SendRespMsg() {

}

func (ctx *Context) SendAsync() {

}

func (ctx *Context) From() auth.User {
	return nil
}
