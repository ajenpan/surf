package noop

import (
	protobuf "google.golang.org/protobuf/proto"
)

type GamePlayer struct {
	SeatID int32
	Score  int64
	Robot  bool
}

func (p *GamePlayer) GetSeatID() int32 {
	return p.SeatID
}
func (p *GamePlayer) GetScore() int64 {
	return p.Score
}
func (p *GamePlayer) IsRobot() bool {
	return p.Robot
}
func (p *GamePlayer) SendMessage(protobuf.Message) error {
	return nil
}
