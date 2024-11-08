package battle

import (
	"time"

	log "github.com/ajenpan/surf/core/log"

	"google.golang.org/protobuf/proto"
)

type SeatID = int32
type GameStatus = int32
type RoleType = int32

const (
	BattleStatus_Idle GameStatus = iota
	BattleStatus_Running
	BattleStatus_Over
)

const (
	RoleType_Player RoleType = iota
	RoleType_Robot
)

type Player interface {
	UID() int64
	SeatID() SeatID
	Score() int64
	Role() RoleType
	SetUserData(any)
	GetUserData() any
}

type AfterCancelFunc func()

type Table interface {
	BattleID() string

	SendMessageToPlayer(p Player, msgid uint32, data proto.Message)

	BroadcastMessage(msgid uint32, data proto.Message)

	ReportBattleStatus(GameStatus)
	ReportBattleEvent(topic string, event proto.Message)

	AfterFunc(time.Duration, func()) AfterCancelFunc
}

type LogicOpts struct {
	Table   Table
	Players []Player
	Conf    []byte
	Log     log.Logger
}

type Logic interface {
	OnInit(opts LogicOpts) error
	OnTick(df time.Duration)
	OnReset()
	OnPlayerConnStatus(p Player, enable bool)
	OnPlayerMessage(p Player, msgid uint32, data []byte)
}
