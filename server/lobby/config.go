package lobby

import "encoding/json"

type Config struct {
	WGameDBDSN string `json:"w_gamedb_dsn"`
	WRedisDSN  string `json:"w_redis_dsn"`
}

func ConfigFromJson(cfg []byte) (*Config, error) {
	conf := &Config{}
	err := json.Unmarshal(cfg, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
