/*******************************************************************************
 * Copyright 2017 Dell Inc.
 * Copyright (c) 2019 Intel Corporation
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
package data

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/core/data/interfaces"
	"github.com/edgexfoundry/edgex-go/internal/core/data/messaging"
	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/bolt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/logger"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
)

// Global variables
var Configuration *ConfigurationStruct
var dbClient interfaces.DBClient
var LoggingClient logger.LoggingClient

var chEvents chan interface{} //A channel for "domain events" sourced from event operations

var ep messaging.EventPublisher
var mdc metadata.DeviceClient
var msc metadata.DeviceServiceClient

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
				LoggingClient = logger.NewClient(internal.CoreDataServiceKey, Configuration.Logging.EnableRemote, logTarget, Configuration.Writable.LogLevel)

				// Initialize service clients
				initializeClients()
			}
		}

		// Only attempt to connect to database if configuration has been populated
		if Configuration != nil {
			err := connectToDatabase()
			if err != nil {
				ch <- err
			} else {
				break
			}
		}
		time.Sleep(time.Second * time.Duration(1))
	}
	close(ch)
	wait.Done()

	return
}

func Init() bool {
	if Configuration == nil || dbClient == nil {
		return false
	}
	chEvents = make(chan interface{}, 100)
	initEventHandlers()

	go telemetry.StartCpuUsageAverage()

	return true
}

func Destruct() {
	if dbClient != nil {
		dbClient.CloseSession()
		dbClient = nil
	}
	if chEvents != nil {
		close(chEvents)
	}
}

func initializeConfiguration(useProfile string) (*ConfigurationStruct, error) {
	configuration := &ConfigurationStruct{}
	err := config.LoadFromFile(useProfile, configuration)
	if err != nil {
		return nil, err
	}

	return configuration, nil
}

func connectToDatabase() error {
	// Create a database client
	var err error
	dbConfig := db.Configuration{
		Host:         Configuration.Databases["Primary"].Host,
		Port:         Configuration.Databases["Primary"].Port,
		Timeout:      Configuration.Databases["Primary"].Timeout,
		DatabaseName: Configuration.Databases["Primary"].Name,
		Username:     Configuration.Databases["Primary"].Username,
		Password:     Configuration.Databases["Primary"].Password,
	}

	dbClient, err = newDBClient(Configuration.Databases["Primary"].Type, dbConfig)
	if err != nil {
		dbClient = nil
		return fmt.Errorf("couldn't create database client: %v", err.Error())
	}

	return err
}

// Return the dbClient interface
func newDBClient(dbType string, config db.Configuration) (interfaces.DBClient, error) {
	switch dbType {
	case db.MongoDB:
		return mongo.NewClient(config)
	case db.BoltDB:
		return bolt.NewClient(config)
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}

func initializeClients() {
	// Create metadata clients
	url := Configuration.Clients["Metadata"].Url() + clients.ApiDeviceRoute
	mdc = metadata.NewDeviceClient(url)

	url = Configuration.Clients["Metadata"].Url() + clients.ApiDeviceServiceRoute
	msc = metadata.NewDeviceServiceClient(url)

	// Create the event publisher
	ep = messaging.NewEventPublisher(messaging.PubSubConfiguration{
		AddressPort: Configuration.MessageQueue.Uri(),
	})
}

func setLoggingTarget() string {
	if Configuration.Logging.EnableRemote {
		return Configuration.Clients["Logging"].Url() + clients.ApiLoggingRoute
	}
	return Configuration.Logging.File
}
