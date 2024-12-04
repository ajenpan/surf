package lobby

import (
	"log/slog"

	"gorm.io/gorm"

	"github.com/ajenpan/surf/core"
	lobbymsg "github.com/ajenpan/surf/msg/lobby"
	"github.com/redis/go-redis/v9"
)

var log = slog.Default()

func NewLobby() *Lobby {
	return &Lobby{}
}

type Lobby struct {
	GameDB *gorm.DB
	Rds    *redis.Client

	loginUsers   map[uint32]*User
	inTableUsers map[uint32]*User
}

func (h *Lobby) ServerType() uint16 {
	return 2
}

func (h *Lobby) ServerName() string {
	return "lobby"
}

func (h *Lobby) OnInit(surf *core.Surf) (err error) {
	return nil
}

func (h *Lobby) OnReady()

func (h *Lobby) OnStop() error {
	return nil
}

func (h *Lobby) GetUserGameInfo(uid uint32) (*lobbymsg.UserBaseInfo, error) {
	info := &lobbymsg.UserBaseInfo{}
	return info, nil
}

func (h *Lobby) GetPropInfo(uid uint32) (map[int]int64, error) {
	return nil, nil
}

func (h *Lobby) getLoginUser(uid uint32) *User {
	return h.loginUsers[uid]
}

func (h *Lobby) addLoginUser(u *User) {
	h.loginUsers[u.UserId] = u
}

func (h *Lobby) getUser(uid uint32) *User {
	u, has := h.loginUsers[uid]
	if has {
		return u
	}

	u, has = h.inTableUsers[uid]
	if has {
		return u
	}

	return nil
}
