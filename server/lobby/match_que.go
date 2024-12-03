package lobby

import (
	"sync"
)

type MatchQue struct {
	expext int32
}

type OnMatchedFunc func(Matcher, []int64)
type OnTimeoutFunc func(Matcher)

type Matcher interface {
	String() string
}

type StaticMatcher struct {
	rwlock sync.RWMutex
	users  []*User
}

// TODO:
func (sm *StaticMatcher) Add(u *User) {
	sm.rwlock.Lock()
	defer sm.rwlock.Unlock()

	sm.users = append(sm.users, u)

	const expert = 4

}
