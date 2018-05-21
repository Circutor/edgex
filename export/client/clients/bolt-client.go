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
	//	"encoding/json"
	"log"
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
		//json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var reg export.Registration

			iter := jsoniter.ConfigFastest.BorrowIterator(encoded)
			defer jsoniter.ConfigFastest.ReturnIterator(iter)
			iter.ReadVal(&reg)
			//		err := json.Unmarshal(encoded, &reg)
			//if err != nil {
			//return err
			if iter.Error != nil {
				return iter.Error
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
	//json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(REGISTRATION_COLLECTION))

		encoded := jsoniter.ConfigFastest.BorrowStream(nil)
		defer jsoniter.ConfigFastest.ReturnStream(encoded)
		encoded.WriteVal(reg)
		if encoded.Error != nil {
			return encoded.Error
		}
		//os.Stdout.Write(encoded.Buffer())

		/*
			b, err := Marshal(group)
			if err != nil {
				fmt.Println("error:", err)
			}
			os.Stdout.Write(b)



			encoded, err := json.Marshal(reg)
			if err != nil {
				return err
			}
		*/
		return b.Put([]byte(reg.ID), encoded.Buffer())
	})

	return reg.ID, err
}

// Update a registration
// UnexpectedError - problem updating in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) UpdateRegistration(reg export.Registration) error {
	reg.Modified = time.Now().UnixNano() / int64(time.Millisecond)

	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
		//json := jsoniter.ConfigCompatibleWithStandardLibrary

		encoded := jsoniter.ConfigFastest.BorrowStream(nil)
		defer jsoniter.ConfigFastest.ReturnStream(encoded)
		encoded.WriteVal(reg)
		if encoded.Error != nil {
			return encoded.Error
		}

		/*encoded, err := json.Marshal(reg)
		if err != nil {
			return err
		}
		*/
		return b.Put([]byte(reg.ID), encoded.Buffer())
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
				return ErrNotFound
			}
			//json := jsoniter.ConfigCompatibleWithStandardLibrary
			encoded := b.Get([]byte(bson.ObjectIdHex(id)))
			if encoded == nil {
				return ErrNotFound
			}
			iter := jsoniter.ConfigFastest.BorrowIterator(encoded)
			defer jsoniter.ConfigFastest.ReturnIterator(iter)
			iter.ReadVal(&reg)
			return iter.Error

			//ret := json.Unmarshal(encoded, &reg)
			//return ret
		})
		return reg, err
	} else {
		return export.Registration{}, ErrNotFound
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
				return ErrNotFound
			}
			//json := jsoniter.ConfigCompatibleWithStandardLibrary
			b.Delete([]byte(bson.ObjectIdHex(id)))
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
	log.Println("Ini RegistrationByName")
	start := time.Now()
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	reg := export.Registration{}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
		err := b.ForEach(func(id, encoded []byte) error {
			var data map[string]interface{}
			iter := jsoniter.ConfigFastest.BorrowIterator(encoded)
			defer jsoniter.ConfigFastest.ReturnIterator(iter)
			iter.ReadVal(&data)
			if iter.Error != nil {
				return iter.Error
			}
			//err := json.Unmarshal(encoded, &data)
			//if err != nil {
			//	return err
			//}

			if data["name"] == name {
				//			iter := jsoniter.ConfigFastest.BorrowIterator(encoded)
				//			defer jsoniter.ConfigFastest.ReturnIterator(iter)
				//			iter.ReadVal(&reg)
				//	err := Unmarshal(jsonBlob, &animals)
				err := json.Unmarshal(encoded, &reg)
				//if iter.Error != nil {
				if err != nil {
					return err
				}
				// The only way to leave the ForEach is using an error, so let's trick it
				return ErrNameFound
			}
			return nil
		})
		if err == nil {
			err = ErrNotFound
		} else if err == ErrNameFound {
			return nil
		}
		return err
	})
	elapsed := time.Since(start)
	log.Println("Total time. Post RegistrationByName", elapsed)

	return reg, err

}

// Delete a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) DeleteRegistrationByName(name string) error {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(REGISTRATION_COLLECTION))
		//json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var data map[string]interface{}
			iter := jsoniter.ConfigFastest.BorrowIterator(encoded)
			defer jsoniter.ConfigFastest.ReturnIterator(iter)
			iter.ReadVal(&data)
			/*if iter.Error != nil {
				return iter.Error
			}
			err := json.Unmarshal(encoded, &data)
			if err != nil {
				return err
			}
			*/
			if data["name"] == name {
				b.Delete([]byte(id))
				return ErrNameFound
			}
			// The only way to leave the ForEach is using an error, so let's trick it
			return nil
		})
		if err == nil {
			err = ErrNotFound
		} else if err == ErrNameFound {
			return nil
		}
		return err
	})
	return err
}
