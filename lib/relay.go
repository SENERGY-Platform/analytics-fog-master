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

package lib

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func processMessage(message MQTT.Message) {
	subTopic := strings.TrimPrefix(message.Topic(), TopicPrefix)
	switch subTopic {
	case "control":
		processControlCommand(message.Payload())
	case "agents":
		processAgentsCommand(message.Payload())
	case "operators":
		processOperatorsCommand(message.Payload())
	}

}

func processControlCommand(message []byte) {
	command := ControlCommand{}
	err := json.Unmarshal(message, &command)
	if err != nil {
		fmt.Println("error:", err)
	}
	if command.Command == "startOperator" {
		startOperator(command)
	}
	if command.Command == "stopOperator" {
		stopOperator(command)
	}
}

func processAgentsCommand(message []byte) {
	agentMessage := AgentMessage{}
	err := json.Unmarshal(message, &agentMessage)
	if err != nil {
		fmt.Println("error:", err)
	}
	switch agentMessage.Type {
	case "register":
		fmt.Println("Registering Agent")
		agentMessage.Conf.Active = true
		if err := DB().Write("agents", agentMessage.Conf.Id, agentMessage.Conf); err != nil {
			fmt.Println("Error", err)
		}
	case "pong":
		fmt.Println("Received Pong: " + agentMessage.Conf.Id)
		agentMessage.Conf.Updated = time.Now().UTC()
		if err := DB().Write("agents", agentMessage.Conf.Id, agentMessage.Conf); err != nil {
			fmt.Println("Error", err)
		}
	}
}

func processOperatorsCommand(message []byte) {
	op := OperatorJob{}
	err := json.Unmarshal(message, &op)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(op)
	if err := DB().Write("operatorJobs", op.Config.OperatorId, op); err != nil {
		fmt.Println("Error", err)
	}
}
