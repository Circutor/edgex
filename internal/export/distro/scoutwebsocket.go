// Copyright (c) 2024 Circutor S.A. All rights reserved.
package distro

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type scoutWebsocket struct {
	connection *websocket.Conn
	readBuffer []byte
}

func NewScoutWebsocket() *scoutWebsocket {
	return &scoutWebsocket{}
}

func (w *scoutWebsocket) Connect(url string) error {
	con, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to websocket: %w", err)
	}

	w.connection = con

	return nil
}

func (w *scoutWebsocket) Read(p []byte) (int, error) {
	if len(w.readBuffer) == 0 {
		_, msg, err := w.connection.ReadMessage()
		if err != nil {
			return 0, fmt.Errorf("failed to read message: %w", err)
		}

		w.readBuffer = msg
	}

	n := copy(p, w.readBuffer)
	w.readBuffer = w.readBuffer[n:]

	return n, nil
}

func (w *scoutWebsocket) Write(p []byte) (int, error) {
	err := w.connection.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, fmt.Errorf("failed to write message: %w", err)
	}

	return len(p), nil
}

func (w *scoutWebsocket) Close() error {
	if err := w.connection.Close(); err != nil {
		return fmt.Errorf("failed to close websocket: %w", err)
	}

	return nil
}
