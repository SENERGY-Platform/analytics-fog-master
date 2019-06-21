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

func startOperator(command ControlCommand) {
	agents := GetAllAgents()
	if len(agents) > 0 {
		out, err := json.Marshal(command)
		if err != nil {
			panic(err)
		}
		for agentId, agent := range agents {
			if agent.Active == true {
				fmt.Println("Trying Agent: " + agent.Id)
				publishMessage(agents[agentId].Id, string(out))
				if checkOperatorDeployed(command.Data.OperatorId) {
					break
				}
			}
		}
	} else {
		fmt.Println("No agents available")
	}
}

func checkOperatorDeployed(operatorId string) (created bool) {
	created = false
	loops := 0
	operatorJob := OperatorJob{}
	for loops < 6 {
		if err := DB().Read("operatorJobs", operatorId, &operatorJob); err != nil {
			fmt.Println("Could not find job")
		} else {
			fmt.Println("Did find job")
			created = true
			break
		}
		loops++
		time.Sleep(30 * time.Second)
	}
	return
}

func stopOperator(command ControlCommand) {
	operatorJob := OperatorJob{}
	if err := DB().Read("operatorJobs", command.Data.OperatorId, &operatorJob); err != nil {
		fmt.Println("Error", err)
	}
	command.Data = operatorJob
	out, err := json.Marshal(command)
	if err != nil {
		panic(err)
	}
	publishMessage(operatorJob.Agent.Id, string(out))
	if err := DB().Delete("operatorJobs", command.Data.OperatorId); err != nil {
		fmt.Println("Error", err)
	}
}
