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
	"github.com/edgexfoundry/edgex-go/pkg/models"
	jsoniter "github.com/json-iterator/go"
)

// Struct that wraps an event to handle DB references
type boltEvent struct {
	Event    models.Event
	Readings []string
}

// Custom marshaling into bolt
func (be boltEvent) MarshalJSON() ([]byte, error) {
	// Turn the readings into DB references
	var readings []string
	for _, reading := range be.Event.Readings {
		readings = append(readings, reading.Id)
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(&struct {
		ID       string   `json:"id"`
		Pushed   int64    `json:"pushed,omitempty"`
		Device   string   `json:"device"`
		Created  int64    `json:"created,omitempty"`
		Modified int64    `json:"modified,omitempty"`
		Origin   int64    `json:"origin,omitempty"`
		Readings []string `json:"readings,omitempty"`
	}{
		ID:       be.Event.ID,
		Pushed:   be.Event.Pushed,
		Device:   be.Event.Device,
		Created:  be.Event.Created,
		Modified: be.Event.Modified,
		Origin:   be.Event.Origin,
		Readings: readings,
	})
}

// Custom unmarshaling out of bolt
func (be *boltEvent) UnmarshalJSON(data []byte) error {
	decoded := new(struct {
		ID       string   `json:"id"`
		Pushed   int64    `json:"pushed"`
		Device   string   `json:"device"`
		Created  int64    `json:"created"`
		Modified int64    `json:"modified"`
		Origin   int64    `json:"origin"`
		Readings []string `json:"readings"`
	})
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	// Copy over the non DB referenced fields
	be.Event.ID = decoded.ID
	be.Event.Pushed = decoded.Pushed
	be.Event.Device = decoded.Device
	be.Event.Created = decoded.Created
	be.Event.Modified = decoded.Modified
	be.Event.Origin = decoded.Origin
	be.Readings = decoded.Readings

	return nil
}
