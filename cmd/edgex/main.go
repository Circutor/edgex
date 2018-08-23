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

	"github.com/edgexfoundry/edgex-go"
	"github.com/edgexfoundry/edgex-go/export/client"
	"github.com/edgexfoundry/edgex-go/export/distro"
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/command"
	"github.com/edgexfoundry/edgex-go/internal/core/data"
	"github.com/edgexfoundry/edgex-go/internal/core/metadata"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"go.uber.org/zap"
)

var bootTimeout int = 30000 //Once we start the V2 configuration rework, this will be config driven

func main() {
	start := time.Now()

	// Create logging client
	loggingClient := logger.NewClient("edgex", false, "")
	loggingClient.Info(fmt.Sprintf("Starting EdgeX %s ", edgex.Version))

	// Create ZAP logging client
	var loggerClientZap *zap.Logger
	loggerClientZap, _ = zap.NewProduction()
	defer loggerClientZap.Sync()

	// Initialize core-data
	// We should initialize core-data before core-metadata to avoid BoltDB client issues
	params := startup.BootParams{UseConsul: false, UseProfile: "core-data", BootTimeout: bootTimeout}
	startup.Bootstrap(params, data.Retry, logBeforeInit)
	if data.Init() == false {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed!", internal.CoreDataServiceKey))
		return
	}

	// Initialize core-metadata
	params = startup.BootParams{UseConsul: false, UseProfile: "core-metadata", BootTimeout: bootTimeout}
	startup.Bootstrap(params, metadata.Retry, logBeforeInit)
	if metadata.Init() == false {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed!", internal.CoreMetaDataServiceKey))
		return
	}

	// Initialize core-command
	params = startup.BootParams{UseConsul: false, UseProfile: "core-command", BootTimeout: bootTimeout}
	startup.Bootstrap(params, command.Retry, logBeforeInit)
	if command.Init() == false {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed!", internal.CoreCommandServiceKey))
		return
	}

	// Initialize export-client
	ecConfiguration := &client.ConfigurationStruct{}
	err := config.LoadFromFile("export-client", ecConfiguration)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}
	err = client.Init(*ecConfiguration, loggerClientZap)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("Could not initialize export-client: %v", err.Error()))
		return
	}

	// Initialize export-distro
	edConfiguration := &distro.ConfigurationStruct{}
	err = config.LoadFromFile("export-distro", edConfiguration)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}
	err = distro.Init(*edConfiguration, loggerClientZap, false)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("Could not initialize export-distro: %v", err.Error()))
		return
	}

	// Make chanels
	errs := make(chan error, 3)
	eventCh := make(chan *models.Event, 10)
	listenForInterrupt(errs)

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
	loggingClient.Info("EdgeX started in: "+time.Since(start).String(), "")

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
