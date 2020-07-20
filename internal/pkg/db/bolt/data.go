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

package bolt

import (
	"fmt"

	"github.com/Circutor/edgex/internal/pkg/db"
	contract "github.com/Circutor/edgex/pkg/models"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	bolt "go.etcd.io/bbolt"
)

/*
Core data client
Has functions for interacting with the core data bolt database
*/

const (
	maxEvents = 50000
)

// ******************************* EVENTS **********************************

// Return all the events
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (bc *BoltClient) Events() ([]contract.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		return true
	}, -1)
}

// Return events up to the max number specified
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (bc *BoltClient) EventsWithLimit(limit int) ([]contract.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		return true
	}, limit)
}

// Add a new event
// UnexpectedError - failed to add to database
// NoValueDescriptor - no existing value descriptor for a reading in the event
func (bc *BoltClient) AddEvent(e contract.Event) (string, error) {
	e.Created = db.MakeTimestamp()
	e.ID = fmt.Sprintf("%013d-", e.Created) + uuid.New().String()
	e.Modified = e.Created

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(db.EventsCollection))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}

		// Deletes older event if its necessary
		numElements := b.Stats().KeyN
		if numElements > maxEvents {
			cursor := b.Cursor()
			cursor.First()
			err := cursor.Delete()
			if err != nil {
				return err
			}
		}

		encoded, err := json.Marshal(e)
		if err != nil {
			return err
		}
		return b.Put([]byte(e.ID), encoded)
	})
	return e.ID, err
}

// Update an event - do NOT update readings
// UnexpectedError - problem updating in database
// NotFound - no event with the ID was found
func (bc *BoltClient) UpdateEvent(e contract.Event) error {
	e.Modified = db.MakeTimestamp()

	return bc.update(db.EventsCollection, e, e.ID)
}

// Get an event by id
func (bc *BoltClient) EventById(id string) (contract.Event, error) {
	var ev contract.Event
	err := bc.getById(&ev, db.EventsCollection, id)
	return ev, err
}

// Get the number of events in bolt
func (bc *BoltClient) EventCount() (int, error) {
	return bc.count(db.EventsCollection)
}

// Get the number of events in bolt for the device
func (bc *BoltClient) EventCountByDeviceId(devid string) (int, error) {
	bstat := 0
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.EventsCollection))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "device").ToString()
			if value == devid {
				bstat++
			}
			return nil
		})
		return err
	})
	return bstat, err
}

// Delete an event by ID and all of its readings
// 404 - Event not found
// 503 - Unexpected problems
func (bc *BoltClient) DeleteEventById(id string) error {
	return bc.deleteById(id, db.EventsCollection)
}

// Get a list of events based on the device id and limit
func (bc *BoltClient) EventsForDeviceLimit(ide string, limit int) ([]contract.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "device").ToString()
		if value == ide {
			return true
		}
		return false
	}, limit)
}

// Get a list of events based on the device id
func (bc *BoltClient) EventsForDevice(ide string) ([]contract.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "device").ToString()
		if value == ide {
			return true
		}
		return false
	}, -1)
}

// Return a list of events whos creation time is between startTime and endTime
// Limit the number of results by limit
func (bc *BoltClient) EventsByCreationTime(startTime, endTime int64, limit int) ([]contract.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "created").ToInt64()
		if (value >= startTime) && (value <= endTime) {
			return true
		}
		return false
	}, limit)
}

// Get Events that are older than the given age (defined by age = now - created)
func (bc *BoltClient) EventsOlderThanAge(age int64) ([]contract.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "created").ToInt64()
		value = (db.MakeTimestamp()) - value
		if value >= age {
			return true
		}
		return false
	}, -1)
}

// Get all of the events that have been pushed
func (bc *BoltClient) EventsPushed() ([]contract.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "pushed").ToInt64()
		if value != 0 {
			return true
		}
		return false
	}, -1)
}

// Delete all of the readings and all of the events
func (bc *BoltClient) ScrubAllEvents() error {
	bc.scrubAll(db.EventsCollection)
	bc.scrubAll(db.ReadingsCollection)
	return nil
}

// Get events for the passed check
func (bc *BoltClient) getEvents(fn func(encoded []byte) bool, limit int) ([]contract.Event, error) {
	var event contract.Event
	var events []contract.Event
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	// Check if limit is not 0
	if limit == 0 {
		return events, nil
	}
	cnt := 0

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.EventsCollection))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &event)
				if err != nil {
					return err
				}
				events = append(events, event)
				if limit > 0 {
					cnt++
					if cnt >= limit {
						return ErrLimReached
					}
				}
			}
			return nil
		})
		if err == ErrLimReached {
			return nil
		}
		return err
	})
	return events, err
}

// ************************ READINGS ************************************

// Return a list of readings sorted by reading id
func (bc *BoltClient) Readings() ([]contract.Reading, error) {
	return bc.getReadings(func(encoded []byte) bool {
		return true
	}, -1)
}

// Post a new reading
func (bc *BoltClient) AddReading(r contract.Reading) (string, error) {
	r.Id = uuid.New().String()
	r.Created = db.MakeTimestamp()
	r.Modified = r.Created

	err := bc.add(db.ReadingsCollection, r, r.Id)
	return r.Id, err
}

// Update a reading
// 404 - reading cannot be found
// 409 - Value descriptor doesn't exist
// 503 - unknown issues
func (bc *BoltClient) UpdateReading(r contract.Reading) error {
	r.Modified = db.MakeTimestamp()

	return bc.update(db.ReadingsCollection, r, r.Id)
}

// Get a reading by ID
func (bc *BoltClient) ReadingById(id string) (contract.Reading, error) {
	var reading contract.Reading
	err := bc.getById(&reading, db.ReadingsCollection, id)
	return reading, err
}

// Get the count of readings in BOLT
func (bc *BoltClient) ReadingCount() (int, error) {
	return bc.count(db.ReadingsCollection)
}

// Delete a reading by ID
// 404 - can't find the reading with the given id
func (bc *BoltClient) DeleteReadingById(id string) error {
	return bc.deleteById(id, db.ReadingsCollection)
}

// Return a list of readings for the given device (id or name)
// Sort the list of readings on creation date
func (bc *BoltClient) ReadingsByDevice(ids string, limit int) ([]contract.Reading, error) {
	return bc.getReadings(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "device").ToString()
		if value == ids {
			return true
		}
		return false
	}, limit)
}

// Return a list of readings for the given value descriptor
// Limit by the given limit
func (bc *BoltClient) ReadingsByValueDescriptor(name string, limit int) ([]contract.Reading, error) {
	return bc.getReadings(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "name").ToString()
		if value == name {
			return true
		}
		return false
	}, limit)
}

// Return a list of readings whos creation time is in-between start and end
// Limit by the limit parameter
func (bc *BoltClient) ReadingsByCreationTime(start, end int64, limit int) ([]contract.Reading, error) {
	return bc.getReadings(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "created").ToInt64()
		if (value >= start) && (value <= end) {
			return true
		}
		return false
	}, limit)
}

// Return a list of readings for a device filtered by the value descriptor and limited by the limit
// The readings are linked to the device through an event
func (bc *BoltClient) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]contract.Reading, error) {
	return bc.getReadings(func(encoded []byte) bool {
		valuedev := jsoniter.Get(encoded, "device").ToString()
		valuename := jsoniter.Get(encoded, "name").ToString()
		if (valuename == valueDescriptor) || (valuedev == deviceId) {
			return true
		}
		return false
	}, limit)
}

// Get readings for the passed check
func (bc *BoltClient) getReadings(fn func(encoded []byte) bool, limit int) ([]contract.Reading, error) {
	r := contract.Reading{}
	rs := []contract.Reading{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	// Check if limit is not 0
	if limit == 0 {
		return rs, nil
	}
	cnt := 0

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.ReadingsCollection))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &r)
				if err != nil {
					return err
				}
				rs = append(rs, r)
				if limit > 0 {
					cnt++
					if cnt >= limit {
						return ErrLimReached
					}
				}
			}
			return nil
		})
		if err == ErrLimReached {
			return nil
		}
		return err
	})
	return rs, err
}
