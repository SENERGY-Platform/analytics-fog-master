package mqtt

import (
	operator "command-line-arguments/home/hannes/Desktop/projects/analytics-fog/analytics-fog-lib/lib/operator/topics.go"

	"github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	log_level "github.com/y-du/go-log-level"
)

func NewMQTTClient(brokerConfig mqtt.BrokerConfig, logger *log_level.Logger) *mqtt.MQTTClient {
	topics := map[string]byte{
		operator.StartOperatorFogTopic:   byte(2),
		operator.StopOperatorFogTopic:   byte(2),
		agent.AgentsTopic:    byte(2),
		operator.StartOperatorResponseFogTopic: byte(2),
		operator.StopOperatorResponseFogTopic: byte(2),
		operator.OperatorControlSyncResponseFogTopic: byte(2),
	}

	return &mqtt.MQTTClient{
		Broker:      brokerConfig,
		TopicConfig: topics,
		Logger:      logger,
	}
}
