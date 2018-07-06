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
	"errors"

	"github.com/coreos/bbolt"
	"github.com/edgexfoundry/edgex-go/core/db"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

var currentBoltClient *BoltClient

type BoltClient struct {
	db *bolt.DB // Bolt database
}

var ErrLimReached error = errors.New("Limit reached")
var ErrObjFound error = errors.New("Object name found")

// Return a pointer to the BoltClient
func NewClient(config db.Configuration) (*BoltClient, error) {

	bdb, err := bolt.Open("./database.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	boltClient := &BoltClient{db: bdb}
	currentBoltClient = boltClient
	return boltClient, nil
}

func (bc *BoltClient) Connect() error {
	return nil
}

func (bc *BoltClient) CloseSession() {
	bc.db.Close()
}

// Get the current Bolt Client
func getCurrentBoltClient() (*BoltClient, error) {
	if currentBoltClient == nil {
		return nil, errors.New("No current bolt client, please create a new client before requesting it")
	}

	return currentBoltClient, nil
}

// Add an element
func (bc *BoltClient) add(bucket string, element interface{}, id bson.ObjectId) error {
	return bc.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(bucket))
		if b == nil {
			return db.ErrUnsupportedDatabase
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
		b, _ := tx.CreateBucketIfNotExists([]byte(bucket))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		if b.Get([]byte(id)) == nil {
			return db.ErrNotFound
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
		return db.ErrInvalidObjectId
	}
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b, _ := tx.CreateBucketIfNotExists([]byte(col))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		return b.Delete([]byte(bson.ObjectIdHex(id)))
	})
	return err
}

// Get an element by ID
func (bc *BoltClient) getById(v interface{}, c string, gid string) error {
	// Check if id is a hexstring
	if !bson.IsObjectIdHex(gid) {
		return db.ErrInvalidObjectId
	}
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(c))
		if b == nil {
			return db.ErrNotFound
		}
		encoded := b.Get([]byte(bson.ObjectIdHex(gid)))
		if encoded == nil {
			return db.ErrNotFound
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
			return db.ErrNotFound
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
			return db.ErrNotFound
		} else if err == ErrObjFound {
			return nil
		}
		return err
	})
	return err
}

// Count number of elements
func (bc *BoltClient) count(bucket string) (int, error) {
	bstat := 0
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b != nil {
			bstat = b.Stats().KeyN
		}
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
