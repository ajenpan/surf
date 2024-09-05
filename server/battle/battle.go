package battle

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type SeatID int32
type GameStatus int32
type RoleType int32

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
	SeatID() SeatID
	Score() int64 //game jetton
	Role() RoleType
	SetUserData(any)
	GetUserData() any
}

type Table interface {
	SendMessageToPlayer(p Player, msgid uint32, data proto.Message)

	BroadcastMessage(msgid uint32, data proto.Message)

	ReportBattleStatus(GameStatus)
	ReportBattleEvent(topic string, event proto.Message)

	AfterFunc(time.Duration, func())
}

type Logic interface {
	OnInit(c Table, players []Player, conf interface{}) error
	OnTick(df time.Duration)
	OnReset()
	OnPlayerConnStatus(p Player, enable bool)
	OnPlayerMessage(p Player, msgid uint32, data []byte)
}
