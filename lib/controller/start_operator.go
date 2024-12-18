package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/storage"
)

func (controller *Controller) operatorIsAlreadyDeployedOrStopping(command operatorEntities.StartOperatorControlCommand) (bool, error) {
	ctx := context.Background()
	operator, err := controller.DB.GetOperator(ctx, command.OperatorIDs.PipelineId, command.OperatorIDs.OperatorId)
	if err != nil {
		if errors.Is(err, storage.NotFoundErr) {
			return false, nil
		}
		return false, fmt.Errorf("Cant load operator to check deployment state: %w", err)
	}

	requestedOperatorID := command.OperatorIDs.OperatorId
	requestedPipelineID := command.OperatorIDs.PipelineId
	opState := operator.DeploymentState
	if opState == "starting" {
		logging.Logger.Debug(fmt.Sprintf("Operator %s (Pipeline: %s) is starting. Dont start until starting process finishes", requestedOperatorID, requestedPipelineID))
		return true, nil
	}

	if opState == "started" {
		logging.Logger.Debug(fmt.Sprintf("Operator %s (Pipeline: %s) is already started.", requestedOperatorID, requestedPipelineID))
		return true, nil
	}

	if opState == "stopping" {
		logging.Logger.Debug(fmt.Sprintf("Operator %s (Pipeline: %s) is stopping. Dont start until stopping process finishes", requestedOperatorID, requestedPipelineID))
		return true, nil
	}

	return false, nil
}

func (controller *Controller) startOperator(command operatorEntities.StartOperatorControlCommand) error {
	/*
		To start an operator, first an agent has to be selected.
		Then, the start command is sent to the agent.
		As we the request is async, we will get the response from the agent eventually
		Caution: Duplicate start commands will lead to duplicate deployments.
		Keep in mind, that there is sync process with the platform which will also lead to new start commands
	*/
	operatorID := command.OperatorIDs.OperatorId
	pipelineID := command.OperatorIDs.PipelineId

	operatorIsDeployed, err := controller.operatorIsAlreadyDeployedOrStopping(command)
	if err != nil && !errors.Is(err, storage.NotFoundErr) {
		logging.Logger.Error("Cant check deployment state", "error", err)
		return err
	}
	if operatorIsDeployed {
		return nil
	}

	ctx := context.Background()
	agents, err := controller.DB.GetAllAgents(ctx)
	if len(agents) == 0 {
		logging.Logger.Error("No agents available")
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

	commandValue, err := json.Marshal(command)
	if err != nil {
		logging.Logger.Error("Error marshalling start command", "error", err)
		return err
	}
	logging.Logger.Debug(fmt.Sprintf("Try to start operator %s (Pipeline: %s) at agent %s", operatorID, pipelineID, agent.Id))
	controller.Client.Publish(agentEntities.GetStartOperatorAgentTopic(agent.Id), string(commandValue), 2)

	logging.Logger.Debug("Set operator state to starting")
	operator := operatorEntities.Operator{
		DeploymentState:             "starting",
		StartOperatorControlCommand: command,
		AgentId:                     agent.Id,
		TimeOfLastHeartbeat:         time.Now(),
	}
	if err := controller.DB.CreateOrUpdateOperator(ctx, operator); err != nil {
		logging.Logger.Error(fmt.Sprintf("Error saving operator  %s (Pipeline: %s) after receiving start command: %s", operatorID, pipelineID, err.Error()))
		return err
	}

	logging.Logger.Debug(fmt.Sprintf("Operator %s (Pipeline: %s) was started", operatorID, pipelineID))
	return nil
}

func (controller *Controller) SelectAgent(activeAgents []agentEntities.Agent) agentEntities.Agent {
	randomIdx := rand.Intn(len(activeAgents))
	return activeAgents[randomIdx]
}
