package guandan

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	gdpoker "github.com/ajenpan/poker_algorithm/guandan"
	logger "github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/utils/calltable"
	gutils "github.com/ajenpan/surf/game/utils"
	"github.com/ajenpan/surf/server/battle"
)

func init() {
	battle.RegisterGame("guandan", NewLogic)
}

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
	seatId   int32

	getPowerAt time.Time

	powerDeadLine time.Time
	actionPower   *NotifyPlayerActionPower
	actionReq     *ReqPlayerAction
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

	CT *calltable.CallTable[uint32]

	lastStage StageType

	lastActionPlayer *Player
	currActionPlayer *Player
	rawDeck          []byte
}

// func (gd *Guandan) NewStageInfo() []*StageInfo {
// 	ret := []*StageInfo{}
// 	ret = append(ret, &StageInfo{})
// 	return ret
// }

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

func (g *Guandan) OnReqGameInfo(player *Player, msg *ReqGameInfo) {
	g.sendMsg(player, g.info)
}

func (g *Guandan) OnReqPlayerAction(player *Player, msg *ReqPlayerAction) {
	flag := int32(0)
	resp := &RespPlayerAction{
		Flag: flag,
	}
	g.sendMsg(player, resp)

	if flag != 0 {
		return
	}

	if msg.ActionDetail.ActionType == ActionType_action_out_card {
		player.outcards = append(player.outcards, msg.ActionDetail.GetOutCards())
	}

	notify := &NotifyPlayerAction{
		ActionDetail: msg.ActionDetail,
		SeatId:       player.seatId,
	}
	g.broadcastMsg(notify)

	g.setNextOutCardPlayer(player)
}

func (g *Guandan) OnReset() {

}
func (g *Guandan) OnPlayerConnStatus(p battle.Player, enable bool) {
	player := g.playerConv(p)
	player.online = enable
}

func (g *Guandan) OnPlayerMessage(p battle.Player, msgid uint32, data []byte) {
	method := g.CT.Get(msgid)
	if method == nil {
		return
	}

	req, ok := method.GetRequest().(proto.Message)
	if !ok {
		return
	}
	err := proto.Unmarshal(data, req)
	if err != nil {
		return
	}
	player := g.playerConv(p)
	method.Call(g, player, req)
}

func (g *Guandan) OnTick(duration time.Duration) {
	g.gameTime += duration
	g.stageTime += duration

	currStage := g.getStage()
	timeout := false

	stageDownTime := g.getStageDowntime(g.info.Stage)
	if stageDownTime > 0 {
		timeout = g.stageTime >= stageDownTime
	}

	switch currStage {
	case StageType_StageNone:
		if timeout || g.allPlayerOnline() {
			g.changeLogicStep(g.nextStage())
		}
	case StageType_StageGameStart:
		if timeout {
			g.changeLogicStep(g.nextStage())
		}
	case StageType_StageDoubleRate:
		if timeout || g.allDoubleRateSet() {
			g.changeLogicStep(g.nextStage())

			// 发牌
			g.doDealingCards()
		}
	case StageType_StageDealingCards:
		if timeout {
			g.changeLogicStep(g.nextStage())
			// 开始游戏
			g.setOutCardPlayer(g.getPlayerBySeatId(0))
		}
	case StageType_StageGaming:
		g.onGaming()
	case StageType_StageTally:

	case StageType_StageFinalResult:
	}
}

func (g *Guandan) getStageDowntime(s StageType) time.Duration {
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

func (g *Guandan) allPlayerOnline() bool {
	for _, p := range g.players {
		if !p.online {
			return false
		}
	}
	return true
}

func (g *Guandan) allDoubleRateSet() bool {
	for _, p := range g.players {
		if p.DoubleRate == 0 {
			return false
		}
	}
	return true
}

func (g *Guandan) doDealingCards() {
	deck := gdpoker.NewDeck()
	deck.Shuffle()
	g.rawDeck = deck.Bytes()

	for _, p := range g.players {
		p.rawHandCards = deck.DealCards(27)
		p.HandCards = p.rawHandCards.Bytes()
		notice := &NotifyPlayerHandCards{
			SeatId: p.SeatId,
			Cards:  p.HandCards,
		}
		g.sendMsg(p, notice)
	}
}

func (g *Guandan) setOutCardPlayer(p *Player) {
	g.currActionPlayer = p
	p.getPowerAt = time.Now()
	p.powerDeadLine = p.getPowerAt.Add(time.Second * time.Duration(g.conf.OutCardTimeSec))

	notice := &NotifyPlayerActionPower{
		SeatId:   p.SeatId,
		Action:   ActionType_action_out_card,
		Deadline: p.powerDeadLine.Unix(),
		TimeDown: int32(g.conf.OutCardTimeSec),
	}

	p.actionReq = nil
	p.actionPower = notice

	g.broadcastMsg(notice)
}

func (g *Guandan) setNextOutCardPlayer(currPlayer *Player) {
	currPlayer.actionPower = nil
	currPlayer.actionReq = nil

	currSeatId := currPlayer.seatId
	var nextplayer *Player = nil

	for i := int32(0); i < 4; i++ {
		nextseatid := currSeatId + i + 1
		if nextseatid >= int32(len(g.players)) {
			nextseatid = 0
		}
		temp := g.getPlayerBySeatId(nextseatid)
		if temp.rawHandCards.Size() != 0 {
			nextplayer = temp
			break
		}
	}

	if nextplayer == nil {
		return
	}

	g.setOutCardPlayer(nextplayer)
}

func (g *Guandan) onGaming() {
	currplayer := g.currActionPlayer
	isTimeout := time.Now().After(currplayer.powerDeadLine)
	if !isTimeout {
		return
	}
	g.playerActionTimeoutHelp(currplayer)
}

func (g *Guandan) playerActionTimeoutHelp(curr *Player) {
	if curr.actionPower.Action == ActionType_action_out_card {
		action := &ReqPlayerAction{}
		g.OnReqPlayerAction(curr, action)
		return
	}
}

func (g *Guandan) nextStage() StageType {
	curr := g.getStage()

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
		g.Errorf("unknown stage:%v", curr)
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

func (g *Guandan) sendMsg(p *Player, msg proto.Message) {
	msgid, err := gutils.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	if err != nil {
		g.Errorf("get message msgid error:%v", err)
		return
	}
	g.table.SendMessageToPlayer(p.raw, msgid, msg)
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
