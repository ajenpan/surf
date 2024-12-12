package core

import (
	"encoding/json"

	"github.com/ajenpan/surf/core/auth"
)

type NodeState = int32

const (
	NodeState_Closed   NodeState = 0
	NodeState_Init     NodeState = 1
	NodeState_Draining NodeState = 10 // maintaining existing connections but not accepting new ones
	NodeState_Running  NodeState = 100
)

type nodeRegistryData struct {
	Status NodeState       `json:"status"`
	Node   auth.NodeInfo   `json:"node"`
	Meta   registryMeta    `json:"meta"`
	Data   json.RawMessage `json:"data"`
}

type registryMeta struct {
	HttpListenAddr string `json:"http_listen_addr"`
	WsListenAddr   string `json:"ws_listen_addr"`
	TcpListenAddr  string `json:"tcp_listen_addr"`
}
