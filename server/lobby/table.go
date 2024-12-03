package lobby

import (
	"time"

	msgLobby "github.com/ajenpan/surf/msg/lobby"
)

type Table struct {
	idx          uint64
	status       msgLobby.TableStatus
	battleNodeId uint32
	BattleId     string
	createAt     time.Time
	deadline     time.Time
	users        []*User
	context      []byte
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
