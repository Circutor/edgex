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
package interfaces

import (
	contract "github.com/Circutor/edgex/pkg/models"
)

type DBClient interface {
	CloseSession()

	// ********************** EVENT FUNCTIONS *******************************
	// Return all the events
	// UnexpectedError - failed to retrieve events from the database
	Events() ([]contract.Event, error)

	// Return events up to the number specified
	// UnexpectedError - failed to retrieve events from the database
	EventsWithLimit(limit int) ([]contract.Event, error)

	// Add a new event
	// UnexpectedError - failed to add to database
	AddEvent(e contract.Event) (string, error)

	// Update an event - do NOT update readings
	// UnexpectedError - problem updating in database
	// NotFound - no event with the ID was found
	UpdateEvent(e contract.Event) error

	// Get an event by id
	EventById(id string) (contract.Event, error)

	// Get the number of events in Core Data
	EventCount() (int, error)

	// Get the number of events in Core Data for the device specified by id
	EventCountByDeviceId(id string) (int, error)

	// Update an event by ID
	// Set the pushed variable to the current time
	// 404 - Event not found
	// 503 - Unexpected problems
	//UpdateEventById(id string) error

	// Delete an event by ID and all of its readings
	// 404 - Event not found
	// 503 - Unexpected problems
	DeleteEventById(id string) error

	// Get a list of events that haven't been pushed yet to export/server based on the limit
	EventsUnpushedLimit(limit int) ([]contract.Event, error)

	// Get a list of events based on the device id and limit
	EventsForDeviceLimit(id string, limit int) ([]contract.Event, error)

	// Get a list of events based on the device id
	EventsForDevice(id string) ([]contract.Event, error)

	// Delete all of the events by the device id (and the readings)
	//DeleteEventsByDeviceId(id string) error

	// Return a list of events whos creation time is between startTime and endTime
	// Limit the number of results by limit
	EventsByCreationTime(startTime, endTime int64, limit int) ([]contract.Event, error)

	// Remove all the events that are older than the given age
	// Return the number of events removed
	//RemoveEventByAge(age int64) (int, error)

	// Get events that are older than a age
	EventsOlderThanAge(age int64) ([]contract.Event, error)

	// Remove all the events that have been pushed
	//func (dbc *DBClient) ScrubEvents()(int, error)

	// Get events that have been pushed (pushed field is not 0)
	EventsPushed() ([]contract.Event, error)

	// Delete all readings and events
	ScrubAllEvents() error
}
