package game

import (
	"log/slog"
	"time"

	"google.golang.org/protobuf/proto"
)

type SeatID = int32
type GameStatus = int32
type PlayerType = int32

const (
	GameStatus_Idle GameStatus = iota
	GameStatus_Running
	GameStatus_Over
)

const (
	PlayerType_Human PlayerType = 0
	PlayerType_Robot PlayerType = 1
)

type Player interface {
	UID() int64
	SeatID() SeatID
	Score() int64
	Role() PlayerType
	SetUserData(any)
	GetUserData() any
}

type AfterCancelFunc func()

type Table interface {
	BattleID() string

	SendMessageToPlayer(p Player, syn uint32, msgid uint32, data proto.Message)

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
	Log     *slog.Logger
}

type PlayerEnterSubType = int32

const (
	PlayerEnterSubType_join PlayerEnterSubType = iota
	PlayerEnterSubType_timeout
)

type PlayerLeaveSubType = int32

const (
	PlayerLeaveSubType_leave PlayerLeaveSubType = iota
	PlayerLeaveSubType_disconnect
)

type Logic interface {
	// Called when the logic is created.
	OnInit(opts LogicOpts) error

	// Called every tick. 'delta' is the elapsed time since the previous tick.
	OnTick(delta time.Duration)

	// Called when a player enters the game.
	OnPlayerEnter(p Player, subtype PlayerEnterSubType, extra []byte)

	// Called when a player leaves the game.
	OnPlayerLeave(p Player, subtype PlayerLeaveSubType, extra []byte)

	// Called when a player sends a message.
	OnPlayerMessage(p Player, syn uint32, msgid uint32, data []byte)

	OnReset()
}
