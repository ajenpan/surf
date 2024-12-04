package lobby

import (
	"time"

	msgLobby "github.com/ajenpan/surf/msg/lobby"
	"google.golang.org/protobuf/proto"
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
	ConnID    string
	ChannelId int32
	OSType    string
	IP        string
	LoginAt   time.Time
	Sender    func(msg proto.Message) error
}

type User struct {
	UserId uint32

	BaseInfo UserBaseInfo
	GameInfo UserGameInfo
	PlayInfo UserPlayInfo
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
