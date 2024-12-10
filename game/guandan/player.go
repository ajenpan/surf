package guandan

import (
	"time"

	gdpoker "github.com/ajenpan/poker_algorithm/guandan"
	battle "github.com/ajenpan/surf/game"
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
