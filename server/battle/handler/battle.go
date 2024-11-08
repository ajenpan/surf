package handler

import (
	"sync"
	"time"

	"github.com/ajenpan/surf/core/event"
	"github.com/ajenpan/surf/server/battle"
	"github.com/ajenpan/surf/server/battle/table"
)

type Battle struct {
	tables        sync.Map
	LogicCreator  *battle.GameLogicCreator
	Publisher     event.Publisher
	userBattleMap sync.Map
}

func New() *Battle {
	h := &Battle{
		LogicCreator: battle.LogicCreator,
	}
	return h
}

func (h *Battle) ServerType() uint16 {
	return 1
}

func (h *Battle) ServerName() string {
	return "battle"
}

// 是否需要保持 uid - battleid 映射?
// 1 uid -> n * battleid.
// 当uid掉线时, 需要遍历所有的battleid, 并且通知battleid.
func (h *Battle) UIDBindBattleID(uid int64, bid string) error {
	battleMap, _ := h.userBattleMap.LoadOrStore(uid, &sync.Map{})
	battleMap.(*sync.Map).Store(bid, time.Now())
	return nil
}

func (h *Battle) UIDUnBindBattleID(uid int64, bid string) {
	battleMap, _ := h.userBattleMap.LoadOrStore(uid, &sync.Map{})
	battleMap.(*sync.Map).Delete(bid)
}

func (h *Battle) LoadBattleByUID(uid int64) []string {
	battleMap, _ := h.userBattleMap.LoadOrStore(uid, &sync.Map{})
	ret := []string{}
	battleMap.(*sync.Map).Range(func(key, value any) bool {
		ret = append(ret, key.(string))
		return true
	})
	return ret
}

func (h *Battle) getBattleById(battleId string) *table.Table {
	if raw, ok := h.tables.Load(battleId); ok {
		return raw.(*table.Table)
	}
	return nil
}
