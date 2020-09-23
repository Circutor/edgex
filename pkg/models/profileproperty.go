/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package models

import "encoding/json"

type ProfileProperty struct {
	Value     PropertyValue `json:"value"`
	Units     Units         `json:"units" yaml:"units,omitempty"`
	MediaType string        `json:"mediaType" yaml:"mediaType,omitempty"`
}

// Custom marshaling to make empty strings null
func (p ProfileProperty) MarshalJSON() ([]byte, error) {
	test := struct {
		Value     PropertyValue `json:"value"`
		Units     *Units        `json:"units,omitempty"`
		MediaType *string       `json:"mediaType,omitempty"`
	}{}

	// Empty strings are null
	test.Value = p.Value
	if p.Units.Type != "" || p.Units.ReadWrite != "" || p.Units.DefaultValue != "" {
		test.Units = &p.Units
	}
	if p.MediaType != "" {
		test.MediaType = &p.MediaType
	}

	return json.Marshal(test)
}

/*
 * To String function for DeviceService
 */
func (pp ProfileProperty) String() string {
	out, err := json.Marshal(pp)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
