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
			Uid:    int64(i),
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
		table.OnPlayerConn(int64(i), nil, true)
	}

	time.Sleep(time.Second * 1)

	// check
	if logic.currStage.StageType != StageType_Stage_GameStart {
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
		t.Logf("seatid:%d, handcards:%v", p.gameInfo.SeatId, p.handCards.Chinese())
		if len(p.gameInfo.HandCards) != 27 {
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
		table.OnPlayerConn(int64(i), nil, true)
	}

	for _, p := range logic.players {
		p.gameInfo.DoubleBet = 1
	}

	// logic.changeLogicStep(StageType_Stage_DealingCards)
	logic.OnTick(time.Second * 1)
}
