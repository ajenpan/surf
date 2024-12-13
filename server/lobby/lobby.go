package lobby

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/ajenpan/surf/core"
	msgBattle "github.com/ajenpan/surf/msg/battle"
	msgLobby "github.com/ajenpan/surf/msg/lobby"
	"github.com/ajenpan/surf/server"
)

func NodeType() core.NodeType {
	return server.NodeType_Lobby
}

func NodeName() string {
	return server.NodeName_Lobby
}

var log = slog.Default()

func New() *Lobby {
	return &Lobby{
		loginUsers:   make(map[uint32]*User),
		inTableUsers: make(map[uint32]*User),
	}
}

type Lobby struct {
	WGameDB *gorm.DB
	WRds    *redis.Client
	surf    *core.Surf

	loginUsers   map[uint32]*User
	inTableUsers map[uint32]*User
	contiTable   map[TableIdxT]*Table
	matchQues    map[int32]DispatchQue

	uLoign UserUniqueLogin
	banker *Banker

	usersIp map[uint32]string
}

func (h *Lobby) OnInit(surf *core.Surf) (err error) {
	cfg := DefaultConf

	h.WRds = core.NewRdsClient(cfg.WRedisDSN)
	h.WGameDB = core.NewMysqlClient(cfg.WGameDBDSN)

	h.uLoign.NodeId = surf.NodeID()
	h.uLoign.NodeType = surf.NodeType()
	h.uLoign.Rds = h.WRds

	core.HandleRequestFromConn(surf, h.OnReqLoginLobby)
	core.HandleRequestFromConn(surf, h.OnReqLogoutLobby)
	core.HandleRequestFromConn(surf, h.OnReqJoinQue)

	h.surf = surf
	return nil
}

func (h *Lobby) OnReady() error {
	return nil
}

func (h *Lobby) OnStop() error {
	return nil
}

func (h *Lobby) GetUserGameInfo(uid uint32) (*msgLobby.UserBaseInfo, error) {
	info := &msgLobby.UserBaseInfo{}
	return info, nil
}

func (h *Lobby) GetPropInfo(uid uint32) (map[int]int64, error) {
	return nil, nil
}

func (h *Lobby) getLoginUser(uid uint32) *User {
	return h.loginUsers[uid]
}

func (h *Lobby) addLoginUser(u *User) {
	h.loginUsers[u.UserId] = u
}

func (h *Lobby) getUser(uid uint32) *User {
	u := h.getLoginUser(uid)
	if u != nil {
		return u
	}
	return h.inTableUsers[uid]
}

func (h *Lobby) delLoginUser(uid uint32) {
	user, has := h.loginUsers[uid]
	if !has {
		return
	}

	delPlace := true

	switch user.PlayInfo.PlayerStatus {
	case msgLobby.PlayerStatus_PlayerNone:
		// do nothing
	case msgLobby.PlayerStatus_PlayerInQueue:
		h.LeaveDispatchQue(user)
	case msgLobby.PlayerStatus_PlayerInTable:
		fallthrough
	case msgLobby.PlayerStatus_PlayerInTableReady:
		table := TableStoreInstance.FindTable(user.PlayInfo.tuid)
		if table != nil {
			h.DismissTable(table)
		}
	case msgLobby.PlayerStatus_PlayerInGaming:
		delPlace = false
		user.ConnInfo.GateNodeId = 0
	default:
		log.Error("uknown status")
	}

	if delPlace {
		h.uLoign.Del(uid)
		// user.PlayInfo.PlayerStatus = msgLobby.PlayerStatus_PlayerNone
	}
}

func (h *Lobby) LeaveDispatchQue(user *User) {

}

func (h *Lobby) DismissTable(table *Table) {

}

func (h *Lobby) getQue(roomid int32) *DispatchQue {
	return nil
}

func (h *Lobby) StoreContiTable(table *Table) {
	h.contiTable[table.idx] = table
}

func (h *Lobby) FindContiTable(tidx TableIdxT) *Table {
	return h.contiTable[tidx]
}

func (h *Lobby) RemoveContiTable(tidx TableIdxT) {
	delete(h.contiTable, tidx)
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
