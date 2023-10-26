package db

import (
	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

type DB interface {
	GetAllAgents() (agents []agentEntities.Agent)
	GetAgent(id string, agent *agentEntities.Agent) error
	SaveAgent(id string, agent agentEntities.Agent) error
	DeleteOperator(operatorID string) error
	GetOperator(operatorID string, operatorJob *operatorEntities.Operator) error
	SaveOperator(operator operatorEntities.Operator) error
}
