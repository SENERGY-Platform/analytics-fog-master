package controller

import (
	"context"
	"encoding/json"
	"fmt"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
)

func (controller *Controller) operatorCanBeStopped(operator operatorEntities.Operator) bool {
	operatorID := operator.OperatorIDs.OperatorId
	logging.Logger.Debug(fmt.Sprintf("Operator State is: %s", operator.DeploymentState))
	if operator.DeploymentState == "starting" {
		logging.Logger.Debug(fmt.Sprintf("Operator %s is starting. Dont stop until finished", operatorID))
		return false
	}

	if operator.DeploymentState == "stopping" {
		logging.Logger.Debug(fmt.Sprintf("Operator %s is already stopping", operatorID))
		return false
	}
	return true
}

func (controller *Controller) stopOperator(command operatorEntities.StopOperatorControlCommand) error {
	operatorID := command.OperatorId
	pipelineID := command.PipelineId
	ctx := context.Background()

	operator, err := controller.DB.GetOperator(ctx, pipelineID, operatorID); 
	if err != nil {
		logging.Logger.Error("Cant load operator %s after receving stop command: %s", operatorID, err)
		return err
	}

	if !controller.operatorCanBeStopped(operator) {
		return nil
	}

	agentID := operator.AgentId
	stopOperatorAgentCommand := operatorEntities.StopOperatorAgentControlCommand{
		ContainerId: operator.ContainerId,
		OperatorIDs: operatorEntities.OperatorIDs{
			OperatorId: operatorID,
			PipelineId: command.PipelineId,
		},
	}
	stopOperatorAgentMsg, err := json.Marshal(stopOperatorAgentCommand)
	if err != nil {
		logging.Logger.Error("Cant marshal stop command", "error", err.Error())
		return err
	}

	logging.Logger.Debug(fmt.Sprintf("Try to stop operator %s at agent %s", operatorID, agentID))
	logging.Logger.Debug(fmt.Sprintf("Send stop command to agent: %s", agentID))
	err = controller.Client.Publish(agentEntities.GetStopOperatorAgentTopic(agentID), string(stopOperatorAgentMsg), 2)
	if err != nil {
		logging.Logger.Error("Cant publish operator stop command", "error", err)
		return err
	}

	// Mark as stopping after publish!
	operator.DeploymentState = "stopping"
	if err := controller.DB.CreateOrUpdateOperator(ctx, operator); err != nil {
		logging.Logger.Error("Error saving operator after receiving stop command", "error", err)
		return err
	}

	return nil
}