package e2e

import (
	"context"
	"testing"
	"time"
	"github.com/hahahannes/e2e-go-utils/lib/streaming/mqtt"
	"github.com/stretchr/testify/assert"
	"github.com/SENERGY-Platform/analytics-fog-master/tests/utils"
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

	ctx, cf := context.WithTimeout(ctx, 15 * time.Second)
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
