package config

import (
	srv_base "github.com/SENERGY-Platform/go-service-base/srv-base"
)

type BrokerConfig struct {
	Host string `json:"broker_host" env_var:"BROKER_HOST"`
	Port string `json:"broker_port" env_var:"BROKER_PORT"`
}

type Config struct {
	Broker BrokerConfig
}

func NewConfig(path string) (*Config, error) {
	cfg := Config{
		Broker: BrokerConfig{
			Port: "1883",
			Host: "localhost",
		},
	}

	err := srv_base.LoadConfig(path, &cfg, nil, nil, nil)
	return &cfg, err
}
