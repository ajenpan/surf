package table

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/event"
	log "github.com/ajenpan/surf/core/log"

	innermsg "github.com/ajenpan/surf/msg/innerproto/battle"
	"github.com/ajenpan/surf/server/battle"
)

type TableOption struct {
	ID             string
	EventPublisher event.Publisher
	Conf           *innermsg.BattleConfigure
	FinishReporter func()
}

func NewTable(opt TableOption) *Table {
	if opt.ID == "" {
		opt.ID = uuid.NewString()
	}

	ret := &Table{
		TableOption: &opt,
		CreateAt:    time.Now(),
		quit:        make(chan bool),
		currStat:    battle.BattleStatus_Idle,
		beforeStat:  battle.BattleStatus_Idle,
	}

	ret.actQue = make(chan func(), 10)

	return ret
}

type TableStatus = int32

const (
	TableStatus_Node     TableStatus = iota
	TableStatus_Inited   TableStatus = iota
	TableStatus_Running  TableStatus = iota
	TableStatus_Finished TableStatus = iota
	TableStatus_Closed   TableStatus = iota
)

type Table struct {
	*TableOption

	CreateAt time.Time
	StartAt  time.Time
	OverAt   time.Time

	rwlock sync.RWMutex

	logic      battle.Logic
	currStat   battle.GameStatus
	beforeStat battle.GameStatus
	status     TableStatus

	actQue chan func()
	quit   chan bool

	Players PlayerStore
}

func (d *Table) Init(logic battle.Logic, players []*Player, logicConf interface{}) error {
	d.rwlock.Lock()
	defer d.rwlock.Unlock()

	if d.logic != nil {
		d.logic.OnReset()
	}

	iplayers := []battle.Player{}

	for _, p := range players {
		d.Players.Store(p)
		iplayers = append(iplayers, p)
	}

	if err := logic.OnInit(d, iplayers, logicConf); err != nil {
		return err
	}

	d.logic = logic

	go func() {
		safecall := func(f func()) {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("panic: %v", err)
				}
			}()
			f()
		}

		d.updateStatus(TableStatus_Inited)

		tk := time.NewTicker(1 * time.Second)

		defer tk.Stop()
		latest := time.Now()

		for {
			select {
			case f := <-d.actQue:
				safecall(f)
			case now := <-tk.C:
				sub := now.Sub(latest)
				latest = now
				d.OnTick(sub)
			case <-d.quit:
				return
			}
		}
	}()

	return nil
}

func (d *Table) Do(f func()) {
	d.actQue <- f
}

func (d *Table) AfterFunc(td time.Duration, f func()) {
	//TODO: upgrade this function perfermance
	time.AfterFunc(td, func() {
		d.Do(f)
	})
}

func (d *Table) OnTick(sub time.Duration) {
	if d.logic != nil {
		d.logic.OnTick(sub)
	}
}

func (d *Table) Close() {
	select {
	case <-d.quit:
		return
	default:
	}

	d.logic.OnReset()

	atomic.StoreInt32(&d.status, TableStatus_Closed)

	close(d.quit)
	close(d.actQue)
}

func (d *Table) ReportBattleStatus(s battle.GameStatus) {
	if d.currStat == s {
		return
	}

	d.beforeStat = d.currStat
	d.currStat = s

	event := &innermsg.BattleStatusChangeEvent{
		StatusBefore: int32(d.beforeStat),
		StatusNow:    int32(s),
		BattleId:     d.ID,
	}

	d.PublishEvent(event)

	switch s {
	case battle.BattleStatus_Idle:
	case battle.BattleStatus_Running:
		d.reportGameStart()
	case battle.BattleStatus_Over:
		d.reportGameOver()

		d.updateStatus(TableStatus_Finished)

		d.AfterFunc(5*time.Second, func() {
			d.FinishReporter()
		})
	}
}

func (d *Table) ReportBattleEvent(topic string, event proto.Message) {
	d.PublishEvent(event)
}

func (d *Table) SendMessageToPlayer(p battle.Player, msgid uint32, msg proto.Message) {
	rp := p.(*Player)

	raw, err := proto.Marshal(msg)
	if err != nil {
		log.Error(err)
		return
	}

	err = rp.Send(msgid, raw)
	if err != nil {
		log.Errorf("send message to player: %v, %s: %v", rp.Uid, string(proto.MessageName(msg)), msg)
	} else {
		log.Debugf("send message to player: %v, %s: %v", rp.Uid, string(proto.MessageName(msg)), msg)
	}
}

func (d *Table) BroadcastMessage(msgid uint32, msg proto.Message) {
	msgname := string(proto.MessageName(msg))
	log.Debugf("BroadcastMessage: %s: %v", msgname, msg)

	raw, err := proto.Marshal(msg)
	if err != nil {
		log.Error(err)
		return
	}

	d.Players.Range(func(p *Player) bool {
		err := p.Send(msgid, raw)
		if err != nil {
			log.Error(err)
		}
		return true
	})
}

func (d *Table) IsPlaying() bool {
	return d.currStat == battle.BattleStatus_Running
}

func (d *Table) reportGameStart() {
	d.StartAt = time.Now()
}

func (d *Table) reportGameOver() {
	d.OverAt = time.Now()
}

func (d *Table) PublishEvent(eventmsg proto.Message) {
	if d.EventPublisher == nil {
		return
	}

	log.Infof("PublishEvent: %s: %v", string(proto.MessageName(eventmsg)), eventmsg)

	raw, err := proto.Marshal(eventmsg)
	if err != nil {
		log.Error(err)
		return
	}
	warp := &event.Event{
		Topic:     string(proto.MessageName(eventmsg)),
		Timestamp: time.Now().Unix(),
		Data:      raw,
	}
	d.EventPublisher.Publish(warp)
}

func (d *Table) OnPlayerMessage(uid uint64, msgid uint32, iraw []byte) {
	d.Do(func() {
		p := d.Players.ByUID(uid)
		if p != nil && d.logic != nil {
			d.logic.OnPlayerMessage(p, msgid, iraw)
		}
	})
}

func (d *Table) updateStatus(s TableStatus) {
	atomic.StoreInt32(&d.status, s)
}

func (d *Table) Status() TableStatus {
	return atomic.LoadInt32(&d.status)
}

func (d *Table) Start() {
	d.actQue <- func() {

	}
}

// func (d *Table) OnPlayerJoin(uid uint64, rds int32, thenFn func(error)) {

// 	d.OnPlayerConn(uid,true)

// 	d.Do(func() {
// 		if d.Status() == TableStatus_Inited {
// 			p := d.Players.ByUID(uid)
// 			p.Ready = rds

// 			var err error
// 			if thenFn != nil {
// 				defer thenFn(err)
// 			}

// 			if rds == 0 {
// 				return
// 			}

// 			rdscnt := 0

// 			players := []battle.Player{}

// 			d.Players.Range(func(p *Player) bool {
// 				players = append(players, p)
// 				if p.Ready > 0 {
// 					rdscnt++
// 				}
// 				return true
// 			})

// 			if rdscnt != len(players) {
// 				return
// 			}

// 			oldstatus := atomic.SwapInt32(&d.status, TableStatus_Running)
// 			if oldstatus != TableStatus_Inited {
// 				return
// 			}
// 			// starterr := d.battle.OnStart(players)
// 			// if starterr != nil {
// 			// 	log.Error(starterr)
// 			// }
// 		}
// 	})
// }

func (d *Table) OnPlayerConn(uid uint64, online bool) {
	d.Do(func() {
		p := d.Players.ByUID(uid)
		if p == nil {
			return
		}

		p.online = online
		d.logic.OnPlayerConnStatus(p, online)
	})
}
