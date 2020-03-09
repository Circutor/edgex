//
// Copyright (c) 2018 Tencent
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package logging

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/edgex-go/pkg/clients/logger"

	"github.com/edgexfoundry/edgex-go/internal/pkg/config"
	"github.com/edgexfoundry/edgex-go/internal/pkg/telemetry"
)

var Configuration *ConfigurationStruct
var dbClient persistence
var LoggingClient logger.LoggingClient

func Retry(useProfile string, timeout int, wait *sync.WaitGroup, ch chan error) {
	LoggingClient = newPrivateLogger()
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
			}
		}

		// Only attempt to connect to database if configuration has been populated
		if Configuration != nil {
			err = getPersistence()
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

	go telemetry.StartCpuUsageAverage()

	return true
}

func Destruct() {
	if dbClient != nil {
		dbClient.closeSession()
		dbClient = nil
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

func getPersistence() error {
	switch Configuration.Writable.Persistence {
	case PersistenceFile:
		dbClient = &fileLog{filename: Configuration.Logging.File}
	case PersistenceDB:
		// TODO: Integrate db layer with internal/pkg/db/ types so we can support other databases
		ms, err := connectToMongo()
		if err != nil {
			return err
		} else {
			dbClient = &mongoLog{session: ms}
		}
	default:
		return errors.New(fmt.Sprintf("unrecognized value Configuration.Persistence: %s", Configuration.Writable.Persistence))
	}
	return nil
}
