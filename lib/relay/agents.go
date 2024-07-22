package relay

import (
	"encoding/json"
	"fmt"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
)

func (relay *RelayController) processAgentsCommand(message []byte) {
	agentMessage := agentEntities.AgentInfoMessage{}
	err := json.Unmarshal(message, &agentMessage)
	if err != nil {
		fmt.Println("error:", err)
	}
	switch agentMessage.Command {
	case "register":
		fmt.Println("Received agent register message")
		err = relay.Master.RegisterAgent(agentMessage.Conf)
	case "pong":
		fmt.Println("Received Pong")
		err = relay.Master.PongAgent(agentMessage)
	}
}
