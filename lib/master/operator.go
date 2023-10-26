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

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/constants"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"

	"time"
)

func (master *Master) StartOperator(command operatorEntities.StartOperatorControlCommand) {
	operator := operatorEntities.Operator{
		StartOperatorMessage: command.Operator,
	}
	if err := master.DB.SaveOperator(operator); err != nil {
		logging.Logger.Error(err)
	}

	agents := master.DB.GetAllAgents()
	if len(agents) > 0 {
		var activeAgents []agentEntities.Agent
		for _, agent := range agents {
			if agent.Active {
				activeAgents = append(activeAgents, agent)
			}
		}
		if len(activeAgents) == 0 {
			logging.Logger.Debug("No active agents available, retrying in 10 seconds")
			time.Sleep(10 * time.Second)
			master.StartOperator(command)
		} else {
			for _, agent := range activeAgents {
				logging.Logger.Debugf("Try to start operator at agent %s", agent.Id)
				deployed := master.StartOperatorAtAgent(command, agent.Id)
				if deployed {
					return
				}
				logging.Logger.Debugf("Agent %s did not deploy operator -> try next agent\n", agent.Id)
			}
		}
	} else {
		logging.Logger.Debug("No agents available")
	}
}

func (master *Master) StartOperatorAtAgent(command operatorEntities.StartOperatorControlCommand, agentId string) (deployed bool) {
	loops := 0
	deployed = false
	commandValue, err := json.Marshal(command)
	if err != nil {
		panic(err)
	}

	for loops < master.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Send start command to agent: %s [%d/%d]", agentId, loops, master.StartOperatorConfig.Retries)
		master.publishMessage(constants.TopicPrefix+agentId, string(commandValue), 2)

		operatorID := command.Operator.Config.OperatorId
		if master.checkOperatorDeployed(operatorID) {
			logging.Logger.Debugf("Agent %s deployed operator successfully\n", agentId)
			deployed = true
			return
		}
		loops++
		time.Sleep(time.Duration(master.StartOperatorConfig.Timeout) * time.Second)
	}

	return

	// TODO send stop message in case it got depoyed after the timeout
}

func (master *Master) checkOperatorDeployed(operatorId string) (created bool) {
	created = false
	loops := 0
	operator := operatorEntities.Operator{}
	for loops < master.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Check if operator was deployed [%d/%d]", loops, master.StartOperatorConfig.Retries)

		if err := master.DB.GetOperator(operatorId, &operator); err != nil {

		} else {
			if operator.Event.Response == constants.OperatorDeployedError {
				logging.Logger.Debugln(operator.Event.ResponseMessage)
				return
			} else if operator.Event.Response == constants.OperatorDeployedSuccessfully {
				// Agent deployed operator -> remove response so that later events like stopping can be set
				created = true
				operator.Event = operatorEntities.OperatorAgentResponse{}
				if err := master.DB.SaveOperator(operator); err != nil {
					logging.Logger.Error(err)
				}
				return
			}
		}
		loops++
		time.Sleep(time.Duration(master.StartOperatorConfig.Timeout) * time.Second)
	}
	return
}

func (master *Master) StopOperator(command operatorEntities.StopOperatorControlCommand) error {
	operator := operatorEntities.Operator{}
	operatorID := command.OperatorId
	loops := 0

	if err := master.DB.GetOperator(operatorID, &operator); err != nil {
		logging.Logger.Errorf("Cant get operator: %s", err)
		return err
	}
	agentID := operator.Agent

	out, err := json.Marshal(command)
	if err != nil {
		logging.Logger.Errorf("Cant marshal stop command")
		return err
	}

	logging.Logger.Debugf("Try to stop operator %s at agent %s", operatorID, agentID)

	for loops < master.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Send stop command to agent: %s [%d/%d]", agentID, loops, master.StartOperatorConfig.Retries)

		master.publishMessage(constants.TopicPrefix+agentID, string(out), 2)

		if master.checkOperatorWasStopped(operatorID) {
			logging.Logger.Debugf("Agent %s stopped operator successfully\n", agentID)
			if err := master.DB.DeleteOperator(operatorID); err != nil {
				return err
			}
		}
		loops++
		time.Sleep(time.Duration(master.StartOperatorConfig.Timeout) * time.Second)
	}

	return nil
}

// TODO periodic checker generic?
func (master *Master) checkOperatorWasStopped(operatorID string) (stopped bool) {
	stopped = false
	loops := 0
	operator := operatorEntities.Operator{}

	for loops < master.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Check if operator was stopped [%d/%d]", loops, master.StartOperatorConfig.Retries)

		if err := master.DB.GetOperator(operatorID, &operator); err != nil {

		} else {
			// TODO wie resposne abspeicerhn ??
			if operator.Event.Response == constants.OperatorDeployedError {
				logging.Logger.Debugln(operator.Event.ResponseMessage)
				return
			} else if operator.Event.Response == constants.OperatorDeployedSuccessfully {
				stopped = true
				return
			}

		}
		loops++
		time.Sleep(time.Duration(master.StartOperatorConfig.Timeout) * time.Second)
	}
	return
}

func (master *Master) publishMessage(topic string, message string, qos int) {
	master.Client.PublishMessage(topic, message, qos)
}

func (master *Master) HandleAgentOperatorResponse(response operatorEntities.OperatorAgentResponse) {
	operator := operatorEntities.Operator{}

	err := master.DB.GetOperator(response.OperatorId, &operator)
	if err != nil {
		logging.Logger.Error(err)
	}

	operator.Event = response
	operator.Agent = response.Agent.Id
	operator.State = "Running"

	if err := master.DB.SaveOperator(operator); err != nil {
		logging.Logger.Error(err)
	}
}