/*******************************************************************************
 * Copyright 2019 Dell Inc.
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

package notifications

import (
	"context"

	"github.com/Circutor/edgex/pkg/clients"
)

type CategoryEnum string

const (
	SECURITY  CategoryEnum = "SECURITY"
	HW_HEALTH CategoryEnum = "HW_HEALTH"
	SW_HEALTH CategoryEnum = "SW_HEALTH"
)

type SeverityEnum string

const (
	CRITICAL SeverityEnum = "CRITICAL"
	NORMAL   SeverityEnum = "NORMAL"
)

type StatusEnum string

const (
	NEW       StatusEnum = "NEW"
	PROCESSED StatusEnum = "PROCESSED"
	ESCALATED StatusEnum = "ESCALATED"
)

// Interface defining behavior for the notifications client
type NotificationsClient interface {
	SendNotification(n Notification, ctx context.Context) error
}

// Type struct for REST-specific implementation of the NotificationsClient interface
type notificationsRestClient struct {
	url string
}

// Struct to represent a notification being sent to the notifications service
type Notification struct {
	Id          string       `json:"id,omitempty"` // Generated by the system, users can ignore
	Slug        string       `json:"slug"`         // A meaningful identifier provided by client
	Sender      string       `json:"sender"`
	Category    CategoryEnum `json:"category"`
	Severity    SeverityEnum `json:"severity"`
	Content     string       `json:"content"`
	Description string       `json:"description,omitempty"`
	Status      StatusEnum   `json:"status,omitempty"`
	Labels      []string     `json:"labels,omitempty"`
	Created     int          `json:"created,omitempty"`  // The creation timestamp
	Modified    int          `json:"modified,omitempty"` // The last modification timestamp
}

func NewNotificationsClient(url string) NotificationsClient {
	n := notificationsRestClient{url: url}
	return &n
}

// Send a notification to the notifications service
func (nc *notificationsRestClient) SendNotification(n Notification, ctx context.Context) error {
	_, err := clients.PostJsonRequest(nc.url, n, ctx)
	return err
}
