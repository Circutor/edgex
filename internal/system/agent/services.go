/*******************************************************************************
 * Copyright 2018 Dell Technologies Inc.
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
 *
 *******************************************************************************/

package agent

import (
	"context"
	"fmt"

	"github.com/edgexfoundry/edgex-go/internal"
	"github.com/edgexfoundry/edgex-go/internal/system/agent/interfaces"
)

const (
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
)

func InvokeOperation(action string, services []string) bool {

	// Loop through requested operation, along with respectively-supplied parameters.
	for _, service := range services {
		LoggingClient.Info("invoking operation %s on service %s", action, service)

		if !isKnownServiceKey(service) {
			LoggingClient.Warn("unknown service %s during invocation", service)
		}

		switch action {

		case START:
			if starter, ok := executorClient.(interfaces.ServiceStarter); ok {
				err := starter.Start(service)
				if err != nil {
					LoggingClient.Error(fmt.Sprintf("error starting service %s: %v", service, err.Error()))
				}
			} else {
				LoggingClient.Warn(fmt.Sprintf("starting not supported with specified executor"))
			}

		case STOP:
			if stopper, ok := executorClient.(interfaces.ServiceStopper); ok {
				err := stopper.Stop(service)
				if err != nil {
					LoggingClient.Error(fmt.Sprintf("error stopping service %s: %v", service, err.Error()))
				}
			} else {
				LoggingClient.Warn(fmt.Sprintf("stopping not supported with specified executor"))
			}

		case RESTART:
			if restarter, ok := executorClient.(interfaces.ServiceRestarter); ok {
				err := restarter.Restart(service)
				if err != nil {
					LoggingClient.Error(fmt.Sprintf("error restarting service %s: %v", service, err.Error()))
				}
			} else {
				LoggingClient.Warn(fmt.Sprintf("restarting not supported with specified executor"))
			}
		}
	}
	return true
}

func getConfig(services []string, ctx context.Context) (ConfigRespMap, error) {

	c := ConfigRespMap{}
	c.Configuration = map[string]interface{}{}

	// Loop through requested actions, along with (any) respectively-supplied parameters.
	for _, service := range services {

		// Check whether SMA does _not_ know of ServiceKey ("service") as being one for one of its ready-made list of clients.
		if !isKnownServiceKey(service) {
			LoggingClient.Info(fmt.Sprintf("service %s not known to SMA as being in the ready-made list of clients", service))
			c.Configuration[service] = fmt.Sprintf("service %s not known to SMA", service)
		} else {
			// Service is known to SMA, so no need to ask the Registry for a ServiceEndpoint associated with `service`
			// Simply use one of the ready-made list of clients.
			LoggingClient.Info(fmt.Sprintf("service %s is known to SMA as being in the ready-made list of clients", service))

			responseJSON, err := generalClients[service].FetchConfiguration(ctx)
			if err != nil {
				c.Configuration[service] = fmt.Sprintf(err.Error())
				LoggingClient.Error(err.Error())
			} else {
				c.Configuration[service] = ProcessResponse(responseJSON)
			}
		}
	}
	return c, nil
}

func getMetrics(services []string, ctx context.Context) (MetricsRespMap, error) {

	m := MetricsRespMap{}
	m.Metrics = map[string]interface{}{}

	// Loop through requested actions, along with (any) respectively-supplied parameters.
	for _, service := range services {

		// Check whether SMA does _not_ know of ServiceKey ("service") as being one for one of its ready-made list of clients.
		if !isKnownServiceKey(service) {
			LoggingClient.Info(fmt.Sprintf("service %s not known to SMA as being in the ready-made list of clients", service))
			m.Metrics[service] = fmt.Sprintf("service %s not known to SMA", service)
		} else {
			// Service is known to SMA, so no need to ask the Registry for a ServiceEndpoint associated with `service`
			// Simply use one of the ready-made list of clients.
			LoggingClient.Info(fmt.Sprintf("service %s is known to SMA as being in the ready-made list of clients", service))

			responseJSON, err := generalClients[service].FetchMetrics(ctx)
			if err != nil {
				m.Metrics[service] = fmt.Sprintf(err.Error())
				LoggingClient.Error(err.Error())
			} else {
				m.Metrics[service] = ProcessResponse(responseJSON)
			}
		}
	}
	return m, nil
}

func getHealth(services []string) (map[string]interface{}, error) {

	health := make(map[string]interface{})

	for _, service := range services {

		if !isKnownServiceKey(service) {
			LoggingClient.Warn("unknown service %s found while getting health", service)
		}

		health[service] = true
	}

	return health, nil
}

func isKnownServiceKey(serviceKey string) bool {
	// create a map because this is the easiest/cleanest way to determine whether something exists in a set
	var services = map[string]struct{}{
		internal.SupportNotificationsServiceKey: {},
		internal.CoreCommandServiceKey:          {},
		internal.CoreDataServiceKey:             {},
		internal.CoreMetaDataServiceKey:         {},
		internal.ExportClientServiceKey:         {},
		internal.ExportDistroServiceKey:         {},
		internal.SupportLoggingServiceKey:       {},
		internal.SupportSchedulerServiceKey:     {},
	}

	_, exists := services[serviceKey]

	return exists
}
