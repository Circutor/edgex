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
	"github.com/edgexfoundry/edgex-go/core/command"
	"github.com/edgexfoundry/edgex-go/core/data"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/edgexfoundry/edgex-go/core/metadata"
	"github.com/edgexfoundry/edgex-go/export/client"
	"github.com/edgexfoundry/edgex-go/export/distro"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
	"go.uber.org/zap"
)

var loggingClient logger.LoggingClient

func main() {
	start := time.Now()

	// Create logging client
	loggingClient = logger.NewClient("edgex", false, "")
	loggingClient.Info(fmt.Sprintf("Starting EdgeX %s ", edgex.Version))

	// Create ZAP logging client
	var loggerClientZap *zap.Logger
	loggerClientZap, _ = zap.NewProduction()
	defer loggerClientZap.Sync()

	// Read core-data configuration
	cdConfiguration := &data.ConfigurationStruct{}
	err := config.LoadFromFile("core-data", cdConfiguration)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}

	// Read core-metadata configuration
	cmConfiguration := &metadata.ConfigurationStruct{}
	err = config.LoadFromFile("core-metadata", cmConfiguration)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}

	// Read core-command configuration
	ccConfiguration := &command.ConfigurationStruct{}
	err = config.LoadFromFile("core-command", ccConfiguration)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}

	// Read export-client configuration
	ecConfiguration := &client.ConfigurationStruct{}
	err = config.LoadFromFile("export-client", ecConfiguration)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}

	// Read export-distro configuration
	edConfiguration := &distro.ConfigurationStruct{}
	err = config.LoadFromFile("export-distro", edConfiguration)
	if err != nil {
		loggingClient.Error(err.Error())
		return
	}

	// Initialize core-data
	// We should initialize core-data before core-metadata to avoid BoltDB client issues
	err = data.Init(*cdConfiguration, loggingClient, false)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("Could not initialize core-data: %v", err.Error()))
		return
	}

	// Initialize core-metadata
	err = metadata.Init(*cmConfiguration, loggingClient)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("Could not initialize core-metadata: %v", err.Error()))
		return
	}

	// Initialize core-command
	command.Init(*ccConfiguration, loggingClient, false)

	// Initialize export-client
	err = client.Init(*ecConfiguration, loggerClientZap)
	if err != nil {
		loggingClient.Error(fmt.Sprintf("Could not initialize export-client: %v", err.Error()))
		return
	}

	// Initialize export-distro
	distro.Init(*edConfiguration, loggerClientZap)

	// Make chanels
	errs := make(chan error, 3)
	eventCh := make(chan *models.Event, 10)
	listenForInterrupt(errs)

	// Start core-data HTTP server
	go func() {
		rd := data.LoadRestRoutes()
		errs <- http.ListenAndServe(":"+strconv.Itoa(cdConfiguration.ServicePort), rd)
	}()

	// Start core-metadata HTTP server
	go func() {
		rm := metadata.LoadRestRoutes()
		errs <- http.ListenAndServe(":"+strconv.Itoa(cmConfiguration.ServicePort), rm)
	}()

	// Start core-command HTTP server
	go func() {
		rc := command.LoadRestRoutes()
		errs <- http.ListenAndServe(":"+strconv.Itoa(ccConfiguration.ServicePort), rc)
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

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
