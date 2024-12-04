package lobby

import "encoding/json"

type Config struct {
	GameDBDSN string `json:"gamedb_dsn"`
}

func FromJsonConf(cfg []byte) (*Config, error) {
	conf := &Config{}
	err := json.Unmarshal(cfg, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
