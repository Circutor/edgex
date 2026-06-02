// Copyright (c) 2025 Circutor S.A. All rights reserved.

package distro

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Circutor/edgex/internal/pkg/correlation/models"
	"github.com/Circutor/edgex/internal/system"
	contract "github.com/Circutor/edgex/pkg/models"
	"github.com/go-stomp/stomp"
	"github.com/go-stomp/stomp/frame"
	"github.com/google/uuid"
)

type scoutDeviceInfo struct {
	DeviceID        string
	FirmwareVersion string
	HardwareVersion string
	GatewayType     string
	SerialNumber    string
	ConnectionMode  string
}

type scoutSender struct {
	address          string
	host             string
	claimID          string
	token            string
	deviceInfo       scoutDeviceInfo
	ws               *scoutWebsocket
	stomp            *stomp.Conn
	httpClient       *http.Client
	mutex            sync.Mutex
	updater          *ScoutUpdater
	registrationName string
}

type scoutEvent struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Timestamp string            `json:"ts"`
	Status    string            `json:"status,omitempty"`
	Index     string            `json:"index,omitempty"`
	Info      map[string]string `json:"info,omitempty"`
}

const (
	contentTypeText = "text/plain"
	timeFormat      = "2006-01-02T15:04:05.000Z07:00"

	mycVersion       = "1.0"
	mycMetricVersion = "1.1"

	attributeTopic = "/exchange/attributes/gateways.%s"
	metricTopic    = "/exchange/metrics/gateways.%s"
	eventTopic     = "/exchange/events/gateways.%s"

	attributeType = "ATTRIBUTE"
	metricType    = "METRIC"
	eventType     = "EVENT"
)

// newScoutSender - create new Scout Stomp sender
func newScoutSender(addr contract.Addressable, enable bool, regName string) sender {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	sender := &scoutSender{
		address:          addr.Address,
		host:             addr.Publisher,
		claimID:          addr.User,
		token:            addr.Password,
		registrationName: regName,
		deviceInfo: scoutDeviceInfo{
			DeviceID:        "MAIN",
			FirmwareVersion: system.GetVersion(),
			HardwareVersion: "1.0",
			GatewayType:     "LINE_EDS_CLOUD",
			SerialNumber:    system.GetSerialNumber(),
			ConnectionMode:  "",
		},
		httpClient: &http.Client{
			Timeout:   5 * time.Second,
			Transport: tr,
		},
		mutex: sync.Mutex{},
	}

	if enable {
		sender.Connect() // TODO: do this later! On boot the connection may not be available, we will try to reconnect later
	}

	return sender
}

// destroyScoutSender - delete old Scout Stomp sender
func destroyScoutSender(oldSender sender) {
	if old, ok := oldSender.(*scoutSender); ok {
		old.Disconnect()
		old.updater.Close()
	}
}

func (sender *scoutSender) Connect() bool {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	LoggingClient.Info("connecting to Scout")

	if sender.IsConnected() {
		LoggingClient.Error("already connected")
		return false
	}

	ws := NewScoutWebsocket()

	if err := ws.Connect(sender.address); err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to connect to websocket: %v", err))
		return false
	}

	stompConn, err := stomp.Connect(ws, stomp.ConnOpt.Login(sender.claimID, sender.token), stomp.ConnOpt.Host(sender.host))
	if err != nil {
		ws.Close()

		LoggingClient.Error(fmt.Sprintf("failed to connect to stomp: %v", err))
		return false
	}

	sender.stomp = stompConn
	sender.ws = ws

	go sender.sendScoutEvent("INFO", "EXPORT_CONNECTED", nil)
	go sender.sendDeviceAttributes()

	go sender.startReverseProxy()

	if sender.updater != nil {
		sender.updater.Close()
		sender.updater = nil
	}

	sender.updater = NewScoutUpdater(sender.address, sender.host, sender.claimID, sender.token)

	LoggingClient.Info("connected to Scout")

	return true
}

func (sender *scoutSender) Send(data []byte, event *models.Event) bool {
	if !sender.IsConnected() {
		LoggingClient.Info("not connected to Scout, trying to connect")
		if !sender.Connect() {
			LoggingClient.Error("failed to connect before sending telemetry")
			return false
		}
	}

	if event == nil || len(event.Readings) == 0 {
		LoggingClient.Warn("no event or readings to send")
		return false
	}

	destination := fmt.Sprintf(metricTopic, sender.claimID)
	dataType := metricType

	if strings.HasPrefix(event.Readings[0].Name, "EVENT_") {
		destination = fmt.Sprintf(eventTopic, sender.claimID)
		dataType = eventType
	}

	if err := sender.sendStomp(destination, dataType, data); err != nil {
		LoggingClient.Error(fmt.Sprintf("failed to send telemetry: %v", err))
		return false
	}

	return true
}

func (sender *scoutSender) IsConnected() bool {
	return sender.stomp != nil
}

func (sender *scoutSender) Disconnect() {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	if sender.stomp != nil {
		_ = sender.stomp.Disconnect()
	}

	if sender.ws != nil {
		sender.ws.Close()
	}

	sender.stomp = nil
	sender.ws = nil

	LoggingClient.Info("disconnected from Scout")
}

func (sender *scoutSender) sendDeviceAttributes() {
	type Attributes map[string]string

	d := struct {
		GatewayAttributes Attributes `json:"gateway_attributes"`
	}{
		GatewayAttributes: Attributes{
			"SERIAL_NUMBER": sender.deviceInfo.SerialNumber,
			"FW_VERSION":    sender.deviceInfo.FirmwareVersion,
			"HW_VERSION":    "1.0",
			"HW_MODEL":      "LINE_EDS_CLOUD",
		},
	}

	data, err := json.Marshal(d)
	if err != nil {
		LoggingClient.Warn("failed to marshal attributes: %w", err)
		return
	}

	destination := fmt.Sprintf(attributeTopic, sender.claimID)

	if err := sender.sendStomp(destination, attributeType, data); err != nil {
		LoggingClient.Warn("failed to send attributes: %w", err)
	}
}

func (sender *scoutSender) sendScoutEvent(status string, eventType string, infos []string) {
	d := struct {
		GatewayID string       `json:"gateway_device_id"`
		Events    []scoutEvent `json:"events"`
	}{
		GatewayID: sender.deviceInfo.DeviceID,
		Events: []scoutEvent{
			{
				ID:        uuid.New().String(),
				Type:      eventType,
				Timestamp: time.Now().Format(timeFormat),
				Status:    status,
			},
		},
	}

	if len(infos) != 0 {
		d.Events[0].Info = make(map[string]string)
		for _, info := range infos {
			values := strings.Split(info, "=")
			if len(values) == 2 {
				d.Events[0].Info[values[0]] = values[1]
			}
		}
	}

	data, err := json.Marshal(d)
	if err != nil {
		LoggingClient.Warn("failed to marshal Scout event: %w", err)
		return
	}

	destination := fmt.Sprintf(eventTopic, sender.claimID)

	if err := sender.sendStomp(destination, eventType, data); err != nil {
		LoggingClient.Warn("failed to send Scout event:", "error", err.Error())
	}
}

func (sender *scoutSender) sendStomp(destination string, messageType string, data []byte) error {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	if sender.stomp == nil {
		return fmt.Errorf("stomp connection is not established")
	}

	if !sender.IsConnected() && !sender.Connect() {
		return fmt.Errorf("failed to connect to Scout before sending data")
	}

	headers := sender.defaultHeaders(messageType)

	if err := sender.stomp.Send(destination, "application/json", data, headers...); err != nil {
		sender.Disconnect()

		return fmt.Errorf("failed to send stomp data: %w", err)
	}

	return nil
}

func (sender *scoutSender) defaultHeaders(messageType string) []func(frame *frame.Frame) error {
	messageVersion := mycVersion
	if messageType == metricType {
		messageVersion = mycMetricVersion
	}

	return []func(*frame.Frame) error{
		stomp.SendOpt.Header("myc-version", messageVersion),
		stomp.SendOpt.Header("myc-message-type", messageType),
		stomp.SendOpt.Header("myc-fw-version", sender.deviceInfo.FirmwareVersion),
		stomp.SendOpt.Header("myc-hw-version", "1.0"),
		stomp.SendOpt.Header("myc-gw-type", sender.deviceInfo.GatewayType),
		stomp.SendOpt.Header("myc-serial-number", sender.deviceInfo.SerialNumber),
		stomp.SendOpt.Header("persistent", "true"),
	}
}
