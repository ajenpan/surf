package surf

import "github.com/ajenpan/surf/tcp"

func New(opt Options) *Surf {
	ret := &Surf{
		Options: &opt,
	}
	return ret
}

type Surf struct {
	*Options

	tcpsvr *tcp.Server
}
