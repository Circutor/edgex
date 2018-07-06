/*******************************************************************************
 * Copyright 2018 Circutor S.A.
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
type boltDeviceService struct {
	models.DeviceService
}

// Custom marshaling into bolt
func (bds boltDeviceService) MarshalJSON() ([]byte, error) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(&struct {
		models.DescribedObject `json:",inline"`
		Id                     bson.ObjectId         `json:"id"`
		Name                   string                `json:"name"`           // time in milliseconds that the device last provided any feedback or responded to any request
		LastConnected          int64                 `json:"lastConnected"`  // time in milliseconds that the device last reported data to the core
		LastReported           int64                 `json:"lastReported"`   // operational state - either enabled or disabled
		OperatingState         models.OperatingState `json:"operatingState"` // operational state - ether enabled or disableddc
		Labels                 []string              `json:"labels"`         // tags or other labels applied to the device service for search or other identification needs
		Addressable            bson.ObjectId         `json:"addressableId"`  // address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
		AdminState             models.AdminState     `json:"adminState"`     // Device Service Admin State
	}{
		DescribedObject: bds.Service.DescribedObject,
		Id:              bds.Service.Id,
		Name:            bds.Service.Name,
		LastConnected:   bds.Service.LastConnected,
		LastReported:    bds.Service.LastReported,
		OperatingState:  bds.Service.OperatingState,
		Labels:          bds.Service.Labels,
		Addressable:     bds.Service.Addressable.Id,
		AdminState:      bds.AdminState,
	})
}

// Custom unmarshaling out of bolt
func (bds *boltDeviceService) UnmarshalJSON(data []byte) error {
	decoded := new(struct {
		models.DescribedObject `json:",inline"`
		Id                     bson.ObjectId         `json:"id,omitempty"`
		Name                   string                `json:"name"`           // time in milliseconds that the device last provided any feedback or responded to any request
		LastConnected          int64                 `json:"lastConnected"`  // time in milliseconds that the device last reported data to the core
		LastReported           int64                 `json:"lastReported"`   // operational state - either enabled or disabled
		OperatingState         models.OperatingState `json:"operatingState"` // operational state - ether enabled or disableddc
		Labels                 []string              `json:"labels"`         // tags or other labels applied to the device service for search or other identification needs
		Addressable            bson.ObjectId         `json:"addressableId"`  // address (MQTT topic, HTTP address, serial bus, etc.) for reaching the service
		AdminState             models.AdminState     `json:"adminState"`     // Device Service Admin State
	})
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	// Copy over the non-DBRef fields
	bds.Service.DescribedObject = decoded.DescribedObject
	bds.Service.Id = decoded.Id
	bds.Service.Name = decoded.Name
	bds.AdminState = decoded.AdminState
	bds.Service.OperatingState = decoded.OperatingState
	bds.Service.LastConnected = decoded.LastConnected
	bds.Service.LastReported = decoded.LastReported
	bds.Service.Labels = decoded.Labels

	m, err := getCurrentBoltClient()
	if err != nil {
		return err
	}

	return m.GetAddressableById(&bds.Service.Addressable, decoded.Addressable.Hex())

}
