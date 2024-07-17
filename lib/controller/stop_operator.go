package controller

import (
	"encoding/json"
	"fmt"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
)

func (controller *Controller) operatorCanBeStopped(operator operatorEntities.Operator) bool {
	operatorID := operator.Config.OperatorIDs.OperatorId
	logging.Logger.Debugf("Operator State is: %s", operator.State)
	if operator.State == "starting" {
		logging.Logger.Debugf("Operator %s is starting. Dont stop until finished", operatorID)
		return false
	}

	if operator.State == "stopping" {
		logging.Logger.Debugf("Operator %s is already stopping.", operatorID)
		return false
	}
	return true
}

func (controller *Controller) stopOperator(command operatorEntities.StopOperatorControlCommand) error {
	operator := operatorEntities.Operator{}
	operatorID := command.OperatorId

	if err := controller.DB.GetOperator(operatorID, &operator); err != nil {
		logging.Logger.Errorf("Cant load operator %s after receving stop command: %s", operatorID, err)
		return err
	}

	if !controller.operatorCanBeStopped(operator) {
		return nil
	}

	agentID := operator.Agent
	stopOperatorAgentCommand := operatorEntities.StopOperatorAgentControlCommand{
		DeploymentReference: operator.DeploymentReference,
		OperatorID: operatorID,
	}
	stopOperatorAgentMsg, err := json.Marshal(stopOperatorAgentCommand)
	if err != nil {
		logging.Logger.Errorf("Cant marshal stop command")
		return err
	}

	logging.Logger.Debugf("Try to stop operator %s at agent %s", operatorID, agentID)
	logging.Logger.Debugf("Send stop command to agent: %s", agentID)
	err = controller.Client.Publish(agentEntities.GetStopOperatorAgentTopic(agentID), string(stopOperatorAgentMsg), 2)
	if err != nil {
		return fmt.Errorf("Cant publish operator stop command: %w", err)
	}

	// Mark as stopping after publish!
	operator.State = "stopping"
	if err := controller.DB.SaveOperator(operator); err != nil {
		logging.Logger.Errorf("Error saving operator after receiving stop command: %s", err)
		return err
	}

	return nil
}