package battle

import (
	"context"

	"github.com/google/uuid"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/errors"

	msgBattle "github.com/ajenpan/surf/msg/battle"
	"github.com/ajenpan/surf/server/battle/table"
)

func (h *Battle) OnStartBattle(ctx context.Context, in *msgBattle.ReqStartBattle, out *msgBattle.RespStartBattle) error {
	var err error
	var resp = &msgBattle.RespStartBattle{}

	logic, err := h.LogicCreator.CreateLogic(in.GameName)
	if err != nil {
		return err
	}

	players, err := table.NewPlayers(in.PlayerInfos)
	if err != nil {
		return err
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
		return err
	}

	h.tables.Store(battleid, d)
	resp.BattleId = battleid
	return nil
}

func (h *Battle) OnJoinBattle(ctx context.Context, in *msgBattle.ReqJoinBattle, out *msgBattle.RespJoinBattle) error {
	var err error

	out.BattleId = in.BattleId
	out.SeatId = in.SeatId
	out.ReadyState = in.ReadyState

	d := h.getBattleById(in.BattleId)
	if d == nil {
		err = errors.New(-1, "battle not found")
		return err
	}

	user, _ := core.CtxToUser(ctx)

	uid := user.UserID()

	sender := func(msgid uint32, raw []byte) error {
		return h.surf.SendAsyncToClient(uid, &msgBattle.BattleMsgToClient{
			BattleId: in.BattleId,
			Msgid:    msgid,
			Data:     raw,
		})
	}

	d.OnPlayerConn(int64(uid), sender, true)
	h.UIDBindBattleID(int64(uid), in.BattleId)
	return nil
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

func (h *Battle) OnQuitBattle(ctx context.Context, in *msgBattle.ReqQuitBattle, out *msgBattle.RespQuitBattle) error {
	user, _ := core.CtxToUser(ctx)
	uid := user.UserID()
	out.BattleId = in.BattleId
	h.UIDUnBindBattleID(int64(uid), in.BattleId)
	return nil
}

func (h *Battle) OnBattleMsgToServer(ctx context.Context, in *msgBattle.BattleMsgToServer) {
	d := h.getBattleById(in.BattleId)
	if d == nil {
		log.Warn("battle not found", "battleid", in.BattleId)
		return
	}
	user, _ := core.CtxToUser(ctx)
	uid := user.UserID()
	d.OnPlayerMessage(int64(uid), in.Syn, in.Msgid, in.Data)
}
