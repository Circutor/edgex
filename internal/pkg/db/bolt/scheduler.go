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

// ******************************* INTERVALS **********************************

// Return all the Interval(s)
// UnexpectedError - failed to retrieve intervals from the database
// Sort the events in descending order by ID
func (bc *BoltClient) Intervals() ([]contract.Interval, error) {
	return bc.getIntervals(-1)
}

// Return Interval(s) up to the max number specified
// UnexpectedError - failed to retrieve intervals from the database
// Sort the intervals in descending order by ID
func (bc *BoltClient) IntervalsWithLimit(limit int) ([]contract.Interval, error) {
	return bc.getIntervals(limit)
}

// Return an Interval by name
// UnexpectedError - failed to retrieve interval from the database
func (bc *BoltClient) IntervalByName(name string) (contract.Interval, error) {
	var mi contract.Interval
	err := bc.getByName(&mi, db.Interval, name)
	return mi, err
}

// Return an Interval by ID
// UnexpectedError - failed to retrieve interval from the database
func (bc *BoltClient) IntervalById(id string) (contract.Interval, error) {
	var mi contract.Interval
	err := bc.getById(&mi, db.Interval, id)
	return mi, err
}

// Add an Interval
// UnexpectedError - failed to add interval into  the database
func (bc *BoltClient) AddInterval(interval contract.Interval) (string, error) {
	// Check if the name is unique
	var dumy contract.Interval
	err := bc.getByName(&dumy, db.Interval, interval.Name)
	if err == nil {
		return interval.ID, db.ErrNotUnique
	}

	interval.ID = uuid.New().String()
	interval.Created = db.MakeTimestamp()
	interval.Modified = interval.Created

	// Add the internal
	err = bc.add(db.Interval, interval, interval.ID)
	return interval.ID, err
}

// Update an Interval
// UnexpectedError - failed to update interval in the database
func (bc *BoltClient) UpdateInterval(interval contract.Interval) error {
	interval.Modified = db.MakeTimestamp()
	return bc.update(db.Interval, interval, interval.ID)
}

// Remove an Interval by ID
// UnexpectedError - failed to remove interval from the database
func (bc *BoltClient) DeleteIntervalById(id string) error {
	return bc.deleteById(id, db.Interval)
}

// ******************************* INTERVAL ACTIONS **********************************

// Return all the Interval Action(s)
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (bc *BoltClient) IntervalActions() ([]contract.IntervalAction, error) {
	return bc.getIntervalActions(func(encoded []byte) bool {
		return true
	}, -1)
}

// Return Interval Action(s) up to the max number specified
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (bc *BoltClient) IntervalActionsWithLimit(limit int) ([]contract.IntervalAction, error) {
	return bc.getIntervalActions(func(encoded []byte) bool {
		return true
	}, limit)
}

// Return Interval Action(s) by interval name
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (bc *BoltClient) IntervalActionsByIntervalName(name string) ([]contract.IntervalAction, error) {
	return bc.getIntervalActions(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "interval").ToString()
		if name == value {
			return true
		}
		return false
	}, -1)
}

// Return Interval Action(s) by target name
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (bc *BoltClient) IntervalActionsByTarget(target string) ([]contract.IntervalAction, error) {
	return bc.getIntervalActions(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "target").ToString()
		if target == value {
			return true
		}
		return false
	}, -1)
}

// Return an Interval Action by ID
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (bc *BoltClient) IntervalActionById(id string) (contract.IntervalAction, error) {
	var ia contract.IntervalAction
	err := bc.getById(&ia, db.IntervalAction, id)
	return ia, err
}

// Return an Interval Action by name
// UnexpectedError - failed to retrieve interval actions from the database
// Sort the interval actions in descending order by ID
func (bc *BoltClient) IntervalActionByName(name string) (contract.IntervalAction, error) {
	var ia contract.IntervalAction
	err := bc.getByName(&ia, db.IntervalAction, name)
	return ia, err
}

// Add a new Interval Action
// UnexpectedError - failed to add interval action into the database
func (bc *BoltClient) AddIntervalAction(action contract.IntervalAction) (string, error) {
	// Check if the name is unique
	var dumy contract.IntervalAction
	err := bc.getByName(&dumy, db.IntervalAction, action.Name)
	if err == nil {
		return action.ID, db.ErrNotUnique
	}

	action.ID = uuid.New().String()
	action.Created = db.MakeTimestamp()
	action.Modified = action.Created

	// Add the internal
	err = bc.add(db.IntervalAction, action, action.ID)
	return action.ID, err
}

// Update an Interval Action
// UnexpectedError - failed to update interval action in the database
func (bc *BoltClient) UpdateIntervalAction(action contract.IntervalAction) error {
	action.Modified = db.MakeTimestamp()
	return bc.update(db.IntervalAction, action, action.ID)
}

// Remove an Interval Action by ID
// UnexpectedError - failed to remove interval action from the database
func (bc *BoltClient) DeleteIntervalActionById(id string) error {
	return bc.deleteById(id, db.IntervalAction)
}

// ******************************* HELPER FUNCTIONS **********************************

// Get intervals
func (bc *BoltClient) getIntervals(limit int) ([]contract.Interval, error) {
	mi := contract.Interval{}
	mis := []contract.Interval{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	// Check if limit is not 0
	if limit == 0 {
		return mis, nil
	}
	cnt := 0

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Interval))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			err := json.Unmarshal(encoded, &mi)
			if err != nil {
				return err
			}
			mis = append(mis, mi)
			if limit > 0 {
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
	return mis, err
}

// Get events for the passed check
func (bc *BoltClient) getIntervalActions(fn func(encoded []byte) bool, limit int) ([]contract.IntervalAction, error) {
	ia := contract.IntervalAction{}
	ias := []contract.IntervalAction{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	// Check if limit is not 0
	if limit == 0 {
		return ias, nil
	}
	cnt := 0

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.IntervalAction))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &ia)
				if err != nil {
					return err
				}
				ias = append(ias, ia)
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
	return ias, err
}

// ******************************* UTILITY FUNCTIONS **********************************

// Removes all of the Interval Action(s)
// Returns number of Interval Action(s) removed
// UnexpectedError - failed to remove all of the Interval and IntervalActions from the database
func (bc *BoltClient) ScrubAllIntervalActions() (int, error) {
	count, err := bc.count(db.IntervalAction)
	if err != nil {
		return 0, err
	}
	return count, bc.scrubAll(db.IntervalAction)
}

// Removes all of the Intervals
// Removes any IntervalAction(s) previously not removed as well
// Returns number Interval(s) removed
// UnexpectedError - failed to remove all of the Interval and IntervalActions from the database
func (bc *BoltClient) ScrubAllIntervals() (int, error) {
	// Ensure we have removed interval actions first
	count, err := bc.ScrubAllIntervalActions()
	if err != nil {
		return 0, err
	}

	count, err = bc.count(db.Interval)
	if err != nil {
		return 0, err
	}
	return count, bc.scrubAll(db.Interval)
}
