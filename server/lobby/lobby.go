package lobby

import (
	"log/slog"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/ajenpan/surf/core"
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
}

func (h *Lobby) OnInit(surf *core.Surf) (err error) {
	cfg := DefaultConf

	h.WRds = core.NewRdsClient(cfg.WRedisDSN)
	h.WGameDB = core.NewMysqlClient(cfg.WGameDBDSN)

	h.uLoign.NodeId = surf.NodeID()
	h.uLoign.NodeType = surf.NodeType()
	h.uLoign.Rds = h.WRds

	core.HandleFunc(surf, h.OnReqLoginLobby)
	core.HandleFunc(surf, h.OnReqLogoutLobby)
	core.HandleFunc(surf, h.OnReqJoinQue)

	h.surf = surf
	return nil
}

func (h *Lobby) OnReady() {

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
