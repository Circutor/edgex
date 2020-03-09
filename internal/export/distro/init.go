//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logger"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

var LoggingClient logger.LoggingClient
var ec coredata.EventClient
var Configuration *ConfigurationStruct

func Retry(useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		// When looping, only handle configuration if it hasn't already been set.
		if Configuration == nil {
			Configuration, err = initializeConfiguration(useProfile)
			if err != nil {
				ch <- err

				// Error occurred when attempting to read from local filesystem. Fail fast.
				close(ch)
				wait.Done()
				return
			} else {
				// Setup Logging
				logTarget := setLoggingTarget()
				LoggingClient = logger.NewClient(internal.ExportDistroServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)

				// Initialize service clients
				initializeClient()
			}
		} else {
			// Once config is initialized, stop looping
			break
		}

		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func Init() bool {
	if Configuration == nil {
		return false
	}

	go telemetry.StartCpuUsageAverage()

	return true
}

func Destruct() {
}

func initializeConfiguration(useProfile string) (*ConfigurationStruct, error) {
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}

func initializeClient() {
	// Create data client
	url := Configuration.Clients["CoreData"].Url() + clients.ApiEventRoute
	ec = coredata.NewEventClient(url)
}

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}
