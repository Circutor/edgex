//
// Copyright (c) 2017 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package distro

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/Circutor/edgex/internal/pkg/correlation"
	"github.com/Circutor/edgex/internal/pkg/telemetry"
	"github.com/Circutor/edgex/pkg/clients"
	"github.com/Circutor/edgex/pkg/models"
	"github.com/gorilla/mux"
)

// Test if the service is working
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func configHandler(w http.ResponseWriter, _ *http.Request) {
	encode(Configuration, w)
}

func replyNotifyRegistrations(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed read body. Error: %s", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	update := models.NotifyUpdate{}
	if err := json.Unmarshal(data, &update); err != nil {
		LoggingClient.Error(fmt.Sprintf("Failed to parse %X", data))
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}
	if update.Name == "" || update.Operation == "" {
		LoggingClient.Error(fmt.Sprintf("Missing json field: %s", update.Name))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if update.Operation != models.NotifyUpdateAdd &&
		update.Operation != models.NotifyUpdateUpdate &&
		update.Operation != models.NotifyUpdateDelete {
		LoggingClient.Error(fmt.Sprintf("Invalid value for operation %s", update.Operation))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	RefreshRegistrations(update)
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	s := telemetry.NewSystemUsage()

	encode(s, w)
}

func getRegistrationConnectedStatus(w http.ResponseWriter, _ *http.Request) {
	connectedList := GetRegistrationConnectionStatus()

	response := struct {
		ConnectedList []connectionStatusResponse `json:"connectedList"`
	}{
		ConnectedList: connectedList,
	}

	encode(response, w)
}

func sendScoutEvent(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	queryParams := r.URL.Query()
	name := queryParams.Get("name")
	offTime := queryParams.Get("offTime")
	offTimeInt := 0
	var err error

	// Validate required parameters
	if name == "" {
		LoggingClient.Error("Missing 'name' parameter")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Missing 'name' parameter")
		return
	}

	if offTime == "" {
		LoggingClient.Error("Missing 'offTime' parameter")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Missing 'offTime' parameter")
		return
	}

	// Validate offTime is a valid integer
	if offTimeInt, err = strconv.Atoi(offTime); err != nil {
		LoggingClient.Error("Invalid 'offTime' parameter: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Invalid 'offTime' parameter: "+err.Error())
		return
	}

	if !SendScoutEventToRegistration(name, offTimeInt) {
		LoggingClient.Error("Failed to send event to Scout registration")
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, "No valid Scout registration found to send event")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper function for encoding things for returning from REST calls
func encode(i interface{}, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	err := enc.Encode(i)
	// Problems encoding
	if err != nil {
		LoggingClient.Error("Error encoding the data: " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// HTTPServer function
func httpServer() http.Handler {
	r := mux.NewRouter()

	// Ping Resource
	r.HandleFunc(clients.ApiPingRoute, pingHandler).Methods(http.MethodGet)

	// Configuration
	r.HandleFunc(clients.ApiConfigRoute, configHandler).Methods(http.MethodGet)

	// Metrics
	r.HandleFunc(clients.ApiMetricsRoute, metricsHandler).Methods(http.MethodGet)

	r.HandleFunc(clients.ApiNotifyRegistrationRoute, replyNotifyRegistrations).Methods(http.MethodPut)

	// Registration connection status
	r.HandleFunc(clients.ApiScoutConnectionRoute, getRegistrationConnectedStatus).Methods(http.MethodGet)

	// Request to send a scout event
	r.HandleFunc(clients.ApiScoutEventRoute, sendScoutEvent).Methods(http.MethodPost)

	r.Use(correlation.ManageHeader)
	r.Use(correlation.OnResponseComplete)
	r.Use(correlation.OnRequestBegin)

	return r
}
