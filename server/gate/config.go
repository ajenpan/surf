package gate

import "encoding/json"

type Config struct {
	RsaPublicKeyFile string `json:"RsaPublicKeyFile"`
	ClientListenAddr string `json:"ClientListenAddr"`
	NodeListenAddr   string `json:"NodeListenAddr"`
}

var DefaultConfig = &Config{
	RsaPublicKeyFile: "http://myali01:9999/publickey",
	ClientListenAddr: ":11000",
	NodeListenAddr:   ":13000",
}

func (c *Config) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}
