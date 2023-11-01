package mqtt

import (
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/config"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/constants"
)

func NewMQTTClient(brokerConfig config.BrokerConfig) *mqtt.MQTTClient {
	topics := map[string]byte{
		constants.ControlTopic:   byte(2),
		constants.AgentsTopic:    byte(2),
		constants.OperatorsTopic: byte(2),
	}

	return &mqtt.MQTTClient{
		Broker:      brokerConfig,
		TopicConfig: topics,
	}
}
