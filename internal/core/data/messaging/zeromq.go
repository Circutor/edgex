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
 *******************************************************************************/
package messaging

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Circutor/edgex/internal/pkg/correlation/models"
	"github.com/go-zeromq/zmq4"
)

// ZeroMQ implementation of the event publisher
type zeroMQEventPublisher struct {
	publisher zmq4.Socket
	mux       sync.Mutex
}

func newZeroMQEventPublisher(config PubSubConfiguration) EventPublisher {
	newPublisher := zmq4.NewPub(context.Background())
	newPublisher.Listen(config.AddressPort)

	return &zeroMQEventPublisher{
		publisher: newPublisher,
	}
}

func (zep *zeroMQEventPublisher) SendEventMessage(e models.Event) error {
	s, err := json.Marshal(&e)
	if err != nil {
		return err
	}
	zep.mux.Lock()
	defer zep.mux.Unlock()

	msg := zmq4.NewMsg(s)
	return zep.publisher.Send(msg)
}
