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

	tidx       TableIdxT
	tuid       TableUIDT
	gameRoomId int32
}

type UserConnInfo struct {
	GateNodeId uint32
	ChannelId  int32
	OSType     string
	IP         string
	LoginAt    time.Time
	Sender     func(msg proto.Message) error
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
			Props: make(map[int32]int64),
		},
		MetaInfo: UserMetaInfo{
			StrMeta: make(map[string]string),
			IntMeta: make(map[string]int64),
		},
	}
}

func (u *User) MutableRespLoginLobby(out *msgLobby.RespLoginLobby) {
	out.BaseInfo = &u.BaseInfo
	out.Props = &u.PropInfo
	out.MetaInfo = &u.MetaInfo
}

func (u *User) Send(msg proto.Message) {
	if u.ConnInfo.Sender != nil {
		err := u.ConnInfo.Sender(msg)
		if err != nil {
			log.Error("user send", "err", err)
		}
	} else {
		log.Debug("mock user send", "msg", msg, "uid", u.UserId)
	}
}

func (u *User) Init() error {
	// type UserGameInfo = msgLobby.UserGameBaseInfo
	// type UserPropInfo = msgLobby.UserPropInfo
	// type UserBaseInfo = msgLobby.UserBaseInfo
	// type UserMetaInfo = msgLobby.UserMetaInfo

	return nil
}
