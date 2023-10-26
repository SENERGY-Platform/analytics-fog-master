package relay

import (
	"encoding/json"
	"fmt"

	controlEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/control"
	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

func (relay *RelayController) processControlCommand(message []byte) {
	command := controlEntities.ControlMessage{}
	err := json.Unmarshal(message, &command)
	if err != nil {
		fmt.Println("error:", err)
	}

	if command.Command == "startOperator" {
		fmt.Println("Recevied start operator message")
		command := operatorEntities.StartOperatorControlCommand{}
		err := json.Unmarshal(message, &command)
		if err != nil {
			fmt.Println("error:", err)
		}
		relay.Master.StartOperator(command)
	}
	if command.Command == "stopOperator" {
		fmt.Println("Recevied stop operator message")
		command := operatorEntities.StopOperatorControlCommand{}
		err := json.Unmarshal(message, &command)
		if err != nil {
			fmt.Println("error:", err)
		}
		relay.Master.StopOperator(command)
	}
}
