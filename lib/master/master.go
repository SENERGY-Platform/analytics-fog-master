package master

import (
	"github.com/SENERGY-Platform/analytics-fog-master/lib/config"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/db"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/mqtt"
)

type Master struct {
	Client              *mqtt.MQTTClient
	DB                  db.DB
	StartOperatorConfig config.StartOperatorConfig
}

func NewMaster(mqttClient *mqtt.MQTTClient, db db.DB, startOperatorConfig config.StartOperatorConfig) *Master {
	return &Master{
		Client:              mqttClient,
		DB:                  db,
		StartOperatorConfig: startOperatorConfig,
	}
}
