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

func (relay *RelayController) processStartOperatorCommand(message []byte) {
	logging.Logger.Debug("Recevied start operator message")
	command := operatorEntities.StartOperatorControlCommand{}
	err := json.Unmarshal(message, &command)
	if err != nil {
		logging.Logger.Error("Error at unmarshalling start operator message", "error", err)
	}
	relay.Master.StartOperator(command)
}

func (relay *RelayController) processStopOperatorCommand(message []byte) {
	logging.Logger.Debug("Recevied stop operator message")
	command := operatorEntities.StopOperatorControlCommand{}
	err := json.Unmarshal(message, &command)
	if err != nil {
		logging.Logger.Error("Error at unmarshalling stop operator message", "error", err)
	}
	relay.Master.StopOperator(command)
}
