package conf

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	WLogDBDSN    string
	WSQUserDBDSN string
	WGameDBDSN   string
	WPropsDBDSN  string
	RLogDBDSN    string
	RConfigDBDSN string
	RPropsDBDSN  string
	RGameDBDSN   string
	RSQUserDBDSN string
	RedisConn    string
	IpRegionDB   string
	DBPUrl       string
	Debug        DebugConf
	Env          string

	DBPDeliverUrl  string
	ChecksumVerify bool
}

type DebugConf struct {
	// EnablePProf bool
	LogLevel string
}

var DefaultConf = &Config{
	ChecksumVerify: false,
}

func ConfInit(filename string, PrintConf bool) (*Config, error) {
	out := &Config{}
	defer func() {
		if PrintConf {
			if data, err := json.Marshal(out); err == nil {
				fmt.Println("the real config value is: ", string(data))
			} else {
				fmt.Println(err)
			}
		}
	}()

	c := viper.New()

	ext := filepath.Ext(filename)

	c.SetConfigType(ext) //don't forgot set the config type

	c.SetConfigFile(filename)
	if err := c.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := c.Unmarshal(out); err != nil {
		return nil, err
	}

	return out, nil
}
