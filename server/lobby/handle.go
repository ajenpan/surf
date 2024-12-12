package lobby

import (
	"fmt"
	"time"

	"github.com/ajenpan/surf/core"
	msgBattle "github.com/ajenpan/surf/msg/battle"
	msgLobby "github.com/ajenpan/surf/msg/lobby"
	"github.com/ajenpan/surf/server"
	"github.com/google/uuid"
)

func (h *Lobby) OnClientConnect(uid uint32, gateNodeId uint32, ip string) {

}

func (h *Lobby) OnClientDisconnect(uid uint32, gateNodeId uint32, reason int32) {

}

func (h *Lobby) OnReqLoginLobby(ctx core.Context, req *msgLobby.ReqLoginLobby) {
	resp := &msgLobby.RespLoginLobby{}
	var err error

	uid := ctx.FromUId()
	if ctx.FromURole() != 0 {
		uid = req.Uid
	}

	user := h.getUser(uid)
	var table *Table = nil

	if user == nil {
		user = NewUser(uid)
	} else {
		table = TableStoreInstance.FindTable(user.PlayInfo.tuid)

		switch user.PlayInfo.PlayerStatus {
		case msgLobby.PlayerStatus_PlayerNone:
			// do nothing
		case msgLobby.PlayerStatus_PlayerInQueue:
			h.LeaveDispatchQue(user)
		case msgLobby.PlayerStatus_PlayerInTable:
			fallthrough
		case msgLobby.PlayerStatus_PlayerInTableReady:
			if table != nil {
				h.DismissTable(table)
			} else {
				user.PlayInfo.PlayerStatus = msgLobby.PlayerStatus_PlayerNone
			}
		case msgLobby.PlayerStatus_PlayerInGaming:
			if req.GameRoomId != int32(user.PlayInfo.gameRoomId) {
				resp.Flag = 1
				ctx.Response(resp, nil)
				return
			}
		default:
			log.Error("uknown status")
		}
	}

	user.ConnInfo.Sender = ctx.Async
	user.GameInfo.GameId = req.GameId
	user.PlayInfo.gameRoomId = req.GameRoomId

	isReconnect := table != nil && user.PlayInfo.PlayerStatus == msgLobby.PlayerStatus_PlayerInGaming

	log.Info("on user login", "uid", uid, "roomid", req.GameRoomId)

	if isReconnect {
		h.addLoginUser(user)

		user.MutableRespLoginLobby(resp)
		ctx.Response(resp, nil)

		notify := table.MutableNotifyDispatchResult()
		user.Send(notify)
		return
	}

	err = user.Init()
	if err != nil {
		ctx.Response(resp, err)
		return
	}

	err = h.uLoign.loadOrStore(uid)
	if err != nil {
		resp.Flag = msgLobby.RespLoginLobby_kInOtherRoom
		ctx.Response(resp, nil)
		return
	}

	h.addLoginUser(user)

	// baseinfo, err := h.GetUserGameInfo(uid)
	// if err != nil {
	// 	log.Error("get user game info error", "error", err, "uid", uid)
	// 	resperr = err
	// 	return
	// }
	// resp.BaseInfo = baseinfo
	user.MutableRespLoginLobby(resp)
	ctx.Response(resp, nil)
}

func (h *Lobby) OnReqJoinQue(ctx core.Context, req *msgLobby.ReqJoinQue) {
	uid := ctx.FromUId()
	user := h.getLoginUser(uid)
	resp := &msgLobby.RespJoinQue{}
	var herr error

	defer func() { ctx.Response(resp, herr) }()

	if user == nil {
		herr = fmt.Errorf("user not found %d", uid)
		return
	}

	que := h.getQue(user.PlayInfo.gameRoomId)

	if que == nil {
		herr = fmt.Errorf("que not found roomid:%d", user.PlayInfo.gameRoomId)
		return
	}

	currState := user.PlayInfo.PlayerStatus
	if currState == msgLobby.PlayerStatus_PlayerInGaming ||
		currState == msgLobby.PlayerStatus_PlayerInQueue {
		herr = fmt.Errorf("player state err %d", currState)
		return
	}
	needJoinQue := true

	if user.PlayInfo.tidx != 0 {
		table := h.FindContiTable(user.PlayInfo.tidx)
		if req.JoinType != msgLobby.ReqJoinQue_Noraml && table != nil {
			needJoinQue = false

			if herr = table.AddContinuePlayer(user); herr != nil {
				return
			}

			ok := table.checkStartCondi()

			if ok {
				h.surf.Do(func() {
					h.RemoveContiTable(table.idx)
					table.keepOnUsers = make(map[uint32]*User)
					h.DoTableStart(table)
				})
				return
			}

			if table.keepOnTimer != nil {
				table.keepOnTimer.Stop()
			}

			table.keepOnTimer = time.AfterFunc(10*time.Second, func() {
				h.surf.Do(func() {
					h.DismissTable(table)
				})
			})

		} else {
			if table != nil {
				h.DismissTable(table)
			}
		}
	}

	if needJoinQue {
		err := que.Add(user)
		if err != nil {
			herr = err
			return
		}
	}
}

func (h *Lobby) OnReqLogoutLobby(ctx core.Context, req *msgLobby.ReqLogoutLobby) {
	uid := ctx.FromUId()
	if ctx.FromURole() != 0 {
		uid = req.Uid
	}
	h.delLoginUser(uid)
	ctx.Response(&msgLobby.RespLogoutLobby{}, nil)
}

func (h *Lobby) newPlayID(_ TableIdxT) string {
	return uuid.NewString()
}

func (h *Lobby) DoTableStart(table *Table) error {
	table.status = msgLobby.TableStatus_TableInCreating
	table.playid = h.newPlayID(table.idx)
	// 提前锁住
	for _, user := range table.users {
		user.PlayInfo.PlayerStatus = msgLobby.PlayerStatus_PlayerInGaming
	}

	onFailed := func(table *Table, flag int32, err error) {
		// 		FLOGE("DoTableStart faild flag:{},tidx:{},tappid:{},tid:{},playid:{},player:{},msg:{}",
		// 		flag, pTable->tableIdx, pTable->logic_appid, pTable->tableid, pTable->playid, fmt::join(pTable->m_players, "-"), errmsg);
		notify := &msgLobby.NotifyDispatchResult{
			Flag: 1,
		}
		table.BroadcastMessage(notify)
		h.DismissTable(table)
	}

	onSuccess := func(table *Table) {

		onDeadline := func() {
			// 	auto pTable = TableStore::Instance().FindTable(tuid);
			// 	if (pTable == nullptr) {
			// 		FLOGW("table not found:{}", tuid);
			// 		return;
			// 	}
			// 	if (pTable->playid != playid) {
			// 		FLOGW("table playid has chg:{},{}", pTable->playid, playid);
			// 		return;
			// 	}
			// 	FLOGW("onTallyTimeOver tidx:{},lappid:{},ltid:{},playid:{},tstatus:{}",
			// 		  pTable->tableIdx, pTable->logic_appid, pTable->tableid, pTable->playid, (int)pTable->GetStatus());
			// 	TableStore::Instance().RemoveTable(tuid);
			// 	DismissTable(pTable);
			// }
		}

		if table.deadlineTimer != nil {
			table.deadlineTimer.Stop()
		}
		table.deadlineTimer = time.AfterFunc(time.Second, onDeadline)

		table.status = msgLobby.TableStatus_TableInInGaming

		table.keepOnTimer = nil
		table.keepOnAt = 0

		notify := table.MutableNotifyDispatchResult()

		for _, user := range table.users {
			h.inTableUsers[user.UserId] = user
			user.Send(notify)
		}

	}

	h.TallyTableFee(table, func(table *Table, err error) {
		if err != nil {
			onFailed(table, 1, err)
			return
		}
		h.NewRemoteTable(table, 3, func(table *Table, err error) {
			if err != nil {
				onFailed(table, 2, err)
				return
			}
			onSuccess(table)
		})
	})
	return nil
}

func (h *Lobby) TallyTableFee(table *Table, fn func(table *Table, err error)) {
	// h.WGameDB.
	// h.WGameDB.raw
	var err error
	if h.banker != nil {
		for _, user := range table.users {
			err = h.banker.UpdateUserProp(user.UserId, 1, -100)
			if err != nil {
				break
			}
		}
	}

	if fn != nil {
		fn(table, err)
	}
}

func (h *Lobby) NewRemoteTable(table *Table, trycnt int, fn func(table *Table, err error)) {
	if trycnt <= 0 {
		fn(table, fmt.Errorf("new remote table failed"))
		return
	}
	req := &msgBattle.ReqStartBattle{
		GameName:    table.game.Name,
		GameConf:    table.game.DefaultConf,
		TableConf:   nil,
		PlayerInfos: make([]*msgBattle.PlayerInfo, 0, len(table.users)),
		Playid:      table.playid,
	}

	for _, user := range table.users {
		req.PlayerInfos = append(req.PlayerInfos, &msgBattle.PlayerInfo{
			Uid:    int64(user.UserId),
			SeatId: user.PlayInfo.SeatId,
			Score:  user.GameInfo.PlayScore,
			Role:   user.UType,
		})
	}

	err := core.SendRequestToNode(h.surf, server.NodeType_Battle, 0, req, func(result *core.ResponseResult, pk *msgBattle.RespStartBattle) {
		if result.Failed() {
			h.NewRemoteTable(table, trycnt-1, fn)
			return
		}
	})

	if err != nil {
		h.NewRemoteTable(table, trycnt-1, fn)
	}
}
