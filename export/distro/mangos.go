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

	"github.com/edgexfoundry/edgex-go/core/domain/models"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
	"go.uber.org/zap"
)

const (
	mangosPort = 5563
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

	logger.Info("Connecting to Mangos...")
	url := fmt.Sprintf("tcp://%s:%d", configuration.DataHost, mangosPort)

	q.AddTransport(ipc.NewTransport())
	q.AddTransport(tcp.NewTransport())
	if err = q.Dial(url); err != nil {
		die("can't dial on sub socket: %s", err.Error())
	}
	// Empty byte array effectively subscribes to everything
	err = q.SetOption(mangos.OptionSubscribe, []byte(""))
	if err != nil {
		die("cannot subscribe: %s", err.Error())
	}
	logger.Info("Connected to Mangos")
	for {
		if msg, erro := q.Recv(); erro != nil {
			die("Cannot recv: %s", erro.Error())
		} else {
			event := parseEvent(string(msg))
			logger.Info("Event received", zap.Any("event", event))
			eventCh <- event
		}
	}

}

func parseEvent(str string) *models.Event {
	event := models.Event{}

	if err := json.Unmarshal([]byte(str), &event); err != nil {
		logger.Error("Failed to parse event", zap.Error(err))
		return nil
	}
	return &event
}
