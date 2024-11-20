package ddz

import (
	"time"

	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/server/battle"
)

// func CreateLogic() battle.Logic {
// 	return CreateDDZ()
// }

func CreateDDZ() *DDZ {
	ret := &DDZ{
		// players: make(map[int32]*Player),
		// info:    &GameInfo{},
		// conf:    &Config{},
		// Logger:  logger.Default.WithFields(map[string]interface{}{"game": "niuniu"}),
	}
	return ret
}

type Player struct {
	// raw battle.Player
	// *GamePlayerInfo
	// rawHandCards *nncard.NNHandCards
}

type DDZ struct {
	table battle.Table
	// conf    *Config
	// info    *GameInfo
	// players map[int32]*NNPlayer // seatid to player

	gameTime  time.Duration
	stageTime time.Duration

	CT *calltable.CallTable
}
