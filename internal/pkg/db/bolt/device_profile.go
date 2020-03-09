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
		commands = append(commands, command.Id)
	}
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(&struct {
		models.DescribedObject `json:",inline"`
		Id                     string                   `json:"id"`
		Name                   string                   `json:"name"`             // Non-database identifier (must be unique)
		Manufacturer           string                   `json:"manufacturer"`     // Manufacturer of the device
		Model                  string                   `json:"model"`            // Model of the device
		Labels                 []string                 `json:"labels,omitempty"` // Labels used to search for groups of profiles
		DeviceResources        []models.DeviceResource  `json:"deviceResources,omitempty"`
		Resources              []models.ProfileResource `json:"resources,omitempty"`
		Commands               []string                 `json:"commands"` // List of commands to Get/Put information for devices associated with this profile
	}{
		DescribedObject: bdp.DescribedObject,
		Id:              bdp.Id,
		Name:            bdp.Name,
		Manufacturer:    bdp.Manufacturer,
		Model:           bdp.Model,
		Labels:          bdp.Labels,
		DeviceResources: bdp.DeviceResources,
		Resources:       bdp.Resources,
		Commands:        commands,
	})
}

// Custom unmarshaling out of bolt
func (bdp *boltDeviceProfile) UnmarshalJSON(data []byte) error {
	decoded := new(struct {
		models.DescribedObject `json:",inline"`
		Id                     string                   `json:"id"`
		Name                   string                   `json:"name"`             // Non-database identifier (must be unique)
		Manufacturer           string                   `json:"manufacturer"`     // Manufacturer of the device
		Model                  string                   `json:"model"`            // Model of the device
		Labels                 []string                 `json:"labels,omitempty"` // Labels used to search for groups of profiles
		DeviceResources        []models.DeviceResource  `json:"deviceResources,omitempty"`
		Resources              []models.ProfileResource `json:"resources,omitempty"`
		Commands               []string                 `json:"commands,omitempty"` // List of commands to Get/Put information for devices associated with this profile
	})

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	// Copy over the fields
	bdp.DescribedObject = decoded.DescribedObject
	bdp.Id = decoded.Id
	bdp.Name = decoded.Name
	bdp.Manufacturer = decoded.Manufacturer
	bdp.Model = decoded.Model
	bdp.Labels = decoded.Labels
	bdp.DeviceResources = decoded.DeviceResources
	bdp.Resources = decoded.Resources

	m, err := getCurrentBoltClient()
	if err != nil {
		return err
	}

	// Get all of the commands from the references
	var commands []models.Command
	for _, cRef := range decoded.Commands {
		c, err := m.GetCommandById(cRef)
		if err != nil {
			return err
		}
		commands = append(commands, c)
	}
	bdp.Commands = commands

	return nil
}
