//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/edgexfoundry/edgex-go/internal/pkg/correlation/models"
	"nanomsg.org/go-mangos"
	"nanomsg.org/go-mangos/protocol/sub"
	"nanomsg.org/go-mangos/transport/ipc"
	"nanomsg.org/go-mangos/transport/tcp"
)

func MangosReceiver(eventCh chan *models.Event) {
	go initMangos(eventCh)
}

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func initMangos(eventCh chan *models.Event) {
	var q mangos.Socket
	var err error

	if q, err = sub.NewSocket(); err != nil {
		die("can't get new sub socket: %s", err)
	}
	defer q.Close()

	LoggingClient.Info("Connecting to incoming mangos at: " + Configuration.MessageQueue.Uri())
	q.AddTransport(ipc.NewTransport())
	q.AddTransport(tcp.NewTransport())
	if err = q.Dial(Configuration.MessageQueue.Uri()); err != nil {
		die("can't dial on sub socket: %s", err.Error())
	}
	// Empty byte array effectively subscribes to everything
	err = q.SetOption(mangos.OptionSubscribe, []byte(""))
	if err != nil {
		die("cannot subscribe: %s", err.Error())
	}
	LoggingClient.Info("Connected to Mangos")

	for {
		msg, err := q.Recv()
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Error getting message: %v", err))
		} else {
			str := string(msg)
			event := parseEvent(str)
			LoggingClient.Debug(fmt.Sprintf("Event received: %s", str))
			eventCh <- event
		}
	}
}

func parseEvent(str string) *models.Event {
	event := models.Event{}

	if err := json.Unmarshal([]byte(str), &event); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to parse event: %v", err))
		return nil
	}
	return &event
}
