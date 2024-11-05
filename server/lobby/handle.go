package lobby

import (
	"github.com/ajenpan/surf/core"
	lobbymsg "github.com/ajenpan/surf/msg/lobby"
)

func (h *Lobby) OnReqLoginLobby(ctx core.Context, in *lobbymsg.ReqLoginLobby) {
	uid := ctx.Caller()
	h.GetUserGameInfo(uid)
}

func (h *Lobby) OnReqDispatchQue(ctx core.Context, in *lobbymsg.ReqDispatchQue) {

}

func (h *Lobby) OnReqLogoutLobby(ctx core.Context, in *lobbymsg.ReqLogoutLobby) {

}
