package niuniu

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	nncard "github.com/ajenpan/poker_algorithm/niuniu"
	protobuf "google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/utils/calltable"
	battle "github.com/ajenpan/surf/game"
	"github.com/ajenpan/surf/game/utils"
)

func NewLogic() battle.Logic {
	return NewNiuniu()
}

func NewNiuniu() *Niuniu {
	ret := &Niuniu{
		players: make(map[int32]*NNPlayer),
		info:    &GameInfo{},
		conf:    &Config{},
		Logger:  slog.Default().With("game", "niuniu"),
	}
	return ret
}

type NNPlayer struct {
	raw battle.Player
	*GamePlayerInfo
	rawHandCards *nncard.NNHandCards
	online       bool
}

type Config struct {
	DowntimeSec int
}

var defaultConf = &Config{
	DowntimeSec: 5,
}

func ParseConfig(raw []byte) (*Config, error) {
	ret := defaultConf
	if len(raw) <= 2 {
		return ret, nil
	}
	err := json.Unmarshal(raw, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type Niuniu struct {
	*slog.Logger

	table battle.Table
	conf  *Config

	info    *GameInfo
	players map[int32]*NNPlayer // seatid to player

	gameTime  time.Duration
	stageTime time.Duration

	CT *calltable.CallTable
}

func (nn *Niuniu) BroadcastMessage(msg protobuf.Message) {
	msgid, err := utils.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	if err != nil {
		nn.Error("get msgid error", "err", err)
		return
	}
	nn.table.BroadcastMessage(msgid, msg)
}

func (nn *Niuniu) Send2Player(p battle.Player, syn uint32, msg protobuf.Message) {
	msgid, err := utils.GetMessageMsgID(msg.ProtoReflect().Descriptor())
	if err != nil {
		nn.Error("get msgid error", "err", err)
		return
	}
	nn.table.SendMessageToPlayer(p, syn, msgid, msg)
}

func (nn *Niuniu) OnPlayerConnStatus(player battle.Player, enable bool) {

}

func (nn *Niuniu) OnPlayerEnter(p battle.Player, subtype battle.PlayerEnterSubType, extra []byte) {
	player := nn.playerConv(p)
	player.online = true

	if player.GameStep < GameStep_BEGIN {
		nn.ChangeLogicStep(GameStep_BEGIN)
	}
}

func (nn *Niuniu) OnPlayerLeave(p battle.Player, subtype battle.PlayerLeaveSubType, extra []byte) {
	player := nn.playerConv(p)
	player.online = false
}

func (nn *Niuniu) OnInit(opts battle.LogicOpts) error {
	if len(opts.Players) < 2 {
		return fmt.Errorf("player is not enrough")
	}

	if opts.Log != nil {
		nn.Logger = opts.Log
	}

	for _, v := range opts.Players {
		if _, err := nn.addPlayer(v); err != nil {
			return err
		}
	}

	var err error
	nn.conf, err = ParseConfig(opts.Conf)
	if err != nil {
		return err
	}

	nn.table = opts.Table
	nn.info = &GameInfo{
		GameStep: GameStep_IDLE,
	}
	nn.gameTime = 0

	return nil
}

func (nn *Niuniu) OnStart([]battle.Player) error {
	if len(nn.players) < 2 {
		return fmt.Errorf("player is not enrough")
	}

	nn.table.ReportBattleStatus(battle.GameStatus_Running)
	nn.ChangeLogicStep(GameStep_BEGIN)
	return nil
}

func (nn *Niuniu) OnCommand(topic string, data []byte) {

}

func (nn *Niuniu) OnPlayerMessage(p battle.Player, syn uint32, msgid uint32, raw []byte) {
	nn.Info("recv msgid", "msgid", msgid)
}

func (nn *Niuniu) OnEvent(topic string, event protobuf.Message) {

}

func (nn *Niuniu) OnReqGameInfo(p battle.Player, req *ReqGameInfo) {
	resp := &RespGameInfo{
		// GameConf: nn.conf.String(),
		GameInfo: nn.info,
	}
	nn.Send2Player(p, 0, resp)
}

func (nn *Niuniu) checkStat(p *NNPlayer, expect GameStep) error {
	if nn.getLogicStep() == expect {
		return fmt.Errorf("game status error")
	}
	if p.GameStep != previousStep(expect) {
		return fmt.Errorf("player status error")
	}
	return nil
}

func (nn *Niuniu) OnReqPlayerBanker(nnPlayer *NNPlayer, req *ReqPlayerBanker) {
	if err := nn.checkStat(nnPlayer, GameStep_BANKER); err != nil {
		return
	}
	notice := &NotifyPlayerBanker{
		SeatId: int32(nnPlayer.raw.SeatID()),
		Rob:    req.Rob,
	}
	nnPlayer.BankerRob = req.Rob
	nn.BroadcastMessage(notice)
}

func (nn *Niuniu) OnReqPlayerBetRate(p battle.Player, pMsg *ReqPlayerBetRate) {
	nnPlayer := nn.playerConv(p)
	if nnPlayer == nil {
		nn.Info("can't find player", "uid", p.UID())
		return
	}

	if err := nn.checkStat(nnPlayer, GameStep_BET); err != nil {
		return
	}

	nnPlayer.BetRate = pMsg.Rate
	nnPlayer.GameStep = GameStep_BET

	notice := &NotifyPlayerBetRate{
		SeatId: int32(p.SeatID()),
		Rate:   pMsg.Rate,
	}
	nn.BroadcastMessage(notice)
}

func (nn *Niuniu) OnReqPlayerOutCard(p battle.Player, pMsg *ReqPlayerOutCard) {
	nnPlayer := nn.playerConv(p)

	if nnPlayer == nil {
		nn.Error("OnPlayerOutCardRequest player is nil")
		return
	}

	if err := nn.checkStat(nnPlayer, GameStep_SHOW_CARDS); err != nil {
		return
	}

	nnPlayer.OutCard = &OutCardInfo{
		Cards: nnPlayer.rawHandCards.Bytes(),
		Type:  CardType(nnPlayer.rawHandCards.Type()),
	}
	nnPlayer.GameStep = GameStep_SHOW_CARDS

	notice := &NotifyPlayerOutCard{
		SeatId:  int32(p.SeatID()),
		OutCard: nnPlayer.OutCard,
	}

	nn.BroadcastMessage(notice)
}

func (nn *Niuniu) addPlayer(p battle.Player) (*NNPlayer, error) {
	ret := &NNPlayer{}
	ret.GamePlayerInfo = &GamePlayerInfo{}
	ret.GamePlayerInfo.SeatId = int32(p.SeatID())
	ret.raw = p
	ret.online = false
	if _, has := nn.players[int32(p.SeatID())]; has {
		return nil, fmt.Errorf("seat repeat")
	}
	nn.players[int32(p.SeatID())] = ret
	return ret, nil
}

func (nn *Niuniu) OnTick(duration time.Duration) {
	nn.gameTime += duration
	nn.stageTime += duration

	switch nn.getLogicStep() {
	case GameStep_UNKNOW:
		// do nothing
	case GameStep_IDLE:
		if nn.checkPlayerStep(GameStep_BEGIN) {
			nn.ChangeLogicStep(GameStep_BANKER)
		}
	case GameStep_BEGIN:
		nn.ChangeLogicStep(GameStep_BANKER)
	case GameStep_BANKER:
		if nn.StepTimeover() || nn.checkPlayerStep(GameStep_BANKER) {
			nn.ChangeLogicStep(GameStep_BANKER_NOTIFY)
		}
	case GameStep_BANKER_NOTIFY:
		if nn.StepTimeover() {
			nn.notifyRobBanker()
			nn.ChangeLogicStep(GameStep_BET)
		}
	case GameStep_BET: // 下注
		if nn.StepTimeover() || nn.checkPlayerStep(GameStep_BET) {
			nn.ChangeLogicStep(GameStep_DEAL_CARDS)
		}
	case GameStep_DEAL_CARDS: // 发牌
		nn.sendCardToPlayer()
		nn.ChangeLogicStep(GameStep_SHOW_CARDS)
	case GameStep_SHOW_CARDS: // 开牌
		if nn.StepTimeover() || nn.checkPlayerStep(GameStep_SHOW_CARDS) {
			nn.ChangeLogicStep(GameStep_TALLY)
		}
	case GameStep_TALLY:
		nn.beginTally()
		nn.NextStep()
	case GameStep_OVER:
		if nn.StepTimeover() {
			nn.table.ReportBattleStatus(battle.GameStatus_Over)
			nn.NextStep()
		}
	default:
		//warn
	}
}

func (nn *Niuniu) OnReset() {

}

func (nn *Niuniu) getLogicStep() GameStep {
	return nn.info.GameStep
}

func (nn *Niuniu) getStageDowntime(s GameStep) time.Duration {
	return time.Duration(nn.conf.DowntimeSec) * time.Second
}

func nextStep(status GameStep) GameStep {
	nextStep := status + 1
	if nextStep > GameStep_OVER {
		nextStep = GameStep_UNKNOW
	}
	return nextStep
}

func previousStep(status GameStep) GameStep {
	previousStatus := status - 1
	if previousStatus < GameStep_UNKNOW {
		previousStatus = GameStep_OVER
	}
	return previousStatus
}

func (nn *Niuniu) NextStep() {
	nn.ChangeLogicStep(nextStep(nn.getLogicStep()))
}

func (nn *Niuniu) ChangeLogicStep(s GameStep) {
	lastStatus := nn.getLogicStep()
	nn.info.GameStep = s

	if lastStatus != s {
		//reset stage time
		nn.stageTime = 0
	}

	donwtime := nn.getStageDowntime(s).Seconds()

	nn.Info("game step changed", "before", lastStatus, "now", s)

	if lastStatus == s {
		nn.Error("set same step", "before", lastStatus, "now", s)
	}

	if lastStatus != GameStep_OVER {
		if lastStatus > s {
			nn.Error("last step is bigger than now", "before", lastStatus, "now", s)
		}
	}

	notice := &NotifyGameStep{
		GameStep: s,
		TimeDown: int32(donwtime),
	}

	nn.BroadcastMessage(notice)
}

func (nn *Niuniu) playerConv(p battle.Player) *NNPlayer {
	return nn.getPlayerBySeatId(int32(p.SeatID()))
}

func (nn *Niuniu) getPlayerBySeatId(seatid int32) *NNPlayer {
	p, ok := nn.players[seatid]
	if ok {
		return p
	}
	return nil
}

func (nn *Niuniu) StepTimeover() bool {
	return nn.stageTime >= nn.getStageDowntime(nn.info.GameStep)
}

func (nn *Niuniu) checkPlayerStep(expect GameStep) bool {
	for _, p := range nn.players {
		if p.GameStep != expect {
			return false
		}
	}
	return true
}

func (nn *Niuniu) checkEndBanker() bool {
	for _, p := range nn.players {
		if p.BankerRob == 0 {
			return false
		}
	}
	return true
}

func (nn *Niuniu) notifyRobBanker() {
	for _, p := range nn.players {
		if p.GameStep != GameStep_BANKER {
			p.GameStep = GameStep_BANKER
		}
	}

	seats := []int32{}
	var maxRob int32 = -1
	for _, p := range nn.players {
		if (p.BankerRob) > maxRob {
			maxRob = p.BankerRob
			seats = seats[:0]
			seats = append(seats, p.SeatId)
		} else if (p.BankerRob) == maxRob {
			seats = append(seats, p.SeatId)
		}
	}

	if len(seats) == 0 {
		nn.Error("选庄错误", "maxrob", maxRob)
	}

	index := rand.Intn(len(seats))
	bankSeatId := seats[index]
	banker, ok := nn.players[int32(bankSeatId)]

	if !ok {
		nn.Error("banker seatid error", "seatid", bankSeatId, "index", index)
		return
	}

	banker.Banker = true
	//庄家不参与下注.提前设置好状态
	banker.GameStep = GameStep_BET

	notice := &NotifyBankerSeat{
		SeatId: bankSeatId,
	}

	nn.BroadcastMessage(notice)
}

func (nn *Niuniu) sendCardToPlayer() {
	deck := nncard.NewNNDeck()
	deck.Shuffle()

	for _, p := range nn.players {
		p.rawHandCards = deck.DealHandCards()
		p.HandCards = p.rawHandCards.Bytes()
		p.GameStep = GameStep_DEAL_CARDS
		notice := &NotifyPlayerHandCards{
			SeatId:    p.SeatId,
			HandCards: p.HandCards,
		}
		nn.Send2Player(p.raw, 0, notice)
	}

	for _, p := range nn.players {
		p.rawHandCards.Calculate()
	}
}

func (nn *Niuniu) beginTally() {
	var banker *NNPlayer = nil

	for _, p := range nn.players {
		if p.Banker {
			banker = p
			break
		}
	}
	if banker == nil {
		nn.Error("bank is nil")
		return
	}

	notify := &NotifyGameTally{}
	// notify.TallInfo = make([]*PlayerTallyNotify_TallyInfo, 0)
	// type tally struct {
	// 	UserId int64
	// 	Coins  int32
	// }

	bankerTally := &NotifyGameTally_TallyInfo{
		SeatId: banker.SeatId,
		//Coins:  chips*cardRate*p.BetRate - 100,
	}

	for _, p := range nn.players {
		if p.Banker {
			continue
		}
		var chips int32 = 5
		var cardRate int32 = 1

		if banker.rawHandCards.Compare(p.rawHandCards) {
			//底注*倍率*牌型倍率
			cardRate += int32(banker.rawHandCards.Type())
			cardRate = -cardRate
		} else {
			cardRate += int32(p.rawHandCards.Type())
		}
		temp := &NotifyGameTally_TallyInfo{
			SeatId: p.SeatId,
			Coins:  chips * cardRate * p.BetRate,
		}
		// notify.TallInfo = append(notify.TallInfo, temp)
		bankerTally.Coins += temp.Coins
	}

	// notify.TallInfo = append(notify.TallInfo, bankerTally)

	nn.BroadcastMessage(notify)
}

func (nn *Niuniu) resetDesk() {
	nn.players = make(map[int32]*NNPlayer)
	nn.ChangeLogicStep(GameStep_IDLE)
}

func (nn *Niuniu) PrintDebufInfo() {
	nn.Debug(nn.info.String())
}
