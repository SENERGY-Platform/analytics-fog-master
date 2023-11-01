package master

import (
	"encoding/json"

	controlEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/control"
	masterLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/master"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/config"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/db"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
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

func (master *Master) Register() {
	// Master must register in case an agent is online before the master, so the agent can register again
	// TODO Master copnfiguration
	// masterConf := conf.GetConf()
	masterConf := masterLib.Configuration{
		Id: "id",
	}
	logging.Logger.Debug("Register master")
	conf, _ := json.Marshal(masterLib.MasterInfoMessage{
		ControlMessage: controlEntities.ControlMessage{
			Command: "register",
		},
		Conf: masterConf,
	})
	master.PublishMessage(masterLib.MasterTopic, string(conf), 2)
}
