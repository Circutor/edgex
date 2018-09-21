/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	edgex "github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/command"
	"github.com/edgexfoundry/edgex-go/internal/core/data"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/export/client"
	"github.com/edgexfoundry/edgex-go/internal/export/distro"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/support/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"go.uber.org/zap"
)

var bootTimeout int = 30000 //Once we start the V2 configuration rework, this will be config driven
var EnableRemoteLogging bool = true
var LoggingFile string = "./logs/edgex.log"

//var LoggingFile string = "/var/log/edgex.log" //para arm compile
var LoggingRemoteURL string = "http://localhost:48061/api/v1/logs"

func main() {
	start := time.Now()
	// Create logging client
	logger.VerifyLogDirectory(LoggingFile)
	loggingClient := logger.NewClient("edgex", EnableRemoteLogging, LoggingRemoteURL)

	// Create ZAP logging client
	var loggerClientZap *zap.Logger
	loggerClientZap, _ = zap.NewProduction()
	defer loggerClientZap.Sync()

	// Initialize support-logging
	params := startup.BootParams{UseConsul: false, UseProfile: "support-logging", BootTimeout: bootTimeout}
	startup.Bootstrap(params, logging.Retry, logBeforeInit)
	if logging.Init() == false {
		time.Sleep(time.Millisecond * time.Duration(15))
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed!", internal.SupportLoggingServiceKey, "error"))
		return
	}

	// Make chanels
	errs := make(chan error, 3)
	eventCh := make(chan *models.Event, 10)
	listenForInterrupt(errs)

	// Start support-logginfg HTTP server
	go func() {
		rsl := fmt.Sprintf(":%d", logging.Configuration.Port)
		errs <- http.ListenAndServe(rsl, logging.HttpServer())
	}()

	loggingClient.Info(fmt.Sprintf("Starting EdgeX %s ", edgex.Version), "start")
	time.Sleep(time.Microsecond * time.Duration(500))
	loggingClient.Info(fmt.Sprintf("Started %s %s ", internal.SupportLoggingServiceKey, edgex.Version), "start")
	time.Sleep(time.Microsecond * time.Duration(500))
	loggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.CoreDataServiceKey, edgex.Version), "start")
	// Initialize core-data
	params = startup.BootParams{UseConsul: false, UseProfile: "core-data", BootTimeout: bootTimeout}
	startup.Bootstrap(params, data.Retry, logBeforeInit)
	if data.Init() == false {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed!", internal.CoreDataServiceKey), "error")
		return
	}

	loggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.ExportClientServiceKey, edgex.Version), "start")
	// Initialize export-client
	ecConfiguration := &client.ConfigurationStruct{}
	err := config.LoadFromFile("export-client", ecConfiguration)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}
	err = client.Init(*ecConfiguration, loggerClientZap)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("Could not initialize export-client: %v", err.Error()), "error")
		return
	}

	loggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.CoreMetaDataServiceKey, edgex.Version), "start")
	// Initialize core-metadata
	// We should initialize core-metadata after core-data and export-client to avoid BoltDB client issues
	params = startup.BootParams{UseConsul: false, UseProfile: "core-metadata", BootTimeout: bootTimeout}
	startup.Bootstrap(params, metadata.Retry, logBeforeInit)
	if metadata.Init() == false {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed!", internal.CoreMetaDataServiceKey), "error")
		return
	}

	loggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.CoreCommandServiceKey, edgex.Version), "start")
	// Initialize core-command
	params = startup.BootParams{UseConsul: false, UseProfile: "core-command", BootTimeout: bootTimeout}
	startup.Bootstrap(params, command.Retry, logBeforeInit)
	if command.Init() == false {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed!", internal.CoreCommandServiceKey), "error")
		return
	}

	loggingClient.Info(fmt.Sprintf("Starting %s %s ", internal.ExportDistroServiceKey, edgex.Version), "start")
	// Initialize export-distro
	edConfiguration := &distro.ConfigurationStruct{}
	err = config.LoadFromFile("export-distro", edConfiguration)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}
	err = distro.Init(*edConfiguration, loggerClientZap, false)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("Could not initialize export-distro: %v", err.Error()), "error")
		return
	}

	// Start core-data HTTP server
	go func() {
		rd := data.LoadRestRoutes()
		errs <- http.ListenAndServe(":"+strconv.Itoa(data.Configuration.ServicePort), rd)
	}()

	// Start core-metadata HTTP server
	go func() {
		rm := metadata.LoadRestRoutes()
		errs <- http.ListenAndServe(":"+strconv.Itoa(metadata.Configuration.ServicePort), rm)
	}()

	// Start core-command HTTP server
	go func() {
		rc := command.LoadRestRoutes()
		errs <- http.ListenAndServe(":"+strconv.Itoa(command.Configuration.ServicePort), rc)
	}()

	// Start export-client HTTP server
	client.StartHTTPServer(*ecConfiguration, errs)

	// All clients and HTTP servers have been started
	time.Sleep(time.Microsecond * time.Duration(500))
	loggingClient.Info("EdgeX started in: "+time.Since(start).String(), "time")

	// There can be another receivers that can be initialiced here
	distro.MangosReceiver(eventCh)
	distro.Loop(errs, eventCh)

	// Destroy all clients
	metadata.Destruct()
	data.Destruct()
	client.Destroy()
}

func logBeforeInit(err error) {
	l := logger.NewClient("edgex", false, "")
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
