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
package mongo

import (
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/mongo/models"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

/*
Core data client
Has functions for interacting with the core data mongo database
*/

// ******************************* EVENTS **********************************

// Return all the events
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (mc MongoClient) Events() ([]contract.Event, error) {
	return mc.mapEvents(mc.getEvents(bson.M{}))

}

// Return events up to the max number specified
// UnexpectedError - failed to retrieve events from the database
// Sort the events in descending order by ID
func (mc MongoClient) EventsWithLimit(limit int) ([]contract.Event, error) {
	return mc.mapEvents(mc.getEventsLimit(bson.M{}, limit))
}

// Add a new event
// UnexpectedError - failed to add to database
// NoValueDescriptor - no existing value descriptor for a reading in the event
func (mc MongoClient) AddEvent(e contract.Event) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	//Add the readings
	if len(e.Readings) > 0 {
		var ui []interface{}
		for i, reading := range e.Readings {
			var r models.Reading
			id, err := r.FromContract(reading)
			if err != nil {
				return "", err
			}

			r.TimestampForAdd()

			ui = append(ui, r)

			e.Readings[i].Id = id
		}
		if err := s.DB(mc.database.Name).C(db.ReadingsCollection).Insert(ui...); err != nil {
			return "", errorMap(err)
		}
	}

	var mapped models.Event
	id, err := mapped.FromContract(e, mc)
	if err != nil {
		return "", err
	}

	mapped.TimestampForAdd()

	if err = s.DB(mc.database.Name).C(db.EventsCollection).Insert(mapped); err != nil {
		return "", errorMap(err)
	}
	return id, nil
}

// Update an event - do NOT update readings
// UnexpectedError - problem updating in database
// NotFound - no event with the ID was found
func (mc MongoClient) UpdateEvent(e contract.Event) error {
	var mapped models.Event
	id, err := mapped.FromContract(e, mc)
	if err != nil {
		return err
	}

	mapped.TimestampForUpdate()

	return mc.updateId(db.EventsCollection, id, mapped)
}

// Get an event by id
func (mc MongoClient) EventById(id string) (contract.Event, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	query, err := idToBsonM(id)
	if err != nil {
		return contract.Event{}, err
	}

	var evt models.Event
	if err := s.DB(mc.database.Name).C(db.EventsCollection).Find(query).One(&evt); err != nil {
		return contract.Event{}, errorMap(err)
	}
	return evt.ToContract(mc)
}

// Get the number of events in Mongo
func (mc MongoClient) EventCount() (int, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	count, err := s.DB(mc.database.Name).C(db.EventsCollection).Find(nil).Count()
	if err != nil {
		return 0, errorMap(err)
	}
	return count, nil
}

// Get the number of events in Mongo for the device
func (mc MongoClient) EventCountByDeviceId(id string) (int, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	query := bson.M{"device": id}
	count, err := s.DB(mc.database.Name).C(db.EventsCollection).Find(query).Count()
	if err != nil {
		return 0, errorMap(err)
	}
	return count, nil
}

// Delete an event by ID and all of its readings
// 404 - Event not found
// 503 - Unexpected problems
func (mc MongoClient) DeleteEventById(id string) error {
	return mc.deleteById(db.EventsCollection, id)
}

// Get a list of events based on the device id and limit
func (mc MongoClient) EventsForDeviceLimit(id string, limit int) ([]contract.Event, error) {
	return mc.mapEvents(mc.getEventsLimit(bson.M{"device": id}, limit))
}

// Get a list of events based on the device id
func (mc MongoClient) EventsForDevice(id string) ([]contract.Event, error) {
	return mc.mapEvents(mc.getEvents(bson.M{"device": id}))
}

// Return a list of events whose creation time is between startTime and endTime
// Limit the number of results by limit
func (mc MongoClient) EventsByCreationTime(startTime, endTime int64, limit int) ([]contract.Event, error) {
	query := bson.M{"created": bson.M{
		"$gte": startTime,
		"$lte": endTime,
	}}
	return mc.mapEvents(mc.getEventsLimit(query, limit))
}

// Get Events that are older than the given age (defined by age = now - created)
func (mc MongoClient) EventsOlderThanAge(age int64) ([]contract.Event, error) {
	expireDate := (db.MakeTimestamp()) - age
	return mc.mapEvents(mc.getEvents(bson.M{"created": bson.M{"$lt": expireDate}}))
}

// Get all of the events that have been pushed
func (mc MongoClient) EventsPushed() ([]contract.Event, error) {
	return mc.mapEvents(mc.getEvents(bson.M{"pushed": bson.M{"$gt": int64(0)}}))
}

// Delete all of the readings and all of the events
func (mc MongoClient) ScrubAllEvents() error {
	s := mc.getSessionCopy()
	defer s.Close()

	_, err := s.DB(mc.database.Name).C(db.ReadingsCollection).RemoveAll(nil)
	if err != nil {
		return errorMap(err)
	}

	_, err = s.DB(mc.database.Name).C(db.EventsCollection).RemoveAll(nil)
	if err != nil {
		return errorMap(err)
	}

	return nil
}

// Get events for the passed query
func (mc MongoClient) getEvents(q bson.M) (me []models.Event, err error) {
	s := mc.getSessionCopy()
	defer s.Close()

	err = s.DB(mc.database.Name).C(db.EventsCollection).Find(q).All(&me)
	if err != nil {
		return []models.Event{}, errorMap(err)
	}
	return
}

// Get events with a limit
// Sort the list before applying the limit so we can return the most recent events
func (mc MongoClient) getEventsLimit(q bson.M, limit int) (me []models.Event, err error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Check if limit is 0
	if limit == 0 {
		return []models.Event{}, nil
	}

	err = s.DB(mc.database.Name).C(db.EventsCollection).Find(q).Sort("-modified").Limit(limit).All(&me)
	if err != nil {
		return []models.Event{}, errorMap(err)
	}
	return
}

// ************************ READINGS ************************************8

func (mc MongoClient) DBRefToReading(dbRef mgo.DBRef) (a models.Reading, err error) {
	if err = mc.database.C(db.ReadingsCollection).Find(bson.M{"_id": dbRef.Id}).One(&a); err != nil {
		return models.Reading{}, errorMap(err)
	}
	return
}

func (mc MongoClient) ReadingToDBRef(r models.Reading) (dbRef mgo.DBRef, err error) {
	s := mc.session.Copy()
	defer s.Close()

	// validate identity provided in contract actually exists and populate missing Id, Uuid field
	var reading models.Reading
	if r.Id.Valid() {
		reading, err = mc.readingById(r.Id.Hex())
	} else {
		reading, err = mc.readingById(r.Uuid)
	}
	if err != nil {
		return dbRef, err
	}

	dbRef = mgo.DBRef{Collection: db.ReadingsCollection, Id: reading.Id}
	return
}

// Return a list of readings sorted by reading id
func (mc MongoClient) Readings() ([]contract.Reading, error) {
	readings, err := mc.getReadings(nil)
	if err != nil {
		return []contract.Reading{}, err
	}

	mapped := make([]contract.Reading, 0)
	for _, r := range readings {
		mapped = append(mapped, r.ToContract())
	}
	return mapped, nil
}

// Post a new reading
func (mc MongoClient) AddReading(r contract.Reading) (string, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var mapped models.Reading
	id, err := mapped.FromContract(r)
	if err != nil {
		return "", err
	}

	mapped.TimestampForAdd()

	err = s.DB(mc.database.Name).C(db.ReadingsCollection).Insert(&mapped)
	if err != nil {
		return "", errorMap(err)
	}
	return id, err
}

// Update a reading
// 404 - reading cannot be found
// 409 - Value descriptor doesn't exist
// 503 - unknown issues
func (mc MongoClient) UpdateReading(r contract.Reading) error {
	var mapped models.Reading
	id, err := mapped.FromContract(r)
	if err != nil {
		return err
	}

	mapped.TimestampForUpdate()

	return mc.updateId(db.ReadingsCollection, id, mapped)
}

// Get a reading by ID
func (mc MongoClient) ReadingById(id string) (contract.Reading, error) {
	res, err := mc.readingById(id)
	if err != nil {
		return contract.Reading{}, err
	}
	return res.ToContract(), nil
}

func (mc MongoClient) readingById(id string) (models.Reading, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	query, err := idToBsonM(id)
	if err != nil {
		return models.Reading{}, err
	}

	var res models.Reading
	if err = s.DB(mc.database.Name).C(db.ReadingsCollection).Find(query).One(&res); err != nil {
		return models.Reading{}, errorMap(err)
	}
	return res, nil
}

// Get the count of readings in Mongo
func (mc MongoClient) ReadingCount() (int, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	count, err := s.DB(mc.database.Name).C(db.ReadingsCollection).Find(bson.M{}).Count()
	if err != nil {
		return 0, errorMap(err)
	}
	return count, nil
}

// Delete a reading by ID
// 404 - can't find the reading with the given id
func (mc MongoClient) DeleteReadingById(id string) error {
	return mc.deleteById(db.ReadingsCollection, id)
}

// Return a list of readings for the given device (id or name)
// Sort the list of readings on creation date
func (mc MongoClient) ReadingsByDevice(id string, limit int) ([]contract.Reading, error) {
	return mapReadings(mc.getReadingsLimit(bson.M{"device": id}, limit))
}

// Return a list of readings for the given value descriptor
// Limit by the given limit
func (mc MongoClient) ReadingsByValueDescriptor(name string, limit int) ([]contract.Reading, error) {
	return mapReadings(mc.getReadingsLimit(bson.M{"name": name}, limit))
}

// Return a list of readings whose creation time is in-between start and end
// Limit by the limit parameter
func (mc MongoClient) ReadingsByCreationTime(start, end int64, limit int) ([]contract.Reading, error) {
	return mapReadings(mc.getReadingsLimit(bson.M{"created": bson.M{"$gte": start, "$lte": end}}, limit))
}

// Return a list of readings for a device filtered by the value descriptor and limited by the limit
// The readings are linked to the device through an event
func (mc MongoClient) ReadingsByDeviceAndValueDescriptor(deviceId, valueDescriptor string, limit int) ([]contract.Reading, error) {
	return mapReadings(mc.getReadingsLimit(bson.M{"device": deviceId, "name": valueDescriptor}, limit))
}

// Return a list of readings that match.
// Sort the list before applying the limit so we can return the most recent readings
func (mc MongoClient) getReadingsLimit(q bson.M, limit int) ([]models.Reading, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	// Check if limit is 0
	if limit == 0 {
		return []models.Reading{}, nil
	}

	var readings []models.Reading
	if err := s.DB(mc.database.Name).C(db.ReadingsCollection).Find(q).Sort("-modified").Limit(limit).All(&readings); err != nil {
		return []models.Reading{}, errorMap(err)
	}
	return readings, nil
}

// Get readings from the database
func (mc MongoClient) getReadings(q bson.M) ([]models.Reading, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var readings []models.Reading
	if err := s.DB(mc.database.Name).C(db.ReadingsCollection).Find(q).All(&readings); err != nil {
		return []models.Reading{}, errorMap(err)
	}
	return readings, nil
}

// Get a reading from the database with the passed query
func (mc MongoClient) getReading(q bson.M) (models.Reading, error) {
	s := mc.getSessionCopy()
	defer s.Close()

	var res models.Reading
	if err := s.DB(mc.database.Name).C(db.ReadingsCollection).Find(q).One(&res); err != nil {
		return models.Reading{}, errorMap(err)
	}
	return res, nil
}

func (mc MongoClient) mapEvents(events []models.Event, errIn error) (ce []contract.Event, err error) {
	if errIn != nil {
		return []contract.Event{}, errIn
	}

	ce = []contract.Event{}
	for _, event := range events {
		contractEvent, err := event.ToContract(mc)
		if err != nil {
			return []contract.Event{}, err
		}
		ce = append(ce, contractEvent)
	}

	return
}

func mapReadings(readings []models.Reading, err error) ([]contract.Reading, error) {
	if err != nil {
		return []contract.Reading{}, err
	}

	mapped := make([]contract.Reading, 0)
	for _, r := range readings {
		mapped = append(mapped, r.ToContract())
	}
	return mapped, nil
}
