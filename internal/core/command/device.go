/*******************************************************************************
 * Copyright 2017 Dell Inc.
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
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/edgexfoundry/edgex-go/pkg/clients"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func issueCommand(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		LoggingClient.Error(err.Error())
	}
	return resp, err
}

func commandByDeviceID(did string, cid string, b string, p bool, ctx context.Context) (string, int) {
	d, err := mdc.Device(did, ctx)

	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return "", chk.StatusCode
		} else {
			return "", http.StatusInternalServerError
		}
	}

	if d.AdminState == models.Locked {
		LoggingClient.Error(d.Name + " is in admin locked state")

		return "", http.StatusLocked
	}

	url := d.Service.Addressable.GetBaseURL() + clients.ApiDeviceRoute + "/" + d.Id + "/" + cid
	if p {
		LoggingClient.Debug("Issuing PUT command to: " + url)
		req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(b))
		if err != nil {
			return "", http.StatusInternalServerError
		}
		resp, err := issueCommand(req)
		if err != nil {
			return "", http.StatusBadGateway
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return buf.String(), resp.StatusCode
	} else {
		LoggingClient.Debug("Issuing GET command to: " + url)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return "", http.StatusInternalServerError
		}
		resp, err := issueCommand(req)
		if err != nil {
			return "", http.StatusBadGateway
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return buf.String(), resp.StatusCode
	}
}

func putDeviceAdminState(did string, as string, ctx context.Context) (int, error) {
	err := mdc.UpdateAdminState(did, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, chk
		} else {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, err
}

func putDeviceAdminStateByName(dn string, as string, ctx context.Context) (int, error) {
	err := mdc.UpdateAdminStateByName(dn, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, chk
		} else {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, err
}

func putDeviceOpState(did string, as string, ctx context.Context) (int, error) {
	err := mdc.UpdateOpState(did, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, chk
		} else {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, err
}

func putDeviceOpStateByName(dn string, as string, ctx context.Context) (int, error) {
	err := mdc.UpdateOpStateByName(dn, as, ctx)
	if err != nil {
		LoggingClient.Error(err.Error())

		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, chk
		} else {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, err
}

func getCommands(ctx context.Context) (int, []models.CommandResponse, error) {
	devices, err := mdc.Devices(ctx)
	if err != nil {
		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, nil, chk
		} else {
			return http.StatusInternalServerError, nil, err
		}
	}
	var cr []models.CommandResponse
	for _, d := range devices {
		cr = append(cr, commandResponseFromDevice(d, Configuration.Service.Url()))
	}
	return http.StatusOK, cr, err

}

func getCommandsByDeviceID(did string, ctx context.Context) (int, models.CommandResponse, error) {
	d, err := mdc.Device(did, ctx)
	if err != nil {
		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, models.CommandResponse{}, chk
		} else {
			return http.StatusInternalServerError, models.CommandResponse{}, err
		}
	}
	return http.StatusOK, commandResponseFromDevice(d, Configuration.Service.Url()), err
}

func getCommandsByDeviceName(dn string, ctx context.Context) (int, models.CommandResponse, error) {
	d, err := mdc.DeviceForName(dn, ctx)
	if err != nil {
		chk, ok := err.(*types.ErrServiceClient)
		if ok {
			return chk.StatusCode, models.CommandResponse{}, err
		} else {
			return http.StatusInternalServerError, models.CommandResponse{}, err
		}
	}
	return http.StatusOK, commandResponseFromDevice(d, Configuration.Service.Url()), err
}

func commandResponseFromDevice(d models.Device, cmdURL string) models.CommandResponse {
	cmdResp := models.CommandResponse{
		Id:             d.Id,
		Name:           d.Name,
		AdminState:     d.AdminState,
		OperatingState: d.OperatingState,
		LastConnected:  d.LastConnected,
		LastReported:   d.LastReported,
		Labels:         d.Labels,
		Location:       d.Location,
	}

	basePath := fmt.Sprintf("%s%s/%s/command/", cmdURL, clients.ApiDeviceRoute, d.Id)

	for _, rp := range d.Profile.Resources {
		var c models.Command
		c.Name = rp.Name

		if len(rp.Get) != 0 {
			var get models.Get
			get.Path = clients.ApiDeviceRoute + "/{deviceId}/" + rp.Name
			get.URL = basePath + rp.Name
			c.Get = &get
		}
		if len(rp.Set) != 0 {
			var put models.Put
			put.Path = clients.ApiDeviceRoute + "/{deviceId}/" + rp.Name
			put.URL = basePath + rp.Name
			c.Put = &put
		}

		cmdResp.Commands = append(cmdResp.Commands, c)
	}

	return cmdResp
}
