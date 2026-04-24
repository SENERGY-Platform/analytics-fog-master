/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"context"

	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/storage"
)

type Controller struct {
	OperatorStartCommands chan operatorEntities.StartOperatorControlCommand
	OperatorStopCommands  chan operatorEntities.StopOperatorControlCommand
	Ctx                   context.Context
	Client                *mqtt.MQTTClient
	DB                    storage.DB
}

func NewController(ctx context.Context, mqttClient *mqtt.MQTTClient, db storage.DB) *Controller {
	return &Controller{
		OperatorStartCommands: make(chan operatorEntities.StartOperatorControlCommand),
		OperatorStopCommands:  make(chan operatorEntities.StopOperatorControlCommand),
		Ctx:                   ctx,
		Client:                mqttClient,
		DB:                    db,
	}
}

func (controller *Controller) Start() {
	for {
		select {
		case startCommand := <-controller.OperatorStartCommands:
			controller.startOperator(startCommand)
		case stopCommand := <-controller.OperatorStopCommands:
			controller.stopOperator(stopCommand)
		case <-controller.Ctx.Done():
			return
		}
	}
}

func (controller *Controller) StartOperator(command operatorEntities.StartOperatorControlCommand) {
	controller.OperatorStartCommands <- command
}

func (controller *Controller) StopOperator(command operatorEntities.StopOperatorControlCommand) {
	controller.OperatorStopCommands <- command
}
