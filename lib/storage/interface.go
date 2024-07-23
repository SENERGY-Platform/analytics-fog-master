package storage

import (
	"context"
	"database/sql/driver"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

type DB interface {
	GetAllAgents(ctx context.Context, txItf driver.Tx) ([]agentEntities.Agent, error)
	GetAgent(ctx context.Context, id string, txItf driver.Tx) (agentEntities.Agent, error)
	CreateOrUpdateAgent(ctx context.Context, agent agentEntities.Agent, txItf driver.Tx) error
	DeleteOperator(ctx context.Context, pipelineID, operatorID string, txItf driver.Tx) error
	GetOperator(ctx context.Context, pipelineID, operatorID string, txItf driver.Tx) (operatorEntities.Operator, error)
	CreateOrUpdateOperator(ctx context.Context, operator operatorEntities.Operator, txItf driver.Tx) error
	GetOperators(ctx context.Context, txItf driver.Tx) ([]operatorEntities.Operator, error)
}
