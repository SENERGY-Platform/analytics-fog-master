package relay

import (
	"encoding/json"
	"fmt"

	operatorEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

func (relay *RelayController) processAgentStartOperatorResponse(message []byte) {
	fmt.Println("Received agent response to start operator command")
	startOperatorResponse := operatorEntities.StartOperatorAgentResponse{}
	err := json.Unmarshal(message, &startOperatorResponse)
	if err != nil {
		fmt.Println("Cant Unmarshal agent response: ", err)
	}
	relay.Master.HandleAgentStartOperatorResponse(startOperatorResponse)
}

func (relay *RelayController) processAgentStopOperatorResponse(message []byte) {
	fmt.Println("Received agent response to stop operator command")
	stopOperatorResponse := operatorEntities.StopOperatorAgentResponse{}
	err := json.Unmarshal(message, &stopOperatorResponse)
	if err != nil {
		fmt.Println("Cant Unmarshal agent response: ", err)
	}
	relay.Master.HandleAgentStopOperatorResponse(stopOperatorResponse)
}