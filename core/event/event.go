package event

import (
	"github.com/ajenpan/surf/core/event/proto"
)

type Event = proto.EventMessage

type Publisher interface {
	Publish(e *Event)
}

type Recver interface {
	OnEvent(e *Event)
}
