package core

import (
	"github.com/ajenpan/surf/core/registry"
)

type NodeConf struct {
	SurfConf   SurfConfig `json:"SurfConf"`
	ServerConf []byte     `json:"ServerConf"`
}

type SurfConfig struct {
	HttpListenAddr    string               `json:"HttpListenAddr"`
	WsListenAddr      string               `yaml:"WsListenAddr"`
	TcpListenAddr     string               `json:"TcpListenAddr"`
	GateAddrList      []string             `json:"GateAddrList"`
	LogLevel          string               `json:"LogLevel"`
	EtcdConf          *registry.EtcdConfig `json:"EtcdConf"`
	PublicKeyFilePath string               `json:"PublicKeyFilePath"`
}
