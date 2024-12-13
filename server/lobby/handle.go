package lobby

import (
	"context"
	"fmt"
	"time"

	"github.com/ajenpan/surf/core"
	msgLobby "github.com/ajenpan/surf/msg/lobby"
	"google.golang.org/protobuf/proto"
)

func (h *Lobby) OnClientConnect(uid uint32, gateNodeId uint32, ip string) {
	h.usersIp[uid] = ip
}

func (h *Lobby) OnClientDisconnect(uid uint32, gateNodeId uint32, reason int32) {
	h.delLoginUser(uid)
}

func (h *Lobby) OnReqLoginLobby(ctx context.Context, req *msgLobby.ReqLoginLobby, resp *msgLobby.RespLoginLobby) error {
	var err error

	uid := core.CtxToUId(ctx)
	if uid == 0 {
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
				return nil
			}
		default:
			log.Error("uknown status")
		}
	}

	user.ConnInfo.Sender = func(msg proto.Message) error { return h.surf.SendAsyncToClient(uid, msg) }
	user.GameInfo.GameId = req.GameId
	user.PlayInfo.gameRoomId = req.GameRoomId

	isReconnect := table != nil && user.PlayInfo.PlayerStatus == msgLobby.PlayerStatus_PlayerInGaming

	log.Info("on user login", "uid", uid, "roomid", req.GameRoomId)

	if isReconnect {
		h.addLoginUser(user)

		user.MutableRespLoginLobby(resp)
		h.surf.Do(func() {
			notify := table.MutableNotifyDispatchResult()
			user.Send(notify)
		})
		return nil
	}

	err = user.Init()
	if err != nil {
		return err
	}

	err = h.uLoign.loadOrStore(uid)
	if err != nil {
		resp.Flag = msgLobby.RespLoginLobby_kInOtherRoom
		return nil
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
	return nil
}

func (h *Lobby) OnReqJoinQue(ctx context.Context, req *msgLobby.ReqJoinQue, resp *msgLobby.RespJoinQue) error {
	uid := core.CtxToUId(ctx)

	var herr error

	user := h.getLoginUser(uid)

	if user == nil {
		herr = fmt.Errorf("user not found %d", uid)
		return herr
	}

	que := h.getQue(user.PlayInfo.gameRoomId)

	if que == nil {
		return fmt.Errorf("que not found roomid:%d", user.PlayInfo.gameRoomId)
	}

	currState := user.PlayInfo.PlayerStatus
	if currState == msgLobby.PlayerStatus_PlayerInGaming ||
		currState == msgLobby.PlayerStatus_PlayerInQueue {
		return fmt.Errorf("player state err %d", currState)
	}
	needJoinQue := true

	if user.PlayInfo.tidx != 0 {
		table := h.FindContiTable(user.PlayInfo.tidx)
		if req.JoinType != msgLobby.ReqJoinQue_Noraml && table != nil {
			needJoinQue = false

			if herr = table.AddContinuePlayer(user); herr != nil {
				return herr
			}

			ok := table.checkStartCondi()

			if ok {
				h.surf.Do(func() {
					h.RemoveContiTable(table.idx)
					table.keepOnUsers = make(map[uint32]*User)
					h.DoTableStart(table)
				})
				return nil
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
			return herr
		}
	}
	return nil
}

func (h *Lobby) OnReqLogoutLobby(ctx context.Context, req *msgLobby.ReqLogoutLobby, resp *msgLobby.RespLogoutLobby) error {
	uid := core.CtxToUId(ctx)
	if uid == 0 {
		uid = req.Uid
	}
	h.delLoginUser(uid)
	return nil
}
