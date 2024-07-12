package controller

import (
	"encoding/json"
	"errors"
	"math/rand"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
)

func (controller *Controller) operatorIsAlreadyDeployedOrStopping(command operatorEntities.StartOperatorControlCommand) bool {
	operator := operatorEntities.Operator{}
	err := controller.DB.GetOperator(operator.StartOperatorControlCommand.Config.OperatorId, &operator)
	if err != nil {
		return false
	}

	requestedOperatorID := command.Config.OperatorIDs.OperatorId
	requestedPipelineID := command.Config.OperatorIDs.PipelineId
	if operator.Config.OperatorIDs.OperatorId != requestedOperatorID && operator.Config.OperatorIDs.PipelineId != requestedPipelineID {
		return false
	}
	
	opState := operator.State
	if opState == "starting" {
		logging.Logger.Debugf("Operator %s (Pipeline: %s) is starting. Dont start until response from agent", requestedOperatorID, requestedPipelineID)
		return true
	} 
	
	if opState == "started" {
		logging.Logger.Debugf("Operator %s (Pipeline: %s) is already started.", requestedOperatorID, requestedPipelineID)
		return true
	}
	
	if opState == "stopping" {
		logging.Logger.Debugf("Operator %s (Pipeline: %s) is stopping. Dont start until done", requestedOperatorID, requestedPipelineID)
		return true
	}

	return false
}

func (controller *Controller) startOperator(command operatorEntities.StartOperatorControlCommand) error {
	/* 
		To start an operator, first an agent has to be selected.
		Then, the start command is sent to the agent.
		As we the request is async, we will get the response from the agent eventually
		Caution: Duplicate start commands will lead to duplicate deployments.
		Keep in mind, that there is sync process with the platform which will also lead to new start commands
	*/
	operatorID := command.Config.OperatorIDs.OperatorId
	pipelineID := command.Config.OperatorIDs.PipelineId

	operatorIsDeployed := controller.operatorIsAlreadyDeployedOrStopping(command)
	if operatorIsDeployed {
		return nil
	}

	agents := controller.DB.GetAllAgents()
	if len(agents) == 0 {
		logging.Logger.Debug("No agents available")
		return errors.New("No agents available")
	}
	
	var activeAgents []agentEntities.Agent
	for _, agent := range agents {
		if agent.Active {
			activeAgents = append(activeAgents, agent)
		}
	}
	if len(activeAgents) == 0 {
		/* It will be retried with the next sync anyways. We could think of enabling this part again, 
		but keep in mind that it could create duplicate deployments if no check is done whether operator is currently starting 
		logging.Logger.Debug("No active agents available, retrying in 10 seconds")
		time.Sleep(10 * time.Second)
		controller.StartOperator(command)*/
		return errors.New("No active agents available")
	} 

	agent := controller.SelectAgent(activeAgents)
	
	logging.Logger.Debugf("Try to start operator %s (Pipeline: %s) at agent %s", operatorID, pipelineID, agent.Id)
	commandValue, err := json.Marshal(command)
	if err != nil {
		logging.Logger.Errorf("Error marshalling start command: %s", err)
		return err
	}
	controller.Client.Publish(agentEntities.GetStartOperatorAgentTopic(agent.Id), string(commandValue), 2)

	operator := operatorEntities.Operator{
		StartOperatorControlCommand: command,
		State:                "starting",
	}
	if err := controller.DB.SaveOperator(operator); err != nil {
		logging.Logger.Errorf("Error saving operator  %s (Pipeline: %s) after receiving start command: %s", operatorID, pipelineID, err)
		return err
	}

	logging.Logger.Debugf("Operator %s (Pipeline: %s) was started", operatorID, pipelineID)
	return nil
}

func (controller *Controller) SelectAgent(activeAgents []agentEntities.Agent) agentEntities.Agent {
	randomIdx := rand.Intn(len(activeAgents))
	return activeAgents[randomIdx]
}