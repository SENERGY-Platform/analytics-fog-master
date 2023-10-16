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
	"encoding/json"
	"fmt"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/constants"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/entities"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/master"

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
	case constants.ControlTopic:
		relay.processControlCommand(message.Payload())
	case constants.AgentsTopic:
		relay.processAgentsCommand(message.Payload())
	case constants.OperatorsTopic:
		relay.processOperatorsCommand(message.Payload())
	}
}

func (relay *RelayController) processControlCommand(message []byte) {
	command := entities.ControlCommand{}
	err := json.Unmarshal(message, &command)
	if err != nil {
		fmt.Println("error:", err)
	}
	if command.Command == "startOperator" {
		relay.Master.StartOperator(command)
	}
	if command.Command == "stopOperator" {
		relay.Master.StopOperator(command)
	}
}

func (relay *RelayController) processAgentsCommand(message []byte) {
	agentMessage := entities.AgentMessage{}
	err := json.Unmarshal(message, &agentMessage)
	if err != nil {
		fmt.Println("error:", err)
	}
	switch agentMessage.Type {
	case "register":
		fmt.Println("Registering Agent")
		err = relay.Master.RegisterAgent(agentMessage.Conf.Id, agentMessage.Conf)
	case "pong":
		fmt.Println("Received Pong: " + agentMessage.Conf.Id)
		err = relay.Master.PongAgent(agentMessage.Conf.Id, agentMessage.Conf)
	}
}

func (relay *RelayController) processOperatorsCommand(message []byte) {
	op := entities.OperatorJob{}
	err := json.Unmarshal(message, &op)
	if err != nil {
		fmt.Println("error:", err)
	}
	if err := relay.Master.DB.SaveOperator(op); err != nil {
		fmt.Println("Error", err)
	}
}

func (relay *RelayController) OnMessageReceived(client MQTT.Client, message MQTT.Message) {
	fmt.Printf("Received message on topic: %s\nMessage: %s\n", message.Topic(), message.Payload())
	go relay.ProcessMessage(message)
}
