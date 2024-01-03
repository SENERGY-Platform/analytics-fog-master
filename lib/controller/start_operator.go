package controller

import (
	"encoding/json"
	"errors"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"

	"time"
)

func (controller *Controller) startOperator(command operatorEntities.StartOperatorControlCommand) error {
	operator := operatorEntities.Operator{
		StartOperatorControlCommand: command,
		State:                "starting",
	}
	if err := controller.DB.SaveOperator(operator); err != nil {
		logging.Logger.Errorf("Error saving operator after receiving start command: %s", err)
		return err
	}

	agents := controller.DB.GetAllAgents()
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
			controller.StartOperator(command)
		} else {
			for _, agent := range activeAgents {
				logging.Logger.Debugf("Try to start operator at agent %s", agent.Id)
				operatorStarted, err := controller.startOperatorAtAgent(command, agent.Id)
				if err != nil {
					logging.Logger.Errorf("Operator could not be started: %s", err)
					return err
				}
				if !operatorStarted {
					logging.Logger.Debugf("Agent %s did not deploy operator -> try next agent\n", agent.Id)
					continue
				}

				logging.Logger.Debugf("Operator was started")
				return nil
			}
		}
	} else {
		logging.Logger.Debug("No agents available")
	}
	return nil
}

func (controller *Controller) startOperatorAtAgent(command operatorEntities.StartOperatorControlCommand, agentId string) (bool, error) {
	loops := 0
	commandValue, err := json.Marshal(command)
	if err != nil {
		logging.Logger.Errorf("Error marshalling start command: %s", err)
		return false, err
	}

	for loops < controller.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Send start command to agent: %s [%d/%d]", agentId, loops, controller.StartOperatorConfig.Retries)
		controller.Client.Publish(agentEntities.GetStartOperatorAgentTopic(agentId), string(commandValue), 2)
		operatorID := command.Config.OperatorId
		operatorStarted, err := controller.checkOperatorDeployed(operatorID) 
		if err != nil {
			return false, err 
		}	

		if operatorStarted {
			logging.Logger.Debugf("Agent %s deployed operator %s successfully\n", agentId, operatorID)
			return true, nil
		}
		loops++
		time.Sleep(time.Duration(controller.StartOperatorConfig.Timeout) * time.Second)
	}

	return false, nil

	// TODO send stop message in case it got depoyed after the timeout
}

func (controller *Controller) checkOperatorDeployed(operatorId string) (created bool, err error) {
	created = false
	loops := 0
	operator := operatorEntities.Operator{}
	for loops < controller.StartOperatorConfig.Retries {
		logging.Logger.Debugf("Check if operator %s was deployed [%d/%d]", operatorId, loops, controller.StartOperatorConfig.Retries)

		if err = controller.DB.GetOperator(operatorId, &operator); err != nil {
			logging.Logger.Errorf("Cant get operator from DB: %s", err)
			return 
		} else {
			logging.Logger.Debugf("Operator State is: %s", operator.State)
			if operator.State == "started" {
				// Agent deployed operator -> remove response so that later events like stopping can be set
				created = true
				operator.Event = operatorEntities.OperatorAgentResponse{}
				if err = controller.DB.SaveOperator(operator); err != nil {
					logging.Logger.Error(err)
					return
				}
				return
			} else if operator.State == "stopping" {
				// operator state can be started, then stop request comes, override state to stopping or stopped, start retry loop does not get stopped as check happened after override
				logging.Logger.Debugf("Operator is stopping -> dont retry")
				err = errors.New("Operator is in stopping state -> no need to retry")
				return
			} else if operator.State == "stopped" {
				logging.Logger.Debugf("Operator is stopped -> restart")
			}
		}
		loops++
		time.Sleep(time.Duration(controller.StartOperatorConfig.Timeout) * time.Second)
	}
	return
}
