package event

type Noop struct {
}

func (Noop) Publish(topic string, data string) error {
	return nil
}

func (*Noop) Register(topic string, fn func(*Event)) error {
	return nil
}
