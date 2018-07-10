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
	"flag"
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
	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/usage"
	"github.com/edgexfoundry/edgex-go/support/logging"
	"github.com/edgexfoundry/edgex-go/support/logging-client"
	"go.uber.org/zap"
)

func main() {
	start := time.Now()
	start2 := start
	var useProfile string

	//-----------------------------------------
	//Support logging
	//-----------------------------------------
	flag.StringVar(&useProfile, "profile-support-logging", "support-logging", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "psl", "support-logging", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	var loggingClient_support_logging logger.LoggingClient
	//Read Configuration
	configuration_sl := &logging.ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration_sl)
	if err != nil {
		loggingClient_support_logging = logger.NewClient(internal.SupportLoggingServiceKey, false, "")
		loggingClient_support_logging.Error(err.Error())
		return
	}

	loggingClient_support_logging = logger.NewClient(internal.SupportLoggingServiceKey, false, configuration_sl.LoggingFile)
	loggingClient_support_logging.Info(fmt.Sprintf("Starting %s %s", internal.SupportLoggingServiceKey, edgex.Version))

	logging.Init(*configuration_sl)

	// Time it took to start service
	loggingClient_support_logging.Info("Service started in: "+time.Since(start).String(), "")

	//-----------------------------------------
	//Metadata
	//-----------------------------------------
	duration := time.Duration(1) * time.Second // Pause for 1 seconds
	time.Sleep(duration)
	start = time.Now()

	flag.StringVar(&useProfile, "profile-core-metadata", "core-metadata", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "pcm", "core-metadata", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	var loggingClient_core_metadata logger.LoggingClient
	//Read Configuration
	configuration_cm := &metadata.ConfigurationStruct{}
	err = config.LoadFromFile(useProfile, configuration_cm)
	if err != nil {
		loggingClient_core_metadata = logger.NewClient(internal.CoreMetaDataServiceKey, false, "")
		loggingClient_core_metadata.Error(err.Error())
		return
	}

	// Setup Logging
	logTarget := configuration_cm.LoggingRemoteURL
	if !configuration_cm.EnableRemoteLogging {
		logTarget = configuration_cm.LoggingFile
	}

	loggingClient_core_metadata = logger.NewClient(internal.CoreMetaDataServiceKey, configuration_cm.EnableRemoteLogging, logTarget)
	loggingClient_core_metadata.Info(fmt.Sprintf("Starting %s %s ", internal.CoreMetaDataServiceKey, edgex.Version))

	err = metadata.Init(*configuration_cm, loggingClient_core_metadata)
	if err != nil {
		loggingClient_core_metadata.Error(fmt.Sprintf("call to init() failed: %v", err.Error()))
		return
	}

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration_cm.ServiceTimeout), "Request timed out")
	loggingClient_core_metadata.Info(configuration_cm.AppOpenMsg, "")

	// Time it took to start service
	loggingClient_core_metadata.Info("Service started in: "+time.Since(start).String(), "")
	loggingClient_core_metadata.Info("Listening on port: " + strconv.Itoa(configuration_cm.ServicePort))

	//-----------------------------------------
	//Data
	//-----------------------------------------
	duration = time.Duration(1) * time.Second // Pause for 1 seconds
	time.Sleep(duration)
	start = time.Now()

	flag.StringVar(&useProfile, "profile-core-data", "core-data", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "pcd", "core-data", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	var loggingClient_core_data logger.LoggingClient
	//Read Configuration
	configuration_cd := &data.ConfigurationStruct{}
	err = config.LoadFromFile(useProfile, configuration_cd)
	if err != nil {
		loggingClient_core_data = logger.NewClient(internal.CoreDataServiceKey, false, "")
		loggingClient_core_data.Error(err.Error())
		return
	}

	// Setup Logging
	logTarget = configuration_cd.LoggingRemoteURL
	if !configuration_cm.EnableRemoteLogging {
		logTarget = configuration_cd.LoggingFile
	}

	loggingClient_core_data = logger.NewClient(internal.CoreDataServiceKey, configuration_cd.EnableRemoteLogging, logTarget)
	loggingClient_core_data.Info(fmt.Sprintf("Starting %s %s ", internal.CoreDataServiceKey, edgex.Version))

	err = data.Init(*configuration_cd, loggingClient_core_data, false)
	if err != nil {
		loggingClient_core_data.Error(fmt.Sprintf("call to init() failed: %v", err.Error()))
		return
	}

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration_cd.ServiceTimeout), "Request timed out")
	loggingClient_core_data.Info(configuration_cd.AppOpenMsg, "")

	// Time it took to start service
	loggingClient_core_data.Info("Service started in: "+time.Since(start).String(), "")
	loggingClient_core_data.Info("Listening on port: " + strconv.Itoa(configuration_cd.ServicePort))

	//-----------------------------------------
	//Command
	//-----------------------------------------
	duration = time.Duration(1) * time.Second // Pause for 1 seconds
	time.Sleep(duration)
	start = time.Now()

	flag.StringVar(&useProfile, "profile-core-command", "core-command", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "pcc", "core-command", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	var loggingClient_core_command logger.LoggingClient
	//Read Configuration
	configuration_cc := &command.ConfigurationStruct{}
	err = config.LoadFromFile(useProfile, configuration_cc)
	if err != nil {
		loggingClient_core_command = logger.NewClient(internal.CoreCommandServiceKey, false, "")
		loggingClient_core_command.Error(err.Error())
		return
	}
	// Setup Logging
	logTarget = configuration_cc.LoggingRemoteURL
	if !configuration_cc.EnableRemoteLogging {
		logTarget = configuration_cc.LogFile
	}
	loggingClient_core_command = logger.NewClient(internal.CoreCommandServiceKey, configuration_cc.EnableRemoteLogging, logTarget)
	loggingClient_core_command.Info(fmt.Sprintf("Starting %s %s ", internal.CoreCommandServiceKey, edgex.Version))

	command.Init(*configuration_cc, loggingClient_core_command, false)

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(configuration_cc.ServiceTimeout), "Request timed out")
	loggingClient_core_command.Info(configuration_cc.AppOpenMsg, "")

	// Time it took to start service
	loggingClient_core_command.Info("Service started in: "+time.Since(start).String(), "")
	loggingClient_core_command.Info("Listening on port: "+strconv.Itoa(configuration_cc.ServicePort), "")

	//-----------------------------------------
	//Export Client
	//-----------------------------------------
	duration = time.Duration(1) * time.Second // Pause for 1 seconds
	time.Sleep(duration)
	start = time.Now()

	var logger_export_client *zap.Logger

	logger_export_client, _ = zap.NewProduction()
	defer logger_export_client.Sync()

	logger_export_client.Info(fmt.Sprintf("Starting %s %s", internal.ExportClientServiceKey, edgex.Version))

	flag.StringVar(&useProfile, "profile-export-client", "export-client", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "pec", "export-client", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	configuration_ec := &client.ConfigurationStruct{}
	err = config.LoadFromFile(useProfile, configuration_ec)
	if err != nil {
		logger_export_client.Error(err.Error(), zap.String("version", edgex.Version))
		return
	}

	err = client.Init(*configuration_ec, logger_export_client)
	if err != nil {
		logger_export_client.Error("Could not initialize export client", zap.Error(err))
		return
	}

	//-----------------------------------------
	//Export distro
	//-----------------------------------------
	duration = time.Duration(1) * time.Second // Pause for 10 seconds
	time.Sleep(duration)
	start = time.Now()

	var logger_export_distro *zap.Logger

	logger_export_distro, _ = zap.NewProduction()
	defer logger_export_distro.Sync()

	logger_export_distro.Info("Starting "+internal.ExportDistroServiceKey, zap.String("version", edgex.Version))

	flag.StringVar(&useProfile, "profile-export-distro", "export-distro", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "ped", "export-distro", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	configuration_ed := &distro.ConfigurationStruct{}
	err = config.LoadFromFile(useProfile, configuration_ed)
	if err != nil {
		logger_export_distro.Error(err.Error(), zap.String("version", edgex.Version))
		return
	}

	err = distro.Init(*configuration_ed, logger_export_distro)

	logger_export_distro.Info("Starting distro")

	//Make chanels
	errs := make(chan error, 3)
	eventCh := make(chan *models.Event, 10)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logging.StartHTTPServer(errs)

	go func() {
		rm := metadata.LoadRestRoutes()
		errs <- http.ListenAndServe(":"+strconv.Itoa(configuration_cm.ServicePort), rm)
	}()

	go func() {
		rd := data.LoadRestRoutes()
		errs <- http.ListenAndServe(":"+strconv.Itoa(configuration_cd.ServicePort), rd)
	}()

	go func() {
		rc := command.LoadRestRoutes()
		errs <- http.ListenAndServe(":"+strconv.Itoa(configuration_cc.ServicePort), rc)
	}()

	client.StartHTTPServer(*configuration_ec, errs)

	fmt.Println("Pack Services started in: " + time.Since(start2).String())

	// There can be another receivers that can be initialiced here
	distro.ZeroMQReceiver(eventCh)
	distro.Loop(errs, eventCh)

	metadata.Destruct()
	data.Destruct()
	client.Destroy()

	logger_export_distro.Warn(fmt.Sprintf("terminating: %v", errs))
}
