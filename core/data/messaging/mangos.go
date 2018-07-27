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

	"github.com/edgexfoundry/edgex-go/core/domain/models"

	mangos "github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

// Configuration struct for Mangos
type MangosConfiguration struct {
	AddressPort string
}

// Mangos implementation of the event publisher
type mangosEventPublisher struct {
	publisher mangos.Socket
	mux       sync.Mutex
}

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func newMangosEventPublisher(config MangosConfiguration) mangosEventPublisher {
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
	return mangosEventPublisher{
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

	if err = mep.publisher.Send([]byte(s)); err != nil {
		die("Failed publishing: %s", err.Error())
	}

	return nil
}