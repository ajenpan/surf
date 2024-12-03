package lobby

import (
	"time"

	msgLobby "github.com/ajenpan/surf/msg/lobby"
)

type UserGameInfo = msgLobby.UserGameBaseInfo
type UserPropInfo = msgLobby.UserPropInfo
type UserBaseInfo = msgLobby.UserBaseInfo
type UserMetaInfo = msgLobby.UserMetaInfo

type UserPlayInfo struct {
	msgLobby.UserPlayingInfo

	DispatchTimeSec uint32
	DispatchCnt     uint32
}

type UserConnInfo struct {
	ConnId    uint32
	ChannelId int32
	OSType    string
	IP        string
	LoginAt   time.Time
}

type User struct {
	UserId uint32

	BaseInfo UserBaseInfo

	GameInfo UserGameInfo
	PropInfo UserPropInfo

	MetaInfo UserMetaInfo
	ConnInfo UserConnInfo
}

func NewUser(uid uint32) *User {
	return &User{
		UserId: uid,
		ConnInfo: UserConnInfo{
			ChannelId: 0,
		},
		BaseInfo: UserBaseInfo{},
		GameInfo: UserGameInfo{},
		PropInfo: UserPropInfo{
			Props: make(map[uint32]int64),
		},
		MetaInfo: UserMetaInfo{
			StrMeta: make(map[string]string),
			IntMeta: make(map[string]int64),
		},
	}
}
