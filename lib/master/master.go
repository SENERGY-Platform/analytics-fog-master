/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	Client                     *mqtt.MQTTClient
	DB                         storage.DB
	OperatorController         *controller.Controller
	AgentSyncInterval          time.Duration
	TimeoutInactiveAgent       float64
	TimeoutStaleOperator       float64
	StaleOperatorCheckInterval time.Duration
}

func NewMaster(mqttClient *mqtt.MQTTClient, db storage.DB, controller *controller.Controller, agentSyncInterval, staleOperatorCheckInterval time.Duration, timeoutInactiveAgent, timeoutStaleOperator float64) *Master {
	logging.Logger.Debug(fmt.Sprintf("%d", staleOperatorCheckInterval))
	return &Master{
		Client:                     mqttClient,
		DB:                         db,
		OperatorController:         controller,
		AgentSyncInterval:          agentSyncInterval,
		TimeoutStaleOperator:       timeoutStaleOperator,
		TimeoutInactiveAgent:       timeoutInactiveAgent,
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
