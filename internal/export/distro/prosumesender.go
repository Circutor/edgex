//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Circutor/edgex/internal/pkg/correlation/models"
	contract "github.com/Circutor/edgex/pkg/models"
)

type prosumeSender struct {
	path       string
	deviceName string
}

func newProsumeSender(addr contract.Addressable) sender {
	sender := prosumeSender{
		path:       addr.Path,
		deviceName: addr.Name,
	}
	return sender
}

// Send reads previous absolute power values and rewrites with new values if they are correct
func (sender prosumeSender) Send(newData []byte, event *models.Event) bool {
	var err error
	type prosumeData struct {
		Timestamp      int64   `json:"Timestamp"`
		ImportedEnergy float64 `json:"Active_energy_imported_kWh"`
		ExportedEnergy float64 `json:"Active_energy_exported_kWh"`
	}
	currentData := prosumeData{}
	json.Unmarshal(newData, &currentData)

	fileData, err := ioutil.ReadFile(sender.path)
	if err != nil {
		LoggingClient.Info(fmt.Sprintf("Could not read previous prosume data file in: %s, creating new file", sender.path))
	} else {
		// Check if new info is valid (aka greater than previous values as per prosume blockchain specs)
		previousData := prosumeData{}
		json.Unmarshal(fileData, &previousData)

		if currentData.Timestamp <= previousData.Timestamp {
			currentData.Timestamp = previousData.Timestamp
		}
		if currentData.ImportedEnergy <= previousData.ImportedEnergy {
			currentData.ImportedEnergy = previousData.ImportedEnergy
		}
		if currentData.ExportedEnergy <= previousData.ExportedEnergy {
			currentData.ExportedEnergy = previousData.ExportedEnergy
		}
	}
	file, err := os.Create(sender.path)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Internal error send prosume - Unable to create prosume data file: %s", err.Error()))
		return false
	}
	// Finally we write data in file
	encoder := json.NewEncoder(file)
	err = encoder.Encode(currentData)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Internal error send prosume - Unable to write new data in prosume file: %s", err.Error()))
		return false
	}
	return true
}
