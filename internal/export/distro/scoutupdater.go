// Copyright (c) 2025 Circutor S.A. All rights reserved.

package distro

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"time"
)

type ScoutUpdater struct {
	rc         ScoutRestClient
	address    string
	host       string
	claimID    string
	token      string
	lastError  string
	haveUpdate bool
	quit       chan bool
}

func NewScoutUpdater(address string, host string, claimID string, token string) *ScoutUpdater {
	m := &ScoutUpdater{
		rc:         *NewScoutRestClient("http://127.0.0.1:20003"), //check this address... shouldnt we use the one from params?
		address:    address,
		host:       host,
		claimID:    claimID,
		token:      token,
		lastError:  "",
		haveUpdate: false,
		quit:       make(chan bool),
	}

	go m.manager()

	return m
}

func (m *ScoutUpdater) Close() {
	m.quit <- true
}

func (m *ScoutUpdater) manager() {
	ticker := time.NewTicker(10 * time.Minute)

	for {
		select {
		case <-ticker.C:
			m.ifNewVersionThenUpdate()
		case <-m.quit:
			ticker.Stop()

			return
		}
	}
}

type FirmwareResponse struct {
	Status string          `json:"status"`
	Data   []FirmwareEntry `json:"data"`
}

type FirmwareEntry struct {
	GatewayType string   `json:"gateway_type"`
	HwVersion   []string `json:"hw_version"`
	MinVersion  string   `json:"min_version"`
	Version     string   `json:"version"` // validate:"semver"`
	Sha256      string   `json:"sha256"`  // validate:"sha256"`
	URL         string   `json:"url"`     // validate:"http_url"`
	Status      string   `json:"status"`
}

func (fe FirmwareEntry) IsValid() error {
	/*if err := validator.New().Struct(fe); err != nil {
		return fmt.Errorf("configuration is not valid: %w", err)
	}
	*/
	if _, err := regexp.Compile(`\d+\.\d+\.\d+`); err != nil {
		return fmt.Errorf("configuration is not valid: %w", err)
	}

	if !isValidSHA256Format(fe.Sha256) {
		return fmt.Errorf("configuration is not valid: sha format invalid")
	}

	if _, err := url.Parse(fe.URL); err != nil {
		return fmt.Errorf("configuration is not valid: %w", err)
	}

	return nil
}

type UpdateRequestFromURL struct {
	Version string `json:"version"`
	Sha256  string `json:"sha256"`
	URL     string `json:"url"`
}

func (m *ScoutUpdater) ifNewVersionThenUpdate() {
	if m.haveUpdate {
		return
	}

	firmware, err := m.getFirmware()
	if err != nil {
		m.logIfNewError("Could not get firmware", err)

		return
	}

	urfu := UpdateRequestFromURL{
		Version: firmware.Version,
		Sha256:  firmware.Sha256,
		URL:     firmware.URL,
	}

	if err = m.rc.Post("/api/v1/update/fromUrl", urfu); err != nil {
		m.logIfNewError("Update request failed", err)

		return
	}

	LoggingClient.Info("Update requested", "version", firmware.Version, "url", firmware.URL)
	m.haveUpdate = true
}

func (m *ScoutUpdater) logIfNewError(msg string, err error) {
	if m.lastError == err.Error() {
		return
	}

	m.lastError = err.Error()
	LoggingClient.Warn(msg, "error", m.lastError)
}

func (m *ScoutUpdater) getFirmware() (*FirmwareEntry, error) {
	url, err := m.getURL()
	if err != nil {
		return nil, fmt.Errorf("could not get URL: %w", err)
	}

	restclient := NewScoutRestClient(url)
	restclient.SetHeader("GATEWAY-TOKEN", m.token)

	var firmware FirmwareResponse

	if err = restclient.Get("", &firmware); err != nil {
		return nil, fmt.Errorf("could not get firmware: %w", err)
	}

	if len(firmware.Data) == 0 {
		return nil, errors.New("no firmware available")
	}

	if err = firmware.Data[0].IsValid(); err != nil {
		return nil, fmt.Errorf("firmware is not valid: %w", err)
	}

	return &firmware.Data[0], nil
}

func (m *ScoutUpdater) getURL() (string, error) {
	parsedURL, err := url.Parse(m.address)
	if err != nil {
		return "", fmt.Errorf("could not parse URL: %w", err)
	}

	baseURL := url.URL{
		Scheme: "https",
		Host:   parsedURL.Host,
		Path:   "api/v2/gateway/" + m.claimID + "/firmware/pending",
	}

	return baseURL.String(), nil
}

func isValidSHA256Format(hash string) bool {
	if len(hash) != sha256.Size*2 { // SHA256 is 32 bytes, hex encoded is 64 chars
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}
