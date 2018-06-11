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
	"encoding/hex"
	"errors"
	"time"

	"github.com/coreos/bbolt"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

var currentBoltClient *BoltClient // Singleton used so that BoltEvent can use it to de-reference readings
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

// Get the current Bolt Client
func getCurrentBoltClient() (*BoltClient, error) {
	if currentBoltClient == nil {
		return nil, errors.New("No current bolt client, please create a new client before requesting it")
	}

	return currentBoltClient, nil
}

func (bc *BoltClient) CloseSession() {
	bc.db.Close()
}

// ******************************* EVENTS **********************************

// Return all the events
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (bc *BoltClient) Events() ([]models.Event, error) {
	events := []models.Event{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		br := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var be BoltEvent
			err := json.Unmarshal(encoded, &be)
			if err != nil {
				return err
			}
			for _, id := range be.Readings {
				encoded := br.Get([]byte(bson.ObjectIdHex(id)))
				if encoded == nil {
					return ErrNotFound
				}
				var reading models.Reading
				err = json.Unmarshal(encoded, &reading)
				if err != nil {
					return err
				}
				be.Event.Readings = append(be.Event.Readings, reading)
			}
			events = append(events, be.Event)
			return nil
		})
		return err
	})
	return events, err
}

// Add a new event
// UnexpectedError - failed to add to database
// NoValueDescriptor - no existing value descriptor for a reading in the event
func (bc *BoltClient) AddEvent(e *models.Event) (bson.ObjectId, error) {
	e.Created = time.Now().UnixNano() / int64(time.Millisecond)
	e.ID = bson.NewObjectId()
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		for i := range e.Readings {
			e.Readings[i].Id = bson.NewObjectId()
			e.Readings[i].Created = time.Now().UnixNano() / int64(time.Millisecond)
			encoded, err := json.Marshal(e.Readings[i])
			if err != nil {
				return err
			}
			err = b.Put([]byte(e.Readings[i].Id), encoded)
			if err != nil {
				return err
			}
		}
		// Handle DB references
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
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		// Handle DB references
		be := BoltEvent{Event: e}
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
	return err
}

// Get an event by id
func (bc *BoltClient) EventById(id string) (models.Event, error) {
	ev := models.Event{}
	if !bson.IsObjectIdHex(id) {
		return ev, ErrInvalidObjectId
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		var be BoltEvent
		encoded := b.Get([]byte(bson.ObjectIdHex(id)))
		if encoded == nil {
			return ErrNotFound
		}
		err := json.Unmarshal(encoded, &be)
		if err != nil {
			return err
		}

		br := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		for _, id := range be.Readings {
			encoded := br.Get([]byte(bson.ObjectIdHex(id)))
			if encoded == nil {
				return ErrNotFound
			}
			var reading models.Reading
			err = json.Unmarshal(encoded, &reading)
			if err != nil {
				return err
			}
			be.Event.Readings = append(be.Event.Readings, reading)
		}
		ev = be.Event
		return nil
	})

	return ev, err
}

// Get the number of events in bolt
func (bc *BoltClient) EventCount() (int, error) {
	var bstat int
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		bstat = b.Stats().KeyN
		return nil
	})
	return bstat, err
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
				return nil
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
	if !bson.IsObjectIdHex(id) {
		return ErrInvalidObjectId
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		var be BoltEvent
		encoded := b.Get([]byte(bson.ObjectIdHex(id)))
		if encoded == nil {
			return ErrNotFound
		}
		err := json.Unmarshal(encoded, &be)
		if err != nil {
			return err
		}

		br := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		for _, id := range be.Readings {
			ret := br.Delete([]byte(bson.ObjectIdHex(id)))
			if ret != nil {
				return ret
			}
		}
		ret := b.Delete([]byte(bson.ObjectIdHex(id)))
		if ret != nil {
			return ret
		}
		return nil
	})
	return err
}

// Get a list of events based on the device id and limit
func (bc *BoltClient) EventsForDeviceLimit(ide string, limit int) ([]models.Event, error) {
	evs := []models.Event{}
	// Check if limit is 0
	if limit == 0 {
		return evs, nil
	}
	cnt := 0
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "device").ToString()
			if value == ide {
				hexid := hex.EncodeToString([]byte(id))
				ev, err := bc.EventById(hexid)
				if err != nil {
					return err
				}
				evs = append(evs, ev)
				cnt++
				if cnt >= limit {
					return ErrLimReached
				}
			}
			return nil
		})
		if err == ErrLimReached {
			return nil
		}
		return err
	})
	return evs, err
}

// Get a list of events based on the device id
func (bc *BoltClient) EventsForDevice(ide string) ([]models.Event, error) {
	evs := []models.Event{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "device").ToString()
			if value == ide {
				hexid := hex.EncodeToString([]byte(id))
				ev, err := bc.EventById(hexid)
				if err != nil {
					return err
				}
				evs = append(evs, ev)
			}
			return nil
		})
		if evs == nil {
			return ErrNotFound
		} else {
			return nil
		}
		return err
	})
	return evs, err
}

// Return a list of events whos creation time is between startTime and endTime
// Limit the number of results by limit
func (bc *BoltClient) EventsByCreationTime(startTime, endTime int64, limit int) ([]models.Event, error) {
	evs := []models.Event{}
	// Check if limit is 0
	if limit == 0 {
		return evs, nil
	}
	cnt := 0
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "created").ToInt64()
			if (value >= int64(startTime)) && (value <= endTime) {
				hexid := hex.EncodeToString([]byte(id))
				ev, err := bc.EventById(hexid)
				if err != nil {
					return err
				}
				evs = append(evs, ev)
				cnt++
				if cnt >= limit {
					return ErrLimReached
				}
			}
			return nil
		})
		if err == ErrLimReached {
			return nil
		}
		return err
	})
	return evs, err
}

// Get Events that are older than the given age (defined by age = now - created)
func (bc *BoltClient) EventsOlderThanAge(age int64) ([]models.Event, error) {
	evs := []models.Event{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "created").ToInt64()
			value = (time.Now().UnixNano() / int64(time.Millisecond)) - value
			if value >= age {
				hexid := hex.EncodeToString([]byte(id))
				ev, err := bc.EventById(hexid)
				if err != nil {
					return err
				}
				evs = append(evs, ev)
			}
			return nil
		})
		if evs == nil {
			return ErrNotFound
		} else {
			return nil
		}
		return err
	})
	return evs, err
}

// Get all of the events that have been pushed
func (bc *BoltClient) EventsPushed() ([]models.Event, error) {
	evs := []models.Event{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			ev := models.Event{}
			value := jsoniter.Get(encoded, "pushed").ToInt64()
			if value != 0 {
				err := json.Unmarshal(encoded, &ev)
				if err != nil {
					return err
				}
				evs = append(evs, ev)
			}
			return nil
		})
		if evs == nil {
			return ErrNotFound
		} else {
			return nil
		}
		return err
	})
	return evs, err
}

// Delete all of the readings and all of the events
func (bc *BoltClient) ScrubAllEvents() error {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte(EVENTS_COLLECTION))
		tx.DeleteBucket([]byte(READINGS_COLLECTION))
		tx.CreateBucketIfNotExists([]byte(EVENTS_COLLECTION))
		tx.CreateBucketIfNotExists([]byte(READINGS_COLLECTION))
		return nil
	})

	return err
}

// ************************ READINGS ************************************

// Return a list of readings sorted by reading id
func (bc *BoltClient) Readings() ([]models.Reading, error) {
	readings := []models.Reading{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			var reading models.Reading
			err := json.Unmarshal(encoded, &reading)
			if err != nil {
				return err
			}
			readings = append(readings, reading)
			return nil
		})
		return err
	})
	return readings, err
}

// Post a new reading
func (bc *BoltClient) AddReading(r models.Reading) (bson.ObjectId, error) {
	r.Id = bson.NewObjectId()
	r.Created = time.Now().UnixNano() / int64(time.Millisecond)
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		encoded, err := json.Marshal(r)
		if err != nil {
			return err
		}
		return b.Put([]byte(r.Id), encoded)
	})
	return r.Id, err
}

// Update a reading
// 404 - reading cannot be found
// 409 - Value descriptor doesn't exist
// 503 - unknown issues
func (bc *BoltClient) UpdateReading(r models.Reading) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	r.Modified = time.Now().UnixNano() / int64(time.Millisecond)
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		encoded, err := json.Marshal(r)
		if err != nil {
			return err
		}
		return b.Put([]byte(r.Id), encoded)
	})
	return err
}

// Get a reading by ID
func (bc *BoltClient) ReadingById(id string) (models.Reading, error) {
	var reading models.Reading
	err := bc.getById(&reading, READINGS_COLLECTION, id)
	return reading, err
}

// Get the count of readings in BOLT
func (bc *BoltClient) ReadingCount() (int, error) {
	var bstat int
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		bstat = b.Stats().KeyN
		return nil
	})
	return bstat, err
}

// Delete a reading by ID
// 404 - can't find the reading with the given id
func (bc *BoltClient) DeleteReadingById(id string) error {
	return bc.deleteById(id, READINGS_COLLECTION)
}

// Return a list of readings for the given device (id or name)
// Sort the list of readings on creation date
func (bc *BoltClient) ReadingsByDevice(ids string, limit int) ([]models.Reading, error) {
	rs := []models.Reading{}
	// Check if limit is 0
	if limit == 0 {
		return rs, nil
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	cnt := 0
	r := models.Reading{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "device").ToString()
			if value == ids {
				err := json.Unmarshal(encoded, &r)
				if err != nil {
					return err
				}
				rs = append(rs, r)
				cnt++
				if cnt >= limit {
					return ErrLimReached
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

// Return a list of readings for the given value descriptor
// Limit by the given limit
func (bc *BoltClient) ReadingsByValueDescriptor(name string, limit int) ([]models.Reading, error) {
	rs := []models.Reading{}
	// Check if limit is 0
	if limit == 0 {
		return rs, nil
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	cnt := 0
	r := models.Reading{}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "name").ToString()
			if value == name {
				err := json.Unmarshal(encoded, &r)
				if err != nil {
					return err
				}
				rs = append(rs, r)
				cnt++
				if cnt >= limit {
					return ErrLimReached
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

// Return a list of readings whose name is in the list of value descriptor names
func (bc *BoltClient) ReadingsByValueDescriptorNames(names []string, limit int) ([]models.Reading, error) {
	rList := []models.Reading{}
	// Check if limit is 0
	if limit == 0 {
		return rList, nil
	}
	cnt := limit
	for _, name := range names {
		rs, err := bc.ReadingsByValueDescriptor(name, cnt)
		if err != nil {
			return rList, err
		}
		for _, r := range rs {
			rList = append(rList, r)
		}
		cnt = cnt - len(rs)
	}
	return rList, nil
}

// Return a list of readings whos creation time is in-between start and end
// Limit by the limit parameter
func (bc *BoltClient) ReadingsByCreationTime(start, end int64, limit int) ([]models.Reading, error) {
	rs := []models.Reading{}
	// Check if limit is 0
	if limit == 0 {
		return rs, nil
	}
	var cnt int
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			r := models.Reading{}
			value := jsoniter.Get(encoded, "created").ToInt64()
			if (value >= int64(start)) && (value <= end) {
				err := json.Unmarshal(encoded, &r)
				if err != nil {
					return err
				}
				rs = append(rs, r)
				cnt++
				if cnt >= limit {
					return ErrLimReached
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

// Return a list of readings for a device filtered by the value descriptor and limited by the limit
// The readings are linked to the device through an event
func (bc *BoltClient) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]models.Reading, error) {
	rs := []models.Reading{}
	// Check if limit is 0
	if limit == 0 {
		return rs, nil
	}
	var cnt int
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(READINGS_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			r := models.Reading{}
			valuedev := jsoniter.Get(encoded, "device").ToString()
			valuename := jsoniter.Get(encoded, "name").ToString()
			if (valuename == valueDescriptor) || (valuedev == deviceId) {

				err := json.Unmarshal(encoded, &r)
				if err != nil {
					return err
				}
				rs = append(rs, r)
				cnt++
				if cnt >= limit {
					return ErrLimReached
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
	var dumy models.ValueDescriptor
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.getByName(&dumy, VALUE_DESCRIPTOR_COLLECTION, v.Name)
	if err == nil {
		return v.Id, ErrNotUnique
	}

	v.Id = bson.NewObjectId()
	v.Created = time.Now().UnixNano() / int64(time.Millisecond)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(VALUE_DESCRIPTOR_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return b.Put([]byte(v.Id), encoded)
	})
	return v.Id, err
}

// Return a list of all the value descriptors
// 513 Service Unavailable - database problems
func (bc *BoltClient) ValueDescriptors() ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(VALUE_DESCRIPTOR_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}

		err := b.ForEach(func(id, encoded []byte) error {
			var vd models.ValueDescriptor
			err := json.Unmarshal(encoded, &vd)
			if err != nil {
				return err
			}
			vds = append(vds, vd)
			return nil
		})
		return err
	})
	return vds, err
}

// Update a value descriptor
// First use the ID for identification, then the name
// TODO: Check for the valid printf formatting
// 404 not found if the value descriptor cannot be found by the identifiers
func (bc *BoltClient) UpdateValueDescriptor(v models.ValueDescriptor) error {
	// See if the name is unique if it changed
	var vd models.ValueDescriptor
	json := jsoniter.ConfigCompatibleWithStandardLibrary
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

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(VALUE_DESCRIPTOR_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return b.Put([]byte(v.Id), encoded)
	})
	return err
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
	vList := []models.ValueDescriptor{}

	for _, name := range names {
		v, err := bc.ValueDescriptorByName(name)
		if err != nil {
			return []models.ValueDescriptor{}, err
		}
		vList = append(vList, v)
	}
	return vList, nil
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
	vds := []models.ValueDescriptor{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(VALUE_DESCRIPTOR_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			vd := models.ValueDescriptor{}
			value := jsoniter.Get(encoded, "uomLabel").ToString()
			if value == uomLabel {
				err := json.Unmarshal(encoded, &vd)
				if err != nil {
					return err
				}
				vds = append(vds, vd)
			}
			return nil
		})
		if vds == nil {
			return ErrNotFound
		} else {
			return nil
		}
		return err
	})
	return vds, err
}

// Return value descriptors based on if it has the label
func (bc *BoltClient) ValueDescriptorsByLabel(label string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(VALUE_DESCRIPTOR_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			vd := models.ValueDescriptor{}
			err := json.Unmarshal(encoded, &vd)
			if err != nil {
				return err
			}
			for i, _ := range vd.Labels {
				value := jsoniter.Get(encoded, "labels", i).ToString()
				if value == label {
					vds = append(vds, vd)
				}
			}
			return nil
		})
		if vds == nil {
			return ErrNotFound
		} else {
			return nil
		}
		return err
	})
	return vds, err
}

// Return value descriptors based on the type
func (bc *BoltClient) ValueDescriptorsByType(t string) ([]models.ValueDescriptor, error) {
	vds := []models.ValueDescriptor{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(VALUE_DESCRIPTOR_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			vd := models.ValueDescriptor{}
			value := jsoniter.Get(encoded, "type").ToString()
			if value == t {
				err := json.Unmarshal(encoded, &vd)
				if err != nil {
					return err
				}
				vds = append(vds, vd)
			}
			return nil
		})
		if vds == nil {
			return ErrNotFound
		} else {
			return nil
		}
		return err
	})
	return vds, err
}

// Delete all value descriptors
func (bc *BoltClient) ScrubAllValueDescriptors() error {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(VALUE_DESCRIPTOR_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			err := bc.deleteById(string(id), VALUE_DESCRIPTOR_COLLECTION)
			if err != nil {
				return err
			}
			return nil
		})
		return err
	})
	return err
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

func (bc *BoltClient) getByName(v interface{}, c string, gn string) error {
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "name").ToString()
			if value == gn {
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
