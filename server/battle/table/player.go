package table

import (
	"fmt"
	"sync"

	protobuf "google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/server/battle"
	bf "github.com/ajenpan/surf/server/battle"
	pb "github.com/ajenpan/surf/server/battle/proto"
)

func NewPlayer(p *pb.PlayerInfo) *Player {
	return &Player{
		PlayerInfo: protobuf.Clone(p).(*pb.PlayerInfo),
	}
}

func NewPlayers(infos []*pb.PlayerInfo) ([]*Player, error) {
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
	*pb.PlayerInfo
	Ready  int32
	sender func(msgname uint32, raw []byte) error
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
	if p.PlayerInfo.IsRobot {
		return bf.RoleType_Robot
	}
	return bf.RoleType_Player
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
