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
	"time"

	agentLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
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
	operator.TimeOfLastHeartbeat = response.Time

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
		operator.TimeOfLastHeartbeat = response.Time
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
	logging.Logger.Debug("check for missing operators")
	ctx := context.Background()
	for _, operatorStartCmd := range(syncMsg) {
		operatorID := operatorStartCmd.OperatorIDs.OperatorId
		pipelineID := operatorStartCmd.OperatorIDs.PipelineId
		_, err := master.DB.GetOperator(ctx, pipelineID, operatorID)
		if err != nil {
			if errors.Is(err, storage.NotFoundErr) {
				logging.Logger.Debug("Start missing operator:" + operatorStartCmd.OperatorIDs.OperatorId)
				master.StartOperator(operatorStartCmd)
			} else {
				logging.Logger.Error("Cant check if operator already exists", "error", err)
			}
			return
		}
	}
	logging.Logger.Debug("Completed missing operators check")
}

func (master *Master) stopOperatorOrphans(syncMsg []operatorEntities.StartOperatorControlCommand) {
	logging.Logger.Debug("check for orphan operators")
	expectedOperators := map[string]map[string]struct{}{}
	for _, operatorStartCmd := range(syncMsg) {
		_, ok := expectedOperators[operatorStartCmd.PipelineId]
		if !ok {
			expectedOperators[operatorStartCmd.PipelineId] = map[string]struct{}{}
		}

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

// Update operator states based on given state from the agents
// Delete operator if agent successfully stopped it
func (master *Master) UpdateOperatorStates(operatorStates []agentLib.OperatorState) error {
	for _, newOperatorState := range(operatorStates) {
		logging.Logger.Debug("Update operator state from pong", "new state", newOperatorState)
		operator := operatorEntities.Operator{}
		operatorID := newOperatorState.OperatorID
		pipelineID := newOperatorState.PipelineID
		ctx := context.Background()
		operator, err := master.DB.GetOperator(ctx, pipelineID, operatorID)
		if err != nil {
			// Also the case when agent sends state of an operator that the master does not know 
			logging.Logger.Error(fmt.Sprintf("Cant load operator %s from DB: %s", operatorID, err.Error()))
			return err
		}
		operator.DeploymentState = newOperatorState.State
		operator.ContainerId = newOperatorState.ContainerID
		if newOperatorState.State == "stopped" {
			logging.Logger.Debug("Operator is stopped -> Delete")
			if err := master.DB.DeleteOperator(ctx, pipelineID, operatorID); err != nil {
				logging.Logger.Error("Cant delete operator %s: %w", operatorID, err)
				return err
			}
			continue
		}
		logging.Logger.Debug(fmt.Sprintf("Update operator: %+v", operator))
		if err := master.DB.CreateOrUpdateOperator(ctx, operator); err != nil {
			logging.Logger.Error("Cant save new operator %s: %w", operatorID, err)
		}
	}
	return nil
}

// An operator deployment state can be hanging in `starting` or `stopping` 
// e.g because the agent went down or the master never received the response 
// The deployment should then be retried after some time
// Timeout should be different than the the agent ping timeout
// Could be that agent started the operator but disconnected shortly after
// Master will mark the agent inactive after the timeout to prevent further scheduling
// Agent reconnects and will notify the master of successfull deployment
// As we dont want duplicate deployments -> the timeout here should be way higher 
func (master *Master) MarkStaleOperators(doneCtx context.Context) error {
	interval := time.Duration(master.StaleOperatorCheckInterval)
	logging.Logger.Debug(fmt.Sprintf("Check operator states each %s", interval))
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-doneCtx.Done():
			ticker.Stop()
			return nil
		case <-ticker.C:
			logging.Logger.Debug("Check for stale operator states")
			ctx := context.Background()
			operators, err := master.DB.GetOperators(ctx)
			if err != nil {
				logging.Logger.Error("Cant load operators", "error", err)
			}
			for _, operator := range(operators) {
				if time.Now().Sub(operator.TimeOfLastHeartbeat).Seconds() < master.TimeoutStaleOperator {
					continue
				}

				deploymentState := operator.DeploymentState
				pipelineID := operator.PipelineId
				operatorID := operator.OperatorId
				logging.Logger.Debug(fmt.Sprintf("Operator state is stale: %s", deploymentState), "pipelineID", pipelineID, "operatorID", operatorID)
				if deploymentState == "starting" {
					logging.Logger.Debug("Delete stale starting operator", "pipelineID", pipelineID, "operatorID", operatorID)
					err := master.DB.DeleteOperator(ctx, pipelineID, operatorID)
					if err != nil {
						logging.Logger.Error("Cant delete stale starting operator", "pipelineID", pipelineID, "operatorID", operatorID, "error", err)
					}
				}
		
				if deploymentState == "stopping" {
					logging.Logger.Debug("Mark stale stopping operator as started", "pipelineID", pipelineID, "operatorID", operatorID)
					operator.DeploymentState = "started"
					err := master.DB.CreateOrUpdateOperator(ctx, operator)
					if err != nil {
						logging.Logger.Error("Cant mark stale stopping operator", "pipelineID", pipelineID, "operatorID", operatorID, "error", err)
					}
				}
			}
		}
	}
}