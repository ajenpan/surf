package lobby

import (
	"github.com/ajenpan/surf/core"
	msgLobby "github.com/ajenpan/surf/msg/lobby"
)

func (h *Lobby) OnReqLoginLobby(ctx core.Context, in *msgLobby.ReqLoginLobby) {
	resp := &msgLobby.RespLoginLobby{}
	var resperr error
	defer ctx.Response(resp, resperr)

	uid := ctx.FromUserID()

	currUser := h.getUser(uid)

	if currUser == nil {
		currUser = NewUser(uid)
	} else {

	}

	baseinfo, err := h.GetUserGameInfo(uid)
	if err != nil {
		log.Error("get user game info error", "error", err, "uid", uid)
		resperr = err
		return
	}

	resp.BaseInfo = baseinfo
}

func (h *Lobby) OnReqDispatchQue(ctx core.Context, in *msgLobby.ReqDispatchQue) {

}

func (h *Lobby) OnReqLogoutLobby(ctx core.Context, in *msgLobby.ReqLogoutLobby) {

}
