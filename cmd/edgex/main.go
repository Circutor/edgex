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
	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	"github.com/edgexfoundry/edgex-go/internal/pkg/startup"
	"github.com/edgexfoundry/edgex-go/internal/support/logging"
	"github.com/edgexfoundry/edgex-go/internal/support/scheduler"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logger"
	"github.com/gorilla/context"
)

var loggingRemoteURL = "http://localhost:48061/api/v1/logs"
var loggingClient logger.LoggingClient

func main() {
	start := time.Now()

	// Make chanels
	errCh := make(chan error, 3)
	eventCh := make(chan *models.Event, 10)
	listenForInterrupt(errCh)

	// Initialize support-logging
	iniSupportLogging(errCh)

	// Wait until support-logging is running
	time.Sleep(time.Millisecond * 500)

	// Create logging client
	loggingClient = logger.NewClient("edgex", true, loggingRemoteURL, logger.InfoLog)
	loggingClient.Info(fmt.Sprintf("Starting EdgeX %s ", edgex.Version))

	// Initialize core-data
	iniCoreData(errCh)

	// Initialize export-client
	iniExportClient(errCh)

	// Initialize support-scheduler
	iniSupportScheduler(errCh)

	// Initialize core-metadata
	// We should initialize core-metadata after core-data, support-scheduler and export-client to avoid BoltDB client issues
	iniCoreMetadata(errCh)

	// Initialize core-command
	iniCoreCommand(errCh)

	// Initialize export-distro
	iniExportDistro()

	// All clients and HTTP servers have been started
	loggingClient.Info("EdgeX started in: " + time.Since(start).String())

	// There can be another receivers that can be initialiced here
	distro.MangosReceiver(eventCh)
	distro.Loop(errCh, eventCh)

	// Destroy all clients
	data.Destruct()
	client.Destruct()
	metadata.Destruct()
	command.Destruct()

	os.Exit(0)
}

// Initialize support-logging
func iniSupportLogging(errCh chan error) {
	params := startup.BootParams{UseProfile: "support-logging", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, logging.Retry, logBeforeInit)
	if logging.Init() == false {
		time.Sleep(time.Millisecond * time.Duration(15))
		fmt.Printf("%s: Service bootstrap failed\n", internal.SupportLoggingServiceKey)
		os.Exit(1)
	}
	// Start support-logging HTTP server
	go func() {
		r := fmt.Sprintf(":%d", logging.Configuration.Service.Port)
		errCh <- http.ListenAndServe(r, logging.HttpServer())
	}()
}

// Initialize support-scheduler
func iniSupportScheduler(errCh chan error) {
	params := startup.BootParams{UseProfile: "support-scheduler", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, scheduler.Retry, logBeforeInit)
	ok := scheduler.Init()
	if !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.SupportSchedulerServiceKey))
		os.Exit(1)
	}

	time.Sleep(time.Millisecond * time.Duration(1000))

	// Bootstrap schedulers
	err := scheduler.LoadScheduler()
	if err != nil {
		scheduler.LoggingClient.Error(fmt.Sprintf("Failed to load schedules and events %s", err.Error()))
	}

	// Start support-scheduler HTTP server
	go func() {
		r := scheduler.LoadRestRoutes()
		errCh <- http.ListenAndServe(":"+strconv.Itoa(scheduler.Configuration.Service.Port), context.ClearHandler(r))
	}()

	// Start the ticker
	scheduler.StartTicker()
}

// Initialize core-data
func iniCoreData(errCh chan error) {
	params := startup.BootParams{UseProfile: "core-data", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, data.Retry, logBeforeInit)
	ok := data.Init()
	if !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.CoreDataServiceKey))
		os.Exit(1)
	}
	http.TimeoutHandler(nil, time.Millisecond*time.Duration(data.Configuration.Service.Timeout), "Request timed out")

	// Start core-data HTTP server
	go func() {
		r := data.LoadRestRoutes()
		errCh <- http.ListenAndServe(":"+strconv.Itoa(data.Configuration.Service.Port), context.ClearHandler(r))
	}()
}

// Initialize core-metadata
func iniCoreMetadata(errCh chan error) {
	params := startup.BootParams{UseProfile: "core-metadata", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, metadata.Retry, logBeforeInit)
	ok := metadata.Init()
	if !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.CoreMetaDataServiceKey))
		os.Exit(1)
	}

	// Start core-metadata HTTP server
	go func() {
		r := metadata.LoadRestRoutes()
		errCh <- http.ListenAndServe(":"+strconv.Itoa(metadata.Configuration.Service.Port), context.ClearHandler(r))
	}()
}

// Initialize core-command
func iniCoreCommand(errCh chan error) {
	params := startup.BootParams{UseProfile: "core-command", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, command.Retry, logBeforeInit)
	ok := command.Init()
	if !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.CoreCommandServiceKey))
		os.Exit(1)
	}

	// Start core-command HTTP server
	go func() {
		r := command.LoadRestRoutes()
		errCh <- http.ListenAndServe(":"+strconv.Itoa(command.Configuration.Service.Port), context.ClearHandler(r))
	}()
}

// Initialize export-client
func iniExportClient(errCh chan error) {
	params := startup.BootParams{UseProfile: "export-client", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, client.Retry, logBeforeInit)
	ok := client.Init()
	if !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.ExportClientServiceKey))
		os.Exit(1)
	}

	// Start export-client HTTP server
	client.StartHTTPServer(errCh)
}

// Initialize export-distro
func iniExportDistro() {
	params := startup.BootParams{UseProfile: "export-distro", BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, distro.Retry, logBeforeInit)
	if ok := distro.Init(); !ok {
		loggingClient.Error(fmt.Sprintf("%s: Service bootstrap failed", internal.ExportDistroServiceKey))
		os.Exit(1)
	}
}

func logBeforeInit(err error) {
	l := logger.NewClient("edgex", false, "", logger.InfoLog)
	l.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
