package surfing

import "surfing/tcp"

func New(opt Options) *Surfing {
	return &Surfing{
		Options: &opt,
	}
}

type Surfing struct {
	*Options

	svr *tcp.Server
}
