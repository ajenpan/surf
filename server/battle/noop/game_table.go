package noop

import (
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/server/battle"
)

func NewGameTable() *GameTable {
	return &GameTable{
		id: uuid.New().String(),
	}
}

type GameTable struct {
	id string
}

func (table *GameTable) SendMessageToPlayer(battle.Player, uint32, proto.Message) {

}

func (table *GameTable) BroadcastMessage(uint32, proto.Message) {

}

func (table *GameTable) PublishEvent(proto.Message) {

}

func (table *GameTable) ReportBattleStatus(battle.GameStatus) {

}

func (table *GameTable) BattleID() string {
	return ""
}

func (table *GameTable) ReportBattleEvent(topic string, event proto.Message) {
}

func (table *GameTable) AfterFunc(time.Duration, func()) battle.AfterCancelFunc {
	return nil
}
