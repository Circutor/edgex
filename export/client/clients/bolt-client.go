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
	REGISTRATION_COLLECTION = "registration"
)

/*
Export client client
Has functions for interacting with the export client bolt database
*/

type BoltClient struct {
	db *bolt.DB // Bolt database
}

// Return a pointer to the MongoClient
func newBoltClient(config DBConfiguration) (*BoltClient, error) {

	db, err := bolt.Open("./export-client.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte(REGISTRATION_COLLECTION))
		return nil
	})

	boltClient := &BoltClient{db: db}
	return boltClient, nil
}

func (mc *BoltClient) CloseSession() {
	mc.db.Close()
}

// ****************************** REGISTRATIONS ********************************

// Return all the registrations
// UnexpectedError - failed to retrieve registrations from the database
func (bc *BoltClient) Registrations() ([]export.Registration, error) {
	return bc.getRegistrations(bson.M{})
}

// Get registrations for the passed query
func (bc *BoltClient) getRegistrations(q bson.M) ([]export.Registration, error) {

	regs := []export.Registration{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
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
		b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		encoded, err := json.Marshal(reg)
		if err != nil {
			return err
		}
		err = b.Put([]byte(reg.ID), encoded)
		if err != nil {
			return ErrNotFound
		}
		return nil
	})

	return reg.ID, err
}

// Update a registration
// UnexpectedError - problem updating in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) UpdateRegistration(reg export.Registration) error {
	_, err := bc.RegistrationById(reg.ID.Hex())
	if err != nil {
		return ErrNotFound
	}

	reg.Modified = time.Now().UnixNano() / int64(time.Millisecond)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		encoded, err := json.Marshal(reg)
		if err != nil {
			return err
		}
		err = b.Put([]byte(reg.ID), encoded)
		if err != nil {
			return ErrNotFound
		}
		return nil
	})
	return err
}

// Get a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) RegistrationById(id string) (export.Registration, error) {
	reg := export.Registration{}
	if bson.IsObjectIdHex(id) {
		err := bc.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
			if b == nil {
				return ErrUnsupportedDatabase
			}
			encoded := b.Get([]byte(bson.ObjectIdHex(id)))
			if encoded == nil {
				return ErrNotFound
			}
			json := jsoniter.ConfigCompatibleWithStandardLibrary
			ret := json.Unmarshal(encoded, &reg)
			return ret
		})
		return reg, err
	} else {
		return reg, ErrNotFound
	}
}

// Delete a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) DeleteRegistrationById(id string) error {
	if bson.IsObjectIdHex(id) {
		err := bc.db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
			if b == nil {
				return ErrUnsupportedDatabase
			}
			err := b.Delete([]byte(bson.ObjectIdHex(id)))
			if err != nil {
				return ErrNotFound
			}
			return nil
		})
		return err
	} else {
		return ErrNotFound
	}
}

// Get a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the name was found
func (bc *BoltClient) RegistrationByName(name string) (export.Registration, error) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	reg := export.Registration{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
		if b == nil {
			return ErrUnsupportedDatabase
		}
		err := b.ForEach(func(id, encoded []byte) error {
			value := jsoniter.Get(encoded, "name").ToString()
			if value == name {
				err := json.Unmarshal(encoded, &reg)
				if err != nil {
					return err
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
	return reg, err
}

// Delete a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) DeleteRegistrationByName(name string) error {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
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
