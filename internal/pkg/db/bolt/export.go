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

package bolt

import (
	bolt "github.com/coreos/bbolt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	contract "github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

const (
	ExportCollection = "export"
)

// ****************************** REGISTRATIONS ********************************

// Return all the registrations
// UnexpectedError - failed to retrieve registrations from the database
func (bc *BoltClient) Registrations() ([]contract.Registration, error) {
	regs := []contract.Registration{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ExportCollection))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			reg := contract.Registration{}
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
func (bc *BoltClient) AddRegistration(reg contract.Registration) (string, error) {
	reg.ID = uuid.New().String()
	reg.Created = db.MakeTimestamp()
	reg.Modified = reg.Created

	err := bc.add(ExportCollection, reg, reg.ID)
	return reg.ID, err
}

// Update a registration
// UnexpectedError - problem updating in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) UpdateRegistration(reg contract.Registration) error {
	reg.Modified = db.MakeTimestamp()

	return bc.update(ExportCollection, reg, reg.ID)
}

// Get a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) RegistrationById(id string) (contract.Registration, error) {
	var reg contract.Registration
	err := bc.getById(&reg, ExportCollection, id)
	return reg, err
}

// Get a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the name was found
func (bc *BoltClient) RegistrationByName(name string) (contract.Registration, error) {
	var reg contract.Registration
	err := bc.getByName(&reg, ExportCollection, name)
	return reg, err
}

// Delete a registration by ID
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) DeleteRegistrationById(id string) error {
	return bc.deleteById(id, ExportCollection)
}

// Delete a registration by name
// UnexpectedError - problem getting in database
// NotFound - no registration with the ID was found
func (bc *BoltClient) DeleteRegistrationByName(name string) error {
	return bc.deleteByName(name, ExportCollection)
}

// Delete all registrations
func (bc *BoltClient) ScrubAllRegistrations() error {
	return bc.scrubAll(ExportCollection)
}
