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
	"context"

	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
)


func (master *Master) PublishMessage(topic string, message string, qos int) {
	master.Client.Publish(topic, message, qos)
}

func (master *Master) HandleAgentStartOperatorResponse(response operatorEntities.StartOperatorAgentResponse) {
	logging.Logger.Debug("Handle agent response to start operator command")
	ctx := context.Background()
	operator, err := master.DB.GetOperator(ctx, response.PipelineId, response.OperatorId)
	if err != nil {
		logging.Logger.Error("Cant load operator from DB: ", err)
		return
	}

	operator.AgentId = response.AgentId
	operator.DeploymentState = response.DeploymentState
	operator.ContainerId = response.ContainerId

	if err := master.DB.CreateOrUpdateOperator(ctx, operator); err != nil {
		logging.Logger.Error("Cant save new operator state: ", err)
	}
}

func (master *Master) HandleAgentStopOperatorResponse(response operatorEntities.StopOperatorAgentResponse) {
	logging.Logger.Debug("Handle agent response to stop operator command")
	operator := operatorEntities.Operator{}
	operatorID := response.OperatorId
	pipelineID := response.PipelineId
	ctx := context.Background()

	operator, err := master.DB.GetOperator(ctx, pipelineID, operatorID)
	if err != nil {
		logging.Logger.Error("Cant load operator %s from DB: %w", operatorID, err)
		return
	}

	newOperatorState := response.DeploymentState

	if newOperatorState == "not stopped" {
		operator.AgentId = response.AgentId
		operator.DeploymentState = newOperatorState
		if err := master.DB.CreateOrUpdateOperator(ctx, operator); err != nil {
			logging.Logger.Error("Cant save new operator %s: %w", operatorID, err)
		}
	}

	if newOperatorState == "stopped" {
		if err := master.DB.DeleteOperator(ctx, pipelineID, operatorID); err != nil {
			logging.Logger.Error("Cant delete operator %s: %w", operatorID, err)
		}
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
	ctx := context.Background()
	for _, operatorStartCmd := range(syncMsg) {
		operatorID := operatorStartCmd.OperatorIDs.OperatorId
		pipelineID := operatorStartCmd.OperatorIDs.PipelineId
		_, err := master.DB.GetOperator(ctx, pipelineID, operatorID)
		if err != nil {
			// operator does not exists
			// TODO: use sqlite as db
			logging.Logger.Debug("Start missing operator:" + operatorStartCmd.OperatorIDs.OperatorId)
			master.StartOperator(operatorStartCmd)
		}
	}
	logging.Logger.Debug("Completed missing operators check")
}

func (master *Master) stopOperatorOrphans(syncMsg []operatorEntities.StartOperatorControlCommand) {
	logging.Logger.Debug("Stop orphan operators")
	expectedOperatorIDs := map[string]string{}
	for _, operatorStartCmd := range(syncMsg) {
		expectedOperatorIDs[operatorStartCmd.OperatorIDs.OperatorId] = ""
	}
	logging.Logger.Debug("Expected operators %+v", expectedOperatorIDs)

	ctx := context.Background()
	currentOperatorIDs, err := master.DB.GetOperatorIDs(ctx)
	if err != nil {
		logging.Logger.Error("Cant load current operators: " + err.Error())
		return
	}
	logging.Logger.Debug("Current operators %+v", currentOperatorIDs)

	for _, operatorID := range(currentOperatorIDs) {
		_, contains := expectedOperatorIDs[operatorID]
		if !contains {
			logging.Logger.Debug("Stop orphan operator " + operatorID)
			stopCommand := operatorEntities.StopOperatorControlCommand{
				OperatorIDs: operatorEntities.OperatorIDs{OperatorId: operatorID},
			}
			master.StopOperator(stopCommand)
		} 
	}
	logging.Logger.Debug("Completed orphan operators check")
}

func (master *Master) SyncOperatorStates(syncMsg []operatorEntities.StartOperatorControlCommand) {
	master.startMissingOperators(syncMsg)
	master.stopOperatorOrphans(syncMsg)
}
