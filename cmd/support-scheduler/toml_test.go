package main

import (
	"testing"

	"gitlab.circutor.com/EDS/edgex-go/internal/pkg/config"
	"gitlab.circutor.com/EDS/edgex-go/internal/support/scheduler"
)

func TestToml(t *testing.T) {
	configuration := &scheduler.ConfigurationStruct{}
	if err := config.VerifyTomlFiles(configuration); err != nil {
		t.Fatalf("%v", err)
	}
	if configuration.Service.StartupMsg == "" {
		t.Errorf("configuration.StartupMsg is zero length.")
	}
}
