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

	ReportBattleStatus(gs GameStatus)

	ReportBattleEvent(topic string, event proto.Message)

	// Schedule a function to be called at table loop after a duration.
	AfterFunc(time.Duration, func()) AfterCancelFunc
}

type LogicOpts struct {
	Table   Table
	Players []Player
	Conf    []byte
	Log     log.Logger
}

type Logic interface {
	// Called when the logic is created.
	OnInit(opts LogicOpts) error

	// Called every tick. 'delta' is the elapsed time since the previous tick.
	OnTick(delta time.Duration)

	// Called when a player connects or disconnects.
	OnPlayerConnStatus(p Player, enable bool)

	// Called when a player sends a message.
	OnPlayerMessage(p Player, msgid uint32, data []byte)

	OnReset()
}
