package lobby

import (
	"fmt"
	"time"

	msgLobby "github.com/ajenpan/surf/msg/lobby"
	"google.golang.org/protobuf/proto"
)

type TableUIDT = uint32
type TableIdxT = int64

type Table struct {
	idx          TableIdxT
	tuid         TableUIDT
	status       msgLobby.TableStatus
	battleNodeId uint32
	BattleId     string
	createAt     time.Time
	deadline     time.Time
	users        []*User
	context      []byte
	playid       string

	keepOnAt    int64
	keepOnUsers map[uint32]*User
	keepOnTimer *time.Timer

	deadlineTimer *time.Timer

	game *GameInfo
}

func (t *Table) getUser(uid uint32) *User {
	var ret *User
	for _, u := range t.users {
		if u.UserId == uid {
			ret = u
			break
		}
	}
	return ret
}

func (t *Table) MutableNotifyDispatchResult() *msgLobby.NotifyDispatchResult {
	return &msgLobby.NotifyDispatchResult{}
}

func (t *Table) AddContinuePlayer(user *User) error {
	if t.status != msgLobby.TableStatus_TableFinished {
		return fmt.Errorf("table state err")
	}

	if _, has := t.keepOnUsers[user.UserId]; has {
		return fmt.Errorf("repeat add")
	}

	t.keepOnUsers[user.UserId] = user

	if t.keepOnAt == 0 {
		t.keepOnAt = time.Now().Unix()
	}

	user.PlayInfo.PlayerStatus = msgLobby.PlayerStatus_PlayerInTableReady
	// BroadcastPlayerStateInfo();
	return nil
}

func (t *Table) checkStartCondi() bool {
	return len(t.keepOnUsers) == len(t.users)
}

func (t *Table) BroadcastMessage(msg proto.Message) {
	for _, user := range t.users {
		user.Send(msg)
	}
}
