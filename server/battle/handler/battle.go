package handler

import (
	"log/slog"
	"sync"
	"time"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/event"
	"github.com/ajenpan/surf/server/battle"
	"github.com/ajenpan/surf/server/battle/table"
)

const ServerType_Battle core.ServerType = 102

var log = slog.Default()

type Battle struct {
	tables        sync.Map
	LogicCreator  *battle.GameLogicCreator
	Publisher     event.Publisher
	userBattleMap sync.Map

	surf *core.Surf
}

func New() *Battle {
	h := &Battle{
		LogicCreator: battle.LogicCreator,
	}
	return h
}

func (h *Battle) OnInit(surf *core.Surf) error {
	h.surf = surf
	confstr := surf.ServerConf()
	log.Info("battle server conf", "conf", confstr)
	return nil
}

func (h *Battle) OnReady() {
	h.surf.UpdateNodeData(core.NodeState_Running, nil)
}

func (h *Battle) OnStop() error {
	return nil
}

func (h *Battle) ServerType() uint16 {
	return ServerType_Battle
}

func (h *Battle) ServerName() string {
	return "battle"
}

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
