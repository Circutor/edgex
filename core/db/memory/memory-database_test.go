//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package memory

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/core/db/test"
)

func TestMemoryDB(t *testing.T) {
	memory := &MemDB{}
	test.TestDataDB(t, memory)
}
