package table

import (
	"fmt"
	"sync"

	protobuf "google.golang.org/protobuf/proto"

	msgBattle "github.com/ajenpan/surf/msg/battle"
	"github.com/ajenpan/surf/server/battle"
	bf "github.com/ajenpan/surf/server/battle"
)

func NewPlayer(p *msgBattle.PlayerInfo) *Player {
	return &Player{
		PlayerInfo: protobuf.Clone(p).(*msgBattle.PlayerInfo),
	}
}

func NewPlayers(infos []*msgBattle.PlayerInfo) ([]*Player, error) {
	ret := make([]*Player, len(infos))
	for i, info := range infos {
		ret[i] = NewPlayer(info)
	}

	// check seatid
	for _, v := range ret {
		if v.SeatId == 0 {
			return nil, fmt.Errorf("seat id is 0")
		}
		if v.Uid == 0 {
			return nil, fmt.Errorf("uid is 0")
		}
	}

	return ret, nil
}

type Player struct {
	*msgBattle.PlayerInfo
	Ready  int32
	sender func(msgname uint32, raw []byte) error
	online bool
	ud     any
}

func (p *Player) SetUserData(ud any) {
	p.ud = ud
}

func (p *Player) GetUserData() any {
	return p.ud
}

func (p *Player) Score() int64 {
	return p.PlayerInfo.Score
}

func (p *Player) UserID() uint64 {
	return p.PlayerInfo.Uid
}

func (p *Player) SeatID() battle.SeatID {
	return battle.SeatID(p.PlayerInfo.SeatId)
}

func (p *Player) Role() bf.RoleType {
	return bf.RoleType(p.PlayerInfo.Role)
}

func (p *Player) Send(msgid uint32, raw []byte) error {
	return p.sender(msgid, raw)
}

type PlayerStore struct {
	byUID    sync.Map
	bySeatID sync.Map
}

func (ps *PlayerStore) ByUID(uid uint64) *Player {
	p, has := ps.byUID.Load(uid)
	if !has {
		return nil
	}
	return p.(*Player)
}

func (ps *PlayerStore) BySeat(seatid uint32) *Player {
	p, has := ps.bySeatID.Load(seatid)
	if !has {
		return nil
	}
	return p.(*Player)
}

func (ps *PlayerStore) Store(p *Player) error {
	uid := p.Uid
	seatid := p.SeatId
	ps.bySeatID.Store(seatid, p)
	ps.byUID.Store(uid, p)
	return nil
}

func (ps *PlayerStore) Range(f func(p *Player) bool) {
	ps.byUID.Range(func(key, value any) bool {
		return f(value.(*Player))
	})
}
func (ps *PlayerStore) ToSlice() []*Player {
	ret := []*Player{}
	ps.Range(func(p *Player) bool {
		ret = append(ret, p)
		return true
	})
	return ret
}
