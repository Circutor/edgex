//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
// Copyright (c) 2018 Dell Technologies, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

// TODO:
// - Event buffer management per sender(do not block distro.Loop on full
//   registration channel)

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Circutor/edgex/internal/pkg/correlation/models"
	contract "github.com/Circutor/edgex/pkg/models"
)

const (
	awsMQTTPort         int    = 8883
	awsThingUpdateTopic string = "$aws/things/%s/shadow/update"
)

var registrationChanges chan contract.NotifyUpdate = make(chan contract.NotifyUpdate, 2)

// RegistrationInfo - registration info
type registrationInfo struct {
	registration contract.Registration
	format       formatter
	compression  transformer
	encrypt      transformer
	sender       sender
	filter       []filterer

	chRegistration chan *contract.Registration
	chEvent        chan *models.Event

	deleteFlag bool
}

func RefreshRegistrations(update contract.NotifyUpdate) {
	// TODO make it not blocking, return bool?
	registrationChanges <- update
}

func newRegistrationInfo() *registrationInfo {
	reg := &registrationInfo{}

	reg.chRegistration = make(chan *contract.Registration)
	reg.chEvent = make(chan *models.Event)
	return reg
}

func (reg *registrationInfo) update(newReg contract.Registration) bool {
	reg.registration = newReg

	reg.format = nil
	switch newReg.Format {
	case contract.FormatJSON:
		reg.format = jsonFormatter{}
	case contract.FormatXML:
		reg.format = xmlFormatter{}
	case contract.FormatSerialized:
		reg.format = jsonFormatter{}
	case contract.FormatIoTCoreJSON:
		reg.format = jsonFormatter{}
	case contract.FormatAzureJSON:
		reg.format = azureFormatter{}
	case contract.FormatAWSJSON:
		reg.format = awsFormatter{}
	case contract.FormatCSV:
		// TODO reg.format = distro.NewCsvFormat()
	case contract.FormatThingsBoardJSON:
		reg.format = thingsboardJSONFormatter{}
	case "DEXMA_JSON":
		reg.format = dexmaJSONFormatter{}
	case contract.FormatNOOP:
		reg.format = noopFormatter{}
	default:
		LoggingClient.Warn(fmt.Sprintf("Format not supported: %s", newReg.Format))
		return false
	}

	reg.compression = nil
	switch newReg.Compression {
	case "":
		fallthrough
	case contract.CompNone:
		reg.compression = nil
	case contract.CompGzip:
		reg.compression = &gzipTransformer{}
	case contract.CompZip:
		reg.compression = &zlibTransformer{}
	default:
		LoggingClient.Warn(fmt.Sprintf("Compression not supported: %s", newReg.Compression))
		return false
	}

	reg.sender = nil
	switch newReg.Destination {
	case contract.DestMQTT, contract.DestAzureMQTT:
		pKey, err := decryptPrivateKey(newReg.Addressable.Password)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error Decrypting Private Key: %s", err.Error()))
			return false
		}
		reg.sender = newMqttSender(newReg.Addressable, pKey)
	case contract.DestAWSMQTT:
		newReg.Addressable.Protocol = "tls"
		newReg.Addressable.Path = ""
		newReg.Addressable.Topic = fmt.Sprintf(awsThingUpdateTopic, newReg.Addressable.Topic)
		newReg.Addressable.Port = awsMQTTPort
		pKey, err := decryptPrivateKey(newReg.Addressable.Password)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error Decrypting Private Key: %s", err.Error()))
			return false
		}
		reg.sender = newMqttSender(newReg.Addressable, pKey)
	case contract.DestIotCoreMQTT:
		pKey, err := decryptPrivateKey(newReg.Addressable.Password)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error Decrypting Private Key: %s", err.Error()))
			return false
		}
		reg.sender = newIoTCoreSender(newReg.Addressable, pKey)
	case contract.DestRest:
		reg.sender = newHTTPSender(newReg.Addressable)
	case "DEXMA_TOPIC":
		reg.sender = newHTTPDexmaSender(newReg.Addressable)
	case contract.DestXMPP:
		reg.sender = newXMPPSender(newReg.Addressable)

	default:
		LoggingClient.Warn(fmt.Sprintf("Destination not supported: %s", newReg.Destination))
		return false
	}

	if reg.sender == nil {
		return false
	}

	reg.encrypt = nil
	switch newReg.Encryption.Algo {
	case "":
		fallthrough
	case contract.EncNone:
		reg.encrypt = nil
	case contract.EncAes:
		reg.encrypt = newAESEncryption(newReg.Encryption)
	default:
		LoggingClient.Warn(fmt.Sprintf("Encryption not supported: %s", newReg.Encryption.Algo))
		return false
	}

	reg.filter = nil

	if len(newReg.Filter.DeviceIDs) > 0 {
		reg.filter = append(reg.filter, newDevIdFilter(newReg.Filter))
		LoggingClient.Debug(fmt.Sprintf("Device ID filter added: %s", newReg.Filter.DeviceIDs))
	}

	if len(newReg.Filter.ValueDescriptorIDs) > 0 {
		reg.filter = append(reg.filter, newValueDescFilter(newReg.Filter))
		LoggingClient.Debug(fmt.Sprintf("Value descriptor filter added: %s", newReg.Filter.ValueDescriptorIDs))
	}

	return true
}

func decryptPrivateKey(encodedKey string) (key []byte, err error) {
	keyIn := []byte(encodedKey)
	decodedPEM, _ := pem.Decode(keyIn)
	pass, err := getShadow()
	if err != nil {
		return
	}

	decrypedPemBlock, err := x509.DecryptPEMBlock(decodedPEM, []byte(pass))
	if err != nil {
		err = errors.New("Error Decrypting PEM Block")
		return
	}

	var parsedKey interface{}
	var bytes []byte
	var privateKey *rsa.PrivateKey
	var ok bool
	if parsedKey, err = x509.ParsePKCS8PrivateKey(decrypedPemBlock); err != nil {
		if parsedKey, err = x509.ParsePKCS1PrivateKey(decrypedPemBlock); err != nil {
			err = errors.New("Couldn't parse Private Key")
			return
		}
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			err = errors.New("Couldn't parse PKCS1 Private Key")
			return
		}
		bytes = x509.MarshalPKCS1PrivateKey(privateKey)
	} else {
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			err = errors.New("Couldn't parse PKCS8 Private Key")
			return
		}
		bytes, err = x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			err = errors.New("Couldn't extract PKCS8 Private Key")
			return
		}
	}

	key = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: bytes,
	})

	return
}

func getShadow() (string, error) {
	_, err := os.Stat("/etc/shadow.toml")
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("Shadow file not found: %v", err)
		}
	}

	contents, err := ioutil.ReadFile("/etc/shadow.toml")
	if err != nil {
		return "", fmt.Errorf("Failed to read shadow file: %v", err)
	}
	var psk struct {
		Shadow string
	}
	err = toml.Unmarshal(contents, &psk)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal shadow file: %v", err)
	}

	return psk.Shadow, nil
}
func (reg registrationInfo) processEvent(event *models.Event) {
	// Valid Event Filter, needed?

	data := event.ToContract()
	for _, f := range reg.filter {
		var accepted bool
		accepted, data = f.Filter(data)
		if !accepted {
			LoggingClient.Info("Event filtered")
			return
		}
	}

	if reg.format == nil {
		LoggingClient.Warn("registrationInfo with nil format")
		return
	}
	formatted := reg.format.Format(data)

	compressed := formatted
	if reg.compression != nil {
		compressed = reg.compression.Transform(formatted)
	}

	encrypted := compressed
	if reg.encrypt != nil {
		encrypted = reg.encrypt.Transform(compressed)
	}

	if reg.sender.Send(encrypted, event) && Configuration.Writable.MarkPushed {
		id := event.ID
		err := ec.MarkPushed(id, context.Background())

		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Failed to mark event as pushed : event ID = %s: %s", id, err))
		}
	}

	LoggingClient.Debug(fmt.Sprintf("Sent event with registration: %s", reg.registration.Name))
}

func registrationLoop(reg *registrationInfo) {
	LoggingClient.Info(fmt.Sprintf("registration loop started: %s", reg.registration.Name))
	for {
		select {
		case event := <-reg.chEvent:
			if reg.registration.Enable {
				reg.processEvent(event)
			}

		case newReg := <-reg.chRegistration:
			if newReg == nil {
				LoggingClient.Info("Terminating registration goroutine")
				return
			} else {
				if reg.update(*newReg) {
					LoggingClient.Info(fmt.Sprintf("Registration %s updated: OK", reg.registration.Name))
				} else {
					LoggingClient.Info(fmt.Sprintf("Registration %s updated: OK, terminating goroutine", reg.registration.Name))
					reg.deleteFlag = true
					return
				}
			}
		}
	}
}

func updateRunningRegistrations(running map[string]*registrationInfo,
	update contract.NotifyUpdate) error {

	switch update.Operation {
	case contract.NotifyUpdateDelete:
		for k, v := range running {
			if k == update.Name {
				v.chRegistration <- nil
				delete(running, k)
				return nil
			}
		}
		return fmt.Errorf("delete update not processed")
	case contract.NotifyUpdateUpdate:
		reg := getRegistrationByName(update.Name)
		if reg == nil {
			return fmt.Errorf("Could not find registration")
		}
		for k, v := range running {
			if k == update.Name {
				v.chRegistration <- reg
				return nil
			}
		}
		return fmt.Errorf("Could not find running registration")
	case contract.NotifyUpdateAdd:
		reg := getRegistrationByName(update.Name)
		if reg == nil {
			return fmt.Errorf("Could not find registration")
		}
		regInfo := newRegistrationInfo()
		if regInfo.update(*reg) {
			running[reg.Name] = regInfo
			go registrationLoop(regInfo)
		}
		return nil
	default:
		return fmt.Errorf("Invalid update operation")
	}
}

// Loop - registration loop
func Loop(errChan chan error, eventCh chan *models.Event) {
	go func() {
		p := fmt.Sprintf(":%d", Configuration.Service.Port)
		LoggingClient.Info(fmt.Sprintf("Starting Export Distro %s", p))
		errChan <- http.ListenAndServe(p, httpServer())
	}()

	registrations := make(map[string]*registrationInfo)

	allRegs, err := getRegistrations()

	for allRegs == nil {
		LoggingClient.Info("Waiting for client microservice")
		select {
		case e := <-errChan:
			LoggingClient.Error(fmt.Sprintf("exit msg: %s", e.Error()))
			if err != nil {
				LoggingClient.Error(fmt.Sprintf("with error: %s", err.Error()))
			}
			return
		case <-time.After(time.Second):
		}
		allRegs, err = getRegistrations()
	}

	// Create new goroutines for each registration
	for _, reg := range allRegs {
		regInfo := newRegistrationInfo()
		if regInfo.update(reg) {
			registrations[reg.Name] = regInfo
			go registrationLoop(regInfo)
		}
	}

	LoggingClient.Info("Starting registration loop")
	for {
		select {
		case e := <-errChan:
			// kill all registration goroutines
			for k, reg := range registrations {
				if !reg.deleteFlag {
					// Do not write in channel that will not be read
					reg.chRegistration <- nil
				}
				delete(registrations, k)
			}
			LoggingClient.Error(fmt.Sprintf("exit msg: %s", e.Error()))
			return

		case update := <-registrationChanges:
			LoggingClient.Info("Registration changes")
			err := updateRunningRegistrations(registrations, update)
			if err != nil {
				LoggingClient.Error(err.Error())
				LoggingClient.Warn(fmt.Sprintf("Error updating registration %s", update.Name))
			}

		case event := <-eventCh:
			for k, reg := range registrations {
				if reg.deleteFlag {
					delete(registrations, k)
				} else {
					// TODO only sent event if it is not blocking
					reg.chEvent <- event
				}
			}
		}
	}
}
