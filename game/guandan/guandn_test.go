package guandan

import (
	"testing"
	"time"

	"github.com/google/uuid"

	msgBattle "github.com/ajenpan/surf/msg/battle"
	"github.com/ajenpan/surf/server/battle/table"
)

func newTestGuandanTable() (*table.Table, *Guandan) {

	logic := NewGuandan()

	players := []*table.Player{}
	for i := 0; i < 4; i++ {
		player := table.NewPlayer(&msgBattle.PlayerInfo{
			Uid:    uint64(i),
			SeatId: int32(i),
		})
		players = append(players, player)
	}

	battleid := uuid.NewString()

	d := table.NewTable(table.TableOptions{
		ID:             battleid,
		Conf:           nil,
		EventPublisher: nil,
		FinishReporter: nil,
	})

	err := d.Init(logic, players, nil)
	if err != nil {
		return nil, nil
	}
	return d, logic
}

func TestGuandanDoStart(t *testing.T) {
	table, logic := newTestGuandanTable()
	if table == nil || logic == nil {
		t.Fatal("newTestGuandanTable2 failed")
		return
	}

	defer table.Close()

	for i := 0; i < 4; i++ {
		table.OnPlayerConn(uint64(i), true)
	}

	time.Sleep(time.Second * 1)

	// check
	if logic.getStage() != StageType_StageGameStart {
		t.Fatal("stage error")
	}
}

func TestGuandanDealingCards1(t *testing.T) {
	table, logic := newTestGuandanTable()
	if table == nil || logic == nil {
		t.Fatal("newTestGuandanTable2 failed")
		return
	}

	defer table.Close()
	logic.doDealingCards()
	for _, p := range logic.players {
		t.Logf("seatid:%d, handcards:%v", p.seatId, p.rawHandCards.Chinese())

		if len(p.HandCards) != 27 {
			t.Fatal("handcards error")
		}
	}
}

func TestGuandanDealingCards(t *testing.T) {
	table, logic := newTestGuandanTable()
	if table == nil || logic == nil {
		t.Fatal("newTestGuandanTable2 failed")
		return
	}

	defer table.Close()

	for i := 0; i < 4; i++ {
		table.OnPlayerConn(uint64(i), true)
	}

	for _, p := range logic.players {
		p.DoubleRate = 1
	}

	logic.changeLogicStep(StageType_StageDealingCards)
	logic.OnTick(time.Second * 1)
}
