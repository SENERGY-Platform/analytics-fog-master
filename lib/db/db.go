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
	"strings"
	"os"
	"path/filepath"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"

	scribble "github.com/nanobox-io/golang-scribble"
)

// TODO https://stackoverflow.com/questions/29981050/concurrent-writing-to-a-file

type FileDatabase struct {
	DB *scribble.Driver
	DataDir string
}

func NewFileDatabase(dataDir string) (*FileDatabase, error) {
	fileDB, err := scribble.New(dataDir, nil)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}

	db := &FileDatabase{
		DB: fileDB,
		DataDir: dataDir,
	}
	return db, nil
}

func (db *FileDatabase) GetAllAgents() (agents []agentEntities.Agent) {
	records, err := db.DB.ReadAll("agents")
	if err != nil {
		fmt.Println("Error", err)
	}
	for _, a := range records {
		agent := agentEntities.Agent{}
		if err := json.Unmarshal([]byte(a), &agent); err != nil {
			fmt.Println("Error", err)
		}
		agents = append(agents, agent)
	}
	return
}

func (db *FileDatabase) GetAgent(id string, agent *agentEntities.Agent) error {
	err := db.DB.Read("agents", id, &agent)
	return err
}

func (db *FileDatabase) SaveAgent(id string, agent agentEntities.Agent) error {
	err := db.DB.Write("agents", id, &agent)
	return err
}

func (db *FileDatabase) DeleteOperator(operatorID string) error {
	err := db.DB.Delete("operatorJobs", operatorID)
	return err
}

func (db *FileDatabase) SaveOperator(operatorJob operatorEntities.Operator) error {
	operatorID := operatorJob.Config.OperatorId
	err := db.DB.Write("operatorJobs", operatorID, operatorJob)
	return err
}

func (db *FileDatabase) GetOperator(operatorID string, operatorJob *operatorEntities.Operator) error {
	err := db.DB.Read("operatorJobs", operatorID, &operatorJob)
	return err
}

func (db *FileDatabase) GetOperatorIDs() (operatorIDs []string, err error) {
	files, err := os.ReadDir(filepath.Join(db.DataDir, "operatorJobs"))
    if err != nil {
		return
    }

    for _, file := range files {
		fileName := file.Name()
        operatorIDs = append(operatorIDs, strings.Split(fileName, ".")[0])
    }
	return 
}