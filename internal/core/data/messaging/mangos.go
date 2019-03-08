/*******************************************************************************
 * Copyright 2017 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * @microservice: core-data-go library
 * @author: Ryan Comer, Dell
 * @version: 0.5.0
 *******************************************************************************/
package messaging

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	"nanomsg.org/go-mangos"
	"nanomsg.org/go-mangos/protocol/pub"
	"nanomsg.org/go-mangos/transport/ipc"
	"nanomsg.org/go-mangos/transport/tcp"
)

// Mangos implementation of the event publisher
type mangosEventPublisher struct {
	publisher mangos.Socket
	mux       sync.Mutex
}

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func newMangosEventPublisher(config PubSubConfiguration) EventPublisher {
	var newPublisher mangos.Socket
	var err error

	if newPublisher, err = pub.NewSocket(); err != nil {
		die("can't get new pub socket: %s", err)
	}
	newPublisher.AddTransport(ipc.NewTransport())
	newPublisher.AddTransport(tcp.NewTransport())

	url := fmt.Sprintf(config.AddressPort)
	if err = newPublisher.Listen(url); err != nil {
		die("can't listen on pub socket: %s", err.Error())
	}
	return &mangosEventPublisher{
		publisher: newPublisher,
	}
}

func (mep *mangosEventPublisher) SendEventMessage(e models.Event) error {
	s, err := json.Marshal(&e)
	if err != nil {
		return err
	}
	mep.mux.Lock()
	defer mep.mux.Unlock()

	err = mep.publisher.Send([]byte(s))
	if err != nil {
		return err
	}

	return nil
}
