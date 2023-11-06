package server

import (
	"github.com/ajenpan/surf/msg"
)

type MessageHead = *msg.Head
type MessageBody = []byte
type Error = msg.Error

type Message struct {
	Head MessageHead
	Body MessageBody
}

func NewMessage() *Message {
	return &Message{
		Head: &msg.Head{},
	}
}
