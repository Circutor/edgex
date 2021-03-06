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
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"

	"github.com/Circutor/edgex/internal/pkg/correlation/models"
	contract "github.com/Circutor/edgex/pkg/models"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type mqttSender struct {
	client MQTT.Client
	topic  string
}

// newMqttSender - create new mqtt sender
func newMqttSender(addr contract.Addressable) sender {
	protocol := strings.ToLower(addr.Protocol)

	opts := MQTT.NewClientOptions()
	broker := protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(false)

	if protocol == "tcps" || protocol == "ssl" || protocol == "tls" {
		var tlsConfig *tls.Config
		if addr.Certificate == "" {
			tlsConfig = &tls.Config{
				ClientCAs:          nil,
				InsecureSkipVerify: true,
			}
		} else {
			cert, err := tls.X509KeyPair([]byte(addr.Certificate), []byte(addr.Password))
			if err != nil {
				LoggingClient.Error(fmt.Sprintf("Failed loading x509 data: %s", err.Error()))
				return nil
			}

			tlsConfig = &tls.Config{
				ClientCAs:          nil,
				InsecureSkipVerify: true,
				Certificates:       []tls.Certificate{cert},
			}
		}

		opts.SetTLSConfig(tlsConfig)

	}

	sender := &mqttSender{
		client: MQTT.NewClient(opts),
		topic:  addr.Topic,
	}

	return sender
}

func (sender *mqttSender) Send(data []byte, event *models.Event) bool {
	if !sender.client.IsConnected() {
		LoggingClient.Info("Connecting to mqtt server")
		if token := sender.client.Connect(); token.Wait() && token.Error() != nil {
			LoggingClient.Error(fmt.Sprintf("Could not connect to mqtt server, drop event. Error: %s", token.Error().Error()))
			return false
		}
	}

	token := sender.client.Publish(sender.topic, 0, false, data)
	// FIXME: could be removed? set of tokens?
	token.Wait()
	if token.Error() != nil {
		LoggingClient.Error(token.Error().Error())
		return false
	} else {
		LoggingClient.Info(fmt.Sprintf("Sent data to mqtt server"))
		return true
	}
}
