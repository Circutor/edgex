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

package command

import (
	"context"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
)

// CommandClient : client to interact with core command
type CommandClient interface {
	Get(id string, cID string, ctx context.Context) (string, error)
	Put(id string, cID string, body string, ctx context.Context) (string, error)
}

type CommandRestClient struct {
	url string
}

// NewCommandClient : Create an instance of CommandClient
func NewCommandClient(url string) CommandClient {
	c := CommandRestClient{url: url}
	return &c
}

// Get : issue GET command
func (cc *CommandRestClient) Get(id string, cID string, ctx context.Context) (string, error) {
	body, err := clients.GetRequest(cc.url+"/"+id+"/command/"+cID, ctx)
	return string(body), err
}

// Put : Issue PUT command
func (cc *CommandRestClient) Put(id string, cID string, body string, ctx context.Context) (string, error) {
	return clients.PutRequest(cc.url+"/"+id+"/command/"+cID, []byte(body), ctx)
}
