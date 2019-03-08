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

	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db/test"
)

func TestBoltDB(t *testing.T) {

	config := db.Configuration{DatabaseName: "coredata.db"}
	bolt, err := NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect with BoltDB: %v", err)
	}
	test.TestDataDB(t, bolt)

	config.DatabaseName = "metadata.db"
	bolt, err = NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect with BoltDB: %v", err)
	}
	test.TestMetadataDB(t, bolt)

	config.DatabaseName = "export.db"
	bolt, err = NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect with BoltDB: %v", err)
	}
	test.TestExportDB(t, bolt)

	config.DatabaseName = "scheduler.db"
	bolt, err = NewClient(config)
	if err != nil {
		t.Fatalf("Could not connect with BoltDB: %v", err)
	}
	test.TestSchedulerDB(t, bolt)
}

func BenchmarkBoltDB(b *testing.B) {

	config := db.Configuration{DatabaseName: "coredata.db"}
	bolt, err := NewClient(config)
	if err != nil {
		b.Fatalf("Could not connect with BoltDB: %v", err)
	}
	test.BenchmarkDB(b, bolt)
}
