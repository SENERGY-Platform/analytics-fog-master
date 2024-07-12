package controller

import (
	"encoding/json"

	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"

	"time"
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

func (controller *Controller) RemoveStoppedContainer(operator operatorEntities.Operator) error {
	if operator.State == "stopped" {
		// operator was stopped by the agent but response did not reach master, so it got not deleted
		if err := controller.DB.DeleteOperator(operator.Config.OperatorIDs.OperatorId); err != nil {
			return err
		}
	}
	return nil
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

	err := controller.RemoveStoppedContainer(operator)
	if err != nil {
		return err
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
	controller.Client.Publish(agentEntities.GetStopOperatorAgentTopic(agentID), string(stopOperatorAgentMsg), 2)
	
	// Mark as stopping after first publish!
	operator.State = "stopping"
	if err := controller.DB.SaveOperator(operator); err != nil {
		logging.Logger.Errorf("Error saving operator after receiving stop command: %s", err)
		return err
	}

	if controller.checkOperatorWasStopped(operatorID) {
		logging.Logger.Debugf("Agent %s stopped operator %s successfully\n", agentID, operatorID)
		if err := controller.DB.DeleteOperator(operatorID); err != nil {
			return err
		}
		logging.Logger.Debugf("Deleted operator: %s successfully\n", operatorID)
		return nil
	}

	return nil
}

func (controller *Controller) checkOperatorWasStopped(operatorID string) (stopped bool) {
	stopped = false
	loops := 0
	operator := operatorEntities.Operator{}

	for loops <= controller.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Check if operator %s was stopped [%d/%d]", operatorID, loops+1, controller.StartOperatorConfig.Retries+1)

		if err := controller.DB.GetOperator(operatorID, &operator); err != nil {

		} else {
			if operator.State == "stopped" {
				stopped = true
				return
			}

		}
		loops++
		time.Sleep(time.Duration(controller.StartOperatorConfig.Timeout) * time.Second)
	}
	return
}