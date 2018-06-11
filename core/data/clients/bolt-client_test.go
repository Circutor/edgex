//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

// +build boltRunning

// This test will only be executed if the tag boltRunning is added when running
// the tests with a command like:
// go test -tags boltRunning

package clients

import (
	"testing"
)

func TestBoltDB(t *testing.T) {

	//t.Log("This test needs to have a running bolt on localhost")

	config := DBConfiguration{
		DbType:       BOLT,
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}

	bolt, err := newBoltClient(config)
	if err != nil {
		t.Fatalf("Could not connect with boltdb: %v", err)
	}

	testDB(t, bolt)
}

func BenchmarkBoltDB(b *testing.B) {

	//b.Log("This benchmark needs to have a running bolt on localhost")

	config := DBConfiguration{
		DbType:       BOLT,
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}

	benchmarkDB(b, config)
}
