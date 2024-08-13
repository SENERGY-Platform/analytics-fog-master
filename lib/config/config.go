package config

import (
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	srv_base "github.com/SENERGY-Platform/go-service-base/srv-base"
	"github.com/y-du/go-log-level/level"
)

type DataBaseConfig struct {
	Timeout int64 `json:"timeout" env_var:"DATABASE_TIMEOUT"`
	Path       string `json:"url" env_var:"DATABASE_PATH"`
}

type Config struct {
	Broker              mqtt.FogBrokerConfig
	Logger              srv_base.LoggerConfig `json:"logger" env_var:"LOGGER_CONFIG"`
	DataDir             string                `json:"data_dir" env_var:"DATA_DIR"`
	DataBase DataBaseConfig
	AgentSyncIntervalSeconds float64 `json:"agent_sync_interval" env_var:"AGENT_SYNC_INTERVAL"`
	StaleOperatorCheckIntervalSeconds float64 `json:"stale_operator_check_interval" env_var:"STALE_OPERATOR_CHECK_INTERVAL"`
	TimeoutStaleOperatorSeconds float64 `json:"timeout_stale_operator" env_var:"TIMEOUT_STALE_OPERATOR"`
	TimeoutInactiveAgentSeconds float64 `json:"timeout_inactive_agent" env_var:"TIMEOUT_INACTIVE_AGENT"`
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
		AgentSyncIntervalSeconds: 120,
		StaleOperatorCheckIntervalSeconds: 120,
		TimeoutInactiveAgentSeconds: 120,
		TimeoutStaleOperatorSeconds: 3600,
		DataBase: DataBaseConfig{
			Timeout: 10000000000,
			Path: "./data/sqlite3.db",
		},
	}

	err := srv_base.LoadConfig(path, &cfg, nil, nil, nil)
	return &cfg, err
}
