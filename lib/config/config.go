package config

import (
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	srv_base "github.com/SENERGY-Platform/go-service-base/srv-base"
	"github.com/y-du/go-log-level/level"
)

type StartOperatorConfig struct {
	Retries int `json:"retries_start_operator" env_var:"RETRIES_START_OPERATOR"`
	Timeout int `json:"timeout_start_operator" env_var:"TIMEOUT_START_OPERATOR"`
}

type DataBaseConfig struct {
	Timeout int `json:"timeout" env_var:"DATABASE_TIMEOUT"`
	ConnectionURL       string `json:"url" env_var:"DATABASE_URL"`
}

type Config struct {
	Broker              mqtt.FogBrokerConfig
	StartOperatorConfig StartOperatorConfig
	Logger              srv_base.LoggerConfig `json:"logger" env_var:"LOGGER_CONFIG"`
	DataDir             string                `json:"data_dir" env_var:"DATA_DIR"`
	DataBase DataBaseConfig
}

func NewConfig(path string) (*Config, error) {
	cfg := Config{
		Broker: mqtt.FogBrokerConfig{
			Port: "1883",
			Host: "localhost",
		},
		Logger: srv_base.LoggerConfig{
			Level:        level.Debug,
			Utc:          true,
			Microseconds: true,
			Terminal:     true,
		},
		DataDir: "./data",
		StartOperatorConfig: StartOperatorConfig{
			Retries: 10,
			Timeout: 10,
		},
	}

	err := srv_base.LoadConfig(path, &cfg, nil, nil, nil)
	return &cfg, err
}
