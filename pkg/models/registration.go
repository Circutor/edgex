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
	"fmt"
)

// Export destination types
const (
	DestMQTT        = "MQTT_TOPIC"
	DestZMQ         = "ZMQ_TOPIC"
	DestIotCoreMQTT = "IOTCORE_TOPIC"
	DestAzureMQTT   = "AZURE_TOPIC"
	DestRest        = "REST_ENDPOINT"
	DestXMPP        = "XMPP_TOPIC"
	DestAWSMQTT     = "AWS_TOPIC"
)

// Compression algorithm types
const (
	CompNone = "NONE"
	CompGzip = "GZIP"
	CompZip  = "ZIP"
)

// Data format types
const (
	FormatJSON            = "JSON"
	FormatXML             = "XML"
	FormatSerialized      = "SERIALIZED"
	FormatIoTCoreJSON     = "IOTCORE_JSON"
	FormatAzureJSON       = "AZURE_JSON"
	FormatAWSJSON         = "AWS_JSON"
	FormatCSV             = "CSV"
	FormatThingsBoardJSON = "THINGSBOARD_JSON"
	FormatNOOP            = "NOOP"
)

const (
	NotifyUpdateAdd    = "add"
	NotifyUpdateUpdate = "update"
	NotifyUpdateDelete = "delete"
)

// Registration - Defines the registration details
// on the part of north side export clients
type Registration struct {
	ID          string            `json:"id"`
	Created     int64             `json:"created"`
	Modified    int64             `json:"modified"`
	Origin      int64             `json:"origin"`
	Name        string            `json:"name"`
	Addressable Addressable       `json:"addressable"`
	Format      string            `json:"format"`
	Filter      Filter            `json:"filter"`
	Encryption  EncryptionDetails `json:"encryption"`
	Compression string            `json:"compression"`
	Enable      bool              `json:"enable"`
	Destination string            `json:"destination"`
}

// Custom marshaling for JSON
// Treat the strings as pointers so they can be null in JSON
func (r Registration) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID          *string            `json:"id,omitempty"`
		Created     int64              `json:"created,omitempty"`
		Modified    int64              `json:"modified,omitempty"`
		Origin      int64              `json:"origin,omitempty"`
		Name        *string            `json:"name,omitempty"`
		Addressable Addressable        `json:"addressable"`
		Format      *string            `json:"format,omitempty"`
		Filter      *Filter            `json:"filter,omitempty"`
		Encryption  *EncryptionDetails `json:"encryption,omitempty"`
		Compression *string            `json:"compression,omitempty"`
		Enable      bool               `json:"enable"`
		Destination *string            `json:"destination,omitempty"`
	}{
		Created:     r.Created,
		Modified:    r.Modified,
		Origin:      r.Origin,
		Addressable: r.Addressable,
		Enable:      r.Enable,
	}

	// Only initialize the non-empty strings (empty are null)
	if r.ID != "" {
		aux.ID = &r.ID
	}
	if r.Name != "" {
		aux.Name = &r.Name
	}
	if r.Format != "" {
		aux.Format = &r.Format
	}
	if len(r.Filter.DeviceIDs) != 0 || len(r.Filter.ValueDescriptorIDs) != 0 {
		aux.Filter = &r.Filter
	}
	if r.Encryption.Algo != "" || r.Encryption.Key != "" || r.Encryption.InitVector != "" {
		aux.Encryption = &r.Encryption
	}
	if r.Compression != "" {
		aux.Compression = &r.Compression
	}
	if r.Destination != "" {
		aux.Destination = &r.Destination
	}

	return json.Marshal(aux)
}

func (reg Registration) Validate() (bool, error) {

	if reg.Name == "" {
		return false, fmt.Errorf("Name is required")
	}

	if reg.Compression == "" {
		reg.Compression = CompNone
	}

	if reg.Compression != CompNone &&
		reg.Compression != CompGzip &&
		reg.Compression != CompZip {
		return false, fmt.Errorf("Compression invalid: %s", reg.Compression)
	}

	if reg.Format != FormatJSON &&
		reg.Format != FormatXML &&
		reg.Format != FormatSerialized &&
		reg.Format != FormatIoTCoreJSON &&
		reg.Format != FormatAzureJSON &&
		reg.Format != FormatAWSJSON &&
		reg.Format != FormatCSV &&
		reg.Format != FormatThingsBoardJSON &&
		reg.Format != FormatNOOP {
		return false, fmt.Errorf("Format invalid: %s", reg.Format)
	}

	if reg.Destination != DestMQTT &&
		reg.Destination != DestZMQ &&
		reg.Destination != DestIotCoreMQTT &&
		reg.Destination != DestAzureMQTT &&
		reg.Destination != DestAWSMQTT &&
		reg.Destination != DestRest {
		return false, fmt.Errorf("Destination invalid: %s", reg.Destination)
	}

	if reg.Encryption.Algo == "" {
		reg.Encryption.Algo = EncNone
	}

	if reg.Encryption.Algo != EncNone &&
		reg.Encryption.Algo != EncAes {
		return false, fmt.Errorf("Encryption invalid: %s", reg.Encryption.Algo)
	}

	return true, nil
}
