package master

import (
	"encoding/json"
	"fmt"
	"time"

	controlEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/control"
	masterLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/master"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/controller"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/storage"
)

type Master struct {
	Client              *mqtt.MQTTClient
	DB                  storage.DB
	OperatorController *controller.Controller
	AgentSyncInterval time.Duration
	TimeoutInactiveAgent float64
	TimeoutStaleOperator float64
	StaleOperatorCheckInterval time.Duration
}

func NewMaster(mqttClient *mqtt.MQTTClient, db storage.DB, controller *controller.Controller, agentSyncInterval, staleOperatorCheckInterval time.Duration, timeoutInactiveAgent, timeoutStaleOperator float64) *Master {
	logging.Logger.Debug(fmt.Sprintf("%d", staleOperatorCheckInterval))
	return &Master{
		Client:              mqttClient,
		DB:                  db,
		OperatorController: controller,
		AgentSyncInterval: agentSyncInterval,
		TimeoutStaleOperator: timeoutStaleOperator,
		TimeoutInactiveAgent: timeoutInactiveAgent,
		StaleOperatorCheckInterval: staleOperatorCheckInterval,
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
