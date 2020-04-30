/*******************************************************************************
 * Copyright 1995-2018 Hitachi Vantara Corporation. All rights reserved.
 *
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
 *
 *******************************************************************************/

package coredata

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/Circutor/edgex/pkg/clients"
	"github.com/Circutor/edgex/pkg/models"
)

type ReadingClient interface {
	Readings(ctx context.Context) ([]models.Reading, error)
	ReadingCount(ctx context.Context) (int, error)
	Reading(id string, ctx context.Context) (models.Reading, error)
	ReadingsForDevice(deviceId string, limit int, ctx context.Context) ([]models.Reading, error)
	ReadingsForNameAndDevice(name string, deviceId string, limit int, ctx context.Context) ([]models.Reading, error)
	ReadingsForName(name string, limit int, ctx context.Context) ([]models.Reading, error)
	ReadingsForUOMLabel(uomLabel string, limit int, ctx context.Context) ([]models.Reading, error)
	ReadingsForLabel(label string, limit int, ctx context.Context) ([]models.Reading, error)
	ReadingsForType(readingType string, limit int, ctx context.Context) ([]models.Reading, error)
	ReadingsForInterval(start int, end int, limit int, ctx context.Context) ([]models.Reading, error)
	Add(readiing *models.Reading, ctx context.Context) (string, error)
	Delete(id string, ctx context.Context) error
}

type ReadingRestClient struct {
	url string
}

func NewReadingClient(url string) ReadingClient {
	r := ReadingRestClient{url: url}
	return &r
}

// Helper method to request and decode a reading slice
func (r *ReadingRestClient) requestReadingSlice(url string, ctx context.Context) ([]models.Reading, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return []models.Reading{}, err
	}

	rSlice := make([]models.Reading, 0)
	err = json.Unmarshal(data, &rSlice)
	return rSlice, err
}

// Helper method to request and decode a reading
func (r *ReadingRestClient) requestReading(url string, ctx context.Context) (models.Reading, error) {
	data, err := clients.GetRequest(url, ctx)
	if err != nil {
		return models.Reading{}, err
	}

	reading := models.Reading{}
	err = json.Unmarshal(data, &reading)
	return reading, err
}

// Get a list of all readings
func (r *ReadingRestClient) Readings(ctx context.Context) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url, ctx)
}

// Get the reading by id
func (r *ReadingRestClient) Reading(id string, ctx context.Context) (models.Reading, error) {
	return r.requestReading(r.url+"/"+id, ctx)
}

// Get reading count
func (r *ReadingRestClient) ReadingCount(ctx context.Context) (int, error) {
	return clients.CountRequest(r.url+"/count", ctx)
}

// Get the readings for a device
func (r *ReadingRestClient) ReadingsForDevice(deviceId string, limit int, ctx context.Context) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url+"/device/"+url.QueryEscape(deviceId)+"/"+strconv.Itoa(limit), ctx)
}

// Get the readings for name and device
func (r *ReadingRestClient) ReadingsForNameAndDevice(name string, deviceId string, limit int, ctx context.Context) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url+"/name/"+url.QueryEscape(name)+"/device/"+url.QueryEscape(deviceId)+"/"+strconv.Itoa(limit), ctx)
}

// Get readings by name
func (r *ReadingRestClient) ReadingsForName(name string, limit int, ctx context.Context) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url+"/name/"+url.QueryEscape(name)+"/"+strconv.Itoa(limit), ctx)
}

// Get readings for UOM Label
func (r *ReadingRestClient) ReadingsForUOMLabel(uomLabel string, limit int, ctx context.Context) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url+"/uomlabel/"+url.QueryEscape(uomLabel)+"/"+strconv.Itoa(limit), ctx)
}

// Get readings for label
func (r *ReadingRestClient) ReadingsForLabel(label string, limit int, ctx context.Context) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url+"/label/"+url.QueryEscape(label)+"/"+strconv.Itoa(limit), ctx)
}

// Get readings for type
func (r *ReadingRestClient) ReadingsForType(readingType string, limit int, ctx context.Context) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url+"/type/"+url.QueryEscape(readingType)+"/"+strconv.Itoa(limit), ctx)
}

// Get readings for interval
func (r *ReadingRestClient) ReadingsForInterval(start int, end int, limit int, ctx context.Context) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url+"/"+strconv.Itoa(start)+"/"+strconv.Itoa(end)+"/"+strconv.Itoa(limit), ctx)
}

// Get readings for device and value descriptor
func (r *ReadingRestClient) ReadingsForDeviceAndValueDescriptor(deviceId string, vd string, limit int, ctx context.Context) ([]models.Reading, error) {
	return r.requestReadingSlice(r.url+"/device/"+url.QueryEscape(deviceId)+"/valuedescriptor/"+url.QueryEscape(vd)+"/"+strconv.Itoa(limit), ctx)
}

// Add a reading
func (r *ReadingRestClient) Add(reading *models.Reading, ctx context.Context) (string, error) {
	return clients.PostJsonRequest(r.url, reading, ctx)
}

// Delete a reading by id
func (r *ReadingRestClient) Delete(id string, ctx context.Context) error {
	return clients.DeleteRequest(r.url+"/id/"+id, ctx)
}
