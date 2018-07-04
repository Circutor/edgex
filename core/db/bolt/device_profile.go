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
type boltDeviceProfile struct {
	models.DeviceProfile
}

// Custom marshaling into bolt
func (bdp boltDeviceProfile) MarshalJSON() ([]byte, error) {
	// Get the commands from the device profile and turn them into DBRef objects
	var commands []string
	for _, command := range bdp.DeviceProfile.Commands {
		commands = append(commands, command.Id.Hex())
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(&struct {
		models.DescribedObject `json:",inline"`
		Id                     bson.ObjectId            `json:"_id,omitempty"`
		Name                   string                   `json:"name"`         // Non-database identifier (must be unique)
		Manufacturer           string                   `json:"manufacturer"` // Manufacturer of the device
		Model                  string                   `json:"model"`        // Model of the device
		Labels                 []string                 `json:"labels"`       // Labels used to search for groups of profiles
		Objects                interface{}              `json:"objects"`      // JSON data that the device service uses to communicate with devices with this profile
		DeviceResources        []models.DeviceObject    `json:"deviceResources"`
		Resources              []models.ProfileResource `json:"resources"`
		Commands               []string                 `json:"commands"` // List of commands to Get/Put information for devices associated with this profile
	}{
		DescribedObject: bdp.DeviceProfile.DescribedObject,
		Id:              bdp.DeviceProfile.Id,
		Name:            bdp.DeviceProfile.Name,
		Manufacturer:    bdp.DeviceProfile.Manufacturer,
		Model:           bdp.DeviceProfile.Model,
		Labels:          bdp.DeviceProfile.Labels,
		Objects:         bdp.DeviceProfile.Objects,
		DeviceResources: bdp.DeviceProfile.DeviceResources,
		Resources:       bdp.DeviceProfile.Resources,
		Commands:        commands,
	})
}

// Custom unmarshaling out of bolt
func (bdp *boltDeviceProfile) UnmarshalJSON(data []byte) error {
	decoded := new(struct {
		models.DescribedObject `json:",inline"`
		Id                     bson.ObjectId            `json:"_id,omitempty"`
		Name                   string                   `json:"name"`         // Non-database identifier (must be unique)
		Manufacturer           string                   `json:"manufacturer"` // Manufacturer of the device
		Model                  string                   `json:"model"`        // Model of the device
		Labels                 []string                 `json:"labels"`       // Labels used to search for groups of profiles
		Objects                interface{}              `json:"objects"`      // JSON data that the device service uses to communicate with devices with this profile
		DeviceResources        []models.DeviceObject    `json:"deviceResources"`
		Resources              []models.ProfileResource `json:"resources"`
		Commands               []string                 `json:"commands"` // List of commands to Get/Put information for devices associated with this profile
	})

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	// Copy over the fields
	bdp.DeviceProfile.DescribedObject = decoded.DescribedObject
	bdp.DeviceProfile.Id = decoded.Id
	bdp.DeviceProfile.Name = decoded.Name
	bdp.DeviceProfile.Manufacturer = decoded.Manufacturer
	bdp.DeviceProfile.Model = decoded.Model
	bdp.DeviceProfile.Labels = decoded.Labels
	bdp.DeviceProfile.Objects = decoded.Objects
	bdp.DeviceProfile.DeviceResources = decoded.DeviceResources
	bdp.DeviceProfile.Resources = decoded.Resources

	m, err := getCurrentBoltClient()
	if err != nil {
		return err
	}
	var commands []models.Command
	// Get all of the commands from the references
	for _, cRef := range decoded.Commands {
		var c models.Command
		err := m.GetCommandById(&c, cRef)
		if err != nil {
			return err
		}
		commands = append(commands, c)
	}
	bdp.Commands = commands

	return nil
}
