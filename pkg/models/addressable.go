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

import (
	"encoding/json"
	"strconv"
	"strings"
)

/*
 * This file is the model for addressable in EdgeX
 * Addressable holds information about a specific address
 *
 * Addressable struct
 */
type Addressable struct {
	BaseObject
	Id          string `json:"id"`
	Name        string `json:"name"`
	Protocol    string `json:"protocol"`    // Protocol for the address (HTTP/TCP)
	HTTPMethod  string `json:"method"`      // Method for connecting (i.e. POST)
	Address     string `json:"address"`     // Address of the addressable
	Port        int    `json:"port"`        // Port for the address
	Path        string `json:"path"`        // Path for callbacks
	Publisher   string `json:"publisher"`   // For message bus protocols
	User        string `json:"user"`        // User id for authentication
	Password    string `json:"password"`    // Password of the user for authentication for the addressable
	Topic       string `json:"topic"`       // Topic for message bus addressables
	Certificate string `json:"certificate"` // Certificate used for authentication
}

// Custom marshaling for JSON
// Treat the strings as pointers so they can be null in JSON
func (a Addressable) MarshalJSON() ([]byte, error) {
	aux := struct {
		BaseObject
		Id          *string `json:"id,omitempty"`
		Name        *string `json:"name,omitempty"`
		Protocol    *string `json:"protocol,omitempty"`    // Protocol for the address (HTTP/TCP)
		HTTPMethod  *string `json:"method,omitempty"`      // Method for connecting (i.e. POST)
		Address     *string `json:"address,omitempty"`     // Address of the addressable
		Port        int     `json:"port,omitempty"`        // Port for the address
		Path        *string `json:"path,omitempty"`        // Path for callbacks
		Publisher   *string `json:"publisher,omitempty"`   // For message bus protocols
		User        *string `json:"user,omitempty"`        // User id for authentication
		Password    *string `json:"password,omitempty"`    // Password of the user for authentication for the addressable
		Topic       *string `json:"topic,omitempty"`       // Topic for message bus addressables
		Certificate *string `json:"certificate,omitempty"` // Certificate used for authentication
	}{
		BaseObject: a.BaseObject,
		Port:       a.Port,
	}

	if a.Id != "" {
		aux.Id = &a.Id
	}

	// Only initialize the non-empty strings (empty are null)
	if a.Name != "" {
		aux.Name = &a.Name
	}
	if a.Protocol != "" {
		aux.Protocol = &a.Protocol
	}
	if a.HTTPMethod != "" {
		aux.HTTPMethod = &a.HTTPMethod
	}
	if a.Address != "" {
		aux.Address = &a.Address
	}
	if a.Path != "" {
		aux.Path = &a.Path
	}
	if a.Publisher != "" {
		aux.Publisher = &a.Publisher
	}
	if a.User != "" {
		aux.User = &a.User
	}
	if a.Password != "" {
		aux.Password = &a.Password
	}
	if a.Topic != "" {
		aux.Topic = &a.Topic
	}
	if a.Certificate != "" {
		aux.Certificate = &a.Certificate
	}

	return json.Marshal(aux)
}

/*
 * String() function for formatting
 */
func (a Addressable) String() string {
	out, err := json.Marshal(a)
	if err != nil {
		return err.Error()
	}
	return string(out)
}

func (a Addressable) GetBaseURL() string {
	protocol := strings.ToLower(a.Protocol)
	address := a.Address
	port := strconv.Itoa(a.Port)
	baseURL := protocol + "://" + address + ":" + port
	return baseURL
}

// Get the callback url for the addressable if all relevant tokens have values.
// If any token is missing, string will be empty
func (a Addressable) GetCallbackURL() string {
	url := ""
	if len(a.Protocol) > 0 && len(a.Address) > 0 && a.Port > 0 && len(a.Path) > 0 {
		url = a.GetBaseURL() + a.Path
	}

	return url
}
