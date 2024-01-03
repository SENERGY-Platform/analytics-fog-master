package relay

import (
	"encoding/json"
	"fmt"

	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

func (relay *RelayController) processStartOperatorCommand(message []byte) {
	fmt.Println("Recevied start operator message")
	command := operatorEntities.StartOperatorControlCommand{}
	err := json.Unmarshal(message, &command)
	if err != nil {
		fmt.Println("Error at unmarshalling start operator message:", err)
	}
	relay.Master.StartOperator(command)
}

func (relay *RelayController) processStopOperatorCommand(message []byte) {
	fmt.Println("Recevied stop operator message")
	command := operatorEntities.StopOperatorControlCommand{}
	err := json.Unmarshal(message, &command)
	if err != nil {
		fmt.Println("Error at unmarshalling stop operator message:", err)
	}
	relay.Master.StopOperator(command)
}
