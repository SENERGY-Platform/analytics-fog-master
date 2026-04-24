/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
func TestSimple(t *testing.T) {
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

	ctx, cf := context.WithTimeout(ctx, 60*time.Second)
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

	ctx, cf := context.WithTimeout(ctx, 60*time.Second)
	defer cf()
	startOperatorTopic := fmt.Sprintf("analytics/agents/%s/control/start", agentID)
	result, err := mqtt.WaitForMQTTMessageReceived(ctx, startOperatorTopic, ".*", func(context.Context) error {
		time.Sleep(time.Duration((agentPingInterval*3 + 10) * float64(time.Second)))
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
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var staleOperatorTimeout float64 = 5
	var staleOperatorInterval float64 = 5

	env, err := utils.NewEnv(ctx, t, 100, staleOperatorInterval, 100, staleOperatorTimeout, "stale", true)
	if err != nil {
		t.Fatalf("cant start env: %s", err)
	}
	if err = env.Start(ctx, t); err != nil {
		t.Fatalf("cant start broker or master: %s", err)
	}

	agentID := "agent1"
	if err = utils.RegisterAgent(env, t, agentID); err != nil {
		t.Fatalf("cant register agent: %s", err)
	}

	startOperatorTopic := fmt.Sprintf("analytics/agents/%s/control/start", agentID)
	operatorID := "op1"
	pipelineID := "pipe1"

	result, err := mqtt.WaitForMQTTMessageReceived(ctx, startOperatorTopic, ".*", func(context.Context) error {
		t.Log("starting operator")
		return utils.StartOperatorAtMaster(env, t, operatorID, pipelineID)
	}, "localhost", env.BrokerPort, false)
	if err != nil {
		t.Fatalf("waiting for first start command: %s", err)
	}
	if result.Error != nil {
		t.Fatalf("first start command: %s", result.Error)
	}

	wait := time.Duration((staleOperatorTimeout*2 + 10) * float64(time.Second))
	t.Logf("waiting %s for operator state to go stale", wait)
	time.Sleep(wait)

	result, err = mqtt.WaitForMQTTMessageReceived(ctx, startOperatorTopic, ".*", func(context.Context) error {
		t.Log("sending operator sync to trigger re-start")
		return utils.SendOperatorSync(env, t, operatorID, pipelineID)
	}, "localhost", env.BrokerPort, false)
	if err != nil {
		t.Fatalf("waiting for second start command: %s", err)
	}
	if result.Error != nil {
		t.Fatalf("second start command: %s", result.Error)
	}

	assert.True(t, result.Received)
}

// TODO Agent registers
// Start operator -> no immediate response
// ping agent -> pong with new operator state
// master updates state
