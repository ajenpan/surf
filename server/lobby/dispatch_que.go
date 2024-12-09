package lobby

type OnMatchedFunc func(Matcher, []int64)
type OnTimeoutFunc func(Matcher)

type Matcher interface {
	String() string
}

type DispatchQue struct {
	expext int32
	worker func(map[uint32]*User) ([][]*User, error)
	que    map[uint32]*User
}

func (sm *DispatchQue) Add(u *User) error {
	sm.que[u.UserId] = u
	return nil
}
