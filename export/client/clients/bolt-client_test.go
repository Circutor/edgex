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
	config := DBConfiguration{
		DbType: BOLT,
	}

	bolt, err := newBoltClient(config)
	if err != nil {
		t.Fatalf("Could not connect with boltdb: %v", err)
	}

	testDB(t, bolt)
}

func BenchmarkBoltDB(b *testing.B) {
	config := DBConfiguration{
		DbType: BOLT,
	}

	bolt, err := newBoltClient(config)
	if err != nil {
		b.Fatalf("Could not connect with boltdb: %v", err)
	}

	benchmarkDB(b, bolt)
}
