package battle

import (
	"github.com/google/uuid"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/errors"

	msgBattle "github.com/ajenpan/surf/msg/battle"
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
	log.Info("battle finished", "battleid", battleid)
}

func (h *Battle) OnReqStartBattle(ctx core.Context, in *msgBattle.ReqStartBattle) {
	var err error
	var resp = &msgBattle.RespStartBattle{}

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
		Logger: log.With("battleid", battleid),
	})

	err = d.Init(logic, players, in.GameConf)
	if err != nil {
		return
	}

	h.tables.Store(battleid, d)
	resp.BattleId = battleid
}

func (h *Battle) OnReqJoinBattle(ctx core.Context, in *msgBattle.ReqJoinBattle) {
	var err error

	out := &msgBattle.RespJoinBattle{
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
		return ctx.Async(&msgBattle.BattleMsgToClient{
			BattleId: in.BattleId,
			Msgid:    msgid,
			Data:     raw,
		})
	}

	d.OnPlayerConn(int64(ctx.FromUId()), sender, true)

	ctx.Response(out, err)

	h.UIDBindBattleID(int64(ctx.FromUId()), in.BattleId)
}

func (h *Battle) OnPlayerDisConn(uid uint32, gateNodeId uint32, reason int32) {
	log.Info("OnPlayerDisConn:", "uid", uid)

	tableids := h.LoadBattleByUID(int64(uid))

	for _, tableid := range tableids {
		d := h.getBattleById(tableid)
		if d == nil {
			continue
		}
		d.OnPlayerConn(int64(uid), nil, false)
	}
}

func (h *Battle) OnReqQuitBattle(ctx core.Context, in *msgBattle.ReqQuitBattle) {
	resp := &msgBattle.RespQuitBattle{
		BattleId: in.BattleId,
	}
	uid := ctx.FromUId()
	h.UIDUnBindBattleID(int64(uid), in.BattleId)
	ctx.Response(resp, nil)
}

func (h *Battle) OnBattleMsgToServer(ctx core.Context, in *msgBattle.BattleMsgToServer) {
	d := h.getBattleById(in.BattleId)
	if d == nil {
		log.Warn("battle not found", "battleid", in.BattleId)
		return
	}
	d.OnPlayerMessage(int64(ctx.FromUId()), in.Syn, in.Msgid, in.Data)
}
