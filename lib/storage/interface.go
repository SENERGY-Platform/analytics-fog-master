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
