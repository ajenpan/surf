package table

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	msgBattle "github.com/ajenpan/surf/msg/battle"
	"github.com/ajenpan/surf/server/battle"
	"github.com/ajenpan/surf/server/battle/noop"
)

type logicwarper struct {
	battle.Logic
	ontick func(time.Duration)
}

func (l *logicwarper) OnTick(d time.Duration) {
	l.ontick(d)
}

func TestTableTicker(t *testing.T) {
	tk := time.Duration(0)

	logic := noop.NewGameLogic()
	logic = &logicwarper{
		Logic: logic,
		ontick: func(d time.Duration) {
			t.Logf("tick %v", d)
			tk += d
		},
	}

	d := NewTable(TableOptions{
		Conf: &msgBattle.TableConfigure{},
	})

	if err := logic.OnInit(battle.LogicOpts{
		Table:   d,
		Players: []battle.Player{},
		Conf:    nil,
	}); err != nil {
		t.Fatal(err)
	}
	d.Init(logic, nil, nil)

	sec := time.Duration(rand.Int31n(10) + 10)
	time.Sleep(time.Second * sec)

	if tk < (sec-1)*time.Second || tk > (sec+1)*time.Second {
		t.Fatal("tick error:", tk)
	}
	fmt.Println("tick", tk)
}
