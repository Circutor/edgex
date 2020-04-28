//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package client

import (
	"testing"

	"gitlab.circutor.com/EDS/edgex-go/internal/pkg/db/test"
)

func TestMemoryDB(t *testing.T) {
	memory := &MemDB{}
	test.TestExportDB(t, memory)
}
