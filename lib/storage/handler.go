package storage

import (
	"database/sql"
	"fmt"
	agentLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
    "context"
    "time"
)

var tLayout = time.RFC3339Nano

type Handler struct {
	db *sql.DB
}

func New(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) createAgent(ctx context.Context, agent agentLib.Agent) error {
    query := `INSERT INTO agents (id, last_update) VALUES (?, ?)`
    _, err := h.db.Exec(query, agent.Id, agent.Updated)
    if err != nil {
        return fmt.Errorf("createAgent: %v", err)
    }
    return nil
}

func (h *Handler) updateAgent(ctx context.Context, agent agentLib.Agent) error {
    query := `UPDATE agents SET last_update = ?, active = ? WHERE id = ?`
    _, err := h.db.Exec(query, agent.Updated, agent.Id)
    if err != nil {
        return fmt.Errorf("updateAgent: %v", err)
    }
    return nil
}

func (h *Handler) deleteAgent(ctx context.Context, id string) error {
    query := `DELETE FROM agents WHERE id = ?`
    _, err := h.db.Exec(query, id)
    if err != nil {
        return fmt.Errorf("deleteAgent: %v", err)
    }
    return nil
}

func (h *Handler) createOperator(ctx context.Context, operator operatorEntities.Operator) error {
    query := `INSERT INTO operators (pipeline_id, operator_id, state, container_id, error, agent_id) VALUES (?, ?, ?, ?, ?, ?)`
    _, err := h.db.Exec(query, operator.OperatorIDs.PipelineId, operator.OperatorIDs.OperatorId, operator.DeploymentState, operator.ContainerId, operator.DeploymentError, operator.AgentId)
    if err != nil {
        return fmt.Errorf("createOperator: %v", err)
    }
    return nil
}

func (h *Handler) updateOperator(ctx context.Context, operator operatorEntities.Operator) error {
    query := `UPDATE operators SET state = ?, container_id = ?, error = ? WHERE pipeline_id = ? AND operator_id = ?`
    _, err := h.db.Exec(query, operator.DeploymentState, operator.ContainerId, operator.DeploymentError, operator.OperatorIDs.PipelineId, operator.OperatorIDs.OperatorId)
    if err != nil {
        return fmt.Errorf("updateOperator: %v", err)
    }
    return nil
}

func (h *Handler) DeleteOperator(ctx context.Context, pipelineID string, operatorID string) error {
    query := `DELETE FROM operators WHERE pipeline_id = ? AND operator_id = ?`
    _, err := h.db.Exec(query, pipelineID, operatorID)
    if err != nil {
        return fmt.Errorf("deleteOperator: %v", err)
    }
    return nil
}

func (h *Handler) GetAllAgents(ctx context.Context) ([]agentLib.Agent, error) {
    agents := []agentLib.Agent{}
    return agents, nil
}

func (h *Handler) GetAgent(ctx context.Context, id string) (agentLib.Agent, error) {
	agent := agentLib.Agent{}
    return agent, nil
}

func (h *Handler) CreateOrUpdateAgent(ctx context.Context, agent agentLib.Agent) error {
    h.createAgent(ctx, agent)
    h.updateAgent(ctx, agent)
    return nil
}

func (h *Handler) GetOperator(ctx context.Context, pipelineID, operatorID string) (operatorEntities.Operator, error) {
	operator := operatorEntities.Operator{}
    return operator, nil
}

func (h *Handler) CreateOrUpdateOperator(ctx context.Context, operator operatorEntities.Operator) error {
    h.createOperator(ctx, operator)
    h.updateOperator(ctx, operator)
    return nil
}

func (h *Handler) GetOperatorIDs(ctx context.Context) ([]string, error) {
    return []string{}, nil
}