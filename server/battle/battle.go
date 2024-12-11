package battle

import (
	"log/slog"
	"sync"
	"time"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/event"
	"github.com/ajenpan/surf/game"
	"github.com/ajenpan/surf/server/battle/table"
)

const NodeType_Battle core.NodeType = 102

func ServerType() core.NodeType {
	return NodeType_Battle
}

func ServerName() string {
	return "battle"
}

var log = slog.Default().With("ntype", "battle")

type Battle struct {
	tables        sync.Map
	LogicCreator  *game.GameLogicCreator
	Publisher     event.Publisher
	userBattleMap sync.Map

	surf *core.Surf
}

func New() *Battle {
	h := &Battle{
		LogicCreator: game.LogicCreator,
	}
	return h
}

func (h *Battle) OnInit(surf *core.Surf) error {
	h.surf = surf
	confstr := surf.ServerConf()
	log.Info("battle server conf", "conf", confstr)

	core.HandleFunc(surf, h.OnReqStartBattle)
	core.HandleFunc(surf, h.OnReqJoinBattle)
	core.HandleFunc(surf, h.OnReqQuitBattle)
	core.HandleFunc(surf, h.OnBattleMsgToServer)
	return nil
}

func (h *Battle) OnReady() {
	h.surf.UpdateNodeData(core.NodeState_Running, nil)
}

func (h *Battle) OnStop() error {
	return nil
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
