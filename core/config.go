package core

type Config struct {
	ControlListenAddr string   `yaml:"ControlListenAddr"`
	HttpListenAddr    string   `yaml:"HttpListenAddr"`
	WsListenAddr      string   `yaml:"WsListenAddr"`
	TcpListenAddr     string   `yaml:"TcpListenAddr"`
	GateAddrList      []string `yaml:"GateAddrList"`
}
