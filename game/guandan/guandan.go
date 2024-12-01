package guandan

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/poker_algorithm/poker"
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
			OutCardTimeSec: 1,
			WildCardRank:   2,
		},
		Logger: slog.Default().With("game", "guandan"),
	}
	return ret
}

const MaxSeatCnt = 4

type msgContext struct {
	player *Player
	syn    uint32
}

type LogicConfig struct {
	OutCardTimeSec int32 // 出牌时间,秒
	WildCardRank   int32 // 集牌
}

type Guandan struct {
	*slog.Logger

	WildCard poker.Card
	table    battle.Table
	conf     *LogicConfig
	// info    *GameInfo
	// players []*Player // seatid to player
	players map[int32]*Player

	gameTime time.Duration

	ct *calltable.CallTable

	lastStage *StageInfo
	currStage *StageInfo

	lastActionPlayer *Player
	currActionPlayer *Player

	rawDeck []byte

	outcardHead       *OutCardInfo
	outcardHeadPlayer *Player

	currOutCardRank uint8 // 当前名次
	gameResult      bool  // 是否结束
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
		g.Error("unknown stage", "stage", curr)
	}
	return StageType_Stage_None
}

func (g *Guandan) newStageInfo(t StageType) *StageInfo {
	switch t {
	case StageType_Stage_None:
		return &StageInfo{
			StageType:  StageType_Stage_None,
			ExitCond:   g.allPlayerOnline,
			TimeToLive: 0,
		}
	case StageType_Stage_GameStart:
		return &StageInfo{
			StageType: StageType_Stage_GameStart,
			OnBeforeEnterFn: func() {
				g.table.ReportBattleStatus(battle.BattleStatus_Running)
			},
			TimeToLive: 1 * time.Second,
		}
	case StageType_Stage_DoubleBet:
		return &StageInfo{
			StageType:  StageType_Stage_DoubleBet,
			TimeToLive: 3 * time.Second,
			ExitCond:   g.allDoubleRateSet,
			OnExitFn:   g.fullPlayerDoubleBet,
		}
	case StageType_Stage_DealingCards:
		return &StageInfo{
			StageType:       StageType_Stage_DealingCards,
			TimeToLive:      10 * time.Second,
			OnBeforeEnterFn: g.doDealingCards,
			OnProcessFn: func(delta time.Duration) {
				if g.rawDeck == nil {
					g.doDealingCards()
				}
			},
		}
	case StageType_Stage_Gaming:
		return &StageInfo{
			StageType:   StageType_Stage_Gaming,
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
			OnBeforeEnterFn: func() {
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

	ct := calltable.NewCallTable()
	ct.AddFunction(g.OnReqGameInfo)
	ct.AddFunction(g.OnReqPlayerDoubleBet)
	ct.AddFunction(g.OnReqPlayerOutCards)

	g.ct = ct

	for _, p := range opts.Players {
		if p.SeatID() < 0 || p.SeatID() >= MaxSeatCnt {
			return fmt.Errorf("seatid out of range, expect 0-3, got:%v", p.SeatID())
		}

		g.players[p.SeatID()] = &Player{
			raw: p,
			gameInfo: &PlayerGameInfo{
				SeatId: p.SeatID(),
				UserId: int32(p.UID()),
				Score:  p.Score(),
			},
		}
	}

	g.WildCard = poker.NewCard(poker.HEART, (poker.CardRank)(g.conf.WildCardRank))
	g.currStage = g.newStageInfo(StageType_Stage_None)
	g.currStage.OnBeforeEnter()

	g.Info("guandan logic init success", "conf", g.conf)
	return nil
}

// handle message
func (g *Guandan) OnReqGameInfo(ctx *msgContext, msg *ReqGameInfo) {
	gameInfo := &GameInfo{
		Stage:    g.currStage.StageType,
		SubStage: 0,
		Conf: &GameConf{
			Wildcard: int32(g.WildCard.Rank()),
		},
		PlayerInfo: make(map[int32]*PlayerGameInfo, len(g.players)),
	}

	if g.currActionPlayer != nil {
		gameInfo.CurrActionPowerSeatid = g.currActionPlayer.gameInfo.SeatId
	}

	for _, p := range g.players {
		gameInfo.PlayerInfo[p.gameInfo.SeatId] = p.gameInfo
	}

	resp := &RespGameInfo{
		GameInfo: gameInfo,
	}
	g.sendMsg(ctx.player, ctx.syn, resp)
}

func (g *Guandan) OnReqPlayerDoubleBet(ctx *msgContext, msg *ReqPlayerDoubleBet) {
	flag := int32(0)
	player := ctx.player

	if player.gameInfo.DoubleBet == 0 {
		player.gameInfo.DoubleBet = msg.DoubleBet
	} else {
		flag = 301
	}

	resp := &RespPlayerDoubleBet{
		Flag: flag,
	}
	g.sendMsg(player, ctx.syn, resp)

	if flag != 0 {
		return
	}

	notify := &NotifyPlayerDoubleBet{
		SeatId:    player.gameInfo.SeatId,
		DoubleBet: player.gameInfo.DoubleBet,
	}
	g.broadcastMsg(notify)
}

func (g *Guandan) OnReqPlayerOutCards(ctx *msgContext, msg *ReqPlayerOutCards) {
	player := ctx.player

	respFn := func(flag int32) {
		resp := &RespPlayerOutCards{
			Flag:      flag,
			HandCards: player.gameInfo.HandCards,
		}
		g.sendMsg(player, ctx.syn, resp)
	}

	flag := g.playerOutCards(player, msg.OutCards, respFn)

	if flag != 0 {
		g.Warn("player out cards error", "player", player.raw.UID(), "flag", flag)
	}
}

func (g *Guandan) checkFinish() bool {
	teamAFinished := g.getPlayerBySeatId(0).handCards.IsEmpty() && g.getPlayerBySeatId(2).handCards.IsEmpty()
	teamBFinished := g.getPlayerBySeatId(1).handCards.IsEmpty() && g.getPlayerBySeatId(3).handCards.IsEmpty()

	if teamAFinished || teamBFinished {
		g.Info("game finished", "teamAFinished", teamAFinished, "teamBFinished", teamBFinished)
		g.gameResult = true
		return true
	}

	return false
}

func (g *Guandan) playerOutCards(player *Player, outcards *OutCardInfo, respFunc func(int32)) int32 {
	flag := func() int32 {
		if player.outcardPower == nil {
			return 101
		}
		if g.currActionPlayer != player {
			return 102
		}
		if outcards.GetDeckType() != DeckType_Deck_Pass {
			cards, err := poker.BytesToCards(outcards.GetCards())
			if err != nil {
				g.Error("parse outcards error", "err", err)
				return 201
			}
			// TODO compare deck power
			// result := gdpoker.GetDeckPower(g.WildCard, cards)
			// if (DeckType)(result.DeckType) != outcards.GetDeckType() {
			// 	g.Error("decktype not match", "expect", result.DeckType, "got", outcards.GetDeckType())
			// 	return 202
			// }

			ok := player.handCards.RemoveCards(cards)
			if !ok {
				g.Error("remove cards error")
				return 203
			}
			g.Info("player out cards", "player", player.raw.UID(), "outcards", cards.Chinese(), "outdecktype", outcards.GetDeckType(),
				"handcards", player.handCards.Chinese())

			player.gameInfo.HandCards = player.handCards.Bytes()
			g.outcardHead = outcards
			g.outcardHeadPlayer = player
		}
		return 0
	}()

	if respFunc != nil {
		respFunc(flag)
	}

	player.outcards = append(player.outcards, outcards)

	notify := &NotifyPlayerOutCards{
		OutCards: outcards,
		SeatId:   player.gameInfo.SeatId,
	}

	g.broadcastMsg(notify)

	if outcards.GetDeckType() != DeckType_Deck_Pass {
		if player.handCards.IsEmpty() {
			player.resultRank = g.currOutCardRank
			g.currOutCardRank += 1
		}

		if g.checkFinish() {
			return flag
		}
	}

	g.setNextOutCardPlayer(player)
	return flag
}

func (g *Guandan) OnReset() {

}

func (g *Guandan) hasGameResult() bool {
	return g.gameResult
}

func (g *Guandan) OnPlayerEnter(p battle.Player, subtype battle.PlayerEnterSubType, extra []byte) {
	player := g.playerConv(p)
	player.online = true
}

func (g *Guandan) OnPlayerLeave(p battle.Player, subtype battle.PlayerLeaveSubType, extra []byte) {
	player := g.playerConv(p)
	player.online = false
}

func (g *Guandan) OnPlayerMessage(p battle.Player, syn uint32, msgid uint32, data []byte) {
	method := g.ct.GetByID(msgid)
	if method == nil {
		g.Error("method not found", "msgid", msgid)
		return
	}

	req, ok := method.GetRequest().(proto.Message)
	if !ok {
		g.Error("method request type error", "msgid", msgid)
		return
	}
	err := proto.Unmarshal(data, req)
	if err != nil {
		g.Error("method request unmarshal error", "msgid", msgid, "err", err)
		return
	}
	player := g.playerConv(p)
	if player == nil {
		g.Error("player not found", "seatid", p.SeatID())
		return
	}
	ctx := &msgContext{
		player: player,
		syn:    syn,
	}
	method.Call(ctx, req)
}

func (g *Guandan) OnTick(delta time.Duration) {
	g.gameTime += delta

	curr := g.currStage
	if curr.CheckExit() {
		nextType := g.nextStage(curr.StageType)
		nextStage := g.newStageInfo(nextType)
		g.changeLogicStep(nextStage)
		return
	}

	curr.OnProcess(delta)
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
	// deck := gdpoker.NewDeck()
	deck := poker.NewDeckWithoutJoker()
	deck.Shuffle()
	g.rawDeck = deck.Bytes()

	g.currOutCardRank = 1

	cnt := deck.Size() / len(g.players)

	for _, p := range g.players {
		p.handCards = deck.DealCards(cnt)
		p.gameInfo.HandCards = p.handCards.Bytes()
		notice := &NotifyPlayerHandCards{
			SeatId: p.gameInfo.SeatId,
			Cards:  p.gameInfo.HandCards,
		}

		g.Info("deal cards", "seatid", p.gameInfo.SeatId, "cards:", p.handCards.Chinese())
		g.sendMsg(p, 0, notice)
	}
}

func (g *Guandan) setFirstOutCardPlayer() {
	// seatid := rand.Int31n(MaxSeatCnt)
	seatid := int32(0)
	g.Info("set first out card player", "seatid", seatid)
	conf := &OutCardConf{EnablePass: false, PowerType: OutCardConf_FirstOut}
	g.setPlayerOutCardPower(g.getPlayerBySeatId(seatid), conf)
}

func (g *Guandan) setPlayerOutCardPower(p *Player, conf *OutCardConf) {
	g.currActionPlayer = p
	p.getPowerAt = time.Now()
	p.powerDeadLine = p.getPowerAt.Add(time.Second * time.Duration(g.conf.OutCardTimeSec))

	notice := &NotifyPlayerOutCardPower{
		SeatId:   p.gameInfo.SeatId,
		Conf:     conf,
		Deadline: p.powerDeadLine.Unix(),
		Downtime: int32(g.conf.OutCardTimeSec),
	}

	p.outcardPower = notice
	g.broadcastMsg(notice)
}

func (g *Guandan) setNextOutCardPlayer(currPlayer *Player) {
	currPlayer.outcardPower = nil
	currSeatId := currPlayer.gameInfo.SeatId

	var nextplayer *Player = nil

	isWindflow := false

	for i := int32(0); i < MaxSeatCnt; i++ {
		nextseatid := (currSeatId + i + 1) % MaxSeatCnt
		target := g.getPlayerBySeatId(nextseatid)
		if target == nil {
			g.Error("get next player error")
			return
		}

		if target.handCards.Size() != 0 {
			nextplayer = target
			break
		}

		if g.outcardHeadPlayer == target {
			nextplayer = g.getPlayerBySeatId((target.raw.SeatID() + 2) % MaxSeatCnt)
			isWindflow = true
			break
		}

	}

	if nextplayer == nil {
		g.Error("no player can out card", "seatid", currSeatId)
		return
	}

	conf := &OutCardConf{}
	if isWindflow {
		conf.EnablePass = false
		conf.PowerType = OutCardConf_Windflow
	} else {
		conf.EnablePass = g.outcardHeadPlayer != nextplayer
	}
	g.setPlayerOutCardPower(nextplayer, conf)
}

func (g *Guandan) onGaming(delta time.Duration) {
	if g.currActionPlayer == nil {
		g.setFirstOutCardPlayer()
		return
	}

	currplayer := g.currActionPlayer
	isTimeout := time.Now().After(currplayer.powerDeadLine)
	if !isTimeout {
		return
	}
	g.playerActionTimeoutHelp(currplayer)
}

func (g *Guandan) playerActionTimeoutHelp(player *Player) {
	if player.outcardPower == nil {
		g.Error("player action timeout help error", "player", player.raw.UID())
		return
	}

	action := &ReqPlayerOutCards{}
	if player.outcardPower.Conf.EnablePass {
		action.OutCards = &OutCardInfo{
			DeckType: DeckType_Deck_Pass,
		}
	} else {
		cards := player.handCards.Back()
		action.OutCards = &OutCardInfo{
			DeckType: DeckType_Deck_Single,
			Cards:    []byte{byte(cards)},
		}
	}

	flag := g.playerOutCards(player, action.OutCards, nil)
	if flag != 0 {
		g.Warn("player out cards error", "player", player.raw.UID(), "flag", flag)
	}

}

func (g *Guandan) broadcastMsg(msg proto.Message) {
	msgid, err := gutils.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	if err != nil {
		g.Error("broadcastMsg get message error", "msgname", msg.ProtoReflect().Descriptor().Name(), "err", err)
		return
	}
	g.Info("broadcast message", "msgid", msgid, "msgname", msg.ProtoReflect().Descriptor().Name(), "msg", msg)
	g.table.BroadcastMessage(msgid, msg)
}

func (g *Guandan) sendMsg(p *Player, syn uint32, msg proto.Message) {
	msgid, err := gutils.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	if err != nil {
		g.Error("sendMsg get message error", "msgname", msg.ProtoReflect().Descriptor().Name(), "err", err)
		return
	}
	g.table.SendMessageToPlayer(p.raw, syn, msgid, msg)
}

func (g *Guandan) playerConv(p battle.Player) *Player {
	return g.getPlayerBySeatId(int32(p.SeatID()))
}

func (g *Guandan) getPlayerBySeatId(seatid int32) *Player {
	if seatid < 0 || seatid >= int32(len(g.players)) {
		g.Error("seatid out of range", "seatid", seatid, "players len", len(g.players))
		return nil
	}
	return g.players[seatid]
}

func (g *Guandan) changeLogicStep(nextStage *StageInfo) bool {
	if nextStage == nil {
		g.Error("next stage is nil")
		return false
	}

	currStage := g.currStage
	if currStage.StageType == nextStage.StageType {
		g.Error("set same step", "before", currStage.StageType, "now", nextStage.StageType)
		return false
	}

	currStage.OnExit()
	g.currStage = nextStage
	g.lastStage = currStage

	// donwtime := g.getStageDowntime(g.currStage.StageType).Seconds()
	donwtime := g.currStage.TimeToLive.Seconds()
	g.Info("game stage changed", "from", g.lastStage.StageType, "to", g.currStage.StageType, "donwtime", donwtime)
	g.currStage.OnBeforeEnter()

	notice := &NotifyGameStage{
		CurrStage: g.currStage.StageType,
		LastStage: g.lastStage.StageType,
		Downtime:  int32(donwtime),
		Deadline:  time.Now().Add(time.Second * time.Duration(donwtime)).Unix(),
	}
	g.broadcastMsg(notice)

	g.currStage.OnEnter()
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
