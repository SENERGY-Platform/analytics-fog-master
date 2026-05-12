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

package lib

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqttLib "github.com/SENERGY-Platform/analytics-fog-lib/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/internal/config"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/controller"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/logging"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/master"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/mqtt"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/relay"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/storage"
	"github.com/SENERGY-Platform/analytics-fog-master/migrations"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
)

func Run(
	ctx context.Context,
	stdout,
	stderr io.Writer,
	cfg config.Config,
) error {
	if err := logging.InitLogger(stdout, true); err != nil {
		log.Printf("error initialising logger: %s", err)
		return err
	}

	logging.Logger.Info(fmt.Sprintf("config: %s", sb_util.ToJsonStr(cfg)))

	db, err := initDB(cfg.DataBase.Path)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {

		}
	}(db)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	storageHandler := storage.New(db)
	mqttClient := initMQTT(cfg)
	operatorController := controller.NewController(ctx, mqttClient, storageHandler)
	go operatorController.Start()

	m := initMaster(cfg, mqttClient, storageHandler, operatorController)
	mqttClient.SetSubscriptionHandler(relay.NewRelayController(m))

	logging.Logger.Info("connecting to MQTT broker")
	err = mqttClient.ConnectMQTTBroker(nil, nil)
	if err != nil {
		return err
	}

	logging.Logger.Info("registering master")
	m.Register()

	logging.Logger.Info("starting agent ping")
	go func() {
		err = m.CheckAgents()
		if err != nil {

		}
	}()

	logging.Logger.Info("starting stale operator check")
	go func() {
		err = m.MarkStaleOperators(ctx)
		if err != nil {

		}
	}()

	logging.Logger.Info("master is ready")

	waitForShutdown(ctx, cancel, mqttClient)

	logging.Logger.Info("shutdown complete")
	return nil
}

func waitForShutdown(ctx context.Context, cancel context.CancelFunc, mqttClient *mqttLib.MQTTClient) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case <-quit:
		logging.Logger.Info("received shutdown signal")
	case <-ctx.Done():
		logging.Logger.Info("context cancelled")
	}

	cancel()
	mqttClient.CloseConnection()
}

func initDB(path string) (*sql.DB, error) {
	logging.Logger.Info("initialising database", "path", path)
	db, err := storage.NewDB(path)
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}
	if err := migrations.MigrateDb(path); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}
	return db, nil
}

func initMQTT(cfg config.Config) *mqttLib.MQTTClient {
	return mqtt.NewMQTTClient(mqttLib.BrokerConfig(cfg.Broker), logging.Logger)
}

func initMaster(
	cfg config.Config,
	mqttClient *mqttLib.MQTTClient,
	storageHandler *storage.Handler,
	operatorController *controller.Controller,
) *master.Master {
	agentSync := time.Duration(cfg.AgentSyncIntervalSeconds * float64(time.Second))
	staleCheck := time.Duration(cfg.StaleOperatorCheckIntervalSeconds * float64(time.Second))
	return master.NewMaster(
		mqttClient,
		storageHandler,
		operatorController,
		agentSync,
		staleCheck,
		cfg.TimeoutInactiveAgentSeconds,
		cfg.TimeoutStaleOperatorSeconds,
	)
}
