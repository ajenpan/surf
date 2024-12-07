package lobby

import (
	"github.com/ajenpan/surf/core"
	msgLobby "github.com/ajenpan/surf/msg/lobby"
)

func (h *Lobby) OnClientConnect() {

}

func (h *Lobby) OnReqLoginLobby(ctx core.Context, req *msgLobby.ReqLoginLobby) {
	resp := &msgLobby.RespLoginLobby{}
	var err error

	uid := ctx.FromUserID()
	if ctx.FromUserRole() != 0 {
		uid = uint32(req.Uid)
	}

	user := h.getUser(uid)
	isReconnect := false

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

			} else {

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

	user.ConnInfo.ConnID = ctx.ConnID()
	user.ConnInfo.Sender = ctx.SendAsync
	user.GameInfo.GameId = req.GameId
	user.PlayInfo.gameRoomId = req.GameRoomId

	isReconnect = table != nil && user.PlayInfo.PlayerStatus == msgLobby.PlayerStatus_PlayerInGaming

	log.Info("on user login")

	if isReconnect {
		h.addLoginUser(user)

		user.MutableRespLoginLobby(resp)
		ctx.Response(resp, nil)

		notify := table.MutableNotifyDispatchResult()
		user.Send(notify)
		return
	}

	err = h.uLoign.loadOrStore(uid)
	if err != nil {
		resp.Flag = msgLobby.RespLoginLobby_kInOtherRoom
		ctx.Response(resp, nil)
		return
	}

	err = user.Init()
	if err != nil {
		ctx.Response(resp, err)
		return
	}

	// baseinfo, err := h.GetUserGameInfo(uid)
	// if err != nil {
	// 	log.Error("get user game info error", "error", err, "uid", uid)
	// 	resperr = err
	// 	return
	// }
	// resp.BaseInfo = baseinfo
	ctx.Response(resp, nil)
}

func (h *Lobby) OnReqDispatchQue(ctx core.Context, in *msgLobby.ReqDispatchQue) {

}

func (h *Lobby) OnReqLogoutLobby(ctx core.Context, in *msgLobby.ReqLogoutLobby) {

}
