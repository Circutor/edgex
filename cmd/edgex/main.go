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
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/support/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/context"
)

var LoggingRemoteURL string = "http://localhost:48061/api/v1/logs"

func main() {
	start := time.Now()

	// Make chanels
	errCh := make(chan error, 3)
	eventCh := make(chan *models.Event, 10)
	listenForInterrupt(errCh)

	// Initialize support-logging
	params := startup.BootParams{UseConsul: false, UseProfile: "support-logging", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, logging.Retry, logBeforeInit)
	if logging.Init(false) == false {
		time.Sleep(time.Millisecond * time.Duration(15))
		fmt.Printf("%s: Service bootstrap failed\n", internal.SupportLoggingServiceKey)
		return
	}
	// Start support-logging HTTP server
	go func() {
		r := fmt.Sprintf(":%d", logging.Configuration.Service.Port)
		errCh <- http.ListenAndServe(r, logging.HttpServer())
	}()

	// Create logging client
	loggingClient := logger.NewClient("edgex", true, LoggingRemoteURL)
	loggingClient.Info(fmt.Sprintf("Starting EdgeX %s ", edgex.Version), "start")

	// Initialize core-data
	params = startup.BootParams{UseConsul: false, UseProfile: "core-data", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, data.Retry, logBeforeInit)
	ok := data.Init(false)
	if !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.CoreDataServiceKey))
		return
	}
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(data.Configuration.Service.Timeout), "Request timed out")

	// Start core-data HTTP server
	go func() {
		r := data.LoadRestRoutes()
		errCh <- http.ListenAndServe(":"+strconv.Itoa(data.Configuration.Service.Port), context.ClearHandler(r))
	}()

	// Initialize export-client
	params = startup.BootParams{UseConsul: false, UseProfile: "export-client", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, client.Retry, logBeforeInit)
	ok = client.Init(false)
	if !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.ExportClientServiceKey))
		return
	}

	// Start export-client HTTP server
	client.StartHTTPServer(errCh)

	// Initialize core-metadata
	// We should initialize core-metadata after core-data and export-client to avoid BoltDB client issues
	params = startup.BootParams{UseConsul: false, UseProfile: "core-metadata", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, metadata.Retry, logBeforeInit)
	ok = metadata.Init(false)
	if !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.CoreMetaDataServiceKey))
		return
	}

	// Start core-metadata HTTP server
	go func() {
		r := metadata.LoadRestRoutes()
		errCh <- http.ListenAndServe(":"+strconv.Itoa(metadata.Configuration.Service.Port), context.ClearHandler(r))
	}()

	// Initialize core-command
	params = startup.BootParams{UseConsul: false, UseProfile: "core-command", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, command.Retry, logBeforeInit)
	ok = command.Init(false)
	if !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.CoreCommandServiceKey))
		return
	}

	// Start core-command HTTP server
	go func() {
		r := command.LoadRestRoutes()
		errCh <- http.ListenAndServe(":"+strconv.Itoa(command.Configuration.Service.Port), context.ClearHandler(r))
	}()

	// Initialize export-distro
	params = startup.BootParams{UseConsul: false, UseProfile: "export-distro", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, distro.Retry, logBeforeInit)
	if ok = distro.Init(false); !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.ExportDistroServiceKey))
		return
	}

	// All clients and HTTP servers have been started
	loggingClient.Info("EdgeX started in: "+time.Since(start).String(), "time")

	// There can be another receivers that can be initialiced here
	distro.MangosReceiver(eventCh)
	distro.Loop(errCh, eventCh)

	// Destroy all clients
	data.Destruct()
	client.Destruct()
	metadata.Destruct()
	command.Destruct()
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
