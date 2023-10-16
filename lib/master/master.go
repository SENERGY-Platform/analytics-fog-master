package master

import (
	"github.com/SENERGY-Platform/analytics-fog-master/lib/db"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/mqtt"
)

type Master struct {
	Client *mqtt.MQTTClient
	DB     db.DB
}

func NewMaster(mqttClient *mqtt.MQTTClient, db db.DB) *Master {
	return &Master{
		Client: mqttClient,
		DB:     db,
	}
}
