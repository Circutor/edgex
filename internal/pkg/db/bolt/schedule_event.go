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
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	jsoniter "github.com/json-iterator/go"
)

// Internal version of the schedule event struct
// Use this to handle DBRef
type boltScheduleEvent struct {
	models.ScheduleEvent
}

// Custom marshaling into bolt
func (bse boltScheduleEvent) MarshalJSON() ([]byte, error) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(&struct {
		Created       int64  `json:"created"`
		Modified      int64  `json:"modified"`
		Origin        int64  `json:"origin"`
		Id            string `json:"id"`
		Name          string `json:"name"`          // non-database unique identifier for a schedule event
		Schedule      string `json:"schedule"`      // Name to associated owning schedule
		AddressableID string `json:"addressableId"` // address {MQTT topic, HTTP address, serial bus, etc.} for the action (can be empty)
		Parameters    string `json:"parameters"`    // json body for parameters
		Service       string `json:"service"`       // json body for parameters
	}{
		Created:       bse.Created,
		Modified:      bse.Modified,
		Origin:        bse.Origin,
		Id:            bse.Id,
		Name:          bse.Name,
		Schedule:      bse.Schedule,
		Parameters:    bse.Parameters,
		Service:       bse.Service,
		AddressableID: bse.Addressable.Id,
	})
}

// Custom unmarshaling out of bolt
func (bse *boltScheduleEvent) UnmarshalJSON(data []byte) error {
	decoded := new(struct {
		Created       int64  `json:"created"`
		Modified      int64  `json:"modified"`
		Origin        int64  `json:"origin"`
		Id            string `json:"id"`
		Name          string `json:"name"`          // non-database unique identifier for a schedule event
		Schedule      string `json:"schedule"`      // Name to associated owning schedule
		AddressableID string `json:"addressableId"` // address {MQTT topic, HTTP address, serial bus, etc.} for the action (can be empty)
		Parameters    string `json:"parameters"`    // json body for parameters
		Service       string `json:"service"`       // json body for parameters
	})
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	// Copy over the non-DBRef fields
	bse.Created = decoded.Created
	bse.Modified = decoded.Modified
	bse.Origin = decoded.Origin
	bse.Id = decoded.Id
	bse.Name = decoded.Name
	bse.Schedule = decoded.Schedule
	bse.Parameters = decoded.Parameters
	bse.Service = decoded.Service

	b, err := getCurrentBoltClient()
	if err != nil {
		return err
	}

	return b.getById(&bse.Addressable, db.Addressable, decoded.AddressableID)
}
