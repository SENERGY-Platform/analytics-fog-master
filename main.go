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
	"errors"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/SENERGY-Platform/analytics-fog-master/lib/config"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/db"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/master"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/relay"
	mqttLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	srv_base "github.com/SENERGY-Platform/go-service-base/srv-base"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/joho/godotenv"
)

func main() {
	ec := 0
	defer func() {
		os.Exit(ec)
	}()

	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file")
	}

	config, err := config.NewConfig("")
	if err != nil {
		fmt.Println(err)
	}

	logFile, err := logging.InitLogger(config.Logger)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		var logFileError *srv_base.LogFileError
		if errors.As(err, &logFileError) {
			ec = 1
			return
		}
	}
	if logFile != nil {
		defer logFile.Close()
	}
	logging.Logger.Debugf("config: %s", sb_util.ToJsonStr(config))

	logging.Logger.Debug("Init DB")
	database, err := db.NewFileDatabase(config.DataDir)
	if err != nil {
		logging.Logger.Error(err)
	}

	watchdog := srv_base.NewWatchdog(logging.Logger, syscall.SIGINT, syscall.SIGTERM)

	fogMQTTConfig := mqttLib.BrokerConfig(config.Broker)

	mqttClient := mqtt.NewMQTTClient(fogMQTTConfig, logging.Logger)
	master := master.NewMaster(mqttClient, database, config.StartOperatorConfig)
	relayController := relay.NewRelayController(master)
	mqttClient.SetRelayController(relayController)
	
	logging.Logger.Debug("Connect MQTT")
	mqttClient.ConnectMQTTBroker(nil, nil)

	logging.Logger.Debug("Register master")
	master.Register()

	logging.Logger.Debug("Start agent ping in background")
	go master.CheckAgents()

	watchdog.RegisterStopFunc(func() error {
		mqttClient.CloseConnection()
		return nil
	})

	watchdog.Start()

	ec = watchdog.Join()
}
