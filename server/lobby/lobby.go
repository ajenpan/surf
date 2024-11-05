package lobby

import (
	"gorm.io/gorm"

	lobbymsg "github.com/ajenpan/surf/msg/lobby"
)

type Lobby struct {
	GameDB *gorm.DB
}

func (h *Lobby) ServerType() uint16 {
	return 2
}

func (h *Lobby) ServerName() string {
	return "lobby"
}

func (h *Lobby) GetUserGameInfo(uid uint32) (*lobbymsg.UserBaseInfo, error) {
	info := &lobbymsg.UserBaseInfo{}
	return info, nil
}

func (h *Lobby) GetPropInfo(uid uint32) (map[int]int64, error) {
	return nil, nil
}
