package relay

import (
	"encoding/json"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"

	agentEntities "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
)

func (relay *RelayController) processAgentsCommand(message []byte) {
	agentMessage := agentEntities.AgentInfoMessage{}
	err := json.Unmarshal(message, &agentMessage)
	if err != nil {
		logging.Logger.Error("Cant unmarshal agent command", "error", err)
	}
	switch agentMessage.Command {
	case "register":
		logging.Logger.Debug("Received agent register message")
		err = relay.Master.RegisterAgent(agentMessage.Conf)
	case "pong":
		logging.Logger.Debug("Received Pong")
		err = relay.Master.PongAgent(agentMessage)
	}
}
