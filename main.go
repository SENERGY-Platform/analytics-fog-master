/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/SENERGY-Platform/analytics-fog-master/lib/config"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/db"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/master"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/relay"

	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config, err := config.NewConfig("")
	if err != nil {
		fmt.Println(err)
	}

	database, err := db.NewFileDatabase()
	if err != nil {
		fmt.Println(err)
	}

	mqttClient := mqtt.NewMQTTClient(config.Broker)
	master := master.NewMaster(mqttClient, database)
	relayController := relay.NewRelayController(master)

	mqttClient.ConnectMQTTBroker(relayController)

	go master.CheckAgents()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	defer mqttClient.CloseConnection()
	<-c
}
