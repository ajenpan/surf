package table

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/ajenpan/surf/core/event"

	"github.com/ajenpan/surf/game"
	msgBattle "github.com/ajenpan/surf/msg/battle"
)

type TableOptions struct {
	ID             string
	EventPublisher event.Publisher
	Conf           *msgBattle.TableConfigure
	FinishReporter func()
	Logger         *slog.Logger
}

func NewTable(opts TableOptions) *Table {
	if opts.ID == "" {
		opts.ID = uuid.NewString()
	}

	ret := &Table{
		log:        opts.Logger,
		opts:       opts,
		createAt:   time.Now(),
		quit:       make(chan bool),
		currStat:   game.GameStatus_Idle,
		beforeStat: game.GameStatus_Idle,
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
	log  *slog.Logger

	createAt time.Time
	startAt  time.Time
	finishAt time.Time

	rwlock sync.RWMutex

	logic      game.Logic
	currStat   game.GameStatus
	beforeStat game.GameStatus
	status     TableStatus

	actQue chan func()
	quit   chan bool

	Players PlayerStore
}

func (d *Table) BattleID() string {
	return d.opts.ID
}

func (d *Table) Init(logic game.Logic, players []*Player, logicConf []byte) error {
	d.rwlock.Lock()
	defer d.rwlock.Unlock()

	if d.logic != nil {
		d.logic.OnReset()
	}

	iplayers := []game.Player{}

	for _, p := range players {
		d.Players.Store(p)
		iplayers = append(iplayers, p)
	}

	if err := logic.OnInit(game.LogicOpts{
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
					d.log.Error("panic", "err", err)
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

	d.AfterFunc(3*time.Second, func() {
		for _, p := range players {
			if !p.online {
				p.online = true
				d.logic.OnPlayerEnter(p, 1, nil)
			}
		}
	})

	return nil
}

func (d *Table) Do(f func()) {
	d.actQue <- f
}

func (d *Table) AfterFunc(td time.Duration, f func()) game.AfterCancelFunc {
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

func (d *Table) ReportBattleStatus(s game.GameStatus) {
	if d.currStat == s {
		return
	}

	d.beforeStat = d.currStat
	d.currStat = s

	switch s {
	case game.GameStatus_Idle:
	case game.GameStatus_Running:
		d.reportGameStart()
	case game.GameStatus_Over:
		d.reportGameOver()

		d.updateStatus(TableStatus_Finished)

		d.AfterFunc(5*time.Second, func() {
			d.log.Info("report battle finished")
			if d.opts.FinishReporter != nil {
				d.opts.FinishReporter()
			}
		})
	}
}

func (d *Table) ReportBattleEvent(topic string, event proto.Message) {
	d.PublishEvent(event)
}

func (d *Table) SendMessageToPlayer(p game.Player, sync uint32, msgid uint32, msg proto.Message) {
	rp := p.(*Player)
	raw, err := proto.Marshal(msg)
	if err != nil {
		d.log.Error("sendToUser marshal msg failed", "err", err)
		return
	}

	err = rp.Send(msgid, raw)

	if err != nil {
		d.log.Error("sendToUser failed", "err", err, "uid", rp.UID(), "msgname", string(proto.MessageName(msg)), "msgid", msgid, "msg", msg)
	} else {
		d.log.Debug("sendToUser ok", "uid", rp.UID(), "msgname", string(proto.MessageName(msg)), "msgid", msgid, "msg", msg)
	}
}

func (d *Table) BroadcastMessage(msgid uint32, msg proto.Message) {
	raw, err := proto.Marshal(msg)
	if err != nil {
		d.log.Error("broadcast marshal failed", "err", err)
		return
	}

	d.log.Debug("broadcast", "msgname", string(proto.MessageName(msg)), "msgid", msgid, "msg", msg)

	d.Players.Range(func(p *Player) bool {
		err := p.Send(msgid, raw)
		if err != nil {
			d.log.Error("broadcast failed", "err", err, "uid", p.UID(), "msgname", string(proto.MessageName(msg)), "msgid", msgid, "msg", msg)
		}
		return true
	})
}

func (d *Table) IsPlaying() bool {
	return d.currStat == game.GameStatus_Running
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

	d.log.Debug("PublishEvent", "msgname", string(proto.MessageName(eventmsg)), "msg", eventmsg)

	raw, err := proto.Marshal(eventmsg)
	if err != nil {
		d.log.Error("PublishEvent marshal failed", "err", err)
		return
	}
	warp := &event.Event{
		Topic:     string(proto.MessageName(eventmsg)),
		Timestamp: time.Now().Unix(),
		Data:      raw,
	}
	d.opts.EventPublisher.Publish(warp)
}

func (d *Table) OnPlayerMessage(uid int64, syn uint32, msgid uint32, iraw []byte) {
	d.Do(func() {
		p := d.Players.ByUID(uid)
		if p != nil && d.logic != nil {
			d.logic.OnPlayerMessage(p, syn, msgid, iraw)
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
		d.logic.OnPlayerEnter(p, 0, nil)
	})
}
