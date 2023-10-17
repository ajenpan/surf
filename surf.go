package surfing

import "github.com/ajenpan/surf/tcp"

func New(opt Options) *Surfing {
	return &Surfing{
		Options: &opt,
	}
}

type Surfing struct {
	*Options

	svr *tcp.Server
}
