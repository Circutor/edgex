//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

// +build boltRunning

// This test will only be executed if the tag boltRunning is added when running
// the tests with a command like:
// go test -tags boltRunning

package bolt

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/core/db"
	"github.com/edgexfoundry/edgex-go/core/db/test"
)

func TestBoltDB(t *testing.T) {

	config := db.Configuration{
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}

	bolt, err := NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect with BoltDB: %v", err)
	}

	test.TestDataDB(t, bolt)
}

func BenchmarkBoltDB(b *testing.B) {

	config := db.Configuration{
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}

	bolt, err := NewClient(config)
	if err != nil {
		b.Fatalf("Could not connect with BoltDB: %v", err)
	}

	test.BenchmarkDB(b, bolt)
}
