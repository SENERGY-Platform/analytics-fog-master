package lib

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
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
	"github.com/SENERGY-Platform/analytics-fog-master/migrations"
	sb_util "github.com/SENERGY-Platform/go-service-base/util"
	"github.com/SENERGY-Platform/go-service-base/watchdog"
)

func Run(
	ctx    context.Context,
	stdout, stderr io.Writer,
	config config.Config,
) error {
	err := logging.InitLogger(stdout, true)
	if err != nil {
		log.Printf("Error init logging: %s", err.Error())
		return err
	}
 
	logging.Logger.Info(fmt.Sprintf("config: %s", sb_util.ToJsonStr(config)))
 
	logging.Logger.Info("Create new database at " + config.DataBase.Path)
	db, err := storage.NewDB(config.DataBase.Path)
	if err != nil {
		logging.Logger.Error("Cant init DB", "error", err.Error())
		return err
	}
	err = migrations.MigrateDb(config.DataBase.Path)
	if err != nil {
		logging.Logger.Error("Cant migrate DB", "error", err.Error())
		return err
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
 
	ec := watchdog.Join()
	if ec != 0 {
		return errors.New("Could not join")
	}
	logging.Logger.Info("Shutdowned graceful")
	return nil
}