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
		h.UIDUnBindBattleID(int64(p.UID()), battleid)
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

	sender := func(msgid uint32, raw []byte) error {
		return ctx.SendAsync(&battlemsg.BattleMsgToClient{
			BattleId: in.BattleId,
			Msgid:    msgid,
			Data:     raw,
		})
	}

	d.OnPlayerConn(int64(ctx.FromUserID()), sender, true)

	ctx.Response(out, err)

	h.UIDBindBattleID(int64(ctx.FromUserID()), in.BattleId)
}

func (h *Battle) OnPlayerDisConn(uid int64) {
	log.Info("OnPlayerDisConn:", uid)
	tableids := h.LoadBattleByUID(uid)

	for _, tableid := range tableids {
		d := h.getBattleById(tableid)
		if d == nil {
			continue
		}
		d.OnPlayerConn(uid, nil, false)
	}
}

func (h *Battle) OnReqQuitBattle(ctx core.Context, in *battlemsg.ReqQuitBattle) {
	resp := &battlemsg.RespQuitBattle{
		BattleId: in.BattleId,
	}
	uid := ctx.FromUserID()
	h.UIDUnBindBattleID(int64(uid), in.BattleId)
	ctx.Response(resp, nil)
}

func (h *Battle) OnBattleMsgToServer(ctx core.Context, in *battlemsg.BattleMsgToServer) {
	d := h.getBattleById(in.BattleId)
	if d == nil {
		log.Warnf("battle %s not found", in.BattleId)
		return
	}
	d.OnPlayerMessage(int64(ctx.FromUserID()), in.Msgid, in.Data)
}
