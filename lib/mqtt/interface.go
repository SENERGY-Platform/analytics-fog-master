package mqtt

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// needed because of import cycle between master-mqtt-relay-master
type RelayController interface {
	ProcessMessage(message MQTT.Message)
	OnMessageReceived(client MQTT.Client, message MQTT.Message)
}
