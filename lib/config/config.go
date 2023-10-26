package config

import (
	srv_base "github.com/SENERGY-Platform/go-service-base/srv-base"
	"github.com/y-du/go-log-level/level"
)

type BrokerConfig struct {
	Host string `json:"broker_host" env_var:"BROKER_HOST"`
	Port string `json:"broker_port" env_var:"BROKER_PORT"`
}

type StartOperatorConfig struct {
	Retries int `json:"retries_start_operator" env_var:"RETRIES_START_OPERATOR"`
	Timeout int `json:"timeout_start_operator" env_var:"TIMEOUT_START_OPERATOR"`
}

type Config struct {
	Broker              BrokerConfig
	StartOperatorConfig StartOperatorConfig
	Logger              srv_base.LoggerConfig `json:"logger" env_var:"LOGGER_CONFIG"`
	DataDir             string                `json:"data_dir" env_var:"DATA_DIR"`
}

func NewConfig(path string) (*Config, error) {
	cfg := Config{
		Broker: BrokerConfig{
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
	}

	err := srv_base.LoadConfig(path, &cfg, nil, nil, nil)
	return &cfg, err
}
