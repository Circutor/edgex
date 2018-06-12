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

package clients

import (
	"errors"
	"time"

	"github.com/coreos/bbolt"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

/*
Core data client
Has functions for interacting with the core data bolt database
*/

type BoltClient struct {
	db *bolt.DB // Bolt database
}

var ErrLimReached error = errors.New("Limit reached")
var ErrObjFound error = errors.New("Object name found")

// Return a pointer to the BoltClient
func newBoltClient(config DBConfiguration) (*BoltClient, error) {

	db, err := bolt.Open("./core-data.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte(EVENTS_COLLECTION))
		tx.CreateBucketIfNotExists([]byte(READINGS_COLLECTION))
		tx.CreateBucketIfNotExists([]byte(VALUE_DESCRIPTOR_COLLECTION))
		return nil
	})

	boltClient := &BoltClient{db: db}
	return boltClient, nil
}

func (bc *BoltClient) CloseSession() {
	bc.db.Close()
}

// ******************************* EVENTS **********************************

// Return all the events
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (bc *BoltClient) Events() ([]models.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		return true
	}, -1)
}

// Add a new event
// UnexpectedError - failed to add to database
// NoValueDescriptor - no existing value descriptor for a reading in the event
func (bc *BoltClient) AddEvent(e *models.Event) (bson.ObjectId, error) {
	e.Created = time.Now().UnixNano() / int64(time.Millisecond)
	e.ID = bson.NewObjectId()
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.Update(func(tx *bolt.Tx) error {

		// Insert readings
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		for i := range e.Readings {
			e.Readings[i].Id = bson.NewObjectId()
			e.Readings[i].Created = e.Created
			e.Readings[i].Device = e.Device
			encoded, err := json.Marshal(e.Readings[i])
			if err != nil {
				return err
			}
			err = b.Put([]byte(e.Readings[i].Id), encoded)
			if err != nil {
				return err
			}
		}

		// Add the event
		be := BoltEvent{Event: *e}
		b = tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		encoded, err := json.Marshal(be)
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
func (bc *BoltClient) UpdateEvent(e models.Event) error {
	e.Modified = time.Now().UnixNano() / int64(time.Millisecond)

	be := BoltEvent{Event: e}
	return bc.update(READINGS_COLLECTION, be, e.ID)
}

// Get an event by id
func (bc *BoltClient) EventById(id string) (models.Event, error) {
	ev := models.Event{}
	if !bson.IsObjectIdHex(id) {
		return ev, ErrInvalidObjectId
	}
	err := bc.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		encoded := b.Get([]byte(bson.ObjectIdHex(id)))
		ev, err = getEvent(encoded, tx)
		return err
	})

	return ev, err
}

// Get the number of events in bolt
func (bc *BoltClient) EventCount() (int, error) {
	return bc.count(EVENTS_COLLECTION)
}

// Get the number of events in bolt for the device
func (bc *BoltClient) EventCountByDeviceId(devid string) (int, error) {
	bstat := 0
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
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
	return bc.deleteById(id, EVENTS_COLLECTION)
}

// Get a list of events based on the device id and limit
func (bc *BoltClient) EventsForDeviceLimit(ide string, limit int) ([]models.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "device").ToString()
		if value == ide {
			return true
		}
		return false
	}, limit)
}

// Get a list of events based on the device id
func (bc *BoltClient) EventsForDevice(ide string) ([]models.Event, error) {
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
func (bc *BoltClient) EventsByCreationTime(startTime, endTime int64, limit int) ([]models.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "created").ToInt64()
		if (value >= startTime) && (value <= endTime) {
			return true
		}
		return false
	}, limit)
}

// Get Events that are older than the given age (defined by age = now - created)
func (bc *BoltClient) EventsOlderThanAge(age int64) ([]models.Event, error) {
	return bc.getEvents(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "created").ToInt64()
		value = (time.Now().UnixNano() / int64(time.Millisecond)) - value
		if value >= age {
			return true
		}
		return false
	}, -1)
}

// Get all of the events that have been pushed
func (bc *BoltClient) EventsPushed() ([]models.Event, error) {
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
	bc.scrubAll(EVENTS_COLLECTION)
	bc.scrubAll(READINGS_COLLECTION)
	return nil
}

// Get events for the passed check
func (bc *BoltClient) getEvents(fn func(encoded []byte) bool, limit int) ([]models.Event, error) {
	events := []models.Event{}

	// Check if limit is not 0
	if limit == 0 {
		return events, nil
	}
	cnt := 0

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				ev, err := getEvent(encoded, tx)
				if err != nil {
					return err
				}
				events = append(events, ev)
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

// Get a single event
func getEvent(encoded []byte, tx *bolt.Tx) (models.Event, error) {
	ev := models.Event{}
	b := tx.Bucket([]byte(READINGS_COLLECTION))
	if b == nil {
		return ev, ErrUnsupportedDatabase
	}
	var be BoltEvent
	if encoded == nil {
		return ev, ErrNotFound
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := json.Unmarshal(encoded, &be)
	if err != nil {
		return ev, err
	}

	for _, id := range be.Readings {
		encoded := b.Get([]byte(bson.ObjectIdHex(id)))
		if encoded == nil {
			return ev, ErrNotFound
		}
		var reading models.Reading
		err = json.Unmarshal(encoded, &reading)
		if err != nil {
			return ev, err
		}
		be.Event.Readings = append(be.Event.Readings, reading)
	}
	ev = be.Event
	return ev, nil
}

// ************************ READINGS ************************************

// Return a list of readings sorted by reading id
func (bc *BoltClient) Readings() ([]models.Reading, error) {
	return bc.getReadings(func(encoded []byte) bool {
		return true
	}, -1)
}

// Post a new reading
func (bc *BoltClient) AddReading(r models.Reading) (bson.ObjectId, error) {
	r.Id = bson.NewObjectId()
	r.Created = time.Now().UnixNano() / int64(time.Millisecond)

	err := bc.add(READINGS_COLLECTION, r, r.Id)
	return r.Id, err
}

// Update a reading
// 404 - reading cannot be found
// 409 - Value descriptor doesn't exist
// 503 - unknown issues
func (bc *BoltClient) UpdateReading(r models.Reading) error {
	r.Modified = time.Now().UnixNano() / int64(time.Millisecond)

	return bc.update(READINGS_COLLECTION, r, r.Id)
}

// Get a reading by ID
func (bc *BoltClient) ReadingById(id string) (models.Reading, error) {
	var reading models.Reading
	err := bc.getById(&reading, READINGS_COLLECTION, id)
	return reading, err
}

// Get the count of readings in BOLT
func (bc *BoltClient) ReadingCount() (int, error) {
	return bc.count(READINGS_COLLECTION)
}

// Delete a reading by ID
// 404 - can't find the reading with the given id
func (bc *BoltClient) DeleteReadingById(id string) error {
	return bc.deleteById(id, READINGS_COLLECTION)
}

// Return a list of readings for the given device (id or name)
// Sort the list of readings on creation date
func (bc *BoltClient) ReadingsByDevice(ids string, limit int) ([]models.Reading, error) {
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
func (bc *BoltClient) ReadingsByValueDescriptor(name string, limit int) ([]models.Reading, error) {
	return bc.getReadings(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "name").ToString()
		if value == name {
			return true
		}
		return false
	}, limit)
}

// Return a list of readings whose name is in the list of value descriptor names
func (bc *BoltClient) ReadingsByValueDescriptorNames(names []string, limit int) ([]models.Reading, error) {
	return bc.getReadings(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "name").ToString()
		for _, name := range names {
			if name == value {
				return true
			}
		}
		return false
	}, limit)
}

// Return a list of readings whos creation time is in-between start and end
// Limit by the limit parameter
func (bc *BoltClient) ReadingsByCreationTime(start, end int64, limit int) ([]models.Reading, error) {
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
func (bc *BoltClient) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]models.Reading, error) {
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
func (bc *BoltClient) getReadings(fn func(encoded []byte) bool, limit int) ([]models.Reading, error) {
	r := models.Reading{}
	rs := []models.Reading{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	// Check if limit is not 0
	if limit == 0 {
		return rs, nil
	}
	cnt := 0

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
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

// ************************* VALUE DESCRIPTORS *****************************

// Add a value descriptor
// 409 - Formatting is bad or it is not unique
// 503 - Unexpected
// TODO: Check for valid printf formatting
func (bc *BoltClient) AddValueDescriptor(v models.ValueDescriptor) (bson.ObjectId, error) {
	// See if the name is unique
	var dumy models.ValueDescriptor
	err := bc.getByName(&dumy, VALUE_DESCRIPTOR_COLLECTION, v.Name)
	if err == nil {
		return v.Id, ErrNotUnique
	}

	v.Id = bson.NewObjectId()
	v.Created = time.Now().UnixNano() / int64(time.Millisecond)

	// Add the value descriptor
	err = bc.add(VALUE_DESCRIPTOR_COLLECTION, v, v.Id)
	return v.Id, err
}

// Return a list of all the value descriptors
// 513 Service Unavailable - database problems
func (bc *BoltClient) ValueDescriptors() ([]models.ValueDescriptor, error) {
	return bc.getValueDescriptors(func(encoded []byte) bool {
		return true
	})
}

// Update a value descriptor
// First use the ID for identification, then the name
// TODO: Check for the valid printf formatting
// 404 not found if the value descriptor cannot be found by the identifiers
func (bc *BoltClient) UpdateValueDescriptor(v models.ValueDescriptor) error {
	// See if the name is unique if it changed
	var vd models.ValueDescriptor
	err := bc.getByName(&vd, VALUE_DESCRIPTOR_COLLECTION, v.Name)
	if err != ErrNotFound {
		if err != nil {
			return err
		}
		// IDs are different -> name not unique
		if vd.Id != v.Id {
			return ErrNotUnique
		}
	}
	v.Modified = time.Now().UnixNano() / int64(time.Millisecond)

	return bc.update(VALUE_DESCRIPTOR_COLLECTION, v, v.Id)
}

// Delete the value descriptor based on the id
// Not found error if there isn't a value descriptor for the ID
// ValueDescriptorStillInUse if the value descriptor is still referenced by readings
func (bc *BoltClient) DeleteValueDescriptorById(id string) error {
	return bc.deleteById(id, VALUE_DESCRIPTOR_COLLECTION)
}

// Return a value descriptor based on the name
// Can return null if no value descriptor is found
func (bc *BoltClient) ValueDescriptorByName(name string) (models.ValueDescriptor, error) {
	var vd models.ValueDescriptor
	err := bc.getByName(&vd, VALUE_DESCRIPTOR_COLLECTION, name)
	return vd, err
}

// Return all of the value descriptors based on the names
func (bc *BoltClient) ValueDescriptorsByName(names []string) ([]models.ValueDescriptor, error) {
	return bc.getValueDescriptors(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "name").ToString()
		for _, name := range names {
			if name == value {
				return true
			}
		}
		return false
	})
}

// Return a value descriptor based on the id
// Return NotFoundError if there is no value descriptor for the id
func (bc *BoltClient) ValueDescriptorById(id string) (models.ValueDescriptor, error) {
	var vd models.ValueDescriptor
	err := bc.getById(&vd, VALUE_DESCRIPTOR_COLLECTION, id)
	return vd, err
}

// Return all the value descriptors that match the UOM label
func (bc *BoltClient) ValueDescriptorsByUomLabel(uomLabel string) ([]models.ValueDescriptor, error) {
	return bc.getValueDescriptors(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "uomLabel").ToString()
		if value == uomLabel {
			return true
		}
		return false
	})
}

// Return value descriptors based on if it has the label
func (bc *BoltClient) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	return bc.getValueDescriptors(func(encoded []byte) bool {
		labels := jsoniter.Get(encoded, "labels").GetInterface().([]interface{})
		for _, value := range labels {
			if label == value.(string) {
				return true
			}
		}
		return false
	})
}

// Return value descriptors based on the type
func (bc *BoltClient) ValueDescriptorsByType(t string) ([]models.ValueDescriptor, error) {
	return bc.getValueDescriptors(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "type").ToString()
		if value == t {
			return true
		}
		return false
	})
}

// Delete all value descriptors
func (bc *BoltClient) ScrubAllValueDescriptors() error {
	return bc.scrubAll(VALUE_DESCRIPTOR_COLLECTION)
}

// Get value descriptors for the passed check
func (bc *BoltClient) getValueDescriptors(fn func(encoded []byte) bool) ([]models.ValueDescriptor, error) {
	vd := models.ValueDescriptor{}
	vds := []models.ValueDescriptor{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(VALUE_DESCRIPTOR_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &vd)
				if err != nil {
					return err
				}
				vds = append(vds, vd)
			}
			return nil
		})
		return err
	})
	return vds, err
}

// Add an element
func (bc *BoltClient) add(bucket string, element interface{}, id bson.ObjectId) error {
	return bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		encoded, err := json.Marshal(element)
		if err != nil {
			return err
		}
		return b.Put([]byte(id), encoded)
	})
}

// Update an element
func (bc *BoltClient) update(bucket string, element interface{}, id bson.ObjectId) error {
	return bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		if b.Get([]byte(id)) == nil {
			return ErrNotFound
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		encoded, err := json.Marshal(element)
		if err != nil {
			return err
		}
		return b.Put([]byte(id), encoded)
	})
}

// Delete from the collection based on ID
func (bc *BoltClient) deleteById(id string, col string) error {
	// Check if id is a hexstring
	if !bson.IsObjectIdHex(id) {
		return ErrInvalidObjectId
	}
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(col))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		return b.Delete([]byte(bson.ObjectIdHex(id)))
	})
	return err
}

// Get an element by ID
func (bc *BoltClient) getById(v interface{}, c string, gid string) error {
	// Check if id is a hexstring
	if !bson.IsObjectIdHex(gid) {
		return ErrInvalidObjectId
	}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		encoded := b.Get([]byte(bson.ObjectIdHex(gid)))
		if encoded == nil {
			return ErrNotFound
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := json.Unmarshal(encoded, v)
		return err
	})
	return err
}

// Get an element by name
func (bc *BoltClient) getByName(v interface{}, c string, gn string) error {
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "name").ToString()
			if value == gn {
				json := jsoniter.ConfigCompatibleWithStandardLibrary
				err := json.Unmarshal(encoded, &v)
				if err != nil {
					return err
				}
				return ErrObjFound
			}
			return nil
		})
		if err == nil {
			return ErrNotFound
		} else if err == ErrObjFound {
			return nil
		}
		return err
	})
	return err
}

// Count number of elements
func (bc *BoltClient) count(bucket string) (int, error) {
	var bstat int
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		bstat = b.Stats().KeyN
		return nil
	})
	return bstat, err
}

// Delete all elements in selected bucket
func (bc *BoltClient) scrubAll(bucket string) error {
	return bc.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte(bucket))
		tx.CreateBucketIfNotExists([]byte(bucket))
		return nil
	})
}
