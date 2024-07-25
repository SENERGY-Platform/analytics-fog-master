package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	agentLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
)

var tLayout = time.RFC3339Nano
var NotFoundErr = errors.New("not found")

type Handler struct {
	db *sql.DB
	mu       sync.RWMutex
}

func New(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) createAgent(ctx context.Context, agent agentLib.Agent) error {
	logging.Logger.Debug("Create Agent", "new agent", agent)

    query := `INSERT INTO agents (id, updated, active) VALUES (?, ?, ?)`
    _, err := h.db.ExecContext(ctx, query, agent.Id, agent.Updated, agent.Active)
    if err != nil {
        return fmt.Errorf("createAgent: %v", err)
    }
    return nil
}

func (h *Handler) updateAgent(ctx context.Context, agent agentLib.Agent) error {
	logging.Logger.Debug("Update Agent", "new agent", agent)
    query := `UPDATE agents SET updated = ?, active = ? WHERE id = ?`
    _, err := h.db.ExecContext(ctx, query, agent.Updated, agent.Active, agent.Id)
    if err != nil {
        return fmt.Errorf("updateAgent: %v", err)
    }
    return nil
}

func (h *Handler) DeleteAgent(ctx context.Context, id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	query := `DELETE FROM agents WHERE id = ?`
    _, err := h.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("deleteAgent: %v", err)
    }
    return nil
}

func (h *Handler) createOperator(ctx context.Context, operator operatorEntities.Operator) error {
	logging.Logger.Debug("Create Operator", "new operator", operator)
    query := `INSERT INTO operators (pipeline_id, operator_id, state, container_id, error, agent_id) VALUES (?, ?, ?, ?, ?, ?)`
    _, err := h.db.ExecContext(ctx, query, operator.OperatorIDs.PipelineId, operator.OperatorIDs.OperatorId, operator.DeploymentState, operator.ContainerId, operator.DeploymentError, operator.AgentId)
    if err != nil {
        return fmt.Errorf("createOperator: %v", err)
    }
    return nil
}

func (h *Handler) updateOperator(ctx context.Context, operator operatorEntities.Operator) error {
	logging.Logger.Debug("Update Operator", "new operator", operator)
	timeStr := timeToString(operator.TimeOfLastHeartbeat)
    query := `UPDATE operators SET state = ?, container_id = ?, error = ?, time_of_last_heartbeat = ? WHERE pipeline_id = ? AND operator_id = ?`
    _, err := h.db.ExecContext(ctx, query, operator.DeploymentState, operator.ContainerId, operator.DeploymentError, timeStr, operator.OperatorIDs.PipelineId, operator.OperatorIDs.OperatorId)
    if err != nil {
        return fmt.Errorf("updateOperator: %v", err)
    }
    return nil
}

func (h *Handler) DeleteOperator(ctx context.Context, pipelineID string, operatorID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	query := `DELETE FROM operators WHERE pipeline_id = ? AND operator_id = ?`
    _, err := h.db.ExecContext(ctx, query, pipelineID, operatorID)
    if err != nil {
        return fmt.Errorf("deleteOperator: %v", err)
    }
    return nil
}

func (h *Handler) GetAllAgents(ctx context.Context) ([]agentLib.Agent, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	agents := []agentLib.Agent{}
	rows, err := h.db.QueryContext(ctx, "SELECT id, active, updated FROM agents")
	if err != nil {
		return agents, err
	}
	defer rows.Close()

	for rows.Next() {  
        agent := agentLib.Agent{} 
		if err := rows.Scan(&agent.Id, &agent.Active, &agent.Updated); err != nil {
			return agents, err
		}
        agents = append(agents, agent)
	}
	if err = rows.Err(); err != nil {
        return agents, fmt.Errorf("Agents could not be queried: %w", err)
    }
    return agents, nil
}

func (h *Handler) GetAgent(ctx context.Context, id string) (agentLib.Agent, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	row := h.db.QueryRowContext(ctx, "SELECT id, active, updated FROM agents WHERE id == ?", id)
	agent := agentLib.Agent{}
	if err := row.Scan(&agent.Id, &agent.Active, &agent.Updated); err != nil {
		return agentLib.Agent{}, err
	}

    return agent, nil
}

func (h *Handler) CreateOrUpdateAgent(ctx context.Context, agent agentLib.Agent) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	row := h.db.QueryRowContext(ctx, "SELECT id FROM agents WHERE id == ?", agent.Id)
	var agentId string
	err := row.Scan(&agentId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return h.createAgent(ctx, agent)
		} 
		return err
	}
    
	return h.updateAgent(ctx, agent)
}

func (h *Handler) GetOperator(ctx context.Context, pipelineID, operatorID string) (operatorEntities.Operator, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	operator := operatorEntities.Operator{} 
	var timeOfLastHeartbeat sql.NullString
	row := h.db.QueryRowContext(ctx, "SELECT pipeline_id, operator_id, state, container_id, error, agent_id, time_of_last_heartbeat FROM operators WHERE pipeline_id == ? AND operator_id == ?", pipelineID, operatorID)

	err := row.Scan(&operator.PipelineId, &operator.OperatorId, &operator.DeploymentState, &operator.ContainerId, &operator.DeploymentError, &operator.AgentId, &timeOfLastHeartbeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return operatorEntities.Operator{}, NotFoundErr
		}
		return operatorEntities.Operator{} , err
	}
	parsedHeartbeatTime, err := stringToTime(timeOfLastHeartbeat.String)
	if err != nil {
		return operatorEntities.Operator{}, err
	}
	operator.TimeOfLastHeartbeat = parsedHeartbeatTime
	return operator, nil
}

func (h *Handler) CreateOrUpdateOperator(ctx context.Context, operator operatorEntities.Operator) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	row := h.db.QueryRowContext(ctx, "SELECT operator_id FROM operators WHERE pipeline_id == ? AND operator_id == ?", operator.PipelineId, operator.OperatorId)
	var operator_id string
	err := row.Scan(&operator_id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return h.createOperator(ctx, operator)
		} 
		return err
	}	
	return h.updateOperator(ctx, operator)
}

func (h *Handler) GetOperators(ctx context.Context) ([]operatorEntities.Operator, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	currentOperators := []operatorEntities.Operator{}
	rows, err := h.db.QueryContext(ctx, "SELECT pipeline_id, operator_id, state, container_id, error, agent_id, time_of_last_heartbeat FROM operators")
	if err != nil {
		return currentOperators, err
	}
	defer rows.Close()

	for rows.Next() {  
        operator := operatorEntities.Operator{} 
		var timeOfLastHeartbeat sql.NullString
		if err := rows.Scan(&operator.PipelineId, &operator.OperatorId, &operator.DeploymentState, &operator.ContainerId, &operator.DeploymentError, &operator.AgentId, &timeOfLastHeartbeat); err != nil {
			return []operatorEntities.Operator{}, err
		}
		parsedHeartbeatTime, err := stringToTime(timeOfLastHeartbeat.String)
		if err != nil {
			return []operatorEntities.Operator{}, err
		}
		operator.TimeOfLastHeartbeat = parsedHeartbeatTime
        currentOperators = append(currentOperators, operator)
	}
	if err = rows.Err(); err != nil {
        return []operatorEntities.Operator{}, fmt.Errorf("Operators could not be queried: %w", err)
    }
    return currentOperators, nil
}

func timeToString(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(tLayout)
}

func stringToTime(s string) (time.Time, error) {
	if s != "" {
		return time.Parse(tLayout, s)
	}
	return time.Time{}, nil
}

/*
func (h *Handler) BeginTransaction(ctx context.Context, txItf driver.Tx) (*sql.Tx, error) {
	if txItf != nil {
		tx := txItf.(*sql.Tx)
		return tx, nil
	}

	tx, e := h.db.BeginTx(ctx, nil)
	if e != nil {
		return nil, e
	}
	return tx, nil
}

func (h *Handler) Commit(tx *sql.Tx, txItf driver.Tx) error {
	if txItf == nil {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}*/