//
// Copyright (c) 2017
// Mainflux
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/gorilla/mux"
)

const (
	distroPort int = 48070
)

const (
	typeAlgorithms   = "algorithms"
	typeCompressions = "compressions"
	typeFormats      = "formats"
	typeDestinations = "destinations"

	applicationJson = "application/json; charset=utf-8"
)

func getRegByID(w http.ResponseWriter, r *http.Request) {
	// URL parameters
	vars := mux.Vars(r)
	id := vars["id"]

	reg, err := dbClient.RegistrationById(id)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by id: %s. Error: %s", id, err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&reg)
}

func getRegList(w http.ResponseWriter, r *http.Request) {
	// URL parameters
	vars := mux.Vars(r)
	t := vars["type"]

	var list []string

	switch t {
	case typeAlgorithms:
		list = append(list, models.EncNone)
		list = append(list, models.EncAes)
	case typeCompressions:
		list = append(list, models.CompNone)
		list = append(list, models.CompGzip)
		list = append(list, models.CompZip)
	case typeFormats:
		list = append(list, models.FormatJSON)
		list = append(list, models.FormatXML)
		list = append(list, models.FormatIoTCoreJSON)
		list = append(list, models.FormatAzureJSON)
		list = append(list, models.FormatAWSJSON)
		list = append(list, models.FormatThingsBoardJSON)
		list = append(list, models.FormatNOOP)
	case typeDestinations:
		list = append(list, models.DestMQTT)
		list = append(list, models.DestIotCoreMQTT)
		list = append(list, models.DestAzureMQTT)
		list = append(list, models.DestRest)
		list = append(list, models.DestXMPP)
		list = append(list, models.DestAWSMQTT)
	default:
		LoggingClient.Error("Unknown type: " + t)
		http.Error(w, "Unknown type: "+t, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&list)
}

func getAllReg(w http.ResponseWriter, r *http.Request) {
	reg, err := dbClient.Registrations()
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query all registrations. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&reg)
}

func getRegByName(w http.ResponseWriter, r *http.Request) {
	// URL parameters
	vars := mux.Vars(r)
	name := vars["name"]

	reg, err := dbClient.RegistrationByName(name)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by name. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", applicationJson)
	json.NewEncoder(w).Encode(&reg)
}

func addReg(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query add registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reg := models.Registration{}
	if err := json.Unmarshal(data, &reg); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query add registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if reg.Format == "DEXMA_JSON" && reg.Destination == "DEXMA_TOPIC" {
		if reg.Name == "" {
			LoggingClient.Error(fmt.Sprintf("Failed to validate registrations fields: %X. Error: Name is required", data))
			http.Error(w, "Could not validate json fields", http.StatusBadRequest)
			return
		}
	} else if valid, err := reg.Validate(); !valid {
		LoggingClient.Error(fmt.Sprintf("Failed to validate registrations fields: %X. Error: %s", data, err.Error()))
		http.Error(w, "Could not validate json fields", http.StatusBadRequest)
		return
	}

	_, err = dbClient.RegistrationByName(reg.Name)
	if err == nil {
		LoggingClient.Error("Name already taken: " + reg.Name)
		http.Error(w, "Name already taken", http.StatusBadRequest)
		return
	} else if err != db.ErrNotFound {
		LoggingClient.Error(fmt.Sprintf("Failed to query add registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if reg.Name == "Amazon Web Services (AWS)" || reg.Name == "Google Cloud IoT Core" {
		var keyDirAmd64 string
		//var keyDirArm string
		var certDirAmd64 string
		//var certDirArm string
		var keyRegister string
		var certRegister string

		if reg.Name == "Amazon Web Services (AWS)" {
			keyDirAmd64 = "./keys/aws_private.pem"
			//keyDirArm = "/etc/edgex/aws_private.pem"
			certDirAmd64 = "./keys/aws_cert.pem"
			//certDirArm = "/etc/edgex/aws_cert.pem"
		} else if reg.Name == "Google Cloud IoT Core" {
			keyDirAmd64 = "./keys/giot_private.pem"
			//keyDirArm = "/etc/edgex/giot_private.pem"
			certDirAmd64 = "./keys/giot_cert.pem"
			//certDirArm = "/etc/edgex/giot_cert.pem"
		}
		keyRegister = reg.Addressable.Path
		certRegister = reg.Addressable.Protocol

		LoggingClient.Debug("Check Private Key")
		blockkey, _ := pem.Decode([]byte(keyRegister))
		if blockkey == nil {
			LoggingClient.Error("Error decoding Private Key")
			http.Error(w, "Error decoding Private Key", http.StatusInternalServerError)
			return
		}
		_, err = x509.ParsePKCS8PrivateKey(blockkey.Bytes)
		if err != nil {
			_, err = x509.ParsePKCS1PrivateKey(blockkey.Bytes)
			if err != nil {
				LoggingClient.Error(fmt.Sprintf("Error validating Private Key. Error: %s", err.Error()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		err = ioutil.WriteFile(keyDirAmd64, ([]byte(keyRegister)), 0644)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error Writting Private Key. Error: %s", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		LoggingClient.Debug("Check Certificate")
		block, _ := pem.Decode([]byte(certRegister))
		if block == nil {
			LoggingClient.Error("Error decoding Certificate")
			http.Error(w, "Error decoding Certificate", http.StatusInternalServerError)
			return
		}
		_, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			LoggingClient.Error("Error validating Certificate")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = ioutil.WriteFile(certDirAmd64, ([]byte(certRegister)), 0644)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error Writting Certificate. Error: %s", err.Error()))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reg.Addressable.Path = ""
		if reg.Name == "Amazon Web Services (AWS)" {
			reg.Addressable.Protocol = ""
		} else if reg.Name == "Google Cloud IoT Core" {
			reg.Addressable.Protocol = "tls"
		}
	}

	id, err := dbClient.AddRegistration(reg)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query add registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	notifyUpdatedRegistrations(models.NotifyUpdate{Name: reg.Name,
		Operation: "add"})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(id))
}

func updateReg(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to read update registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var fromReg models.Registration
	if err := json.Unmarshal(data, &fromReg); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to unmarshal update registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the registration exists
	var toReg models.Registration
	if fromReg.ID != "" {
		toReg, err = dbClient.RegistrationById(fromReg.ID)
	} else if fromReg.Name != "" {
		toReg, err = dbClient.RegistrationByName(fromReg.Name)
	} else {
		http.Error(w, "Need id or name", http.StatusBadRequest)
		return
	}

	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query update registration. Error: %s", err.Error()))
		if err == db.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
		}
		return
	}

	if fromReg.Name != "" {
		toReg.Name = fromReg.Name
	}
	if fromReg.Addressable.Name != "" {
		toReg.Addressable = fromReg.Addressable
	}
	if fromReg.Format != "" {
		toReg.Format = fromReg.Format
	}
	if fromReg.Filter.DeviceIDs != nil {
		toReg.Filter.DeviceIDs = fromReg.Filter.DeviceIDs
	}
	if fromReg.Filter.ValueDescriptorIDs != nil {
		toReg.Filter.ValueDescriptorIDs = fromReg.Filter.ValueDescriptorIDs
	}
	if fromReg.Encryption.Algo != "" {
		toReg.Encryption = fromReg.Encryption
	}
	if fromReg.Compression != "" {
		toReg.Compression = fromReg.Compression
	}
	if fromReg.Destination != "" {
		toReg.Destination = fromReg.Destination
	}

	// In order to know if 'enable' parameter have been sent or not, we unmarshal again
	// the registration in a map[string] and then check if the parameter is present or not
	var objmap map[string]*json.RawMessage
	json.Unmarshal(data, &objmap)
	if objmap["enable"] != nil {
		toReg.Enable = fromReg.Enable
	}

	if toReg.Format == "DEXMA_JSON" && toReg.Destination == "DEXMA_TOPIC" {
		if toReg.Name == "" {
			LoggingClient.Error(fmt.Sprintf("Failed to validate registrations fields: %X. Error: Name is required", data))
			http.Error(w, "Could not validate json fields", http.StatusBadRequest)
			return
		}
	} else if valid, err := toReg.Validate(); !valid {
		LoggingClient.Error(fmt.Sprintf("Failed to validate registrations fields: %X. Error: %s", data, err.Error()))
		http.Error(w, "Could not validate json fields", http.StatusBadRequest)
		return
	}

	err = dbClient.UpdateRegistration(toReg)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query update registration. Error: %s", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notifyUpdatedRegistrations(models.NotifyUpdate{Name: toReg.Name,
		Operation: "update"})

	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func delRegByID(w http.ResponseWriter, r *http.Request) {

	// URL parameters
	vars := mux.Vars(r)
	id := vars["id"]

	// Read the registration, the registration name is needed to
	// notify distro of the deletion
	reg, err := dbClient.RegistrationById(id)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by id: %s. Error: %s", id, err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = dbClient.DeleteRegistrationById(id)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by id: %s. Error: %s", id, err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	notifyUpdatedRegistrations(models.NotifyUpdate{Name: reg.Name,
		Operation: "delete"})

	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func delRegByName(w http.ResponseWriter, r *http.Request) {
	// URL parameters
	vars := mux.Vars(r)
	name := vars["name"]

	err := dbClient.DeleteRegistrationByName(name)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to query by name: %s. Error: %s", name, err.Error()))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	notifyUpdatedRegistrations(models.NotifyUpdate{Name: name,
		Operation: "delete"})

	w.Header().Set("Content-Type", applicationJson)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("true"))
}

func notifyUpdatedRegistrations(update models.NotifyUpdate) {
	go func() {
		err := dc.NotifyRegistrations(update, context.Background())
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("error from distro: %s", err.Error()))
		}
	}()
}
