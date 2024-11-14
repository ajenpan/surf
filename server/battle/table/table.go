package table

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/event"
	log "github.com/ajenpan/surf/core/log"

	msgBattle "github.com/ajenpan/surf/msg/battle"
	"github.com/ajenpan/surf/server/battle"
)

type TableOptions struct {
	ID             string
	EventPublisher event.Publisher
	Conf           *msgBattle.TableConfigure
	FinishReporter func()
}

func NewTable(opts TableOptions) *Table {
	if opts.ID == "" {
		opts.ID = uuid.NewString()
	}

	ret := &Table{
		log: log.Default.WithFields(map[string]interface{}{
			"battle": opts.ID,
		}),
		opts:       opts,
		createAt:   time.Now(),
		quit:       make(chan bool),
		currStat:   battle.BattleStatus_Idle,
		beforeStat: battle.BattleStatus_Idle,
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
	opts TableOptions
	log  log.Logger

	createAt time.Time
	startAt  time.Time
	finishAt time.Time

	rwlock sync.RWMutex

	logic      battle.Logic
	currStat   battle.GameStatus
	beforeStat battle.GameStatus
	status     TableStatus

	actQue chan func()
	quit   chan bool

	Players PlayerStore
}

func (d *Table) BattleID() string {
	return d.opts.ID
}

func (d *Table) Init(logic battle.Logic, players []*Player, logicConf []byte) error {
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

	if err := logic.OnInit(battle.LogicOpts{
		Table:   d,
		Players: iplayers,
		Conf:    logicConf,
		Log:     d.log,
	}); err != nil {
		return err
	}

	d.logic = logic

	go func() {
		safecall := func(f func()) {
			defer func() {
				if err := recover(); err != nil {
					d.log.Error("panic: %v", err)
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
			case f, ok := <-d.actQue:
				if !ok {
					return
				}
				if f == nil {
					continue
				}
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

func (d *Table) AfterFunc(td time.Duration, f func()) battle.AfterCancelFunc {
	tk := time.AfterFunc(td, func() {
		d.Do(f)
	})

	return func() {
		tk.Stop()
		tk = nil
	}
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

	switch s {
	case battle.BattleStatus_Idle:
	case battle.BattleStatus_Running:
		d.reportGameStart()
	case battle.BattleStatus_Over:
		d.reportGameOver()

		d.updateStatus(TableStatus_Finished)

		d.AfterFunc(5*time.Second, func() {
			d.log.Infof("report battle finished")
			if d.opts.FinishReporter != nil {
				d.opts.FinishReporter()
			}
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
		d.log.Error(err)
		return
	}

	err = rp.Send(msgid, raw)

	if err != nil {
		d.log.Errorf("sendToUser err:%v uid:%v,msgname:%s,msgid:%d,msg:%v", err, rp.UID(), string(proto.MessageName(msg)), msgid, msg)
	} else {
		d.log.Debugf("sendToUser ok uid:%v,msgname:%s,msgid:%d,msg:%v", rp.UID(), string(proto.MessageName(msg)), msgid, msg)
	}
}

func (d *Table) BroadcastMessage(msgid uint32, msg proto.Message) {

	raw, err := proto.Marshal(msg)
	if err != nil {
		log.Error(err)
		return
	}

	d.log.Debugf("broadcast msgname:%s,msgid:%d,msg:%v", string(proto.MessageName(msg)), msgid, msg)

	d.Players.Range(func(p *Player) bool {
		err := p.Send(msgid, raw)
		if err != nil {
			d.log.Errorf("broadcast err:%v uid:%v,msgname:%s,msgid:%d,msg:%v", err, p.UID(), string(proto.MessageName(msg)), msgid, msg)
		}
		return true
	})
}

func (d *Table) IsPlaying() bool {
	return d.currStat == battle.BattleStatus_Running
}

func (d *Table) reportGameStart() {
	d.startAt = time.Now()
}

func (d *Table) reportGameOver() {
	d.finishAt = time.Now()
}

func (d *Table) PublishEvent(eventmsg proto.Message) {
	if d.opts.EventPublisher == nil {
		return
	}

	d.log.Debugf("PublishEvent msgname:%s,msg:%v", string(proto.MessageName(eventmsg)), eventmsg)

	raw, err := proto.Marshal(eventmsg)
	if err != nil {
		d.log.Error(err)
		return
	}
	warp := &event.Event{
		Topic:     string(proto.MessageName(eventmsg)),
		Timestamp: time.Now().Unix(),
		Data:      raw,
	}
	d.opts.EventPublisher.Publish(warp)
}

func (d *Table) OnPlayerMessage(uid int64, msgid uint32, iraw []byte) {
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

func (d *Table) OnPlayerConn(uid int64, sender PlayerSender, enable bool) {
	d.Do(func() {
		p := d.Players.ByUID(uid)
		if p == nil {
			return
		}

		p.sender = sender
		p.online = enable
		d.logic.OnPlayerConnStatus(p, p.online)
	})
}
