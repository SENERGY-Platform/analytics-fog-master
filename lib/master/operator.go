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
	"errors"
	"fmt"

	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/storage"
)


func (master *Master) PublishMessage(topic string, message string, qos int) {
	master.Client.Publish(topic, message, qos)
}

func (master *Master) HandleAgentStartOperatorResponse(response operatorEntities.StartOperatorAgentResponse) {
	logging.Logger.Debug("Handle agent response to start operator command")
	ctx := context.Background()
	operator, err := master.DB.GetOperator(ctx, response.PipelineId, response.OperatorId)
	if err != nil {
		logging.Logger.Error("Cant load operator from DB", "error", err)
		return
	}

	operator.AgentId = response.AgentId
	operator.DeploymentState = response.DeploymentState
	operator.ContainerId = response.ContainerId

	if err := master.DB.CreateOrUpdateOperator(ctx, operator); err != nil {
		logging.Logger.Error("Cant save new operator state", "error", err)
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
		logging.Logger.Error(fmt.Sprintf("Cant load operator %s from DB", operatorID), "error", err)
		return
	}

	newOperatorState := response.DeploymentState

	if newOperatorState == "not stopped" {
		operator.AgentId = response.AgentId
		operator.DeploymentState = newOperatorState
		if err := master.DB.CreateOrUpdateOperator(ctx, operator); err != nil {
			logging.Logger.Error(fmt.Sprintf("Cant save new operator %s", operatorID), "error", err)
		}
	}

	if newOperatorState == "stopped" {
		if err := master.DB.DeleteOperator(ctx, pipelineID, operatorID); err != nil {
			logging.Logger.Error(fmt.Sprintf("Cant delete operator %s", operatorID), "error", err)
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
			if errors.Is(err, storage.NotFoundErr) {
				logging.Logger.Debug("Start missing operator:" + operatorStartCmd.OperatorIDs.OperatorId)
				master.StartOperator(operatorStartCmd)
			}
			logging.Logger.Error("Cant check if operator already exists", "error", err)
			return
		}
	}
	logging.Logger.Debug("Completed missing operators check")
}

func (master *Master) stopOperatorOrphans(syncMsg []operatorEntities.StartOperatorControlCommand) {
	logging.Logger.Debug("Stop orphan operators")
	expectedOperators := map[string]map[string]struct{}{}
	for _, operatorStartCmd := range(syncMsg) {
		expectedOperators[operatorStartCmd.PipelineId][operatorStartCmd.OperatorId] = struct{}{}
	}
	logging.Logger.Debug(fmt.Sprintf("Expected operators %+v", expectedOperators))

	ctx := context.Background()
	currentOperators, err := master.DB.GetOperators(ctx)
	if err != nil {
		logging.Logger.Error("Cant load current operators: " + err.Error())
		return
	}
	logging.Logger.Debug(fmt.Sprintf("Current operators %+v", currentOperators))

	for _, operator := range(currentOperators) {
		currentPipelineID := operator.PipelineId
		currentOperatorID := operator.OperatorId
		_, contains := expectedOperators[currentPipelineID][currentOperatorID]
		if !contains {
			logging.Logger.Debug(fmt.Sprintf("Stop orphan operator: %s from pipeline: %s", currentOperatorID, currentPipelineID))
			stopCommand := operatorEntities.StopOperatorControlCommand{
				OperatorIDs: operatorEntities.OperatorIDs{
					OperatorId: currentOperatorID,
					PipelineId: currentPipelineID,
				},
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
