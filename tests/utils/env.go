package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	mqttLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/config"

	testLib "github.com/hahahannes/e2e-go-utils/lib"
	"github.com/hahahannes/e2e-go-utils/lib/streaming/mqtt"
)

type ApplicationLogger struct {
	readyLogChannel chan string
	logChannel chan string
}

func (l *ApplicationLogger) Write(p []byte) (n int, err error) {
	msg := string(p)
	select {
		// use select default to not block logging when readylogchannel has no receiver anymore after the check
		case l.readyLogChannel <- msg:
		default:
	}

	select {
		case l.logChannel <- msg:
		default:
	}
	
	return 1, nil
}

func (e *Env) StartAndWait(ctx context.Context, t *testing.T) error {
	readyLogChannel := make(chan string, 1000)
	logChannel := make(chan string, 1000)
	logger := &ApplicationLogger{
		readyLogChannel: readyLogChannel,
		logChannel: logChannel,
	}

	go func() {
		f, err := os.OpenFile(fmt.Sprintf("%s.log", e.TestName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Print("Could not open log file")
		}
		
		for {
			select {
			case log := <- logChannel:
				if _, err := f.Write([]byte(log + "\n")); err != nil {
					fmt.Print("Could not write to log file")
				}
			case <- ctx.Done():
				if err := f.Close(); err != nil {
					fmt.Print("Could not close file")
				}
				return
			}
		}
		
	}()

	runConfig := config.Config{
		Broker: mqttLib.FogBrokerConfig{
			Port: e.BrokerPort,
			Host: "localhost",
		},
		DataDir: "./data",
		AgentSyncIntervalSeconds: e.AgentSyncIntervalSeconds,
		StaleOperatorCheckIntervalSeconds: e.StaleOperatorCheckIntervalSeconds,
		TimeoutInactiveAgentSeconds: e.TimeoutInactiveAgentSeconds,
		TimeoutStaleOperatorSeconds: e.TimeoutStaleOperatorSeconds,
		DataBase: config.DataBaseConfig{
			Path: e.DataBaseURL,
		},
	}
	
	received, err := testLib.WaitForStringReceived(".*Master is ready.*", func (sendCtx context.Context) error {
		return lib.Run(sendCtx, logger, logger, runConfig)
	}, readyLogChannel, 30 * time.Second, false)

	if err != nil {
		return err 
	}

	if received.Received == false {
		return errors.New("Master ready log not received")
	}
	t.Log("Master is ready log received!")

	return nil
}

type Env struct {
	MqttClient *mqtt.MQTTClient
	broker *Mosquitto
	UserID string
	BrokerPort string
	AgentSyncIntervalSeconds float64
	StaleOperatorCheckIntervalSeconds float64
	TimeoutInactiveAgentSeconds float64
	TimeoutStaleOperatorSeconds float64
	TestName string
	DataBaseURL string
	CleanUpAfterTest bool
}

func NewEnv(ctx context.Context, t *testing.T, agentSyncIntervalSeconds, staleOperatorCheckIntervalSeconds, timeoutInactiveAgentSeconds, timeoutStaleOperatorSeconds float64, testName string, cleanUpAfterTest bool) (*Env, error) {
	broker, err := NewMosquitto(ctx)
	if err != nil {
		return &Env{}, err
	}

	env := Env{
		broker: broker,
		UserID: "user",
		AgentSyncIntervalSeconds: agentSyncIntervalSeconds,
		StaleOperatorCheckIntervalSeconds: staleOperatorCheckIntervalSeconds,
		TimeoutInactiveAgentSeconds: timeoutInactiveAgentSeconds,
		TimeoutStaleOperatorSeconds: timeoutStaleOperatorSeconds,
		TestName: testName,
		DataBaseURL: "./db.sqlite3",
		CleanUpAfterTest: cleanUpAfterTest,
	}

	if cleanUpAfterTest {
		t.Cleanup(func() {
			e := os.Remove(env.DataBaseURL) 
			if e != nil { 
				t.Error(fmt.Errorf("Cant delete database: %w", e))
			} 
		})
	}

	return &env, nil
}

func (e *Env) StartBroker(ctx context.Context, t *testing.T) (error) {
	t.Log("Start Mosquitto")
	err, brokerPort := e.broker.StartAndWait(ctx)
	if err != nil {
		return err
	}
	e.BrokerPort = brokerPort
	t.Log("Started Mosquitto")

	return nil
}

func (e *Env) StartMaster(ctx context.Context, t *testing.T) (error) {
	t.Log("Start Master")
	err := e.StartAndWait(ctx, t)
	if err != nil {
		return err
	}
	t.Log("Started Master")
	
	return nil
}

func (e *Env) Start(ctx context.Context, t *testing.T) (error) {
	err := e.StartBroker(ctx, t)
	if err != nil {
		return err
	}
	return e.StartMaster(ctx, t)
}

func (e *Env) PublishToBroker(topic string, payload []byte, t *testing.T) error {
	e.MqttClient = mqtt.NewMQTTClient("localhost", e.BrokerPort, nil, nil, false)
	t.Log("Connect to local broker at " + e.BrokerPort)
	err := e.MqttClient.ConnectMQTTBroker(nil, nil)
	if err != nil {
		return err
	}
	t.Log("Publish to: " + topic)
	return e.MqttClient.Publish(topic, string(payload), 2)

}