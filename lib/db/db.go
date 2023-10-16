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

package db

import (
	"encoding/json"
	"fmt"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/entities"
	scribble "github.com/nanobox-io/golang-scribble"
)

type FileDatabase struct {
	DB *scribble.Driver
}

func NewFileDatabase() (*FileDatabase, error) {
	dir := "./data"
	fileDB, err := scribble.New(dir, nil)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}

	db := &FileDatabase{
		DB: fileDB,
	}
	return db, nil
}

func (db *FileDatabase) GetAllAgents() (agents []entities.Agent) {
	records, err := db.DB.ReadAll("agents")
	if err != nil {
		fmt.Println("Error", err)
	}
	for _, a := range records {
		agent := entities.Agent{}
		if err := json.Unmarshal([]byte(a), &agent); err != nil {
			fmt.Println("Error", err)
		}
		agents = append(agents, agent)
	}
	return
}

func (db *FileDatabase) GetAgent(id string, agent *entities.Agent) error {
	err := db.DB.Read("agents", id, &agent)
	return err
}

func (db *FileDatabase) SaveAgent(id string, agent entities.Agent) error {
	err := db.DB.Write("agents", id, &agent)
	return err
}

func (db *FileDatabase) DeleteOperator(operatorID string) error {
	err := db.DB.Delete("operatorJobs", operatorID)
	return err
}

func (db *FileDatabase) SaveOperator(operator entities.OperatorJob) error {
	err := db.DB.Write("operatorJobs", operator.Config.OperatorId, operator)
	return err
}

// TODO: pipelineID + operatorID oder nur operatorID
func (db *FileDatabase) GetOperator(operatorID string, operatorJob *entities.OperatorJob) error {
	err := db.DB.Read("operatorJobs", operatorID, &operatorJob)
	return err
}
