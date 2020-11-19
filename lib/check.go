/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lib

import (
	"encoding/json"
	"fmt"
	"time"
)

func CheckAgents() {
	for {
		agents := GetAllAgents()
		out, err := json.Marshal(ControlCommand{Command: "ping", Data: OperatorJob{Agent: Agent{Updated: time.Now()}}})
		command := string(out)
		if err != nil {
			panic(err)
		}
		if len(agents) > 0 {
			for agentId := range agents {
				go checkAgent(agents[agentId].Id, &command)
			}
		} else {
			fmt.Println("No agents available")
		}
		time.Sleep(60 * time.Second)
	}
}

func checkAgent(id string, command *string) {
	publishMessage(TopicPrefix+id, *command, 1)
	agent := Agent{}
	for i := 0; i < 3; i++ {
		time.Sleep(5 * time.Second)
		if err := DB().Read("agents", id, &agent); err != nil {
			fmt.Println("Could not find agent record")
			break
		}
		// agent is not reachable
		if time.Now().Sub(agent.Updated).Seconds() > 120 {
			if agent.Active == true {
				agent.Active = false
				if err := DB().Write("agents", id, &agent); err != nil {
					fmt.Println("Could not write agent record")
				}
			}
		} else {
			if agent.Active == false {
				agent.Active = true
				if err := DB().Write("agents", id, &agent); err != nil {
					fmt.Println("Could not write agent record")
				}
			}
			break
		}
	}
}
