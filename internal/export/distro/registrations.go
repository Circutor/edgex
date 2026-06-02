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
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "net/http/pprof"

	"github.com/Circutor/edgex/internal/pkg/correlation/models"
	clients "github.com/Circutor/edgex/pkg/clients/export/distro"
	contract "github.com/Circutor/edgex/pkg/models"
	"github.com/google/uuid"
)

const (
	awsMQTTPort         int           = 8883
	awsThingUpdateTopic string        = "$aws/things/%s/shadow/update"
	pushEventsTimer     time.Duration = 300
)

var registrationChanges chan contract.NotifyUpdate = make(chan contract.NotifyUpdate, 2)

type connectionStatusQuery struct {
	responseChan chan []connectionStatusResponse
}

type connectionStatusResponse struct {
	Connected bool   `json:"connected"`
	Name      string `json:"name"`
}

var connectionStatusQueries chan connectionStatusQuery = make(chan connectionStatusQuery, 10)

type sendEventQuery struct {
	ruleName     string
	offTime      int
	body         string
	responseChan chan bool
}

var sendEventQueries chan sendEventQuery = make(chan sendEventQuery, 10)

// GetRegistrationConnectionStatus queries the connection status of all registrations
func GetRegistrationConnectionStatus() []connectionStatusResponse {
	responseChan := make(chan []connectionStatusResponse)
	query := connectionStatusQuery{
		responseChan: responseChan,
	}

	connectionStatusQueries <- query
	response := <-responseChan

	return response
}

// SendScoutEventToRegistration sends a generic event to the first found registration with destination Scout.
// returns true if the event was sent to at least one registration, false otherwise
func SendScoutEventToRegistration(name string, offTime int, body string) bool {
	responseChan := make(chan bool)
	query := sendEventQuery{
		ruleName:     name,
		offTime:      offTime,
		body:         body,
		responseChan: responseChan,
	}

	sendEventQueries <- query
	response := <-responseChan

	return response
}

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
	// In case new export is Prosume or after boot up, we have to start/stop Prosume client
	if newReg.Destination == contract.DestProsume && newReg.Enable != reg.registration.Enable {
		if newReg.Enable {
			clients.ProsumeClientExec(clients.ProsumeOpStart)
		} else {
			clients.ProsumeClientExec(clients.ProsumeOpStop)
		}
	}
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
	case contract.FormatDEXMAJSON:
		reg.format = dexmaJSONFormatter{}
	case contract.FormatProsume:
		reg.format = prosumeJSONFormatter{}
	case contract.FormatMyCircutorJSON:
		reg.format = myCircutorJSONFormatter{}
	case contract.FormatScoutJSON:
		reg.format = scoutJSONFormatter{}
	case contract.FormatSentiloJSON:
		reg.format = sentiloJSONFormatter{location: reg.registration.Addressable.Topic}
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

	if newReg.Destination != contract.DestScout {
		reg.sender = nil
	}
	switch newReg.Destination {
	case contract.DestMQTT, contract.DestAzureMQTT:
		reg.sender = newMqttSender(newReg.Addressable)
	case contract.DestAWSMQTT:
		newReg.Addressable.Protocol = "tls"
		newReg.Addressable.Path = ""
		newReg.Addressable.Topic = fmt.Sprintf(awsThingUpdateTopic, newReg.Addressable.Topic)
		newReg.Addressable.Port = awsMQTTPort
		reg.sender = newMqttSender(newReg.Addressable)
	case contract.DestIotCoreMQTT:
		reg.sender = newIoTCoreSender(newReg.Addressable)
	case contract.DestRest:
		reg.sender = newHTTPSender(newReg.Addressable)
	case contract.DestDEXMAMQTT:
		reg.sender = newHTTPDexmaSender(newReg.Addressable)
	case contract.DestXMPP:
		reg.sender = newXMPPSender(newReg.Addressable)
	case contract.DestProsume:
		reg.sender = newProsumeSender(newReg.Addressable)
	case contract.DestMyCircutor:
		reg.sender = newMyCircutorSender(newReg.Addressable)
	case contract.DestScout:
		if !newReg.Enable {
			destroyScoutSender(reg.sender)
		}
		reg.sender = newScoutSender(newReg.Addressable, newReg.Enable, newReg.Name)
	case contract.DestSentilo:
		reg.sender = newSentiloSender(newReg.Addressable)
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
	} else if reg.registration.Destination == contract.DestProsume && reg.registration.Addressable.Name != event.Device {
		return
	}
	formatted := reg.format.Format(data)
	if formatted == nil {
		if Configuration.Writable.MarkPushed {
			id := event.ID
			err := ec.MarkPushed(id, context.Background())
			if err != nil {
				LoggingClient.Error(fmt.Sprintf("Failed to mark event as pushed : event ID = %s: %s", id, err))
			}
		}
		return
	}

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

	tickerRegistration := time.NewTicker(time.Second * 10)
	timerPush := time.NewTimer(pushEventsTimer * time.Second)

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
		case <-timerPush.C:
			if Configuration.Writable.MarkPushed && reg.registration.Enable {
				events, err := ec.EventsUnpushed(context.Background(), 100)
				if err != nil {
					LoggingClient.Error(fmt.Sprintf("Failed getting events to send non-pushed %s", err.Error()))
				}

				if len(events) != 0 {
					LoggingClient.Info("Pushing unpushed events")
					for i := range events {
						correlationID := uuid.New()
						ev := models.Event{CorrelationId: correlationID.String(), Event: events[i]}
						reg.processEvent(&ev)
					}
				}
			}
			timerPush.Reset(pushEventsTimer * time.Second)
		case <-tickerRegistration.C:
			if reg.registration.Destination == contract.DestScout && reg.registration.Enable {
				if scoutSender, ok := reg.sender.(*scoutSender); ok {
					if !scoutSender.IsConnected() {
						scoutSender.Connect()
					}
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
				// In case export to delete is Prosume we stop client
				if v.registration.Destination == contract.DestProsume {
					clients.ProsumeClientExec(clients.ProsumeOpStop)
				}
				v.chRegistration <- nil
				delete(running, k)
				return nil
			}
		}
		return fmt.Errorf("delete update not processed")
	case contract.NotifyUpdateUpdate:
		reg := getRegistrationByName(update.Name)
		if reg == nil {
			return fmt.Errorf("could not find registration")
		}
		for k, v := range running {
			if k == update.Name {
				v.chRegistration <- reg
				return nil
			}
		}
		return fmt.Errorf("could not find running registration")
	case contract.NotifyUpdateAdd:
		reg := getRegistrationByName(update.Name)
		if reg == nil {
			return fmt.Errorf("could not find registration")
		}
		regInfo := newRegistrationInfo()
		if regInfo.update(*reg) {
			running[reg.Name] = regInfo
			go registrationLoop(regInfo)
		}
		return nil
	default:
		return fmt.Errorf("invalid update operation")
	}
}

// Loop - registration loop
func Loop(errChan chan error, eventCh chan *models.Event) {
	go func() {
		p := fmt.Sprintf(":%d", Configuration.Service.Port)
		LoggingClient.Info(fmt.Sprintf("Starting Export Distro %s", p))
		errChan <- http.ListenAndServe(p, httpServer())
	}()

	startProfile()

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

	// Create a map containing each rule and its expiration time
	registrationMap := make(map[string]time.Time)
	registrationCheckTicker := time.NewTicker(time.Second)

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
			if event.Restricted {
				// Mark as pushed in order to remove it from the list of events to push
				ec.MarkPushed(event.ID, context.Background())
			} else {
				for k, reg := range registrations {
					if reg.deleteFlag {
						delete(registrations, k)
					} else {
						// TODO only sent event if it is not blocking
						reg.chEvent <- event
					}
				}
			}

		case query := <-connectionStatusQueries:
			response := make([]connectionStatusResponse, 0)

			for regName, info := range registrations {
				if info.registration.Destination != contract.DestScout {
					continue
				}

				newReg := connectionStatusResponse{Name: regName, Connected: false}
				if scoutSender, ok := info.sender.(*scoutSender); ok {
					newReg.Connected = scoutSender.IsConnected()
				}

				response = append(response, newReg)
			}

			query.responseChan <- response

		case eventQuery := <-sendEventQueries:
			newTime := time.Now().Add(time.Duration(eventQuery.offTime) * time.Second)
			if _, ok := registrationMap[eventQuery.ruleName]; ok { // If the rule already exists, it has not expired so there is no need to send event
				registrationMap[eventQuery.ruleName] = newTime // just update the expiration time and return
				eventQuery.responseChan <- true

				continue
			}

			// If the rule does not exist or has expired, we send the activation event and set its expiration time
			registrationMap[eventQuery.ruleName] = newTime

			success := false
			for _, info := range registrations {
				if info.registration.Destination != contract.DestScout {
					continue
				}

				if scoutSender, ok := info.sender.(*scoutSender); ok {
					infos := []string{"EVENT_CUSTOM_NAME=" + eventQuery.ruleName}
					if len(eventQuery.body) > 0 {
						infos = append(infos, strings.Split(eventQuery.body, ",")...)
					}
					scoutSender.sendScoutEvent("ON", "ALARM_DEVICE", infos)
					success = true

				}

			}

			eventQuery.responseChan <- success

		case <-registrationCheckTicker.C:
			for ruleName, expirationTime := range registrationMap {
				if time.Now().Before(expirationTime) {
					continue
				}

				for _, info := range registrations {
					if info.registration.Destination != contract.DestScout {
						continue
					}

					if scoutSender, ok := info.sender.(*scoutSender); ok {
						infos := []string{"EVENT_CUSTOM_NAME=" + ruleName}
						scoutSender.sendScoutEvent("OFF", "ALARM_DEVICE", infos)
					}
				}
				// We delete the rule so next time the event is received we will send an "ON" event
				delete(registrationMap, ruleName)
			}
		}
	}
}

// Start starts the pprof server
// In order to get a pprof sample, you need to run one of the following commands (or similar):
// - curl http://localhost:6060/debug/pprof/heap > heap.pprof
// - go tool pprof http://localhost:6060/debug/pprof/goroutine
// - ...
// And then you can analyze the heap.pprof file with the pprof tool:
// - go tool pprof heap.pprof
// - go tool pprof --trim_path=/builds/circutor/firmware/concentrator/cnc-service --source_path=/home/jduran/go/src/cnc cnc-service goroutine.pprof
// For more information, please visit https://pkg.go.dev/net/http/pprof
func startProfile() {
	go func() {
		server := &http.Server{
			Addr:              "localhost:6060",
			ReadHeaderTimeout: 3 * time.Second,
		}

		err := server.ListenAndServe()
		if err != nil {
			LoggingClient.Error("Could not start pprof server", "error", err)
		}
	}()
}
