package controller

import (
	"context"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/config"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/storage"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

type Controller struct {
	OperatorStartCommands chan operatorEntities.StartOperatorControlCommand
	OperatorStopCommands chan operatorEntities.StopOperatorControlCommand
	Ctx context.Context
	Client              *mqtt.MQTTClient
	DB                  storage.DB
	StartOperatorConfig config.StartOperatorConfig
}

func NewController(ctx context.Context, mqttClient *mqtt.MQTTClient, db storage.DB, startOperatorConfig config.StartOperatorConfig) *Controller {
	return &Controller{
		OperatorStartCommands: make(chan operatorEntities.StartOperatorControlCommand),
		OperatorStopCommands: make(chan operatorEntities.StopOperatorControlCommand),
		Ctx: ctx, 
		Client:              mqttClient,
		DB:                  db,
		StartOperatorConfig: startOperatorConfig,
	}
}

func (controller *Controller) Start() {
	for {
		select {
		case startCommand := <- controller.OperatorStartCommands:
			controller.startOperator(startCommand)
		case stopCommand := <- controller.OperatorStopCommands:
			controller.stopOperator(stopCommand)
		case <- controller.Ctx.Done():
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