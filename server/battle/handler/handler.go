package handler

import (
	"github.com/google/uuid"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/errors"
	log "github.com/ajenpan/surf/core/log"
	battlemsg "github.com/ajenpan/surf/msg/battle"
	"github.com/ajenpan/surf/server/battle/table"
)

func (h *Battle) reportBattleFinished(battleid string) {
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

	d := table.NewTable(table.TableOptions{
		ID:             battleid,
		Conf:           in.TableConf,
		EventPublisher: h.Publisher,
		FinishReporter: func() {
			h.reportBattleFinished(battleid)
		},
	})

	err = d.Init(logic, players, in.GameConf)
	if err != nil {
		return
	}

	h.tables.Store(battleid, d)
	resp.BattleId = battleid
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

	h.UIDBindBattleID(uint64(ctx.Caller()), in.BattleId)
}

func (h *Battle) OnPlayerDisConn(uid uint64) {
	log.Info("OnPlayerDisConn:", uid)
	tableids := h.LoadBattleByUID(uid)

	for _, tableid := range tableids {
		d := h.getBattleById(tableid)
		if d == nil {
			continue
		}
		d.OnPlayerConn(uid, false)
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
