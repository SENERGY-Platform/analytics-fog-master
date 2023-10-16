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

package master

import (
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/constants"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/entities"

	"time"
)

func (master *Master) StartOperator(command entities.ControlCommand) {
	agents := master.DB.GetAllAgents()
	if len(agents) > 0 {
		out, err := json.Marshal(command)
		if err != nil {
			panic(err)
		}
		var activeAgents []entities.Agent
		for _, agent := range agents {
			if agent.Active {
				activeAgents = append(activeAgents, agent)
			}
		}
		if len(activeAgents) == 0 {
			fmt.Println("No active agents available, retrying in 10 seconds")
			time.Sleep(10 * time.Second)
			master.StartOperator(command)
		} else {
			for agentId, agent := range activeAgents {
				loops := 0
				for loops < 3 {
					fmt.Println("Trying Agent: " + agent.Id)
					master.publishMessage(constants.TopicPrefix+agents[agentId].Id, string(out), 2)
					if master.checkOperatorDeployed(command.Data.Config.OperatorId + "-" + command.Data.Config.PipelineId) {
						break
					}
					loops++
					time.Sleep(5 * time.Second)
				}
			}
		}
	} else {
		fmt.Println("No agents available")
	}
}

func (master *Master) checkOperatorDeployed(operatorId string) (created bool) {
	created = false
	loops := 0
	operatorJob := entities.OperatorJob{}
	for loops < 5 {
		if err := master.DB.GetOperator(operatorId, &operatorJob); err != nil {

		} else {
			if operatorJob.Response == "Error" {
				fmt.Println(operatorJob.ResponseMessage)
			}
			created = true
			break
		}
		loops++
		fmt.Println("Could not find job in time")
		time.Sleep(10 * time.Second)
	}
	return
}

func (master *Master) StopOperator(command entities.ControlCommand) error {
	operatorJob := entities.OperatorJob{}
	if err := master.DB.GetOperator(command.Data.Config.OperatorId, &operatorJob); err != nil {
		return err
	}
	command.Data = operatorJob
	out, err := json.Marshal(command)
	if err != nil {
		return err
	}
	master.publishMessage(constants.TopicPrefix+operatorJob.Agent.Id, string(out), 2)

	if err := master.DB.DeleteOperator(command.Data.Config.OperatorId); err != nil {
		return err
	}

	return nil
}

func (master *Master) publishMessage(topic string, message string, qos int) {
	master.Client.PublishMessage(topic, message, qos)
}
