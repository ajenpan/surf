package handler

import (
	"sync"

	"github.com/google/uuid"

	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/errors"
	"github.com/ajenpan/surf/core/event"
	log "github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
	innermsg "github.com/ajenpan/surf/msg/innerproto/battle"
	openmsg "github.com/ajenpan/surf/msg/openproto/battle"

	"github.com/ajenpan/surf/server/battle"
	"github.com/ajenpan/surf/server/battle/table"
)

type Battle struct {
	tables       sync.Map
	LogicCreator *battle.GameLogicCreator
	Publisher    event.Publisher
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

func (h *Battle) OnStartBattleRequest(ctx core.Context, in *innermsg.StartBattleRequest) {
	var err error
	var resp = &innermsg.StartBattleResponse{}

	defer func() {
		ctx.Response(resp, err)
	}()

	logic, err := h.LogicCreator.CreateLogic(in.GameName)
	if err != nil {
		return
	}

	players, err := table.NewPlayers(in.PlayerInfos)
	if err != nil {
		return
	}

	battleid := uuid.NewString()

	d := table.NewTable(table.TableOption{
		ID:             battleid,
		Conf:           in.TableConf,
		EventPublisher: h.Publisher,
		FinishReporter: func() {
			h.onBattleFinished(battleid)
		},
	})

	err = d.Init(logic, players, in.GameConf)
	if err != nil {
		return
	}

	h.tables.Store(battleid, d)

	resp.BattleId = d.ID
}

func (h *Battle) onBattleFinished(battleid string) {
	d := h.getBattleById(battleid)
	if d == nil {
		return
	}

	d.Players.Range(func(p *table.Player) bool {
		h.UIDUnBindBattleID(uint64(p.Uid), battleid)
		return true
	})

	d.Close()
	h.tables.Delete(battleid)
}

func (h *Battle) UIDBindBattleID(uid uint64, bid string) error {
	// TODO:
	return nil
}

func (h *Battle) UIDUnBindBattleID(uid uint64, bid string) {

}

func (h *Battle) LoadBattleByUID(uid uint64) map[string]table.Table {
	// TODO:
	return nil
}

func (h *Battle) OnJoinBattleRequest(ctx core.Context, in *openmsg.JoinBattleRequest) {
	var err error

	out := &openmsg.JoinBattleResponse{
		BattleId:   in.BattleId,
		SeatId:     in.SeatId,
		ReadyState: in.ReadyState,
	}

	d := h.getBattleById(in.BattleId)
	if d == nil {
		err = errors.New(-1, "battle not found")
		ctx.Response(out, err)
		return
	}

	d.OnPlayerConn(uint64(ctx.Caller()), true)

	ctx.Response(out, nil)

	// TODO:
	// 是否需要保持 uid - battleid 映射?
	// 1 uid -> n * battleid.
	// 当uid掉线时, 需要遍历所有的battleid, 并且通知battleid.
	// h.UIDBingBID(socket.Uid, in.BattleId)
	// return out, nil
}

// func (h *Battle) OnBattleMessageWrap(s *tcp.Socket, msg *proto.LoigcMessageWrap) {
// 	b := h.getBattleById(msg.BattleId)
// 	if b == nil {
// 		return
// 	}
// 	b.OnPlayerMessage(s.Uid, (msg.Msgid), msg.Data)
// }

func (h *Battle) getBattleById(battleId string) *table.Table {
	if raw, ok := h.tables.Load(battleId); ok {
		return raw.(*table.Table)
	}
	return nil
}

func (h *Battle) OnConn(s network.Conn, online bool) {

	log.Info("OnConn:", online)

	// tables := h.LoadBattleByUID(s.Uid)
	// for _, t := range tables {
	// 	t.OnPlayerConn(s.Uid, online)
	// }
}

func (h *Battle) OnMessage(s network.Conn, ss *network.HVPacket) {
	// ctype := ss.GetType()

	// if ctype == 6 {
	// 	head := ss.GetHead()

	// 	msgid := binary.LittleEndian.Uint32(head)
	// 	method := h.ct.Get(msgid)
	// 	if method == nil {
	// 		return
	// 	}

	// 	req := method.GetRequest()
	// 	defer method.PutRequest(req)

	// 	body := ss.GetBody()
	// 	err := h.marshal.Unmarshal(body, req)
	// 	if err != nil {
	// 		log.Errorf("marshal msgid:%d,error:%w", msgid, err)
	// 		return
	// 	}

	// 	ctx := WithTcpSocket(context.Background(), s)
	// 	res := method.Call(ctx, req)
	// 	if len(res) == 0 {
	// 		return
	// 	}
	// 	if res[0].IsNil() {
	// 		return
	// 	}
	// 	err, ok := res[0].Interface().(error)
	// 	if ok && err != nil {
	// 		log.Errorf("call msgid:%d,error:%w", msgid, err)
	// 		return
	// 	}
	// }
}
