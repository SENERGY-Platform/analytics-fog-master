package relay

import (
	"encoding/json"
	"fmt"

	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

func (relay *RelayController) processOperatorsCommand(message []byte) {
	fmt.Println("Received agent response to operator command")
	startOperatorResponse := operatorEntities.OperatorAgentResponse{}
	err := json.Unmarshal(message, &startOperatorResponse)
	if err != nil {
		fmt.Println("error:", err)
	}
	relay.Master.HandleAgentOperatorResponse(startOperatorResponse)
}
