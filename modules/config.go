package modules

import (
	"encoding/json"
	"os"
)

func ReadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	config := &Config{}
	return config, decoder.Decode(config)
}

type Config struct {
	MySQL *MySQLConfig `json:"mysql"`
	ZhiHu *ZhiHuConfig `json:"zhiHu"`
}

type MySQLConfig struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type ZhiHuConfig struct {
	Mode        int    `json:"mode"`
	Cookie      string `json:"cookie"`
	OwnURLToken string `json:"ownURLToken"`
}
