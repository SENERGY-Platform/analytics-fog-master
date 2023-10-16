package db

import (
	"github.com/SENERGY-Platform/analytics-fog-master/lib/entities"
)

type DB interface {
	GetAllAgents() (agents []entities.Agent)
	GetAgent(id string, agent *entities.Agent) error
	SaveAgent(id string, agent entities.Agent) error
	DeleteOperator(operatorID string) error
	GetOperator(operatorID string, operatorJob *entities.OperatorJob) error
	SaveOperator(operator entities.OperatorJob) error
}
