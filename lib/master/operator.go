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
	"errors"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"

	"time"
)

func (master *Master) StartOperator(command operatorEntities.StartOperatorControlCommand) error {
	operator := operatorEntities.Operator{
		StartOperatorMessage: command.Operator,
	}
	if err := master.DB.SaveOperator(operator); err != nil {
		logging.Logger.Errorf("Error saving operator after receiving start command: %s", err)
		return err
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
				err := master.StartOperatorAtAgent(command, agent.Id)
				if err != nil {
					logging.Logger.Debugf("Agent %s did not deploy operator -> try next agent\n", agent.Id)
					return err
				}
				return nil
			}
		}
	} else {
		logging.Logger.Debug("No agents available")
	}
	return nil
}

func (master *Master) StartOperatorAtAgent(command operatorEntities.StartOperatorControlCommand, agentId string) (err error) {
	loops := 0
	commandValue, err := json.Marshal(command)
	if err != nil {
		logging.Logger.Errorf("Error marshalling start command: %s", err)
		return err
	}

	for loops < master.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Send start command to agent: %s [%d/%d]", agentId, loops, master.StartOperatorConfig.Retries)
		master.PublishMessageToAgent(string(commandValue), agentId, 2)

		operatorID := command.Operator.Config.OperatorId
		if master.checkOperatorDeployed(operatorID) {
			logging.Logger.Debugf("Agent %s deployed operator %s successfully\n", agentId, operatorID)
			return nil
		}
		loops++
		time.Sleep(time.Duration(master.StartOperatorConfig.Timeout) * time.Second)
	}

	return errors.New("Retries exceeded. Operator was not deployed.")

	// TODO send stop message in case it got depoyed after the timeout
}

func (master *Master) checkOperatorDeployed(operatorId string) (created bool) {
	created = false
	loops := 0
	operator := operatorEntities.Operator{}
	for loops < master.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Check if operator %s was deployed [%d/%d]", operatorId, loops, master.StartOperatorConfig.Retries)

		if err := master.DB.GetOperator(operatorId, &operator); err != nil {

		} else {
			if operator.Event.Response == operatorEntities.OperatorDeployedError {
				logging.Logger.Debugln(operator.Event.ResponseMessage)
				return
			} else if operator.Event.Response == operatorEntities.OperatorDeployedSuccessfully {
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
		logging.Logger.Errorf("Cant load operator %s after receving stop command: %s", operatorID, err)
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

		master.PublishMessageToAgent(string(out), agentID, 2)

		if master.checkOperatorWasStopped(operatorID) {
			logging.Logger.Debugf("Agent %s stopped operator %s successfully\n", agentID, operatorID)
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
		logging.Logger.Debugf("Check if operator %s was stopped [%d/%d]", operatorID, loops, master.StartOperatorConfig.Retries)

		if err := master.DB.GetOperator(operatorID, &operator); err != nil {

		} else {
			// TODO wie resposne abspeicerhn ??
			if operator.Event.Response == operatorEntities.OperatorDeployedError {
				logging.Logger.Debugln(operator.Event.ResponseMessage)
				return
			} else if operator.Event.Response == operatorEntities.OperatorDeployedSuccessfully {
				stopped = true
				return
			}

		}
		loops++
		time.Sleep(time.Duration(master.StartOperatorConfig.Timeout) * time.Second)
	}
	return
}

func (master *Master) PublishMessage(topic string, message string, qos int) {
	master.Client.Publish(topic, message, qos)
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
