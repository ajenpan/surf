package guandan

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"

	gdpoker "github.com/ajenpan/poker_algorithm/guandan"
	"github.com/ajenpan/poker_algorithm/poker"
	"github.com/ajenpan/surf/core/log"
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
		players: make(map[int32]*Player),
		conf: &LogicConfig{
			OutCardTimeSec: 15,
			WildCardRank:   2,
		},
		Logger: logger.Default.WithFields(map[string]interface{}{"game": "guandan"}),
	}
	return ret
}

const MaxSeatCnt = 4

type Player struct {
	raw    battle.Player
	online bool

	handCards *gdpoker.GDCards
	gameInfo  *PlayerGameInfo

	outcards []*OutCardInfo

	getPowerAt time.Time

	powerDeadLine time.Time
	actionPower   *NotifyPlayerActionPower
}

type LogicConfig struct {
	OutCardTimeSec int32 // 出牌时间,秒
	WildCardRank   int32 // 集牌
}

type Guandan struct {
	logger.Logger

	WildCard poker.Card
	table    battle.Table
	conf     *LogicConfig
	// info    *GameInfo
	// players []*Player // seatid to player
	players map[int32]*Player

	gameTime time.Duration

	CT *calltable.CallTable[uint32]

	lastStage *StageInfo
	currStage *StageInfo

	lastActionPlayer *Player
	currActionPlayer *Player

	rawDeck []byte
}

func (g *Guandan) nextStage(curr StageType) StageType {
	switch curr {
	case StageType_Stage_None:
		return StageType_Stage_GameStart
	case StageType_Stage_GameStart:
		return StageType_Stage_DoubleBet
	case StageType_Stage_DoubleBet:
		return StageType_Stage_DealingCards
	case StageType_Stage_DealingCards:
		return StageType_Stage_Gaming
	case StageType_Stage_Gaming:
		return StageType_Stage_Tally
	case StageType_Stage_Tally:
		return StageType_Stage_FinalResult
	case StageType_Stage_FinalResult:
		return StageType_Stage_None
	default:
		g.Errorf("unknown stage:%v", curr)
	}
	return StageType_Stage_None
}

func (g *Guandan) getStageInfo(t StageType) *StageInfo {
	switch t {
	case StageType_Stage_None:
		return &StageInfo{
			StageType:  StageType_Stage_None,
			ExitCond:   g.allPlayerOnline,
			TimeToLive: 10 * time.Second,
		}
	case StageType_Stage_GameStart:
		return &StageInfo{
			StageType: StageType_Stage_GameStart,
			OnEnterFn: func() {
				g.table.ReportBattleStatus(battle.BattleStatus_Running)
			},
			TimeToLive: 1 * time.Second,
		}
	case StageType_Stage_DoubleBet:
		return &StageInfo{
			StageType:  StageType_Stage_DoubleBet,
			TimeToLive: 15 * time.Second,
			ExitCond:   g.allDoubleRateSet,
			OnExitFn:   g.fullPlayerDoubleBet,
		}
	case StageType_Stage_DealingCards:
		return &StageInfo{
			StageType:  StageType_Stage_DealingCards,
			TimeToLive: 10 * time.Second,
			OnEnterFn:  g.doDealingCards,
		}
	case StageType_Stage_Gaming:
		return &StageInfo{
			StageType:   StageType_Stage_Gaming,
			OnEnterFn:   g.setFirstOutCardPlayer,
			OnProcessFn: g.onGaming,
			ExitCond:    g.hasGameResult,
		}
	case StageType_Stage_Tally:
		return &StageInfo{
			StageType:  StageType_Stage_Tally,
			TimeToLive: 5 * time.Second,
		}
	case StageType_Stage_FinalResult:
		return &StageInfo{
			StageType: StageType_Stage_FinalResult,
			OnEnterFn: func() {
				g.table.ReportBattleStatus(battle.BattleStatus_Over)
			},
		}
	default:
		return nil
	}
}

func (g *Guandan) OnInit(opts battle.LogicOpts) error {
	g.table = opts.Table
	if len(opts.Conf) > 2 {
		err := json.Unmarshal(opts.Conf, g.conf)
		if err != nil {
			return fmt.Errorf("unmarshal config error:%v", err)
		}
	}

	if opts.Log != nil {
		g.Logger = opts.Log
	}

	if len(opts.Players) != 4 {
		return fmt.Errorf("player count not match, expect 4, got:%v", len(opts.Players))
	}

	g.CT = calltable.NewCallTable[uint32]()

	for _, p := range opts.Players {
		if p.SeatID() < 0 || p.SeatID() >= MaxSeatCnt {
			return fmt.Errorf("seatid out of range, expect 0-3, got:%v", p.SeatID())
		}

		g.players[p.SeatID()] = &Player{
			raw: p,
			gameInfo: &PlayerGameInfo{
				SeatId:      p.SeatID(),
				CurrOutCard: nil,
			},
		}
	}

	g.WildCard = poker.NewCard(poker.HEART, (poker.CardRank)(g.conf.WildCardRank))
	g.currStage = g.getStageInfo(StageType_Stage_None)
	g.currStage.OnEnter()

	g.Infof("guandan logic init success, conf:%v", g.conf)
	return nil
}

// handle message
func (g *Guandan) OnReqGameInfo(player *Player, msg *ReqGameInfo) {
	resp := &RespGameInfo{
		// todo:
	}
	g.sendMsg(player, resp)
}

func (g *Guandan) OnReqPlayerDoubleBet(player *Player, msg *ReqPlayerDoubleBet) {
	flag := int32(0)

	if player.actionPower == nil {
		flag = 101
	} else {
		if player.gameInfo.DoubleBet == 0 {
			player.gameInfo.DoubleBet = msg.DoubleBet
		} else {
			flag = 301
		}
	}

	resp := &RespPlayerDoubleBet{
		Flag: flag,
	}
	g.sendMsg(player, resp)

	if flag != 0 {
		return
	}

	notify := &NotifyPlayerDoubleBet{
		SeatId:    player.gameInfo.SeatId,
		DoubleBet: player.gameInfo.DoubleBet,
	}
	g.broadcastMsg(notify)
}

func (g *Guandan) OnReqPlayerOutCards(player *Player, msg *ReqPlayerOutCards) {

	flag := func() int32 {
		if player.actionPower == nil {
			return 101
		}
		outcards := msg.GetOutCards()
		cards, err := poker.BytesToCards(outcards.GetCards())
		if err != nil {
			g.Errorf("parse outcards error:%v", err)
			return 201
		}
		result := gdpoker.GetDeckPower(g.WildCard, cards)
		if (DeckType)(result.DeckType) != outcards.GetDeckType() {
			g.Errorf("decktype not match, expect:%v, got:%v", result.DeckType, outcards.GetDeckType())
			return 202
		}
		ok := player.handCards.RemoveCards(cards)
		if !ok {
			g.Errorf("remove cards error")
			return 203
		}

		player.gameInfo.HandCards = player.handCards.Bytes()
		player.outcards = append(player.outcards, outcards)

		return 0
	}()

	resp := &RespPlayerOutCards{
		Flag:      flag,
		HandCards: player.gameInfo.HandCards,
	}

	g.sendMsg(player, resp)

	if flag != 0 {
		return
	}

	notify := &NotifyPlayerOutCards{
		OutCards: msg.OutCards,
		SeatId:   player.gameInfo.SeatId,
	}

	g.broadcastMsg(notify)

	g.setNextOutCardPlayer(player)
}

func (g *Guandan) OnReset() {

}

func (g *Guandan) hasGameResult() bool {
	return false
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

func (g *Guandan) OnTick(delta time.Duration) {
	g.gameTime += delta

	curr := g.currStage
	if curr.CheckExit() {
		nextType := g.nextStage(curr.StageType)
		nextStage := g.getStageInfo(nextType)
		g.changeLogicStep(nextStage)
		return
	}

	curr.OnProcess(delta)
}

func (g *Guandan) getStageDowntime(s StageType) time.Duration {
	switch s {
	case StageType_Stage_None:
		return 10 * time.Second
	case StageType_Stage_GameStart:
		return 1 * time.Second
	case StageType_Stage_DoubleBet:
		return 15 * time.Second
	case StageType_Stage_DealingCards:
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
		if p.gameInfo.DoubleBet == 0 {
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
		p.handCards = deck.DealCards(27)
		p.gameInfo.HandCards = p.handCards.Bytes()
		notice := &NotifyPlayerHandCards{
			SeatId: p.gameInfo.SeatId,
			Cards:  p.gameInfo.HandCards,
		}
		g.sendMsg(p, notice)
	}
}

func (g *Guandan) setFirstOutCardPlayer() {
	g.setOutCardPlayer(g.getPlayerBySeatId(0), &ActionPower_OutCardConf{EnablePass: false})
}

func (g *Guandan) setOutCardPlayer(p *Player, conf *ActionPower_OutCardConf) {
	g.currActionPlayer = p
	p.getPowerAt = time.Now()
	p.powerDeadLine = p.getPowerAt.Add(time.Second * time.Duration(g.conf.OutCardTimeSec))

	notice := &NotifyPlayerActionPower{
		SeatId: p.gameInfo.SeatId,
		ActionPower: &ActionPower{
			ActionType: ActionType_Action_OutCard,
			Conf: &ActionPower_Outcard{
				Outcard: conf,
			},
		},
		Deadline: p.powerDeadLine.Unix(),
		Downtime: int32(g.conf.OutCardTimeSec),
	}

	p.actionPower = notice

	g.broadcastMsg(notice)
}

func (g *Guandan) setNextOutCardPlayer(currPlayer *Player) {
	currPlayer.actionPower = nil

	currSeatId := currPlayer.gameInfo.SeatId

	var nextplayer *Player = nil

	for i := int32(0); i < 4; i++ {
		nextseatid := (currSeatId + i + 1) % MaxSeatCnt
		temp := g.getPlayerBySeatId(nextseatid)
		if temp == nil {
			log.Error("get next player error")
			return
		}
		if temp.handCards.Size() != 0 {
			nextplayer = temp
			break
		}
	}

	if nextplayer == nil {
		return
	}

	g.setOutCardPlayer(nextplayer, &ActionPower_OutCardConf{EnablePass: true})
}

func (g *Guandan) onGaming(time.Duration) {
	currplayer := g.currActionPlayer
	isTimeout := time.Now().After(currplayer.powerDeadLine)
	if !isTimeout {
		return
	}
	g.playerActionTimeoutHelp(currplayer)
}

func (g *Guandan) playerActionTimeoutHelp(player *Player) {
	if player.actionPower == nil {
		g.Errorf("player action timeout help error, player:%v", player.raw.UID())
		return
	}

	if player.actionPower.ActionPower.ActionType == ActionType_Action_OutCard {
		action := &ReqPlayerOutCards{}
		if player.actionPower.ActionPower.Conf.(*ActionPower_Outcard).Outcard.EnablePass {
			action.OutCards = &OutCardInfo{
				DeckType: DeckType_Deck_Pass,
			}
		} else {
			cards := player.handCards.PopBack()
			action.OutCards = &OutCardInfo{
				DeckType: DeckType_Deck_Single,
				Cards:    []byte{byte(cards)},
			}
		}
		g.OnReqPlayerOutCards(player, action)
		return
	}
}

func (g *Guandan) broadcastMsg(msg proto.Message) {
	msgid, err := gutils.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	if err != nil {
		g.Errorf("broadcastMsg get message:%s msgid error:%v", msg.ProtoReflect().Descriptor().Name(), err)
		return
	}
	g.table.BroadcastMessage(msgid, msg)
}

func (g *Guandan) sendMsg(p *Player, msg proto.Message) {
	msgid, err := gutils.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	if err != nil {
		g.Errorf("sendMsg get message:%s msgid error:%v", msg.ProtoReflect().Descriptor().Name(), err)
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

func (g *Guandan) changeLogicStep(nextStage *StageInfo) bool {
	if nextStage == nil {
		g.Errorf("next stage is nil")
		return false
	}

	currStage := g.currStage
	if currStage.StageType == nextStage.StageType {
		g.Errorf("set same step before:%v, now:%v", currStage.StageType, nextStage.StageType)
		return false
	}

	g.currStage = nextStage
	g.lastStage = currStage

	currStage.OnExit()

	donwtime := g.getStageDowntime(g.currStage.StageType).Seconds()

	g.Infof("game stage changed, from:%v,to:%v,donwtime:%v",
		g.lastStage.StageType, g.currStage.StageType, donwtime)

	notice := &NotifyGameStage{
		CurrStage: g.currStage.StageType,
		LastStage: g.lastStage.StageType,
		Downtime:  int32(donwtime),
		Deadline:  time.Now().Add(time.Second * time.Duration(donwtime)).Unix(),
	}

	g.broadcastMsg(notice)

	nextStage.OnEnter()
	return true
}

// 填充玩家的加倍倍数
func (g *Guandan) fullPlayerDoubleBet() {
	for _, p := range g.players {
		if p.gameInfo.DoubleBet != 0 {
			continue
		}
		p.gameInfo.DoubleBet = 1
		notice := &NotifyPlayerDoubleBet{
			SeatId:    p.gameInfo.SeatId,
			DoubleBet: p.gameInfo.DoubleBet,
		}
		g.broadcastMsg(notice)
	}
}
