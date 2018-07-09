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
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

// Internal version of the device service struct
// Use this to handle DB references
type boltDevice struct {
	models.Device
}

// Custom marshaling into bolt
func (bd boltDevice) MarshalJSON() ([]byte, error) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(&struct {
		models.DescribedObject `json:",inline"`
		Id                     bson.ObjectId         `json:"_id,omitempty"`
		Name                   string                `json:"name"`           // Unique name for identifying a device
		AdminState             models.AdminState     `json:"adminState"`     // Admin state (locked/unlocked)
		OperatingState         models.OperatingState `json:"operatingState"` // Operating state (enabled/disabled)
		Addressable            bson.ObjectId         `json:"addressableId"`  // Addressable for the device - stores information about it's address
		LastConnected          int64                 `json:"lastConnected"`  // Time (milliseconds) that the device last provided any feedback or responded to any request
		LastReported           int64                 `json:"lastReported"`   // Time (milliseconds) that the device reported data to the core microservice
		Labels                 []string              `json:"labels"`         // Other labels applied to the device to help with searching
		Location               interface{}           `json:"location"`       // Device service specific location (interface{} is an empty interface so it can be anything)
		ServiceID              string                `json:"serviceId"`      // Associated Device Service - One per device
		ProfileID              string                `json:"profileId"`
	}{
		DescribedObject: bd.DescribedObject,
		Id:              bd.Id,
		Name:            bd.Name,
		AdminState:      bd.AdminState,
		OperatingState:  bd.OperatingState,
		Addressable:     bd.Addressable.Id,
		LastConnected:   bd.LastConnected,
		LastReported:    bd.LastReported,
		Labels:          bd.Labels,
		ServiceID:       bd.Service.Id.Hex(),
		ProfileID:       bd.Profile.Id.Hex(),
	})
}

// Custom unmarshaling out of bolt
func (bd *boltDevice) UnmarshalJSON(data []byte) error {
	decoded := new(struct {
		models.DescribedObject `json:",inline"`
		Id                     bson.ObjectId         `json:"_id,omitempty"`
		Name                   string                `json:"name"`           // Unique name for identifying a device
		AdminState             models.AdminState     `json:"adminState"`     // Admin state (locked/unlocked)
		OperatingState         models.OperatingState `json:"operatingState"` // Operating state (enabled/disabled)
		Addressable            bson.ObjectId         `json:"addressableId"`  // Addressable for the device - stores information about it's address
		LastConnected          int64                 `json:"lastConnected"`  // Time (milliseconds) that the device last provided any feedback or responded to any request
		LastReported           int64                 `json:"lastReported"`   // Time (milliseconds) that the device reported data to the core microservice
		Labels                 []string              `json:"labels"`         // Other labels applied to the device to help with searching
		Location               interface{}           `json:"location"`       // Device service specific location (interface{} is an empty interface so it can be anything)
		ServiceID              string                `json:"serviceId"`      // Associated Device Service - One per device
		ProfileID              string                `json:"profileId"`
	})
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	// Copy the fields
	bd.DescribedObject = decoded.DescribedObject
	bd.Id = decoded.Id
	bd.Name = decoded.Name
	bd.AdminState = decoded.AdminState
	bd.OperatingState = decoded.OperatingState
	bd.LastConnected = decoded.LastConnected
	bd.LastReported = decoded.LastReported
	bd.Labels = decoded.Labels
	bd.Location = decoded.Location

	m, err := getCurrentBoltClient()
	if err != nil {
		return err
	}

	var a models.Addressable
	err = m.GetAddressableById(&a, decoded.Addressable.Hex())
	if err != nil {
		return err
	}

	var ds models.DeviceService
	err = m.GetDeviceServiceById(&ds, decoded.ServiceID)
	if err != nil {
		return err
	}

	var dp models.DeviceProfile
	err = m.GetDeviceProfileById(&dp, decoded.ProfileID)
	if err != nil {
		return err
	}

	bd.Addressable = a
	bd.Profile = dp
	bd.Service = ds

	return nil
}
