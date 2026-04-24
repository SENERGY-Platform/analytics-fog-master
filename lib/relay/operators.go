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

package relay

import (
	"encoding/json"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"

	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

func (relay *RelayController) processAgentStartOperatorResponse(message []byte) {
	logging.Logger.Debug("Received agent response to start operator command")
	startOperatorResponse := operatorEntities.StartOperatorAgentResponse{}
	err := json.Unmarshal(message, &startOperatorResponse)
	if err != nil {
		logging.Logger.Error("Cant Unmarshal agent response", "error", err)
	}
	relay.Master.HandleAgentStartOperatorResponse(startOperatorResponse)
}

func (relay *RelayController) processAgentStopOperatorResponse(message []byte) {
	logging.Logger.Debug("Received agent response to stop operator command")
	stopOperatorResponse := operatorEntities.StopOperatorAgentResponse{}
	err := json.Unmarshal(message, &stopOperatorResponse)
	if err != nil {
		logging.Logger.Error("Cant Unmarshal agent response", "error", err)
	}
	relay.Master.HandleAgentStopOperatorResponse(stopOperatorResponse)
}

func (relay *RelayController) processOperatorControlSync(message []byte) {
	logging.Logger.Debug("Received operator control sync message")
	syncMessage := []operatorEntities.StartOperatorControlCommand{}
	err := json.Unmarshal(message, &syncMessage)
	if err != nil {
		logging.Logger.Error("Cant unmarshal upstream sync message", "error", err)
	}
	relay.Master.SyncOperatorStates(syncMessage)
}
