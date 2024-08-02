package route

import "time"

type ConnUserData struct {
	CreateAt time.Time
}

func NewConnUserData() *ConnUserData {
	ret := &ConnUserData{
		CreateAt: time.Now(),
	}
	return ret
}
