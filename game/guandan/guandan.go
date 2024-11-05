package guandan

import (
	"time"

	logger "github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/server/battle"
)

// func CreateLogic() battle.Logic {
// 	return CreateDDZ()
// }

func CreateDDZ() *Guandan {
	ret := &Guandan{
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

type Guandan struct {
	logger.Logger

	table battle.Table
	// conf    *Config
	// info    *GameInfo
	// players map[int32]*NNPlayer // seatid to player

	gameTime  time.Duration
	stageTime time.Duration

	CT *calltable.CallTable[int]
}
