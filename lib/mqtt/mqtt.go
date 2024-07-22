package mqtt

import (
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
)

func NewMQTTClient(brokerConfig mqtt.BrokerConfig, logger *slog.Logger) *mqtt.MQTTClient {
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
		OnConnectHandler: OnConnectFog,
		SubscribeInitial: true,
	}
}

func OnConnectFog(client MQTT.Client) {
}