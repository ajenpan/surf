package niuniu

type robotSocket struct {
}

func NewRobotSocket() *robotSocket {
	ret := &robotSocket{}
	return ret
}

func (s *robotSocket) ID() string {
	return "s.sid"
}

func (s *robotSocket) Close() {}

func (s *robotSocket) TypeName() string {
	return "RobotGameSession"
}
