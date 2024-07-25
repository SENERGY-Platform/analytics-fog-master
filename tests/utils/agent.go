package utils
import (
	"encoding/json"
	"testing"

	agentLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/control"
)

func RegisterAgent(env *Env, t *testing.T, agentID string) error {
	register := agentLib.AgentInfoMessage{
		ControlMessage: control.ControlMessage{
			Command: "register",
		},
		CurrentOperatorStates: []agentLib.OperatorState{},
		Conf: agentLib.Configuration{
			Id: agentID,
		},
	}
	msg, err := json.Marshal(register)
	if err != nil {
		return err
	}
	err = env.PublishToBroker("analytics/agents", msg, t)
	return err
}

