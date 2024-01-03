package master

import (
	"encoding/json"

	controlEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/control"
	masterLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/master"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/db"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/controller"

)

type Master struct {
	Client              *mqtt.MQTTClient
	DB                  db.DB
	OperatorController *controller.Controller
}

func NewMaster(mqttClient *mqtt.MQTTClient, db db.DB, controller *controller.Controller) *Master {
	return &Master{
		Client:              mqttClient,
		DB:                  db,
		OperatorController: controller,
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
