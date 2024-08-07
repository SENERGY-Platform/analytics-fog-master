/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package relay

import (
	"fmt"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/master"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"


	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type RelayController struct {
	Master *master.Master
}

func NewRelayController(master *master.Master) *RelayController {
	return &RelayController{
		Master: master,
	}
}

func (relay *RelayController) ProcessMessage(message MQTT.Message) {
	switch message.Topic() {
	case operator.StartOperatorFogTopic:
		relay.processStartOperatorCommand(message.Payload())
	case operator.StopOperatorFogTopic:
		relay.processStopOperatorCommand(message.Payload())
	case agent.AgentsTopic:
		relay.processAgentsCommand(message.Payload())
	case operator.StartOperatorResponseFogTopic:
		relay.processAgentStartOperatorResponse(message.Payload())
	case operator.StopOperatorResponseFogTopic:
		relay.processAgentStopOperatorResponse(message.Payload())
	case operator.OperatorControlSyncResponseFogTopic:
		relay.processOperatorControlSync(message.Payload())
	}
}

func (relay *RelayController) OnMessageReceived(client MQTT.Client, message MQTT.Message) {
	logging.Logger.Debug(fmt.Sprintf("Received message on topic: %s Message: %s", message.Topic(), message.Payload()))
	go relay.ProcessMessage(message)
}
