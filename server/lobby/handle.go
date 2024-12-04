package lobby

import (
	"github.com/ajenpan/surf/core"
	msgLobby "github.com/ajenpan/surf/msg/lobby"
)

func (h *Lobby) OnReqLoginLobby(ctx core.Context, in *msgLobby.ReqLoginLobby) {
	resp := &msgLobby.RespLoginLobby{}
	var resperr error

	uid := ctx.FromUserID()

	currUser := h.getUser(uid)

	reconnect := false

	if currUser == nil {
		currUser = NewUser(uid)
	} else {
		switch currUser.PlayInfo.PlayerStatus {
		case msgLobby.PlayerStatus_PlayerNone:
			// do nothing
		case msgLobby.PlayerStatus_PlayerInQueue:
			// todo leave queue
		case msgLobby.PlayerStatus_PlayerInTable:
			fallthrough
		case msgLobby.PlayerStatus_PlayerInTableReady:
			// todo leave table
		case msgLobby.PlayerStatus_PlayerInGaming:
			// reconnect
			reconnect = true
		}
	}

	currUser.ConnInfo.ConnID = ctx.ConnID()
	currUser.ConnInfo.Sender = ctx.SendAsync

	// todo: how to get ip?
	currUser.ConnInfo.IP = ""

	if reconnect {
		// todo reconnect
		ctx.Response(resp, resperr)

		// todo send table info
		// ctx.SendAsync()
		return
	}

	baseinfo, err := h.GetUserGameInfo(uid)
	if err != nil {
		log.Error("get user game info error", "error", err, "uid", uid)
		resperr = err
		return
	}

	resp.BaseInfo = baseinfo
	ctx.Response(resp, resperr)
}

func (h *Lobby) OnReqDispatchQue(ctx core.Context, in *msgLobby.ReqDispatchQue) {

}

func (h *Lobby) OnReqLogoutLobby(ctx core.Context, in *msgLobby.ReqLogoutLobby) {

}
