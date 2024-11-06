package noop

import (
	protobuf "google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/server/battle"
)

type GamePlayer struct {
	seat_id int32
	score   int64

	ud   any
	role battle.RoleType
}

func (p *GamePlayer) SeatID() int32 {
	return p.seat_id
}

func (p *GamePlayer) Score() int64 {
	return p.score
}

func (p *GamePlayer) Role() battle.RoleType {
	return p.role
}

func (p *GamePlayer) SendMessage(protobuf.Message) error {
	return nil
}

func (p *GamePlayer) SetUserData(ud any) {
	p.ud = ud
}

func (p *GamePlayer) GetUserData() any {
	return p.ud
}
