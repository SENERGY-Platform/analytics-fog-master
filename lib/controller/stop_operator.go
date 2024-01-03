package controller

import (
	"encoding/json"
	"errors"

	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"

	"time"
)

func (controller *Controller) stopOperator(command operatorEntities.StopOperatorControlCommand) error {
	operator := operatorEntities.Operator{}
	operatorID := command.OperatorId
	loops := 0

	if err := controller.DB.GetOperator(operatorID, &operator); err != nil {
		logging.Logger.Errorf("Cant load operator %s after receving stop command: %s", operatorID, err)
		return err
	}

	logging.Logger.Debugf("Operator State is: %s", operator.State)
	if operator.State == "starting" {
		// There might be nothing to stop
		logging.Logger.Debugf("Operator %s is in starting state", operatorID)
		return errors.New("Operator is starting. Try again after it is started.")
	}

	agentID := operator.Agent

	operator.State = "stopping"
	if err := controller.DB.SaveOperator(operator); err != nil {
		logging.Logger.Errorf("Error saving operator after receiving stop command: %s", err)
		return err
	}

	stopOperatorAgent := operatorEntities.StopOperatorAgentControlCommand{
		DeploymentReference: operator.DeploymentReference,
		OperatorID: operatorID,
	}
	stopOperatorAgentMsg, err := json.Marshal(stopOperatorAgent)
	if err != nil {
		logging.Logger.Errorf("Cant marshal stop command")
		return err
	}

	logging.Logger.Debugf("Try to stop operator %s at agent %s", operatorID, agentID)

	for loops < controller.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Send stop command to agent: %s [%d/%d]", agentID, loops, controller.StartOperatorConfig.Retries)
		controller.Client.Publish(agentEntities.GetStopOperatorAgentTopic(agentID), string(stopOperatorAgentMsg), 2)

		if controller.checkOperatorWasStopped(operatorID) {
			logging.Logger.Debugf("Agent %s stopped operator %s successfully\n", agentID, operatorID)
			return nil
			/*if err := master.DB.DeleteOperator(operatorID); err != nil {
				return err
			}*/
		}
		loops++
		time.Sleep(time.Duration(controller.StartOperatorConfig.Timeout) * time.Second)
	}

	return nil
}

// TODO periodic checker generic?
func (controller *Controller) checkOperatorWasStopped(operatorID string) (stopped bool) {
	stopped = false
	loops := 0
	operator := operatorEntities.Operator{}

	for loops < controller.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Check if operator %s was stopped [%d/%d]", operatorID, loops, controller.StartOperatorConfig.Retries)

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