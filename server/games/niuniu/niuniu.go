package niuniu

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	nncard "github.com/ajenpan/poker_algorithm/niuniu"
	"github.com/sirupsen/logrus"
	protobuf "google.golang.org/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/server/battle"
)

func CreateLogic() battle.Logic {
	return CreateNiuniu()
}

func CreateNiuniu() *Niuniu {
	ret := &Niuniu{
		players: make(map[int32]*NNPlayer),
		info:    &GameInfo{},
		conf:    &Config{},
	}
	return ret
}

func GetMessageMsgID(msg protoreflect.MessageDescriptor) uint32 {
	MSGIDDesc := msg.Enums().ByName("MSGID")
	if MSGIDDesc == nil {
		return 0
	}
	IDDesc := MSGIDDesc.Values().ByName("ID")
	if IDDesc == nil {
		return 0
	}
	return uint32(IDDesc.Number())
}

func init() {
	battle.RegisterGame("niuniu", "1.0.0", CreateLogic)
}

type NNPlayer struct {
	raw battle.Player
	*GamePlayerInfo
	rawHandCards *nncard.NNHandCards
}

type Config struct {
	Downtime time.Duration
}

func ParseConfig(raw []byte) (*Config, error) {
	ret := &Config{}
	err := json.Unmarshal(raw, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type Niuniu struct {
	table battle.Table
	conf  *Config
	log   *logrus.Logger

	info    *GameInfo
	players map[int32]*NNPlayer // seatid to player

	gameTime  time.Duration
	stageTime time.Duration

	CT *calltable.CallTable[int]
}

func (nn *Niuniu) BroadcastMessage(msg protobuf.Message) {

}

func (nn *Niuniu) Send2Player(p battle.Player, msg protobuf.Message) {

}

func (nn *Niuniu) OnPlayerJoin(players []battle.Player) error {
	if len(players) == 0 {
		return nil
	}
	for _, v := range players {
		if _, err := nn.addPlayer(v); err != nil {
			return err
		}
	}
	return nil
}

func (nn *Niuniu) OnInit(d battle.Table, conf interface{}) error {
	switch v := conf.(type) {
	case []byte:
		var err error
		nn.conf, err = ParseConfig(v)
		if err != nil {
			return err
		}
	case *Config:
		nn.conf = v
	default:
		return fmt.Errorf("unknow config type ")
	}
	nn.table = d
	nn.info = &GameInfo{
		Status: GameStep_IDLE,
	}
	nn.gameTime = 0
	return nil
}

func (nn *Niuniu) OnStart([]battle.Player) error {
	if len(nn.players) < 2 {
		return fmt.Errorf("player is not enrough")
	}

	nn.table.ReportBattleStatus(battle.BattleStatus_Running)
	nn.ChangeLogicStep(GameStep_BEGIN)
	return nil
}

func (nn *Niuniu) OnCommand(topic string, data []byte) {

}

func (nn *Niuniu) OnPlayerMessage(p battle.Player, msgid uint32, raw []byte) {
	nn.log.Infof("recv msgid:%d", msgid)
}

func (nn *Niuniu) OnEvent(topic string, event protobuf.Message) {

}

func (nn *Niuniu) GameDeskInfoRequest(p battle.Player, req *GameDeskInfoRequest) {
	resp := &GameDeskInfoResponse{
		Info: nn.info,
	}
	nn.Send2Player(p, resp)
}

func (nn *Niuniu) checkStat(p *NNPlayer, expect GameStep) error {
	if nn.getLogicStep() == expect {
		return fmt.Errorf("game status error")
	}
	if p.Status != previousStep(expect) {
		return fmt.Errorf("player status error")
	}
	return nil
}

func (nn *Niuniu) OnPlayerBankerRequest(nnPlayer *NNPlayer, req *PlayerBanker) {
	if err := nn.checkStat(nnPlayer, GameStep_BANKER); err != nil {
		return
	}
	notice := &PlayerBankerNotify{
		SeatId: int32(nnPlayer.raw.SeatID()),
		Rob:    req.Rob,
	}
	nnPlayer.BankerRob = req.Rob
	nn.BroadcastMessage(notice)
}

func (nn *Niuniu) OnPlayerBetRateRequest(p battle.Player, pMsg *PlayerBetRate) {
	nnPlayer := nn.playerConv(p)
	if nnPlayer == nil {
		nn.log.Infof("can't find player uid :%d", p.SeatID())
		return
	}

	if err := nn.checkStat(nnPlayer, GameStep_BET); err != nil {
		return
	}

	nnPlayer.BetRate = pMsg.Rate
	nnPlayer.Status = GameStep_BET

	notice := &PlayerBetRateNotify{
		SeatId: int32(p.SeatID()),
		Rate:   pMsg.Rate,
	}
	nn.BroadcastMessage(notice)
}

func (nn *Niuniu) OnPlayerOutCardRequest(p battle.Player, pMsg *PlayerOutCard) {
	nnPlayer := nn.playerConv(p)

	if nnPlayer == nil {
		nn.log.Errorf("OnPlayerOutCardRequest player is nil")
		return
	}

	if err := nn.checkStat(nnPlayer, GameStep_SHOW_CARDS); err != nil {
		return
	}

	nnPlayer.OutCard = &OutCardInfo{
		Cards: nnPlayer.rawHandCards.Bytes(),
		Type:  CardType(nnPlayer.rawHandCards.Type()),
	}
	nnPlayer.Status = GameStep_SHOW_CARDS

	notice := &PlayerOutCardNotify{
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
		fallthrough
	case GameStep_IDLE:
		//do nothing, when the game create but not start
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
			nn.table.ReportBattleStatus(battle.BattleStatus_Over)
			nn.NextStep()
		}
	default:
		//warn
	}
}

func (nn *Niuniu) OnReset() {

}

func (nn *Niuniu) getLogicStep() GameStep {
	return nn.info.Status
}

func (nn *Niuniu) getStageDowntime(s GameStep) time.Duration {
	//TODO:
	return nn.conf.Downtime
}

func nextStep(status GameStep) GameStep {
	nextStep := status + 1
	if nextStep > GameStep_OVER {
		nextStep = GameStep_IDLE
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
	nn.info.Status = s

	if lastStatus != s {
		//reset stage time
		nn.stageTime = 0
	}

	donwtime := nn.getStageDowntime(s).Seconds()

	nn.log.Infof("game step changed, before:%v, now:%v ", lastStatus, s)

	if lastStatus == s {
		nn.log.Errorf("set same step before:%v, now:%v", lastStatus, s)
	}

	if lastStatus != GameStep_OVER {
		if lastStatus > s {
			nn.log.Errorf("last step is bigger than now before:%v, now:%v", lastStatus, s)
		}
	}

	notice := &GameStatusNotify{
		GameStatus: s,
		TimeDown:   int32(donwtime),
	}

	nn.BroadcastMessage(notice)

	nn.Debug()
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
	return nn.stageTime >= nn.getStageDowntime(nn.info.Status)
}

func (nn *Niuniu) checkPlayerStep(expect GameStep) bool {
	for _, p := range nn.players {
		if p.Status != expect {
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
		if p.Status != GameStep_BANKER {
			p.Status = GameStep_BANKER
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
		nn.log.Errorf("选庄错误 maxrob:%d", maxRob)
	}

	index := rand.Intn(len(seats))
	bankSeatId := seats[index]
	banker, ok := nn.players[int32(bankSeatId)]

	if !ok {
		nn.log.Errorf("banker seatid error. seatid:%d,index:%d", bankSeatId, index)
		return
	}

	banker.Banker = true
	//庄家不参与下注.提前设置好状态
	banker.Status = GameStep_BET

	notice := &BankerSeatNotify{
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
		p.Status = GameStep_DEAL_CARDS
		notice := &PlayerHandCardsNotify{
			SeatId:    p.SeatId,
			HandCards: p.HandCards,
		}
		nn.Send2Player(p.raw, notice)
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
		nn.log.Errorf("bank is nil")
		return
	}

	notify := &PlayerTallyNotify{}
	// notify.TallInfo = make([]*PlayerTallyNotify_TallyInfo, 0)
	// type tally struct {
	// 	UserId int64
	// 	Coins  int32
	// }

	bankerTally := &PlayerTallyNotify_TallyInfo{
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
		temp := &PlayerTallyNotify_TallyInfo{
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
	for _, p := range nn.players {
		p.GamePlayerInfo.Reset()
		p.GamePlayerInfo.Status = GameStep_IDLE
		p.GamePlayerInfo.SeatId = p.SeatId
	}
	nn.ChangeLogicStep(GameStep_IDLE)
}

func (nn *Niuniu) Debug() {
	// nn.log.Debug(nn.info.String())
	fmt.Println(nn.info.String())
}
