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
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
)


func (master *Master) PublishMessage(topic string, message string, qos int) {
	master.Client.Publish(topic, message, qos)
}

func (master *Master) HandleAgentStartOperatorResponse(response operatorEntities.StartOperatorAgentResponse) {
	logging.Logger.Debugf("Handle agent response to start operator command")
	operator := operatorEntities.Operator{}

	err := master.DB.GetOperator(response.OperatorId, &operator)
	if err != nil {
		logging.Logger.Errorf("Cant load operator from DB: ", err)
		return
	}

	operator.Agent = response.Agent.Id
	operator.State = response.OperatorState
	operator.DeploymentReference = response.ContainerId

	if err := master.DB.SaveOperator(operator); err != nil {
		logging.Logger.Errorf("Cant save new operator state: ", err)
	}
}

func (master *Master) HandleAgentStopOperatorResponse(response operatorEntities.StopOperatorAgentResponse) {
	logging.Logger.Debugf("Handle agent response to start operator command")
	operator := operatorEntities.Operator{}

	err := master.DB.GetOperator(response.OperatorId, &operator)
	if err != nil {
		logging.Logger.Errorf("Cant load operator from DB: ", err)
		return
	}

	operator.Agent = response.Agent.Id
	operator.State = response.OperatorState

	if err := master.DB.SaveOperator(operator); err != nil {
		logging.Logger.Errorf("Cant save new operator state: ", err)
	}
}


func (master *Master) StartOperator(command operatorEntities.StartOperatorControlCommand) {
	master.OperatorController.StartOperator(command)
}

func (master *Master) StopOperator(command operatorEntities.StopOperatorControlCommand) {
	master.OperatorController.StopOperator(command)
}

func (master *Master) startMissingOperators(syncMsg []operatorEntities.StartOperatorControlCommand) {
	logging.Logger.Debug("Start missing operators")
	for _, operatorStartCmd := range(syncMsg) {
		err := master.DB.GetOperator(operatorStartCmd.Config.OperatorID)
		if err != nil {
			// operator does not exists
			// TODO: better way to check
			master.StartOperator(operatorStartCmd)
		}
	}
}

func (master *Master) stopOperatorOrphans(syncMsg []operatorEntities.StartOperatorControlCommand) {
	logging.Logger.Debug("Stop orphan operators")
	expectedOperatorIDs := map[string]string{}
	for _, operatorStartCmd := range(syncMsg) {
		expectedOperatorIDs[operatorStartCmd.Config.OperatorID] = ""
	}

	for _, operator := range(master.DB.GetOperators()) {
		operatorID := operator.Config.OperatorID
		_, contains := expectedOperatorIDs[operatorID]
		if !contains {
			stopCommand := operatorEntities.StopOperatorControlCommand{
				OperatorID: operatorID,
			}
			master.StopOperator(stopCommand)
		} 
	}
}

func (master *Master) SyncOperatorStates(syncMsg []operatorEntities.StartOperatorControlCommand) {
	master.startMissingOperators(syncMsg)
	master.stopOperatorOrphans(syncMsg)
}
