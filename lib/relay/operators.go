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
