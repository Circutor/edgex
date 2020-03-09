//
// Copyright (c) 2017
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

const (
	tcpsPrefix      = "tcps"
	sslPrefix       = "ssl"
	tlsPrefix       = "tls"
	projectsPrefix  = "/projects/"
	locationsPrefix = "/locations/"
	devicesPrefix   = "/devices/"
)

// newIoTCoreSender returns new Google IoT Core sender instance.
func newIoTCoreSender(addr models.Addressable) sender {
	protocol := strings.ToLower(addr.Protocol)
	broker := fmt.Sprintf("%s%s", addr.GetBaseURL(), addr.Path)
	deviceID := extractDeviceID(addr.Publisher)
	projectID := extractProjectID(addr.Publisher)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetAutoReconnect(true)
	opts.SetProtocolVersion(4)

	if validateProtocol(protocol) {
		c := Configuration.Certificates["GIOT"]
		cert, err := tls.LoadX509KeyPair(c.Cert, c.Key)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Failed loading x509 data: %s", err.Error()))
			return nil
		}

		opts.SetTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
		})

		now := time.Now()
		t := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(time.Hour * 24).Unix(),
			Audience:  projectID,
		})
		password, err := t.SignedString(cert.PrivateKey)
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Could not generate JWT: %s", err.Error()))
			return nil
		}
		opts.SetPassword(password)
	}

	if addr.Topic == "" {
		addr.Topic = fmt.Sprintf("/devices/%s/events", deviceID)
	}

	return &mqttSender{
		client: MQTT.NewClient(opts),
		topic:  addr.Topic,
	}
}

func extractDeviceID(addr string) string {
	return addr[strings.Index(addr, devicesPrefix)+len(devicesPrefix):]
}

func extractProjectID(addr string) string {
	return addr[len(projectsPrefix)-1 : strings.Index(addr, locationsPrefix)]
}

func validateProtocol(protocol string) bool {
	return protocol == tcpsPrefix || protocol == sslPrefix || protocol == tlsPrefix
}
