// Copyright (c) 2025 Circutor S.A. All rights reserved.

package distro

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type ScoutRestClient struct {
	baseURL string
	headers map[string]string
	client  *http.Client
}

func NewScoutRestClient(baseURL string) *ScoutRestClient {
	return &ScoutRestClient{
		baseURL: baseURL,
		headers: make(map[string]string),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *ScoutRestClient) SetHeader(key, value string) {
	c.headers[key] = value
}

func (c *ScoutRestClient) Get(path string, v interface{}) error {
	path = c.baseURL + path

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	if err != nil {
		return fmt.Errorf("could not create get request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("get request failed: %w", err)
	}

	if resp == nil {
		return errors.New("unexpected nil response")
	}

	defer resp.Body.Close()

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusAccepted) {
		return fmt.Errorf("get request rejected with error code: %d", resp.StatusCode)
	}

	if v == nil {
		return nil
	}

	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return fmt.Errorf("could not decode response: %w", err)
	}

	return nil
}

func (c *ScoutRestClient) Post(path string, v interface{}) error {
	path = c.baseURL + path

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("could not marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not download file: %w", err)
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusAccepted) {
		return fmt.Errorf("update request failed: %s", resp.Status)
	}

	return nil
}
