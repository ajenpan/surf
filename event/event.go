package event

import (
	"surfing/event/proto"
)

type Event = proto.EventMessage

type EventAgent interface {
	Register(topic string, fn func(*Event)) error
	Publish(topic string, data string) error
}
