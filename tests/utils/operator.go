package utils

import (
	"encoding/json"
	"testing"

	operatorLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

func StartOperatorAtMaster(env *Env, t *testing.T, operatorID, pipelineID string) error {
	startCommand := operatorLib.StartOperatorControlCommand{
		OperatorIDs: operatorLib.OperatorIDs{
			OperatorId: operatorID,
			PipelineId: pipelineID,
		},
	}
	msg, err := json.Marshal(startCommand)
	if err != nil {
		return err
	}
	err = env.PublishToBroker("analytics/operator/control/start", msg, t)
	return err
}

func SendOperatorSync(env *Env, t *testing.T, operatorID, pipelineID string) error {
	startCommand := []operatorLib.StartOperatorControlCommand{
		{
		OperatorIDs: operatorLib.OperatorIDs{
			OperatorId: operatorID,
			PipelineId: pipelineID,
		},
		},
	}
	msg, err := json.Marshal(startCommand)
	if err != nil {
		return err
	}
	err = env.PublishToBroker("analytics/operator/control/sync/response", msg, t)
	return err
}