//
// Copyright (c) 2017 Cavium
// Copyright (c) 2018 Dell Technologies, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Circutor/edgex/internal/pkg/correlation/models"
	"github.com/go-zeromq/zmq4"
)

func ZeroMQReceiver(eventCh chan *models.Event) {
	go initZmq(eventCh)
}

func initZmq(eventCh chan *models.Event) {
	sub := zmq4.NewSub(context.Background())
	defer sub.Close()

	LoggingClient.Info("Connecting to incoming 0MQ at: " + Configuration.MessageQueue.Uri())
	err := sub.Dial(Configuration.MessageQueue.Uri())
	if err != nil {
		LoggingClient.Error("Could not dial: %v", err)
		return
	}

	LoggingClient.Info("Connected to inbound 0MQ")
	err = sub.SetOption(zmq4.OptionSubscribe, "")
	if err != nil {
		LoggingClient.Error("Could not subscribe: %v", err)
		return
	}

	for {
		msg, err := sub.Recv()
		if err != nil {
			LoggingClient.Error("Could not receive message: %v", err)
		} else {
			for _, str := range msg.Frames {
				event := parseEvent(string(str))
				LoggingClient.Debug(fmt.Sprintf("Event received: %s", str))
				eventCh <- event
			}
		}
	}
}

func parseEvent(str string) *models.Event {
	event := models.Event{}

	if err := json.Unmarshal([]byte(str), &event); err != nil {
		LoggingClient.Error(err.Error())
		LoggingClient.Warn("Failed to parse event")
		return nil
	}

	return &event
}
