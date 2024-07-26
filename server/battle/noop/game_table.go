package noop

import (
	"github.com/ajenpan/surf/server/battle"

	"google.golang.org/protobuf/proto"
)

func NewGameTable() *GameDesk {
	return &GameDesk{}
}

type GameDesk struct {
}

func (gd *GameDesk) SendMessageToPlayer(battle.Player, proto.Message) {

}

func (gd *GameDesk) BroadcastMessage(proto.Message) {

}

func (gd *GameDesk) PublishEvent(proto.Message) {

}

func (gd *GameDesk) ReportBattleStatus(battle.GameStatus) {
}
