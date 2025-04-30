// Copyright (c) 2025 Circutor S.A. All rights reserved.

package distro

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-stomp/stomp"
	"github.com/go-stomp/stomp/frame"
)

const (
	messageTTL   = "5000"  // 5 seconds for message TTL
	queueExpires = "15000" // 15 seconds for queue auto-expiry
)

func (sender *scoutSender) startReverseProxy() {
	requestQueue := sender.claimID + "-proxy-request"

	sub, err := sender.subscribe(requestQueue)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to subscribe to request queue: %v", err))
		sender.Disconnect()

		return
	}

	LoggingClient.Info(fmt.Sprintf("Subscribed to request queue: %v", requestQueue))

	for msg := range sub.C {
		if msg.Err != nil {
			LoggingClient.Error(fmt.Sprintf("Error receiving message: %v", msg.Err))
			sender.Disconnect()

			return
		}

		var httpMsg *HTTPMessage

		LoggingClient.Info(fmt.Sprintf("Received message to destination: %v", msg.Destination))

		httpMsg, err := ParseHTTPRequest(string(msg.Body))
		if err != nil {
			LoggingClient.Error(fmt.Sprintf("Failed to parse HTTP request: %v", err))

			continue
		}

		httpMsg.RequestID = msg.Header.Get("request-id")

		go sender.handleRequest(httpMsg)
	}
}

func (sender *scoutSender) subscribe(requestQueue string) (*stomp.Subscription, error) {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	if sender.stomp == nil {
		return nil, errors.New("STOMP client is not connected")
	}

	sub, err := sender.stomp.Subscribe(requestQueue, stomp.AckAuto)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to queue %s: %w", requestQueue, err)
	}

	return sub, nil
}

func (sender *scoutSender) handleRequest(msg *HTTPMessage) {
	req, err := msg.ToHTTPRequest("https://127.0.0.1/")
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to convert HTTP message to request: %v", err))
		sender.sendError(msg.RequestID, err)

		return
	}

	req.Header.Set("Accept-Encoding", "identity")

	resp, err := sender.httpClient.Do(req)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error sending request: %v", err))
		sender.sendError(msg.RequestID, err)

		return
	}
	defer resp.Body.Close()

	sender.processResponse(msg.RequestID, resp)
}

func (sender *scoutSender) processResponse(requestID string, resp *http.Response) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error reading response body: %v", err))
		sender.sendError(requestID, err)

		return
	}

	originalContentType := resp.Header.Get("Content-Type")
	var responseBody string

	if isBinaryContent(originalContentType) {
		responseBody = base64.StdEncoding.EncodeToString(body)

		resp.Header.Set("X-Content-Transfer-Encoding", "base64")
		resp.Header.Set("X-Original-Content-Type", originalContentType)
	} else {
		responseBody = string(body)
	}

	resp.Header.Set("Content-Length", strconv.Itoa(len(responseBody)))

	httpResp := HTTPResponse{
		RequestID:     requestID,
		StatusCode:    resp.StatusCode,
		StatusMessage: resp.Status,
		Headers:       resp.Header,
		Body:          []byte(responseBody),
	}

	sender.sendReply(httpResp)
}

func (sender *scoutSender) sendError(requestID string, err error) {
	httpResp := HTTPResponse{
		RequestID:     requestID,
		StatusCode:    http.StatusInternalServerError,
		StatusMessage: "Internal Server Error",
		Body:          []byte(err.Error()),
	}

	sender.sendReply(httpResp)
}

func (sender *scoutSender) sendReply(response HTTPResponse) {
	sender.mutex.Lock()
	defer sender.mutex.Unlock()

	if sender.stomp == nil {
		LoggingClient.Error("STOMP client is not connected")

		return
	}

	responseQueue := "/queue/" + response.RequestID + "-proxy-response"

	err := sender.stomp.Send(responseQueue, contentTypeText, []byte(response.FormatHTTPResponse()), getStandardHeaderOpts(response.RequestID)...)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Error sending reply: %v", err))

		return
	}

	LoggingClient.Info(fmt.Sprintf("Reply sent: requestID %v statusCode %v", response.RequestID, response.StatusCode))
}

func getStandardHeaderOpts(requestID string) []func(*frame.Frame) error {
	return []func(*frame.Frame) error{
		stomp.SendOpt.Header("request-id", requestID),
		stomp.SendOpt.Header("auto-delete", "true"),
		stomp.SendOpt.Header("durable", "false"),
		stomp.SendOpt.Header("exclusive", "false"),
		stomp.SendOpt.Header("x-expires", queueExpires),
		stomp.SendOpt.Header("x-message-ttl", messageTTL),
	}
}

func isBinaryContent(contentType string) bool {
	return strings.Contains(contentType, "font") ||
		strings.Contains(contentType, "image") ||
		strings.Contains(contentType, "audio") ||
		strings.Contains(contentType, "video") ||
		strings.Contains(contentType, "application/octet-stream") ||
		strings.Contains(contentType, "application/pdf")
}
