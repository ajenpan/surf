package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type uniqueLoginInfoType struct {
	LoginAt  int64  `json:"login_at"`
	NodeId   uint32 `json:"node_id"`
	NodeType uint16 `json:"node_type"`
	UId      uint32 `json:"uid"`
}

type UserUniqueLogin struct {
	Rds *redis.Client

	NodeId   uint32
	NodeType uint16
}

func (h *UserUniqueLogin) loginInfoRdsKey(uid uint32) string {
	return fmt.Sprintf("lobby/userLoginAt/%d", uid)
}

func (h *UserUniqueLogin) Load(uid uint32) *uniqueLoginInfoType {
	rds := h.Rds
	lobbykey := h.loginInfoRdsKey(uid)
	result := rds.Get(context.Background(), lobbykey)
	if result.Err() != nil {
		log.Error("get user login info error", "error", result.Err(), "uid", uid)
		return nil
	}
	var info = &uniqueLoginInfoType{}
	err := json.Unmarshal([]byte(result.Val()), info)
	if err != nil {
		log.Error("get user login info unmarshal error", "error", err, "uid", uid)
		return nil
	}
	return info
}

func (h *UserUniqueLogin) Store(uid uint32) error {
	info := &uniqueLoginInfoType{
		UId:      uid,
		LoginAt:  time.Now().Unix(),
		NodeId:   h.NodeId,
		NodeType: h.NodeType,
	}
	rds := h.Rds
	lobbykey := h.loginInfoRdsKey(uid)
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return rds.Set(context.Background(), lobbykey, data, 0).Err()
}

func (h *UserUniqueLogin) Del(uid uint32) error {
	rds := h.Rds
	lobbykey := h.loginInfoRdsKey(uid)
	return rds.Del(context.Background(), lobbykey).Err()
}

func (h *UserUniqueLogin) loadOrStore(uid uint32) error {
	rds := h.Rds
	result := rds.SetNX(context.Background(), fmt.Sprintf("lobby/lock/userlogin/%d", uid), 1, 2*time.Second)
	if result.Err() != nil {
		return result.Err()
	}
	info := h.Load(uid)
	if info != nil {
		if info.NodeId == h.NodeId {
			return nil
		} else {
			return fmt.Errorf("user %d is already in lobby", uid)
		}
	}
	if !result.Val() {
		return fmt.Errorf("user %d is already in lobby", uid)
	}
	return h.Store(uid)
}
