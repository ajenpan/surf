package guandan

import (
	"time"

	gdpoker "github.com/ajenpan/poker_algorithm/guandan"
	"github.com/ajenpan/surf/server/battle"
)

type Player struct {
	raw    battle.Player
	online bool

	handCards *gdpoker.GDCards
	gameInfo  *PlayerGameInfo

	outcards []*OutCardInfo

	getPowerAt time.Time

	powerDeadLine time.Time
	outcardPower  *NotifyPlayerOutCardPower

	resultRank uint8
}
