//
// Copyright (c) 2017
// Cavium
// Mainflux
// IOTech
//
// SPDX-License-Identifier: Apache-2.0

package distro

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	contract "github.com/Circutor/edgex/pkg/models"
	"github.com/google/uuid"
)

type jsonFormatter struct {
}

// Azure IoT message feedback codes
type feedbackCode int

const scaleKilo = 0.001
const (
	none feedbackCode = iota
)

func (jsonTr jsonFormatter) Format(event *contract.Event) []byte {

	b, err := json.Marshal(event)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing JSON. Error: %s", err.Error()))
		return nil
	}
	return b
}

type xmlFormatter struct {
}

func (xmlTr xmlFormatter) Format(event *contract.Event) []byte {
	b, err := xml.Marshal(event)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing XML. Error: %s", err.Error()))
		return nil
	}
	return b
}

type thingsboardJSONFormatter struct {
}

// ThingsBoard JSON formatter
// https://thingsboard.io/docs/reference/gateway-mqtt-api/#telemetry-upload-api
func (thingsboardjsonTr thingsboardJSONFormatter) Format(event *contract.Event) []byte {

	type Device struct {
		Ts     int64             `json:"ts"`
		Values map[string]string `json:"values"`
	}

	values := make(map[string]string)
	for _, reading := range event.Readings {
		values[reading.Name] = reading.Value
	}

	var deviceValues []Device
	deviceValues = append(deviceValues, Device{Ts: event.Origin, Values: values})

	device := make(map[string][]Device)
	device[event.Device] = deviceValues

	b, err := json.Marshal(device)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing ThingsBoard JSON. Error: %s", err.Error()))
		return nil
	}
	return b
}

type dexmaJSONFormatter struct {
}

// Dexma JSON formatter
// https://support.dexma.com/hc/es/articles/360013772759-API-de-inserci%C3%B3n
func (dexmajsonTr dexmaJSONFormatter) Format(event *contract.Event) []byte {

	type Value struct {
		P int     `json:"p"`
		V float64 `json:"v"`
	}

	type Device struct {
		Did    string  `json:"did"`
		Sqn    int     `json:"sqn"`
		Ts     string  `json:"ts"`
		Values []Value `json:"values"`
	}

	var value Value
	var values []Value

	for _, reading := range event.Readings {
		var err error
		value.P, err = strconv.Atoi(reading.Name)
		if err != nil {
			value.P = transformDexmaParam(reading.Name)
		}
		if value.P == 0 {
			LoggingClient.Error(fmt.Sprintf("Error on Dexma parameter %s: invalid name", reading.Name))
		} else {
			value.V, err = strconv.ParseFloat(reading.Value, 64)
			if err != nil {
				LoggingClient.Error(fmt.Sprintf("Error on Dexma parameter %s: could not parse value %s", reading.Name, reading.Value))
			} else {
				values = append(values, value)
			}
		}
	}
	if len(values) == 0 {
		return nil
	}

	var devices []Device
	time := time.Unix(event.Origin/1000, 0).Format(time.RFC3339)
	devices = append(devices, Device{Did: event.Device, Sqn: 1, Ts: time, Values: values})

	b, err := json.Marshal(devices)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing Dexma JSON. Error: %s", err.Error()))
		return nil
	}
	return b
}

type prosumeJSONFormatter struct {
}

// Prosume JSON formatter
// https://prosume.io
func (prosumeJson prosumeJSONFormatter) Format(event *contract.Event) []byte {
	var err error
	type prosumeData struct {
		Timestamp      int64   `json:"Timestamp"`
		ImportedEnergy float64 `json:"Active_energy_imported_kWh"`
		ExportedEnergy float64 `json:"Active_energy_exported_kWh"`
	}

	data := prosumeData{}
	for _, reading := range event.Readings {
		if strings.Contains(reading.Name, "ENERGY_P_CON_TOT_ABS") {
			tempInt, _ := strconv.ParseUint(reading.Value, 10, 64)
			data.ImportedEnergy = float64(tempInt) * scaleKilo
			data.Timestamp = event.Created
		} else if strings.Contains(reading.Name, "ENERGY_P_GEN_TOT_ABS") {
			tempInt, _ := strconv.ParseUint(reading.Value, 10, 64)
			data.ExportedEnergy = float64(tempInt) * scaleKilo
			data.Timestamp = event.Created
		}
	}

	if data.ImportedEnergy == 0 && data.ExportedEnergy == 0 {
		return nil
	}

	if data.Timestamp == 0 {
		data.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	}

	b, err := json.Marshal(data)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing Prosume JSON. Error: %s", err.Error()))
		return nil
	}
	return b
}

type myCircutorJSONFormatter struct {
}

// MyCircutor JSON formatter
// https://www.mycircutor.com
func (circutorJson myCircutorJSONFormatter) Format(event *contract.Event) []byte {
	var err error
	type myCircutorData struct {
		ID         string           `json:"id"`
		Ts         string           `json:"ts"`
		Telemetry  map[string]int64 `json:"telemetry"`
		DeviceType string           `json:"deviceType"`
	}

	data := myCircutorData{}
	data.ID = event.Device
	if event.Origin == 0 {
		data.Ts = time.Now().Format(time.RFC3339)
	} else {
		data.Ts = time.Unix(0, event.Origin*int64(time.Millisecond)).Format(time.RFC3339)
	}

	data.Telemetry = make(map[string]int64)
	for _, reading := range event.Readings {
		data.Telemetry[reading.Name], _ = strconv.ParseInt(reading.Value, 10, 64)
	}

	if len(event.Device) > 0 {
		data.DeviceType = "DEVICE"
	} else {
		data.DeviceType = "GATEWAY"
	}

	b, err := json.Marshal(data)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error MyCircutor JSON. Error: %s", err.Error()))
		return nil
	}

	return b
}

type scoutJSONFormatter struct {
}

// Scout JSON Formatter
func (scoutJson scoutJSONFormatter) Format(event *contract.Event) []byte {
	type Metrics map[string]float32

	type Telemetry struct {
		Time    string  `json:"ts"`
		Metrics Metrics `json:"data"`
	}

	type Event struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Time   string `json:"ts"`
		Status string `json:"status"`
		Index  string `json:"index"`
	}

	sendData := struct {
		DeviceID  string      `json:"gateway_device_id"`
		Telemetry []Telemetry `json:"metrics,omitempty"`
		Events    []Event     `json:"events,omitempty"`
	}{
		DeviceID:  event.Device,
		Telemetry: make([]Telemetry, 0, len(event.Readings)),
		Events:    make([]Event, 0, len(event.Readings)),
	}

	ts := time.Now().Format(timeFormat)
	if event.Created != 0 {
		ts = time.Unix(0, event.Created*int64(time.Millisecond)).Format(timeFormat)
	} else if event.Origin != 0 {
		ts = time.Unix(0, event.Origin*int64(time.Millisecond)).Format(timeFormat)
	}

	sendData.Telemetry = append(sendData.Telemetry, Telemetry{
		Time:    ts,
		Metrics: Metrics{},
	})

	for _, r := range event.Readings {
		if strings.HasPrefix(r.Name, "EVENT_") {
			val := "OFF"
			if r.Value == "true" {
				val = "ON"
			}

			idx := "NONE"

			spl := strings.Split(r.Name, ".")

			if len(spl) >= 2 {
				_, err := strconv.Atoi(spl[len(spl)-1])
				if err == nil {
					idx = spl[len(spl)-1]
				}
			}

			sendData.Events = append(sendData.Events, Event{
				ID:     event.ID,
				Type:   strings.TrimPrefix(r.Name, "EVENT_"),
				Time:   ts,
				Status: val,
				Index:  idx,
			})

			continue
		}

		val, err := strconv.ParseFloat(r.Value, 32)
		if err == nil {
			sendData.Telemetry[0].Metrics[r.Name] = float32(val)
		}

		if r.AvgValue != "" {
			avgVal, err := strconv.ParseFloat(r.AvgValue, 32)
			if err == nil {
				sendData.Telemetry[0].Metrics[r.Name+"_AVG_10m"] = float32(avgVal)
			}
		}

		if r.MaxValue != "" {
			maxVal, err := strconv.ParseFloat(r.MaxValue, 32)
			if err == nil {
				sendData.Telemetry[0].Metrics[r.Name+"_MAX_10m"] = float32(maxVal)
			}
		}

		if r.MinValue != "" {
			minVal, err := strconv.ParseFloat(r.MinValue, 32)
			if err == nil {
				sendData.Telemetry[0].Metrics[r.Name+"_MIN_10m"] = float32(minVal)
			}
		}
	}

	if len(sendData.Telemetry[0].Metrics) == 0 {
		sendData.Telemetry = nil
	}

	b, err := json.Marshal(sendData)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error Scout JSON. Error: %s", err.Error()))
		return nil
	}

	return b
}

type sentiloJSONFormatter struct {
	location string
}

// Sentilo JSON formatter
// https://sentilo.readthedocs.io/en/latest/api_docs/general_model.html
func (sentilo sentiloJSONFormatter) Format(event *contract.Event) []byte {
	var err error

	type observation struct {
		Value     string `json:"value"`
		Timestamp string `json:"timestamp"`
	}

	type sensor struct {
		Sensor       string        `json:"sensor"`
		Location     string        `json:"location"`
		Observations []observation `json:"observations"`
	}

	type sensors struct {
		Sensors []sensor `json:"sensors"`
	}

	data := sensors{}

	data.Sensors = make([]sensor, len(event.Readings))
	for i, reading := range event.Readings {
		data.Sensors[i].Sensor = reading.Name
		data.Sensors[i].Location = sentilo.location
		data.Sensors[i].Observations = make([]observation, 1)
		data.Sensors[i].Observations[0].Value = reading.Value
		if event.Created != 0 {
			t := time.Unix(0, event.Created*int64(time.Millisecond))
			z, _ := t.Zone()
			data.Sensors[i].Observations[0].Timestamp = fmt.Sprintf("%02d/%02d/%dT%02d:%02d:%02d%s\n", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second(), z)
		} else if event.Origin != 0 {
			t := time.Unix(0, event.Origin*int64(time.Millisecond))
			z, _ := t.Zone()
			data.Sensors[i].Observations[0].Timestamp = fmt.Sprintf("%02d/%02d/%dT%02d:%02d:%02d%s\n", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second(), z)
		} else {
			t := time.Now()
			z, _ := t.Zone()
			data.Sensors[i].Observations[0].Timestamp = fmt.Sprintf("%02d/%02d/%dT%02d:%02d:%02d%s\n", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute(), t.Second(), z)
		}
	}

	b, err := json.Marshal(data)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error Sentilo JSON. Error: %s", err.Error()))
		return nil
	}
	return b
}

// Azure IoT Hub message
// https://docs.microsoft.com/en-us/azure/iot-hub/iot-hub-devguide-messages-construct
type connAuthMethod struct {
	Scope  string `json:"scope"`
	Type   string `json:"type"`
	Issuer string `json:"issuer"`
}

// AzureMessage represents Azure IoT Hub message.
type AzureMessage struct {
	ID             string            `json:"id"`
	SequenceNumber int64             `json:"sequenceNumber"`
	To             string            `json:"To"`
	Created        time.Time         `json:"CreationTimeUtc"`
	Expire         time.Time         `json:"ExpiryTimeUtc"`
	Enqueued       time.Time         `json:"EnqueuedTime"`
	CorrelationID  string            `json:"CorrelationId"`
	UserID         string            `json:"userId"`
	Ack            feedbackCode      `json:"ack"`
	ConnDevID      string            `json:"connectionDeviceId"`
	ConnDevGenID   string            `json:"connectionDeviceGenerationId"`
	ConnAuthMethod connAuthMethod    `json:"connectionAuthMethod,omitempty"`
	Body           []byte            `json:"body"`
	Properties     map[string]string `json:"properties"`
}

// newAzureMessage creates a new Azure message and sets
// Body and default fields values.
func newAzureMessage() (*AzureMessage, error) {
	msg := &AzureMessage{
		Ack:        none,
		Properties: make(map[string]string),
		Created:    time.Now(),
	}

	id := uuid.New()
	msg.ID = id.String()

	correlationID := uuid.New()
	msg.CorrelationID = correlationID.String()

	return msg, nil
}

// AddProperty method ads property performing key check.
func (am *AzureMessage) AddProperty(key, value string) error {
	am.Properties[key] = value
	return nil
}

// azureFormatter is used to convert Event to Azure message and
// Azure message to bytes.
type azureFormatter struct {
}

// Format method does all foramtting job.
func (af azureFormatter) Format(event *contract.Event) []byte {
	am, err := newAzureMessage()
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error creating a new Azure message: %s", err))
		return []byte{}
	}
	am.ConnDevID = event.Device
	am.UserID = fmt.Sprint(event.Origin)
	data, err := json.Marshal(event)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing Event data: %s", err))
		return []byte{}
	}
	am.Body = data
	msg, err := json.Marshal(am)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error parsing Azure Message data: %s", err))
		return []byte{}
	}
	return msg
}

// converting event to AWS shadow message in bytes
type awsFormatter struct {
}

func (af awsFormatter) Format(event *contract.Event) []byte {
	reported := map[string]interface{}{}

	for _, reading := range event.Readings {
		value, err := strconv.ParseFloat(reading.Value, 64)

		if err != nil {
			strVal := reading.Value
			// not a valid numerical reading value, see if it's boolean
			if strings.Compare(strings.ToLower(strVal), "true") == 0 {
				reported[reading.Name] = true
			} else if strings.Compare(strings.ToLower(strVal), "false") == 0 {
				reported[reading.Name] = false
			} else {
				reported[reading.Name] = strVal
			}

			continue
		}

		reported[reading.Name] = value
	}

	currState := map[string]interface{}{
		"state": map[string]interface{}{
			"reported": reported,
		},
	}

	msg, err := json.Marshal(currState)

	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error generating AWS shadow document: %s", err))
		return []byte{}
	}

	return msg
}

type noopFormatter struct {
}

func (noopFmt noopFormatter) Format(event *contract.Event) []byte {
	return []byte{}
}

// BIoTMessage represents Brightics IoT(Samsung SDS IoT platform)  messages.
type BIoTMessage struct {
	Version    string `json:"version"`
	MsgType    string `json:"msgType"`
	FuncType   string `json:"funcType"`
	SId        string `json:"sId"`
	TpId       string `json:"tpId"`
	TId        string `json:"tId"`
	MsgCode    string `json:"msgCode"`
	MsgId      string `json:"msgId"`
	MsgDate    int64  `json:"msgDate"`
	ResCode    string `json:"resCode"`
	ResMsg     string `json:"resMsg"`
	Severity   string `json:"severity"`
	Dataformat string `json:"dataformat"`
	EncType    string `json:"encType"`
	AuthToken  string `json:"authToken"`
	Data       []byte `json:"data"`
}

// newBIoTMessage creates a new Brightics IoT message and sets
// Body and default fields values.
func newBIoTMessage() (*BIoTMessage, error) {
	msg := &BIoTMessage{
		Severity: "1",
		MsgType:  "Q",
	}

	id := uuid.New()
	msg.MsgId = id.String()

	return msg, nil
}

// brighticsiotFormatter is used to convert Event to BIoT message and
// BIoT message to bytes.
type biotFormatter struct {
}

// Format method does all foramtting job.
func (af biotFormatter) Format(event *contract.Event) []byte {
	bm, err := newBIoTMessage()
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error creating a new BIoT message: %s", err))
		return []byte{}
	}
	bm.TpId = event.Device
	bm.TId = fmt.Sprint(event.Origin)
	rawdata, err := json.Marshal(event)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error parsing Event data to BIoTMessage : %s", err))
		return []byte{}
	}
	bm.Data = rawdata
	msg, err := json.Marshal(bm)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("error parsing BIoTMessage to data: %s", err))
		return []byte{}
	}
	return msg
}
