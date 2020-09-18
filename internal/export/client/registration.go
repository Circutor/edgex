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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Circutor/edgex/internal/pkg/db"
	"github.com/Circutor/edgex/pkg/models"
	"github.com/gorilla/mux"
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
	if reg.Addressable.Password != "" {
		reg.Addressable.Password, _ = Decrypt(reg.Addressable.Password)
	}
	if reg.Addressable.Certificate != "" {
		reg.Addressable.Certificate, _ = Decrypt(reg.Addressable.Certificate)
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

	for i := range reg {
		if reg[i].Addressable.Password != "" {
			reg[i].Addressable.Password, _ = Decrypt(reg[i].Addressable.Password)
		}
		if reg[i].Addressable.Certificate != "" {
			reg[i].Addressable.Certificate, _ = Decrypt(reg[i].Addressable.Certificate)
		}
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

	if reg.Addressable.Password != "" {
		reg.Addressable.Password, _ = Decrypt(reg.Addressable.Password)
	}
	if reg.Addressable.Certificate != "" {
		reg.Addressable.Certificate, _ = Decrypt(reg.Addressable.Certificate)
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

	fillRegister(&reg)

	if reg.Format == "DEXMA_JSON" {
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

	if reg.Addressable.Certificate != "" {
		if reg.Format == "AWS_JSON" || reg.Format == "IOTCORE_JSON" {
			err = checkCertificate(reg.Addressable.Certificate)
			if err != nil {
				LoggingClient.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		encCert, _ := Encrypt(reg.Addressable.Certificate)
		reg.Addressable.Certificate = encCert
	}
	if reg.Addressable.Password != "" {
		if reg.Format == "AWS_JSON" || reg.Format == "IOTCORE_JSON" {
			err = checkKey(reg.Addressable.Password)
			if err != nil {
				LoggingClient.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		encKey, _ := Encrypt(reg.Addressable.Password)
		reg.Addressable.Password = encKey
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

func checkCertificate(certRegister string) (err error) {
	block, _ := pem.Decode([]byte(certRegister))
	if block == nil {
		err = errors.New("Error decoding Certificate")
		return
	}
	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		err = errors.New("Error validating Certificate")
		return
	}

	return nil
}

func checkKey(keyRegister string) (err error) {
	blockkey, _ := pem.Decode([]byte(keyRegister))
	if blockkey == nil {
		err = errors.New("Error decoding Private Key")
		return
	}
	_, err = x509.ParsePKCS8PrivateKey(blockkey.Bytes)
	if err != nil {
		_, err = x509.ParsePKCS1PrivateKey(blockkey.Bytes)
		if err != nil {
			err = errors.New("Error validating Private Key")
			return
		}
	}

	return nil
}

func fillRegister(reg *models.Registration) (err error) {
	switch reg.Format {
	case "THINGSBOARD_JSON":
		reg.Addressable.Protocol = "TCP"
		reg.Addressable.Publisher = "Circutor"
		reg.Addressable.Topic = "v1/gateway/telemetry"
		reg.Destination = "MQTT_TOPIC"
	case "DEXMA_JSON":
		reg.Addressable.Protocol = "HTTP"
		reg.Addressable.HTTPMethod = "POST"
		reg.Addressable.Topic = "readings"
		reg.Destination = "DEXMA_TOPIC"
	case "AZURE_JSON":
		reg.Addressable.Protocol = "tls"
		reg.Addressable.User = "EDS-Cloud.azure-devices.net/" + reg.Addressable.Publisher
		reg.Addressable.Topic = "devices/DeviceId/messages/events/"
		reg.Destination = "AZURE_TOPIC"
	case "AWS_JSON":
		reg.Destination = "AWS_TOPIC"
	case "IOTCORE_JSON":
		reg.Addressable.Protocol = "tls"
		reg.Addressable.Path = ""
		reg.Addressable.Address = "mqtt.googleapis.com"
		reg.Addressable.Port = 8883
		reg.Addressable.User = "unused"
		reg.Destination = "IOTCORE_TOPIC"
	default:
		err = errors.New("Not valid protocol")
	}

	return
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

	if fromReg.Addressable.Id != "" {
		toReg.Addressable.Id = fromReg.Addressable.Id
	}
	if fromReg.Addressable.Name != "" {
		toReg.Addressable.Name = fromReg.Addressable.Name
	}
	if fromReg.Addressable.Protocol != "" {
		toReg.Addressable.Protocol = fromReg.Addressable.Protocol
	}
	if fromReg.Addressable.HTTPMethod != "" {
		toReg.Addressable.HTTPMethod = fromReg.Addressable.HTTPMethod
	}
	if fromReg.Addressable.Address != "" {
		toReg.Addressable.Address = fromReg.Addressable.Address
	}
	if fromReg.Addressable.Port != 0 {
		toReg.Addressable.Port = fromReg.Addressable.Port
	}
	if fromReg.Addressable.Path != "" {
		toReg.Addressable.Path = fromReg.Addressable.Path
	}
	if fromReg.Addressable.Publisher != "" {
		toReg.Addressable.Publisher = fromReg.Addressable.Publisher
	}
	if fromReg.Addressable.User != "" {
		toReg.Addressable.User = fromReg.Addressable.User
	}
	if fromReg.Addressable.Topic != "" {
		toReg.Addressable.Topic = fromReg.Addressable.Topic
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

	if fromReg.Addressable.Certificate != "" {
		toReg.Addressable.Certificate = fromReg.Addressable.Certificate
		if toReg.Format == "AWS_JSON" || toReg.Format == "IOTCORE_JSON" {
			err = checkCertificate(toReg.Addressable.Certificate)
			if err != nil {
				LoggingClient.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		encCert, _ := Encrypt(toReg.Addressable.Certificate)
		toReg.Addressable.Certificate = encCert
	}
	if fromReg.Addressable.Password != "" {
		toReg.Addressable.Password = fromReg.Addressable.Password
		if toReg.Format == "AWS_JSON" || toReg.Format == "IOTCORE_JSON" {
			err = checkKey(toReg.Addressable.Password)
			if err != nil {
				LoggingClient.Error(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		encKey, _ := Encrypt(toReg.Addressable.Password)
		toReg.Addressable.Password = encKey
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
