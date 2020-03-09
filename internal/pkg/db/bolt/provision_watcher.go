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

// Internal version of the provision watcher struct
// Use this to handle DBRef
type boltProvisionWatcher struct {
	models.ProvisionWatcher
}

// Custom marshaling into bolt
func (bpw boltProvisionWatcher) MarshalJSON() ([]byte, error) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(&struct {
		models.BaseObject `json:",inline"`
		Id                string                `json:"id"`
		Name              string                `json:"name"`           // Unique name for identifying a device
		Identifiers       map[string]string     `json:"identifiers"`    // set of key value pairs that identify type of of address (MAC, HTTP,...) and address to watch for (00-05-1B-A1-99-99, 10.0.0.1,...)
		OperatingState    models.OperatingState `json:"operatingState"` // Operating state (enabled/disabled)
		ServiceID         string                `json:"serviceId"`      // Associated Device Service - One per device
		ProfileID         string                `json:"profileId"`
	}{
		BaseObject:     bpw.BaseObject,
		Id:             bpw.Id,
		Name:           bpw.Name,
		Identifiers:    bpw.Identifiers,
		OperatingState: bpw.OperatingState,
		ServiceID:      bpw.Service.Id,
		ProfileID:      bpw.Profile.Id,
	})
}

// Custom unmarshaling out of bolt
func (bpw *boltProvisionWatcher) UnmarshalJSON(data []byte) error {
	decoded := new(struct {
		models.BaseObject `json:",inline"`
		Id                string                `json:"id"`
		Name              string                `json:"name"`           // Unique name for identifying a device
		Identifiers       map[string]string     `json:"identifiers"`    // set of key value pairs that identify type of of address (MAC, HTTP,...) and address to watch for (00-05-1B-A1-99-99, 10.0.0.1,...)
		OperatingState    models.OperatingState `json:"operatingState"` // Operating state (enabled/disabled)
		ServiceID         string                `json:"serviceId"`      // Associated Device Service - One per device
		ProfileID         string                `json:"profileId"`
	})

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	// Copy the fields
	bpw.BaseObject = decoded.BaseObject
	bpw.Id = decoded.Id
	bpw.Name = decoded.Name
	bpw.Identifiers = decoded.Identifiers
	bpw.OperatingState = decoded.OperatingState

	b, err := getCurrentBoltClient()
	if err != nil {
		return err
	}

	var ds models.DeviceService
	ds, err = b.GetDeviceServiceById(decoded.ServiceID)
	if err != nil {
		return err
	}

	var dp models.DeviceProfile
	dp, err = b.GetDeviceProfileById(decoded.ProfileID)
	if err != nil {
		return err
	}

	bpw.Profile = dp
	bpw.Service = ds

	return nil
}
