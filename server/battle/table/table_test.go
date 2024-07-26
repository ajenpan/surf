package table

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ajenpan/surf/server/battle"
	"github.com/ajenpan/surf/server/battle/noop"
	"github.com/ajenpan/surf/server/battle/proto"
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

	d := NewTable(TableOption{
		Conf: &proto.BattleConfigure{},
	})

	if err := logic.OnInit(d, nil); err != nil {
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
