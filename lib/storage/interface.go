package storage

import (
	"context"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

type DB interface {
	GetAllAgents(ctx context.Context) ([]agentEntities.Agent, error)
	GetAgent(ctx context.Context, id string) (agentEntities.Agent, error)
	CreateOrUpdateAgent(ctx context.Context, agent agentEntities.Agent) error
	DeleteOperator(ctx context.Context, pipelineID, operatorID string) error
	GetOperator(ctx context.Context, pipelineID, operatorID string) (operatorEntities.Operator, error)
	CreateOrUpdateOperator(ctx context.Context, operator operatorEntities.Operator) error
	GetOperators(ctx context.Context) ([]operatorEntities.Operator, error)
}
