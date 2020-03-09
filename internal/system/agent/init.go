/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
package agent

import (
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/general"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logger"
)

// Global variables
var Configuration *ConfigurationStruct
var generalClients map[string]general.GeneralClient
var LoggingClient logger.LoggingClient

// Note that executorClient is the empty interface so that we may type-cast it
// to whatever operation we need it to do at runtime.
var executorClient interface{}

var services = map[string]string{
	internal.SupportNotificationsServiceKey: "Notifications",
	internal.CoreCommandServiceKey:          "Command",
	internal.CoreDataServiceKey:             "CoreData",
	internal.CoreMetaDataServiceKey:         "Metadata",
	internal.ExportClientServiceKey:         "Export",
	internal.ExportDistroServiceKey:         "Distro",
	internal.SupportLoggingServiceKey:       "Logging",
	internal.SupportSchedulerServiceKey:     "Scheduler",
}

func Retry(useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	until := time.Now().Add(time.Millisecond * time.Duration(timeout))
	for time.Now().Before(until) {
		var err error
		// When looping, only handle configuration if it hasn't already been set.
		// Note, too, that the SMA-managed services are bootstrapped by the SMA.
		// Read in those setting, too, which specifies details for those services
		// (Those setting were _previously_ to be found in a now-defunct TOML manifest file).
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
				LoggingClient = logger.NewClient(internal.SystemManagementAgentServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)

				// Initialize service clients
				initializeClients()
			}
		}

		// Exit the loop if the dependencies have been satisfied.
		if Configuration != nil {
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

func initializeClients() {

	generalClients = make(map[string]general.GeneralClient)

	for serviceKey, serviceName := range services {
		url := Configuration.Clients[serviceName].Url()
		generalClients[serviceKey] = general.NewGeneralClient(url)
	}
}

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}
