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
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/Circutor/edgex/internal"
	"github.com/Circutor/edgex/internal/pkg/correlation/models"
	"github.com/Circutor/edgex/pkg/clients"
	contract "github.com/Circutor/edgex/pkg/models"
)

type sentiloSender struct {
	url    string
	method string
	token  string
}

// newSentiloSender - create sentilo http sender
func newSentiloSender(addr contract.Addressable) sender {
	sender := sentiloSender{
		url:    addr.Protocol + "://" + addr.Address + addr.Path,
		method: addr.HTTPMethod,
		token:  addr.Password,
	}

	return sender
}

// Send will send the optionally filtered, compressed, encypted contract.Event via HTTP POST
// The model.Event is provided in order to obtain the necessary correlation-id.
func (sender sentiloSender) Send(data []byte, event *models.Event) bool {

	switch sender.method {
	case http.MethodPut:
		req, err := http.NewRequest(http.MethodPut, sender.url, bytes.NewReader(data))
		if err != nil {
			return false
		}
		req.Header.Set("Content-Type", mimeTypeJSON)
		req.Header.Set("IDENTITY_KEY", sender.token)

		client := &http.Client{}
		begin := time.Now()

		response, err := client.Do(req)
		if err != nil {
			LoggingClient.Error(err.Error(), clients.CorrelationHeader, event.CorrelationId, internal.LogDurationKey, time.Since(begin).String())
			return false
		}
		defer response.Body.Close()
		LoggingClient.Info(fmt.Sprintf("Pushed event correctly to %s, response: %s", req.URL.Host, response.Status), clients.CorrelationHeader, event.CorrelationId, internal.LogDurationKey, time.Since(begin).String())
	default:
		LoggingClient.Info(fmt.Sprintf("Unsupported method: %s", sender.method))
		return false
	}

	LoggingClient.Debug(fmt.Sprintf("Sent data: %X", data))
	return true
}
