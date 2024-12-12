package lobby

import "encoding/json"

type Config struct {
	WGameDBDSN string `json:"w_gamedb_dsn"`
	WRedisDSN  string `json:"w_redis_dsn"`
}

// redis://<user>:<password>@<host>:<port>/<db_number>
var DefaultConf = &Config{
	WGameDBDSN: "root:123456@tcp(test122:3306)/game?charset=utf8mb4&parseTime=True&loc=Local",
	WRedisDSN:  "redis://:@test122:6379/0",
}

func ConfigFromJson(cfg []byte) (*Config, error) {
	conf := &Config{}
	err := json.Unmarshal(cfg, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
