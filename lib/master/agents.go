package master

import (
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/constants"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/entities"
	"time"
)

func (master *Master) CheckAgents() {
	for {
		agents := master.DB.GetAllAgents()
		out, err := json.Marshal(entities.ControlCommand{Command: "ping", Data: entities.OperatorJob{Agent: entities.Agent{Updated: time.Now()}}})
		command := string(out)
		if err != nil {
			panic(err)
		}
		if len(agents) > 0 {
			for agentId := range agents {
				go master.checkAgent(agents[agentId].Id, &command)
			}
		} else {
			fmt.Println("No agents available")
		}
		time.Sleep(60 * time.Second)
	}
}

func (master *Master) checkAgent(id string, command *string) {
	master.publishMessage(constants.TopicPrefix+id, *command, 1)
	agent := entities.Agent{}
	for i := 0; i < 3; i++ {
		time.Sleep(5 * time.Second)
		if err := master.DB.GetAgent(id, &agent); err != nil {
			fmt.Println("Could not find agent record")
			break
		}
		// agent is not reachable
		if time.Now().Sub(agent.Updated).Seconds() > 120 {
			if agent.Active == true {
				agent.Active = false
				if err := master.DB.SaveAgent(id, agent); err != nil {
					fmt.Println("Could not write agent record")
				}
			}
		} else {
			if agent.Active == false {
				agent.Active = true
				if err := master.DB.SaveAgent(id, agent); err != nil {
					fmt.Println("Could not write agent record")
				}
			}
			break
		}
	}
}

func (master *Master) RegisterAgent(id string, agent entities.Agent) error {
	agent.Active = true
	if err := master.DB.SaveAgent(id, agent); err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}

func (master *Master) PongAgent(id string, agent entities.Agent) error {
	agent.Updated = time.Now().UTC()
	if err := master.DB.SaveAgent(id, agent); err != nil {
		fmt.Println("Error", err)
		return err
	}
	return nil
}
