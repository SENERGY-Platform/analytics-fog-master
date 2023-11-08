package master

import (
	"encoding/json"

	agentLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	controlEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/control"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"

	"time"
)

func (master *Master) CheckAgents() {
	for {
		agents := master.DB.GetAllAgents()
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
	agent := agentLib.Agent{}
	for i := 0; i < 3; i++ {
		time.Sleep(5 * time.Second)
		if err := master.DB.GetAgent(id, &agent); err != nil {
			logging.Logger.Error("Could not find agent record")
			break
		}

		if time.Now().Sub(agent.Updated).Seconds() > 120 {
			if agent.Active == true {
				logging.Logger.Debugf("Agent %s not reachable -> mark unactive\n", id)
				agent.Active = false
				if err := master.DB.SaveAgent(id, agent); err != nil {
					logging.Logger.Error("Could not write agent record ", err)
				}
			}
		} else {
			if agent.Active == false {
				logging.Logger.Debugf("Agent %s reachable again -> mark active\n", id)
				agent.Active = true
				if err := master.DB.SaveAgent(id, agent); err != nil {
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
	}

	if err := master.DB.SaveAgent(id, agent); err != nil {
		logging.Logger.Error(err)
		return err
	}
	return nil
}

func (master *Master) PongAgent(agentConf agentLib.Configuration) error {
	id := agentConf.Id
	agent := agentLib.Agent{
		Id:      id,
		Active:  true,
		Updated: time.Now().UTC(),
	}
	if err := master.DB.SaveAgent(id, agent); err != nil {
		logging.Logger.Error(err)
		return err
	}
	return nil
}

func (master *Master) PublishMessageToAgent(message string, agentID string, qos int) {
	master.PublishMessage(agentLib.AgentsTopic+"/"+agentID, message, qos)
}
