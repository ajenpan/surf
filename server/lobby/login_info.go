package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type loginInfoType struct {
	LoginAt    int64  `json:"login_at"`
	NodeId     uint32 `json:"node_id"`
	ServerType uint16 `json:"server_type"`
}

func getLoginInfoRdsKey(uid uint32) string {
	return fmt.Sprintf("lobby/userLoginAt/%d", uid)
}

func (h *Lobby) getUserLoginInfo(uid uint32) *loginInfoType {
	rds := h.Rds
	lobbykey := getLoginInfoRdsKey(uid)
	result := rds.Get(context.Background(), lobbykey)
	if result.Err() != nil {
		log.Error("get user login info error", "error", result.Err(), "uid", uid)
		return nil
	}
	var info = &loginInfoType{}
	err := json.Unmarshal([]byte(result.Val()), info)
	if err != nil {
		log.Error("get user login info unmarshal error", "error", err, "uid", uid)
		return nil
	}
	return info
}

func (h *Lobby) setUserLoginInfo(uid uint32, info *loginInfoType) error {
	rds := h.Rds
	lobbykey := getLoginInfoRdsKey(uid)
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return rds.Set(context.Background(), lobbykey, data, 0).Err()
}

func (h *Lobby) delUserLoginInfo(uid uint32) error {
	rds := h.Rds
	lobbykey := getLoginInfoRdsKey(uid)
	return rds.Del(context.Background(), lobbykey).Err()
}

func (h *Lobby) loadOrSetLobbyNode(uid uint32, currinfo *loginInfoType) error {
	rds := h.Rds

	info := h.getUserLoginInfo(uid)
	if info != nil {
		if info.NodeId == currinfo.NodeId {
			return nil
		} else {
			return fmt.Errorf("user %d is already in lobby", uid)
		}
	}

	result := rds.SetNX(context.Background(), fmt.Sprintf("lobby/lock/userlogin/%d", uid), 1, 2*time.Second)
	if result.Err() != nil {
		return result.Err()
	}

	if !result.Val() {
		return fmt.Errorf("user %d is already in lobby", uid)
	}

	return h.setUserLoginInfo(uid, currinfo)
}
