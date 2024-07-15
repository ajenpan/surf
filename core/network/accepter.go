package network

import (
	"net"
)

type Accepter func(net.Conn)
