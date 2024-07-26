package event

type NoopPublisher struct{}

func (NoopPublisher) Publish(e *Event) {}

type NoopRecver struct{}

func (*NoopRecver) OnEvent(e *Event) {}
