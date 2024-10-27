package handler

import (
	"sync"

	"github.com/google/uuid"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/event"
	log "github.com/ajenpan/surf/core/log"
	battlemsg "github.com/ajenpan/surf/msg/battle"

	"github.com/ajenpan/surf/server/battle"
	"github.com/ajenpan/surf/server/battle/table"
)

type Battle struct {
	tables       sync.Map
	LogicCreator *battle.GameLogicCreator
	Publisher    event.Publisher
}

func New() *Battle {
	h := &Battle{
		LogicCreator: battle.LogicCreator,
	}
	return h
}

func (h *Battle) ServerType() uint16 {
	return 1
}

func (h *Battle) ServerName() string {
	return "battle"
}

func (h *Battle) OnReqStartBattle(ctx core.Context, in *battlemsg.ReqStartBattle) {
	var err error
	var resp = &battlemsg.RespStartBattle{}

	defer func() {
		ctx.Response(resp, err)
	}()

	logic, err := h.LogicCreator.CreateLogic(in.GameName)
	if err != nil {
		return
	}

	players, err := table.NewPlayers(in.PlayerInfos)
	if err != nil {
		return
	}

	battleid := uuid.NewString()

	d := table.NewTable(table.TableOption{
		ID:             battleid,
		Conf:           in.TableConf,
		EventPublisher: h.Publisher,
		FinishReporter: func() {
			h.onBattleFinished(battleid)
		},
	})

	err = d.Init(logic, players, in.GameConf)
	if err != nil {
		return
	}

	h.tables.Store(battleid, d)

	resp.BattleId = d.ID
}

func (h *Battle) onBattleFinished(battleid string) {
	d := h.getBattleById(battleid)
	if d == nil {
		return
	}

	d.Players.Range(func(p *table.Player) bool {
		h.UIDUnBindBattleID(uint64(p.Uid), battleid)
		return true
	})
	h.tables.Delete(battleid)
	d.Close()
	log.Infof("battle %s finished", battleid)
}

func (h *Battle) UIDBindBattleID(uid uint64, bid string) error {
	// TODO:
	return nil
}

func (h *Battle) UIDUnBindBattleID(uid uint64, bid string) {
	// TODO:
}

func (h *Battle) LoadBattleByUID(uid uint64) map[string]*table.Table {
	// TODO:
	return nil
}

func (h *Battle) OnReqJoinBattle(ctx core.Context, in *battlemsg.ReqJoinBattle) {
	var err error

	out := &battlemsg.RespJoinBattle{
		BattleId:   in.BattleId,
		SeatId:     in.SeatId,
		ReadyState: in.ReadyState,
	}

	d := h.getBattleById(in.BattleId)
	if d == nil {
		err = errors.New(-1, "battle not found")
		ctx.Response(out, err)
		return
	}

	d.OnPlayerConn(uint64(ctx.Caller()), true)

	ctx.Response(out, err)

	// 是否需要保持 uid - battleid 映射?
	// 1 uid -> n * battleid.
	// 当uid掉线时, 需要遍历所有的battleid, 并且通知battleid.
	h.UIDBindBattleID(uint64(ctx.Caller()), in.BattleId)
}

func (h *Battle) OnPlayerDisConn(uid uint64) {
	log.Info("OnPlayerDisConn:", uid)

	tables := h.LoadBattleByUID(uid)
	for _, t := range tables {
		t.OnPlayerConn(uid, false)
	}
}

func (h *Battle) OnReqQuitBattle(ctx core.Context, in *battlemsg.ReqQuitBattle) {
	resp := &battlemsg.RespQuitBattle{
		BattleId: in.BattleId,
	}
	uid := ctx.Caller()
	h.UIDUnBindBattleID(uint64(uid), in.BattleId)
	ctx.Response(resp, nil)
}

func (h *Battle) OnBattleMsgToServer(ctx core.Context, in *battlemsg.BattleMsgToServer) {
	d := h.getBattleById(in.BattleId)
	if d == nil {
		log.Warnf("battle %s not found", in.BattleId)
		return
	}
	d.OnPlayerMessage(uint64(ctx.Caller()), in.Msgid, in.Data)
}

func (h *Battle) getBattleById(battleId string) *table.Table {
	if raw, ok := h.tables.Load(battleId); ok {
		return raw.(*table.Table)
	}
	return nil
}
