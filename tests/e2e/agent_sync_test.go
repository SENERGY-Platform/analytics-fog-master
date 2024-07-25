package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/SENERGY-Platform/analytics-fog-master/tests/utils"
	"github.com/hahahannes/e2e-go-utils/lib/streaming/mqtt"
	"github.com/stretchr/testify/assert"
)

// Agent registers and start command will be forwarded to it
// Master receives response and marks operator as started
// TODO agent response and then check for `started` state
func TestAgentSync(t *testing.T) {
	ctx := context.Background()
	env, err := utils.NewEnv(ctx, t, 30, 30, 30, 30, "normal_start", true)
	if err != nil {
		t.Errorf("Cant start env: %s", err.Error())
		return 
	}
	err = env.Start(ctx, t)
	if err != nil {
		t.Errorf("Cant start broker or master: %s", err.Error())
		return 
	}

	agentID := "agent1"
	err = utils.RegisterAgent(env, t, agentID)
	if err != nil {
		t.Errorf("Cant register agent: %s", err.Error())
		return
	}

	ctx, cf := context.WithTimeout(ctx, 60 * time.Second)
	defer cf()
	startOperatorTopic := fmt.Sprintf("analytics/agents/%s/control/start", agentID)
	result, err := mqtt.WaitForMQTTMessageReceived(ctx, startOperatorTopic, ".*", func(context.Context) error {
		operatorID := "op1"
		pipelineID := "pipe1"
		return utils.StartOperatorAtMaster(env, t, operatorID, pipelineID)
	}, "localhost", env.BrokerPort, false)
	if err != nil {
		t.Error(err)
		return 
	}
	if result.Error != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, result.Received, true)
}



// Agent registers -> Timeout for pong -> mark agent inactive 
// Start command will be ignored by Master
func TestInactiveAgent(t *testing.T) {
	ctx := context.Background()
	var agentPingInterval float64 = 5
	var timeoutInactiveAgent float64 = 5

	env, err := utils.NewEnv(ctx, t, agentPingInterval, 100, timeoutInactiveAgent, 100, "inactive_agent", true)
	if err != nil {
		t.Errorf("Cant start env: %s", err.Error())
		return 
	}
	err = env.Start(ctx, t)
	if err != nil {
		t.Errorf("Cant start broker or master: %s", err.Error())
		return 
	}

	agentID := "agent1"
	err = utils.RegisterAgent(env, t, agentID)
	if err != nil {
		t.Errorf("Cant register agent: %s", err.Error())
		return
	}

	ctx, cf := context.WithTimeout(ctx, 60 * time.Second)
	defer cf()
	startOperatorTopic := fmt.Sprintf("analytics/agents/%s/control/start", agentID)
	result, err := mqtt.WaitForMQTTMessageReceived(ctx, startOperatorTopic, ".*", func(context.Context) error {
		time.Sleep(time.Duration((agentPingInterval * 3 + 10) * float64(time.Second)))
		operatorID := "op1"
		pipelineID := "pipe1"
		return utils.StartOperatorAtMaster(env, t, operatorID, pipelineID)
	}, "localhost", env.BrokerPort, false)
	if err != nil {
		t.Error(err)
		return 
	}
	if result.Error != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, result.Received, false, "Master shall ignore the start command")
}

// Agent registers -> Timeout for stale operator 
// Start operator -> (should be deleted) -> start command with next sync
// Stop operator -> (should be marked as started)  -> stop command with next sync
func TestStaleOperators(t *testing.T) {
	ctx := context.Background()
	var staleOperatorTimeout float64 = 5
	var staleOperatorInterval float64 = 5 

	env, err := utils.NewEnv(ctx, t, 100, staleOperatorInterval, 100, staleOperatorTimeout, "stale", true)
	if err != nil {
		t.Errorf("Cant start env: %s", err.Error())
		return 
	}
	err = env.Start(ctx, t)
	if err != nil {
		t.Errorf("Cant start broker or master: %s", err.Error())
		return 
	}

	agentID := "agent1"
	err = utils.RegisterAgent(env, t, agentID)
	if err != nil {
		t.Errorf("Cant register agent: %s", err.Error())
		return
	}

	startOperatorTopic := fmt.Sprintf("analytics/agents/%s/control/start", agentID)

	operatorID := "op1"
	pipelineID := "pipe1"
	// We have to wait for the first start command message to arrive in mqtt topic so that 
	// the message wont slip in when I test for the second start command below
	result, err := mqtt.WaitForMQTTMessageReceived(ctx, startOperatorTopic, ".*", func(context.Context) error {
		t.Log("Start operator")
		err = utils.StartOperatorAtMaster(env, t, operatorID, pipelineID)
		if err != nil {
			return fmt.Errorf("Cant start operator: %s", err.Error())
		}
		return nil
	}, "localhost", env.BrokerPort, false)
	if err != nil {
		t.Error(err)
		return 
	}
	if result.Error != nil {
		t.Error(err)
		return
	}
	
	ctx, cf := context.WithTimeout(ctx, 60 * time.Second)
	defer cf()
	result, err = mqtt.WaitForMQTTMessageReceived(ctx, startOperatorTopic, ".*", func(context.Context) error {
		wait := time.Duration((staleOperatorTimeout * 2 + 10) * float64(time.Second))
		t.Log(fmt.Sprintf("Wait %s so that operator state is too old", wait))
		time.Sleep(wait)
		t.Log("Simulate incoming operator sync to trigger new start command")
		err = utils.SendOperatorSync(env, t, operatorID, pipelineID)
		return err
	}, "localhost", env.BrokerPort, false)
	if err != nil {
		t.Error(err)
		return 
	}
	if result.Error != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, result.Received, true)
}

// TODO Agent registers
// Start operator -> no immediate response
// ping agent -> pong with new operator state
// master updates state
