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
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	mqttLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/config"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/controller"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/master"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/relay"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/storage"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/SENERGY-Platform/go-service-base/watchdog"
	"github.com/joho/godotenv"
)

func main() {
	ec := 0
	defer func() {
		os.Exit(ec)
	}()

	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: ", err)
		ec = 1
		return
	}

	config, err := config.NewConfig("")
	if err != nil {
		log.Printf("Cant load config: ", err)
		ec = 1
		return
	}
	
	err = logging.InitLogger(os.Stdout, true)
	if err != nil {
		log.Printf("Error init logging: %s", err.Error())
		ec = 1
		return
	}

	logging.Logger.Info(fmt.Sprintf("config: %s", sb_util.ToJsonStr(config)))

	logging.Logger.Info("Create new database at " + config.DataBase.ConnectionURL)
	db, err := storage.NewDB(config.DataBase.ConnectionURL)
	if err != nil {
		logging.Logger.Error("Cant init DB", "error", err.Error())
		ec = 1
		return
	}
	defer db.Close()

	storageHandler := storage.New(db)

	watchdog := watchdog.New(syscall.SIGINT, syscall.SIGTERM)

	fogMQTTConfig := mqttLib.BrokerConfig(config.Broker)

	mqttClient := mqtt.NewMQTTClient(fogMQTTConfig, logging.Logger)

	ctx, cancel := context.WithCancel(context.Background())
	operatorController := controller.NewController(ctx, mqttClient, storageHandler)
	go operatorController.Start()
	watchdog.RegisterStopFunc(func() error {
		cancel()
		return nil
	})

	master := master.NewMaster(mqttClient, storageHandler, operatorController, time.Duration(config.AgentSyncIntervalSeconds * float64(time.Second)), time.Duration(config.StaleOperatorCheckIntervalSeconds * float64(time.Second)), config.TimeoutInactiveAgentSeconds, config.TimeoutStaleOperatorSeconds)
	relayController := relay.NewRelayController(master)
	mqttClient.SetSubscriptionHandler(relayController)
	
	logging.Logger.Info("Connect MQTT")
	mqttClient.ConnectMQTTBroker(nil, nil)

	logging.Logger.Info("Register master")
	master.Register()

	logging.Logger.Info("Start agent ping in background")
	go master.CheckAgents()

	logging.Logger.Info("Start periodic check for stale operators")
	staleOperatorCtx, staleOperatorCancel := context.WithCancel(context.Background())
	go master.MarkStaleOperators(staleOperatorCtx)

	watchdog.RegisterStopFunc(func() error {
		staleOperatorCancel()
		return nil
	})

	watchdog.RegisterStopFunc(func() error {
		mqttClient.CloseConnection()
		return nil
	})

	logging.Logger.Info("Master is ready")
	watchdog.Start()

	ec = watchdog.Join()
}
