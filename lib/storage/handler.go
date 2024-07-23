package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
	agentLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

var tLayout = time.RFC3339Nano

type Handler struct {
	db *sql.DB
}

func New(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) createAgent(ctx context.Context, agent agentLib.Agent, txItf driver.Tx) error {
	logging.Logger.Debug("Create Agent", "new agent", agent)
    tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return err
	}
    query := `INSERT INTO agents (id, updated, active) VALUES (?, ?, ?)`
    _, err = tx.ExecContext(ctx, query, agent.Id, agent.Updated, agent.Active)
    if err != nil {
        return fmt.Errorf("createAgent: %v", err)
    }
    return h.Commit(tx, txItf)
}

func (h *Handler) updateAgent(ctx context.Context, agent agentLib.Agent, txItf driver.Tx) error {
	logging.Logger.Debug("Update Agent", "new agent", agent)
    tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return err
	}
    query := `UPDATE agents SET updated = ?, active = ? WHERE id = ?`
    _, err = tx.ExecContext(ctx, query, agent.Updated, agent.Active, agent.Id)
    if err != nil {
        return fmt.Errorf("updateAgent: %v", err)
    }
    return h.Commit(tx, txItf)
}

func (h *Handler) DeleteAgent(ctx context.Context, id string, txItf driver.Tx) error {
    tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return err
	}
	query := `DELETE FROM agents WHERE id = ?`
    _, err = tx.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("deleteAgent: %v", err)
    }
    return h.Commit(tx, txItf)
}

func (h *Handler) createOperator(ctx context.Context, operator operatorEntities.Operator, txItf driver.Tx) error {
    tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return err
	}
    query := `INSERT INTO operators (pipeline_id, operator_id, state, container_id, error, agent_id) VALUES (?, ?, ?, ?, ?, ?)`
    _, err = tx.ExecContext(ctx, query, operator.OperatorIDs.PipelineId, operator.OperatorIDs.OperatorId, operator.DeploymentState, operator.ContainerId, operator.DeploymentError, operator.AgentId)
    if err != nil {
        return fmt.Errorf("createOperator: %v", err)
    }
    return h.Commit(tx, txItf)
}

func (h *Handler) updateOperator(ctx context.Context, operator operatorEntities.Operator, txItf driver.Tx) error {
    tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return err
	}
    query := `UPDATE operators SET state = ?, container_id = ?, error = ? WHERE pipeline_id = ? AND operator_id = ?`
    _, err = tx.ExecContext(ctx, query, operator.DeploymentState, operator.ContainerId, operator.DeploymentError, operator.OperatorIDs.PipelineId, operator.OperatorIDs.OperatorId)
    if err != nil {
        return fmt.Errorf("updateOperator: %v", err)
    }
    return h.Commit(tx, txItf)
}

func (h *Handler) DeleteOperator(ctx context.Context, pipelineID string, operatorID string, txItf driver.Tx) error {
    tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return err
	}
	query := `DELETE FROM operators WHERE pipeline_id = ? AND operator_id = ?`
    _, err = tx.ExecContext(ctx, query, pipelineID, operatorID)
    if err != nil {
        return fmt.Errorf("deleteOperator: %v", err)
    }
    return h.Commit(tx, txItf)
}

func (h *Handler) GetAllAgents(ctx context.Context, txItf driver.Tx) ([]agentLib.Agent, error) {
    agents := []agentLib.Agent{}
    tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return agents, err
	}

	rows, err := tx.QueryContext(ctx, "SELECT id, active, updated FROM agents")
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
	err = h.Commit(tx, txItf)
	if err != nil {
		return agents, err
	}
    return agents, nil
}

func (h *Handler) GetAgent(ctx context.Context, id string, txItf driver.Tx) (agentLib.Agent, error) {
	agent := agentLib.Agent{}
    return agent, nil
}

func (h *Handler) CreateOrUpdateAgent(ctx context.Context, agent agentLib.Agent, txItf driver.Tx) error {
    logging.Logger.Debug("Create or Update Agent", "agent ID", agent.Id)
	tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, "SELECT COUNT(*) FROM agents WHERE id == ?", agent.Id)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int

	for rows.Next() {   
		if err := rows.Scan(&count); err != nil {
			return err
		}
	}
	if err = rows.Err(); err != nil {
        return fmt.Errorf("Agent Count could not be queried: %w", err)
    }

	err = rows.Close()
	if err != nil {
		return err
	}

	if count == 0 {
        h.createAgent(ctx, agent, tx)
    } else {
        h.updateAgent(ctx, agent, tx)
    }
    
    return h.Commit(tx, txItf)
}

func (h *Handler) GetOperator(ctx context.Context, pipelineID, operatorID string, txItf driver.Tx) (operatorEntities.Operator, error) {
	operator := operatorEntities.Operator{}
	tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return operator, err
	}
	rows, err := tx.QueryContext(ctx, "SELECT pipeline_id, operator_id, state, container_id, error, agent_id FROM operators WHERE pipeline_id == ? AND operator_id == ?", pipelineID, operatorID)
	if err != nil {
		return operator, err
	}
	defer rows.Close()

	operatorExists := false
	for rows.Next() {  
		operatorExists = true
        operator := operatorEntities.Operator{} 
		if err := rows.Scan(&operator.PipelineId, &operator.OperatorId, &operator.DeploymentState, &operator.ContainerId, &operator.DeploymentError, &operator.AgentId); err != nil {
			return operator, err
		}
	}
	if err = rows.Err(); err != nil {
        return operator, fmt.Errorf("Operator could not be queried: %w", err)
    }
	if !operatorExists {
		return operator, fmt.Errorf("Operator not found")
	}
	err = h.Commit(tx, txItf)
	if err != nil {
		return operator, err
	}
    return operator, nil
}

func (h *Handler) CreateOrUpdateOperator(ctx context.Context, operator operatorEntities.Operator, txItf driver.Tx) error {
    tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, "SELECT COUNT(*) FROM operators WHERE pipeline_id == ? AND operator_id == ?", operator.PipelineId, operator.OperatorId)
	if err != nil {
		return err
	}
	defer rows.Close()

	var count int

	for rows.Next() {   
		if err := rows.Scan(&count); err != nil {
			return err
		}
	}
	if err = rows.Err(); err != nil {
        return fmt.Errorf("Operators count could not be queried: %w", err)
    }
	err = rows.Close()
	if err != nil {
		return err
	}
	
	if count == 0 {
        h.createOperator(ctx, operator, tx)
    } else {
        h.updateOperator(ctx, operator, tx)
    }
    
    return h.Commit(tx, txItf)
}

func (h *Handler) GetOperators(ctx context.Context, txItf driver.Tx) ([]operatorEntities.Operator, error) {
    currentOperators := []operatorEntities.Operator{}
    tx, err := h.BeginTransaction(ctx, txItf)
	if err != nil {
		return currentOperators, err
	}

	rows, err := tx.QueryContext(ctx, "SELECT pipeline_id, operator_id, state, container_id, error, agent_id FROM operators")
	if err != nil {
		return currentOperators, err
	}
	defer rows.Close()

	for rows.Next() {  
        operator := operatorEntities.Operator{} 
		if err := rows.Scan(&operator.PipelineId, &operator.OperatorId, &operator.DeploymentState, &operator.ContainerId, &operator.DeploymentError, &operator.AgentId); err != nil {
			return currentOperators, err
		}
        currentOperators = append(currentOperators, operator)
	}
	if err = rows.Err(); err != nil {
        return currentOperators, fmt.Errorf("Operators could not be queried: %w", err)
    }
	err = h.Commit(tx, txItf)
	if err != nil {
		return currentOperators, err
	}
    return currentOperators, nil
}

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
}