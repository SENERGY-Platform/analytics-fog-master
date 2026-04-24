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
	"testing"
	"time"

	"github.com/SENERGY-Platform/analytics-fog-master/tests/utils"
	"github.com/hahahannes/e2e-go-utils/lib/streaming/mqtt"
	"github.com/stretchr/testify/assert"
)

// Master shall register itself at startup
func TestRegistration(t *testing.T) {
	ctx := context.Background()
	env, err := utils.NewEnv(ctx, t, 10, 10, 10, 10, "registration", true)
	if err != nil {
		t.Errorf("Cant start env: %s", err.Error())
		return
	}

	err = env.StartBroker(ctx, t)
	if err != nil {
		t.Errorf("Cant start broker: %s", err.Error())
		return
	}

	t.Log("Run registration test")

	registrationTopic := "analytics/master"

	ctx, cf := context.WithTimeout(ctx, 15*time.Second)
	defer cf()
	result, err := mqtt.WaitForMQTTMessageReceived(ctx, registrationTopic, ".*", func(context.Context) error {
		err = env.StartMaster(ctx, t)
		return err
	}, "localhost", env.BrokerPort, true)
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
