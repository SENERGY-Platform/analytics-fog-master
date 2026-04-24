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

package utils

import (
	"encoding/json"
	"testing"

	operatorLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

func StartOperatorAtMaster(env *Env, t *testing.T, operatorID, pipelineID string) error {
	startCommand := operatorLib.StartOperatorControlCommand{
		OperatorIDs: operatorLib.OperatorIDs{
			OperatorId: operatorID,
			PipelineId: pipelineID,
		},
	}
	msg, err := json.Marshal(startCommand)
	if err != nil {
		return err
	}
	err = env.PublishToBroker("analytics/operator/control/start", msg, t)
	return err
}

func SendOperatorSync(env *Env, t *testing.T, operatorID, pipelineID string) error {
	startCommand := []operatorLib.StartOperatorControlCommand{
		{
			OperatorIDs: operatorLib.OperatorIDs{
				OperatorId: operatorID,
				PipelineId: pipelineID,
			},
		},
	}
	msg, err := json.Marshal(startCommand)
	if err != nil {
		return err
	}
	err = env.PublishToBroker("analytics/operator/control/sync/response", msg, t)
	return err
}
