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

package mqtt

import (
	"log/slog"

	"github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func NewMQTTClient(brokerConfig mqtt.BrokerConfig, logger *slog.Logger) *mqtt.MQTTClient {
	topics := map[string]byte{
		operator.StartOperatorFogTopic:               byte(2),
		operator.StopOperatorFogTopic:                byte(2),
		agent.AgentsTopic:                            byte(2),
		operator.StartOperatorResponseFogTopic:       byte(2),
		operator.StopOperatorResponseFogTopic:        byte(2),
		operator.OperatorControlSyncResponseFogTopic: byte(2),
	}

	return &mqtt.MQTTClient{
		Broker:           brokerConfig,
		TopicConfig:      topics,
		Logger:           logger,
		OnConnectHandler: OnConnectFog,
		SubscribeInitial: true,
	}
}

func OnConnectFog(client MQTT.Client) {
}
