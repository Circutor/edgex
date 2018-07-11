/*******************************************************************************
 * Copyright 2018 Dell Inc.
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

	bolt "github.com/coreos/bbolt"
	"github.com/edgexfoundry/edgex-go/export"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

const (
	ExportCollection = "export"
)

/*
Export client client
Has functions for interacting with the export client bolt database
*/

type BoltClient struct {
	db *bolt.DB // Bolt database
}

// Return a pointer to the BoltClient
func newBoltClient(config DBConfiguration) (*BoltClient, error) {

	db, err := bolt.Open(config.DatabaseName, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte(ExportCollection))
		return nil
	})

	boltClient := &BoltClient{db: db}
	return boltClient, nil
}

// Close current session
func (bc *BoltClient) CloseSession() {
	bc.db.Close()
}

// ****************************** REGISTRATIONS ********************************

// Return all the registrations
// UnexpectedError - failed to retrieve registrations from the database
func (bc *BoltClient) Registrations() ([]export.Registration, error) {
	regs := []export.Registration{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ExportCollection))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var reg export.Registration
			err := json.Unmarshal(encoded, &reg)
			if err != nil {
				return err
			}
			regs = append(regs, reg)
			return nil
		})
		return err
	})
	return regs, err
}

// Add a new registration
// UnexpectedError - failed to add to database
func (bc *BoltClient) AddRegistration(reg *export.Registration) (bson.ObjectId, error) {
	reg.ID = bson.NewObjectId()
	reg.Created = time.Now().UnixNano() / int64(time.Millisecond)

	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ExportCollection))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		encoded, err := json.Marshal(reg)
		if err != nil {
			return err
		}
		return b.Put([]byte(reg.ID), encoded)
	})

	return reg.ID, err
}

// Update a registration
// UnexpectedError - problem updating in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) UpdateRegistration(reg export.Registration) error {

	reg.Modified = time.Now().UnixNano() / int64(time.Millisecond)

	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ExportCollection))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		if b.Get([]byte(reg.ID)) == nil {
			return ErrNotFound
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		encoded, err := json.Marshal(reg)
		if err != nil {
			return err
		}
		return b.Put([]byte(reg.ID), encoded)
	})

	return err
}

// Get a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) RegistrationById(id string) (export.Registration, error) {
	if !bson.IsObjectIdHex(id) {
		return export.Registration{}, ErrInvalidObjectId
	}

	reg := export.Registration{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ExportCollection))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		encoded := b.Get([]byte(bson.ObjectIdHex(id)))
		if encoded == nil {
			return ErrNotFound
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		return json.Unmarshal(encoded, &reg)
	})
	return reg, err
}

// Get a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the name was found
func (bc *BoltClient) RegistrationByName(name string) (export.Registration, error) {
	reg := export.Registration{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ExportCollection))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "name").ToString()
			if value == name {
				json := jsoniter.ConfigCompatibleWithStandardLibrary
				err := json.Unmarshal(encoded, &reg)
				if err != nil {
					return err
				}
				return errors.New("Object name found")
			}
			return nil
		})
		if err == nil {
			return ErrNotFound
		} else if err.Error() == "Object name found" {
			return nil
		}
		return err
	})
	return reg, err
}

// Delete a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) DeleteRegistrationById(id string) error {
	if !bson.IsObjectIdHex(id) {
		return ErrInvalidObjectId
	}

	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ExportCollection))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		return b.Delete([]byte(bson.ObjectIdHex(id)))
	})

	return err
}

// Delete a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) DeleteRegistrationByName(name string) error {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ExportCollection))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "name").ToString()
			if value == name {
				err := b.Delete([]byte(id))
				if err != nil {
					return ErrNotFound
				}
				err = errors.New("Object name found")
				return err
			}
			return nil
		})
		if err == nil {
			return ErrNotFound
		} else if err.Error() == "Object name found" {
			return nil
		}
		return err
	})
	return err
}

// Delete all registrations
func (bc *BoltClient) ScrubAllRegistrations() error {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte(ExportCollection))
		tx.CreateBucketIfNotExists([]byte(ExportCollection))
		return nil
	})

	return err
}
