package core

import "encoding/json"

type Config struct {
	SurfConf   SurfConfig      `json:"surf"`
	ServerConf json.RawMessage `json:"server_conf"`
	NodeConf   json.RawMessage `json:"node_conf"`
}

type SurfConfig struct {
	HttpListenAddr string   `yaml:"HttpListenAddr"`
	WsListenAddr   string   `yaml:"WsListenAddr"`
	TcpListenAddr  string   `yaml:"TcpListenAddr"`
	GateAddrList   []string `yaml:"GateAddrList"`
}
