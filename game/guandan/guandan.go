package guandan

import (
	"encoding/json"
	"fmt"
	"time"

	gdpoker "github.com/ajenpan/poker_algorithm/guandan"

	"google.golang.org/protobuf/proto"

	logger "github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/utils/calltable"
	gutils "github.com/ajenpan/surf/game/utils"
	"github.com/ajenpan/surf/server/battle"
)

func NewLogic() battle.Logic {
	return NewGuandan()
}

func NewGuandan() *Guandan {
	ret := &Guandan{
		players: []*Player{},
		info:    &GameInfo{},
		conf: &Config{
			OutCardTimeSec: 15,
			GhostCardNum:   2,
		},
		Logger: logger.Default.WithFields(map[string]interface{}{"game": "guandan"}),
	}
	return ret
}

type Player struct {
	raw          battle.Player
	online       bool
	rawHandCards *gdpoker.GDCards

	*PlayerGameInfo
	outcards []*OutCardInfo

	getPowerAt    time.Time
	powerDeadLine time.Time
	seatId        int32
}

type Config struct {
	OutCardTimeSec int32 // 出牌时间,秒
	GhostCardNum   int32 // 鬼牌数量
}

type Guandan struct {
	logger.Logger

	table   battle.Table
	conf    *Config
	info    *GameInfo
	players []*Player // seatid to player

	gameTime  time.Duration
	stageTime time.Duration

	CT *calltable.CallTable[int]

	lastStage StageType

	currActionPlayer *Player
	rawDeck          []byte
}

func (gd *Guandan) OnInit(opts battle.LogicOpts) error {
	gd.table = opts.Table
	if len(opts.Conf) > 2 {
		err := json.Unmarshal(opts.Conf, gd.conf)
		if err != nil {
			return fmt.Errorf("unmarshal config error:%v", err)
		}
	}

	if opts.Log != nil {
		gd.Logger = opts.Log
	}

	for i, p := range opts.Players {
		gd.players = append(gd.players, &Player{
			raw: p,
			PlayerGameInfo: &PlayerGameInfo{
				SeatId:      int32(i),
				CurrOutCard: nil,
			},
		})
	}

	return nil
}

// handle message

func (g *Guandan) OnReqGameInfo(p battle.Player, msg *ReqGameInfo) {
	player := g.playerConv(p)
	g.sendMsg(player, g.info)
}

func (g *Guandan) OnReqPlayerAction(p battle.Player, msg *ReqPlayerAction) {

}
func (g *Guandan) OnReset() {

}
func (g *Guandan) OnPlayerConnStatus(p battle.Player, enable bool) {
	player := g.playerConv(p)
	player.online = enable
}

func (g *Guandan) OnPlayerMessage(p battle.Player, msgid uint32, data []byte) {
	// player := g.playerConv(p)

}

func (gd *Guandan) OnTick(duration time.Duration) {
	gd.gameTime += duration
	gd.stageTime += duration

	currStage := gd.getStage()
	timeout := false

	stageDownTime := gd.getStageDowntime(gd.info.Stage)
	if stageDownTime > 0 {
		timeout = gd.stageTime >= stageDownTime
	}

	switch currStage {
	case StageType_StageNone:
		if timeout || gd.allPlayerOnline() {
			gd.changeLogicStep(gd.nextStage())
		}
	case StageType_StageGameStart:
		if timeout {
			gd.changeLogicStep(gd.nextStage())
		}
	case StageType_StageDoubleRate:
		if timeout || gd.allDoubleRateSet() {
			gd.changeLogicStep(gd.nextStage())

			// 发牌
			gd.doDealingCards()
		}
	case StageType_StageDealingCards:
		if timeout {
			gd.changeLogicStep(gd.nextStage())
			// 开始游戏
			gd.setOutCardPlayer(gd.getPlayerBySeatId(0))
		}
	case StageType_StageGaming:
		gd.onGaming()
	case StageType_StageTally:
	case StageType_StageFinalResult:
	}
}

func (gd *Guandan) getStageDowntime(s StageType) time.Duration {
	switch s {
	case StageType_StageNone:
		return 10 * time.Second
	case StageType_StageGameStart:
		return 1 * time.Second
	case StageType_StageDoubleRate:
		return 15 * time.Second
	case StageType_StageDealingCards:
		return 4 * time.Second
	}
	return 0
}

func (gd *Guandan) allPlayerOnline() bool {
	for _, p := range gd.players {
		if !p.online {
			return false
		}
	}
	return true
}

func (gd *Guandan) allDoubleRateSet() bool {
	for _, p := range gd.players {
		if p.DoubleRate == 0 {
			return false
		}
	}
	return true
}

func (gd *Guandan) doDealingCards() {
	deck := gdpoker.NewDeck()
	deck.Shuffle()
	gd.rawDeck = deck.Bytes()

	for _, p := range gd.players {
		p.rawHandCards = deck.DealHandCards()
		p.HandCards = p.rawHandCards.Bytes()
		notice := &NotifyPlayerHandCards{
			SeatId: p.SeatId,
			Cards:  p.HandCards,
		}
		gd.sendMsg(p, notice)
	}
}

func (gd *Guandan) setOutCardPlayer(p *Player) {
	gd.currActionPlayer = p
	p.getPowerAt = time.Now()
	p.powerDeadLine = p.getPowerAt.Add(time.Second * time.Duration(gd.conf.OutCardTimeSec))

	notice := &NotifyPlayerActionPower{
		SeatId:   p.SeatId,
		Action:   ActionType_action_out_card,
		Deadline: p.powerDeadLine.Unix(),
		TimeDown: int32(gd.conf.OutCardTimeSec),
	}

	gd.broadcastMsg(notice)
}

func (gd *Guandan) onGaming() {
	currplayer := gd.currActionPlayer
	isTimeout := time.Now().After(currplayer.powerDeadLine)
	if isTimeout {

	}
}

func (gd *Guandan) nextStage() StageType {
	curr := gd.getStage()

	switch curr {
	case StageType_StageNone:
		return StageType_StageGameStart
	case StageType_StageGameStart:
		return StageType_StageDoubleRate
	case StageType_StageDoubleRate:
		return StageType_StageDealingCards
	case StageType_StageDealingCards:
		return StageType_StageGaming
	case StageType_StageGaming:
		return StageType_StageTally
	case StageType_StageTally:
		return StageType_StageFinalResult
	case StageType_StageFinalResult:
		return StageType_StageNone
	default:
		gd.Errorf("unknown stage:%v", curr)
	}
	return StageType_StageNone
}

func (g *Guandan) broadcastMsg(msg proto.Message) {
	msgid, err := gutils.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	if err != nil {
		g.Errorf("get message msgid error:%v", err)
		return
	}
	g.table.BroadcastMessage(msgid, msg)
}

func (gd *Guandan) sendMsg(p *Player, msg proto.Message) {
	msgid, err := gutils.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	if err != nil {
		gd.Errorf("get message msgid error:%v", err)
		return
	}
	gd.table.SendMessageToPlayer(p.raw, msgid, msg)
}

func (g *Guandan) playerConv(p battle.Player) *Player {
	return g.getPlayerBySeatId(int32(p.SeatID()))
}

func (g *Guandan) getPlayerBySeatId(seatid int32) *Player {
	if seatid < 0 || seatid >= int32(len(g.players)) {
		g.Errorf("seatid out of range, seatid:%v, players len:%v", seatid, len(g.players))
		return nil
	}
	return g.players[seatid]
}

func (g *Guandan) getStage() StageType {
	return g.info.Stage
}

func (g *Guandan) changeLogicStep(s StageType) bool {
	lastStatus := g.getStage()
	g.info.Stage = s
	if lastStatus == s {
		g.Errorf("set same step before:%v, now:%v", lastStatus, s)
		return false
	}

	g.lastStage = lastStatus
	g.stageTime = 0

	donwtime := g.getStageDowntime(s).Seconds()

	g.Infof("game step changed, before:%v, now:%v, time down:%v", lastStatus, s, donwtime)

	notice := &NotifyGameStage{
		CurrStage: s,
		LastStage: lastStatus,
		TimeDown:  int32(donwtime),
		Deadline:  time.Now().Add(time.Second * time.Duration(donwtime)).Unix(),
	}

	g.broadcastMsg(notice)
	return true
}
