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
		TableOption:      &opt,
		CreateAt:         time.Now(),
		quit:             make(chan bool, 1),
		currBattleStat:   battle.BattleStatus_Idle,
		beforeBattleStat: battle.BattleStatus_Idle,
	}

	ret.actQue = make(chan func(), 100)

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

	battle           battle.Logic
	currBattleStat   battle.GameStatus
	beforeBattleStat battle.GameStatus
	actQue           chan func()

	quit   chan bool
	status TableStatus

	Players PlayerStore
}

func (d *Table) Init(logic battle.Logic, players []*Player, logicConf interface{}) error {
	d.rwlock.Lock()
	defer d.rwlock.Unlock()

	if d.battle != nil {
		d.battle.OnReset()
	}

	for _, p := range players {
		d.Players.Store(p)
	}

	if err := logic.OnInit(d, logicConf); err != nil {
		return err
	}

	d.battle = logic

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

		select {
		case f := <-d.actQue:
			safecall(f)
		case now := <-tk.C:
			sub := now.Sub(latest)
			latest = now
			d.OnTick(sub)
		case <-d.quit:
			break
		}
	}()

	return nil
}

func (d *Table) PushAction(f func()) {
	d.actQue <- f
}

func (d *Table) AfterFunc(td time.Duration, f func()) {
	//TODO: upgrade this function perfermance
	time.AfterFunc(td, func() {
		d.PushAction(f)
	})
}

func (d *Table) OnTick(sub time.Duration) {
	if d.battle != nil {
		d.battle.OnTick(sub)
	}
}

func (d *Table) Close() {
	select {
	case <-d.quit:
		return
	default:
	}

	d.battle.OnReset()

	atomic.StoreInt32(&d.status, TableStatus_Closed)

	close(d.quit)
	close(d.actQue)
}

func (d *Table) ReportBattleStatus(s battle.GameStatus) {
	if d.currBattleStat == s {
		return
	}

	d.beforeBattleStat = d.currBattleStat
	d.currBattleStat = s

	event := &innermsg.BattleStatusChangeEvent{
		StatusBefore: int32(d.beforeBattleStat),
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
	return d.currBattleStat == battle.BattleStatus_Running
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
	d.actQue <- func() {
		p := d.Players.ByUID(uid)
		if p != nil && d.battle != nil {
			d.battle.OnPlayerMessage(p, msgid, iraw)
		}
	}
}

func (d *Table) updateStatus(s TableStatus) {
	atomic.StoreInt32(&d.status, s)
}

func (d *Table) Status() TableStatus {
	return atomic.LoadInt32(&d.status)
}

func (d *Table) OnPlayerReady(uid uint64, rds int32, then func(error)) {
	d.actQue <- func() {

		if d.Status() == TableStatus_Inited {
			p := d.Players.ByUID(uid)
			p.Ready = rds

			var err error
			if then != nil {
				defer then(err)
			}

			if rds == 0 {
				return
			}

			rdscnt := 0

			players := []battle.Player{}

			d.Players.Range(func(p *Player) bool {
				players = append(players, p)
				if p.Ready > 0 {
					rdscnt++
				}
				return true
			})

			if rdscnt != len(players) {
				return
			}

			starterr := d.battle.OnStart(players)

			if starterr != nil {
				log.Error(starterr)
			}

			d.updateStatus(TableStatus_Running)
		}

	}
}

func (d *Table) OnPlayerConn(uid uint64, online bool) {

}
