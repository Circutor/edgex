package main

import (
	"github.com/Circutor/edgex/internal/pkg/config"
	"github.com/Circutor/edgex/internal/support/scheduler"
	"testing"
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
