package lobby

import (
	"fmt"
	"time"

	"github.com/ajenpan/surf/core"
	msgLobby "github.com/ajenpan/surf/msg/lobby"
)

func (h *Lobby) OnClientConnect() {}

func (h *Lobby) OnReqLoginLobby(ctx core.Context, req *msgLobby.ReqLoginLobby) {
	resp := &msgLobby.RespLoginLobby{}
	var err error

	uid := ctx.FromUserID()
	if ctx.FromUserRole() != 0 {
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

	user.ConnInfo.ConnID = ctx.ConnID()
	user.ConnInfo.Sender = ctx.SendAsync
	user.GameInfo.GameId = req.GameId
	user.PlayInfo.gameRoomId = req.GameRoomId

	isReconnect := table != nil && user.PlayInfo.PlayerStatus == msgLobby.PlayerStatus_PlayerInGaming

	log.Info("on user login")

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

func (h *Lobby) OnReqDispatchQue(ctx core.Context, req *msgLobby.ReqDispatchQue) {
	uid := ctx.FromUserID()
	user := h.getLoginUser(uid)
	resp := &msgLobby.RespDispatchQue{}
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
		if req.JoinType != 0 && table != nil {
			needJoinQue = false

			if herr = table.AddContinuePlayer(user); herr != nil {
				return
			}

			ok := table.checkStartCondi()
			if ok {
				// DoTableStart()

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
	uid := ctx.FromUserID()
	if ctx.FromUserRole() != 0 {
		uid = req.Uid
	}
	h.delLoginUser(uid)
	ctx.Response(&msgLobby.RespLogoutLobby{}, nil)
}

func (h *Lobby) DoTableStart(t *Table) error {
	return nil
}
