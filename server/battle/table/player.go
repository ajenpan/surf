package table

import (
	"fmt"
	"sync"

	protobuf "google.golang.org/protobuf/proto"

	msgBattle "github.com/ajenpan/surf/msg/battle"
	"github.com/ajenpan/surf/server/battle"
	bf "github.com/ajenpan/surf/server/battle"
)

type PlayerSender = func(msgid uint32, raw []byte) error

func NewPlayer(p *msgBattle.PlayerInfo) *Player {
	return &Player{
		PlayerInfo: protobuf.Clone(p).(*msgBattle.PlayerInfo),
	}
}

func NewPlayers(infos []*msgBattle.PlayerInfo) ([]*Player, error) {
	ret := make([]*Player, len(infos))

	uidMap := make(map[int64]struct{})
	seatidMap := make(map[int32]struct{})

	for i, info := range infos {
		if info.Uid == 0 {
			return nil, fmt.Errorf("uid is 0")
		}

		ret[i] = NewPlayer(info)

		uidMap[info.Uid] = struct{}{}
		if info.SeatId >= 0 {
			seatidMap[info.SeatId] = struct{}{}
		}
	}
	if len(uidMap) != len(ret) {
		return nil, fmt.Errorf("uid is not unique")
	}
	if len(seatidMap) != len(ret) {
		return nil, fmt.Errorf("seatid is not unique")
	}
	return ret, nil
}

type Player struct {
	*msgBattle.PlayerInfo

	Ready  int32
	sender PlayerSender
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

func (p *Player) UID() int64 {
	return p.PlayerInfo.Uid
}

func (p *Player) SeatID() battle.SeatID {
	return battle.SeatID(p.PlayerInfo.SeatId)
}

func (p *Player) Role() bf.RoleType {
	return bf.RoleType(p.PlayerInfo.Role)
}

func (p *Player) Send(msgid uint32, raw []byte) error {
	if p.sender == nil {
		// return fmt.Errorf("sender is nil")
		return nil
	}
	return p.sender(msgid, raw)
}

type PlayerStore struct {
	byUID    sync.Map
	bySeatID sync.Map
}

func (ps *PlayerStore) ByUID(uid int64) *Player {
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
	seatid := p.SeatId
	ps.bySeatID.Store(seatid, p)
	ps.byUID.Store(p.UID(), p)
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
