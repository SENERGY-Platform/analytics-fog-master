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
