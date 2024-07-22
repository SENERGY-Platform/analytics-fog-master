package master

import (
	"context"
	"encoding/json"

	agentLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	controlEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/control"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"

	"time"
)

func (master *Master) CheckAgents() error {
	ctx := context.Background()
	for {
		agents, err := master.DB.GetAllAgents(ctx)
		if err != nil {
			return err
		}
		if len(agents) > 0 {
			for agentId := range agents {
				go master.checkAgent(agents[agentId].Id)
			}
		} else {
			logging.Logger.Debug("No agents available")
		}
		time.Sleep(60 * time.Second)
	}
}

func (master *Master) checkAgent(id string) {
	out, err := json.Marshal(agentLib.Ping{
		ControlMessage: controlEntities.ControlMessage{
			Command: "ping",
		},
		Updated: time.Now(),
	})
	command := string(out)
	if err != nil {
		panic(err)
	}
	master.PublishMessageToAgent(command, id, 1)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		time.Sleep(5 * time.Second)
		agent, err := master.DB.GetAgent(ctx, id)
		if err != nil {
			logging.Logger.Error("Could not find agent record")
			break
		}

		if time.Now().Sub(agent.Updated).Seconds() > 120 {
			if agent.Active == true {
				logging.Logger.Debug("Agent %s not reachable -> mark unactive\n", id)
				agent.Active = false
				if err := master.DB.CreateOrUpdateAgent(ctx, agent); err != nil {
					logging.Logger.Error("Could not write agent record ", err)
				}
			}
		} else {
			if agent.Active == false {
				logging.Logger.Debug("Agent %s reachable again -> mark active\n", id)
				agent.Active = true
				if err := master.DB.CreateOrUpdateAgent(ctx, agent); err != nil {
					logging.Logger.Error("Could not write agent record ", err)
				}
			}
			break
		}
	}
}

func (master *Master) RegisterAgent(agentConf agentLib.Configuration) error {
	// TODO after poing Active field gets removed??
	// TODO ignore register when agent exists
	id := agentConf.Id

	agent := agentLib.Agent{
		Id:     id,
		Active: true,
		Updated: time.Now().UTC(),
	}
	ctx := context.Background()
	if err := master.DB.CreateOrUpdateAgent(ctx, agent); err != nil {
		logging.Logger.Error(err.Error())
		return err
	}
	return nil
}

func (master *Master) PongAgent(pongMessage agentLib.AgentInfoMessage) error {
	id := pongMessage.Conf.Id
	agent := agentLib.Agent{
		Id:      id,
		Active:  true,
		Updated: time.Now().UTC(),
	}
	ctx := context.Background()
	if err := master.DB.CreateOrUpdateAgent(ctx, agent); err != nil {
		logging.Logger.Error(err.Error())
		return err
	}

	return master.UpdateOperatorStates(pongMessage.CurrentOperatorStates)
}

func (master *Master) UpdateOperatorStates(operatorStates []agentLib.OperatorState) error {
	for _, newOperatorState := range(operatorStates) {
		operator := operatorEntities.Operator{}
		operatorID := newOperatorState.OperatorID
		pipelineID := newOperatorState.PipelineID
		ctx := context.Background()
		operator, err := master.DB.GetOperator(ctx, pipelineID, operatorID)
		if err != nil {
			logging.Logger.Error("Cant load operator %s from DB: %w", operatorID, err)
			return err
		}
		operator.DeploymentState = newOperatorState.State
		operator.ContainerId = newOperatorState.ContainerID
		if newOperatorState.State == "stopped" {
			if err := master.DB.DeleteOperator(ctx, pipelineID, operatorID); err != nil {
				logging.Logger.Error("Cant delete operator %s: %w", operatorID, err)
				return err
			}
			continue
		}
		if err := master.DB.CreateOrUpdateOperator(ctx, operator); err != nil {
			logging.Logger.Error("Cant save new operator %s: %w", operatorID, err)
		}
	}
	return nil
}


func (master *Master) PublishMessageToAgent(message string, agentID string, qos int) {
	master.PublishMessage(agentLib.AgentsTopic+"/"+agentID, message, qos)
}
