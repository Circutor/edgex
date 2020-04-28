//
// Copyright (c) 2018
// Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"testing"

	"gitlab.circutor.com/EDS/edgex-go/internal/core/data"
	"gitlab.circutor.com/EDS/edgex-go/internal/pkg/config"
)

func TestToml(t *testing.T) {
	configuration := &data.ConfigurationStruct{}
	if err := config.VerifyTomlFiles(configuration); err != nil {
		t.Fatalf("%v", err)
	}
	if configuration.Service.StartupMsg == "" {
		t.Errorf("configuration.StartupMsg is zero length.")
	}
}
