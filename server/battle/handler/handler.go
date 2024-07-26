package handler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	protobuf "google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/event"
	log "github.com/ajenpan/surf/core/log"
	"github.com/ajenpan/surf/core/network"
	"github.com/ajenpan/surf/core/utils/calltable"
	"github.com/ajenpan/surf/core/utils/marshal"
	"github.com/ajenpan/surf/server/battle"
	"github.com/ajenpan/surf/server/battle/proto"
	"github.com/ajenpan/surf/server/battle/table"
)

type Battle struct {
	tables sync.Map

	LogicCreator *battle.GameLogicCreator
	ct           *calltable.CallTable[uint32]
	marshal      marshal.Marshaler
	Publisher    event.Publisher

	createCounter int32
}

func New() *Battle {
	h := &Battle{
		LogicCreator: &battle.GameLogicCreator{},
	}
	h.ct = calltable.ExtractAsyncMethodByMsgID(proto.File_service_battle_proto_battle_proto.Messages(), h)
	return h
}

func (h *Battle) CreateBattle(ctx context.Context, in *proto.StartBattleRequest) (*proto.StartBattleResponse, error) {
	logic, err := h.LogicCreator.CreateLogic(in.GameName)
	if err != nil {
		return nil, err
	}

	atomic.AddInt32(&h.createCounter, 1)

	battleid := uuid.NewString()

	d := table.NewTable(table.TableOption{
		ID:             battleid,
		Conf:           in.BattleConf,
		EventPublisher: h.Publisher,
		FinishReporter: func() {
			h.onBattleFinished(battleid)
		},
	})

	players, err := table.NewPlayers(in.PlayerInfos)
	if err != nil {
		return nil, err
	}

	err = d.Init(logic, players, in.BattleConf)
	if err != nil {
		return nil, err
	}

	h.tables.Store(battleid, d)

	out := &proto.StartBattleResponse{
		BattleId: d.ID,
	}
	return out, nil
}

func (h *Battle) onBattleFinished(battleid string) {
	d := h.getBattleById(battleid)
	if d == nil {
		return
	}

	d.Players.Range(func(p *table.Player) bool {
		h.UIDUnBingBID(uint64(p.Uid), battleid)
		return true
	})

	d.Close()
	h.tables.Delete(battleid)
}

func (h *Battle) OnEvent(topc string, msg protobuf.Message) {

}

func (h *Battle) UIDBingBID(uid uint64, bid string) error {
	// TODO:
	return nil
}

func (h *Battle) UIDUnBingBID(uid uint64, bid string) {

}

func (h *Battle) LoadBattleByUID(uid uint64) []*table.Table {
	// TODO:
	return nil
}

func (h *Battle) JoinBattle(ctx context.Context, in *proto.JoinBattleRequest) (*proto.JoinBattleResponse, error) {
	out := &proto.JoinBattleResponse{
		BattleId:   in.BattleId,
		SeatId:     in.SeatId,
		ReadyState: in.ReadyState,
	}

	d := h.getBattleById(in.BattleId)
	if d == nil {
		return nil, fmt.Errorf("battle not found")
	}

	// socket := GetTcpSocket(ctx)

	// d.OnPlayerReady(socket.Uid, in.ReadyState)

	// 是否需要保持 uid - battleid 映射?
	// 1 uid -> n * battleid.
	// 当uid掉线时, 需要遍历所有的battleid, 并且通知battleid.
	// h.UIDBingBID(socket.Uid, in.BattleId)

	return out, nil
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
