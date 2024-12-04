package lobby

import (
	"github.com/ajenpan/surf/core"
	"github.com/ajenpan/surf/core/auth"
)

func Run() (err error) {
	// new server
	// new surf	and auth
	// init server with config
	// start surf
	uinfo := &auth.UserInfo{}
	surf, err := core.NewSurf(uinfo, &core.Config{}, NewLobby())
	if err != nil {
		return err
	}
	return surf.Run()
}
