//
// Copyright (c) 2018 Cavium
//
// SPDX-License-Identifier: Apache-2.0
//

package clients

import (
	"strconv"
	"testing"

	"github.com/edgexfoundry/edgex-go/export"
)

func testDB(t *testing.T, db DBClient) {
	r := export.Registration{}
	r.Name = "name"

	// Remove previous registrations
	db.ScrubAllRegistrations()

	id, err := db.AddRegistration(&r)
	if err != nil {
		t.Fatalf("Error adding registration %v: %v", r, err)
	}

	regs, err := db.Registrations()
	if err != nil {
		t.Fatalf("Error getting registrations %v", err)
	}
	if len(regs) != 1 {
		t.Fatalf("There should be only one registration instead of %d", len(regs))
	}
	r2, err := db.RegistrationById(id.Hex())
	if err != nil {
		t.Fatalf("Error getting registrations by id %v", err)
	}
	if r2.ID.Hex() != id.Hex() {
		t.Fatalf("Id does not match %s - %s", r2.ID, id)
	}
	_, err = db.RegistrationById("INVALID")
	if err == nil {
		t.Fatalf("Registration should not be found")
	}

	r3, err := db.RegistrationByName(r.Name)
	if err != nil {
		t.Fatalf("Error getting registrations by name %v", err)
	}
	if r3.Name != r.Name {
		t.Fatalf("Id does not match %s - %s", r2.ID, id)
	}
	_, err = db.RegistrationByName("INVALID")
	if err == nil {
		t.Fatalf("Registration should not be found")
	}

	err = db.DeleteRegistrationById("INVALID")
	if err == nil {
		t.Fatalf("Registration should not be deleted")
	}

	err = db.DeleteRegistrationById(id.Hex())
	if err != nil {
		t.Fatalf("Registration should be deleted: %v", err)
	}

	id, err = db.AddRegistration(&r)
	if err != nil {
		t.Fatalf("Error adding registration %v: %v", r, err)
	}

	r.ID = id
	r.Name = "name2"
	err = db.UpdateRegistration(r)
	if err != nil {
		t.Fatalf("Error updating registration %v", err)
	}

	err = db.DeleteRegistrationByName("INVALID")
	if err == nil {
		t.Fatalf("Registration should not be deleted")
	}

	err = db.DeleteRegistrationByName(r.Name)
	if err != nil {
		t.Fatalf("Registration should be deleted: %v", err)
	}

	err = db.UpdateRegistration(r)
	if err == nil {
		t.Fatalf("Update should return error")
	}

	db.CloseSession()
	// Calling CloseSession twice to test that there is no panic when closing an
	// already closed db
	db.CloseSession()
}

func TestMemoryDB(t *testing.T) {
	memory := &memDB{}
	testDB(t, memory)
}

func BenchmarkMemoryDB(b *testing.B) {
	memory := &memDB{}
	benchmarkDB(b, memory)
}

func benchmarkDB(b *testing.B, db DBClient) {
	// Remove previous registrations
	db.ScrubAllRegistrations()

	var registrations []string

	b.Run("AddRegistration", func(b *testing.B) {
		b.N = 1000
		registration := export.Registration{}
		for i := 0; i < b.N; i++ {
			registration.Name = "test" + strconv.Itoa(i)
			id, err := db.AddRegistration(&registration)
			if err != nil {
				b.Fatalf("Error adding registration %v: %v", id, err)
			}
			registrations = append(registrations, id.Hex())
		}
	})

	b.Run("Registrations", func(b *testing.B) {
		b.N = 10
		for i := 0; i < b.N; i++ {
			_, err := db.Registrations()
			if err != nil {
				b.Fatalf("Error registrations: %v", err)
			}
		}
	})

	b.Run("RegistrationById", func(b *testing.B) {
		b.N = 1000
		for i := 0; i < b.N; i++ {
			_, err := db.RegistrationById(registrations[i])
			if err != nil {
				b.Fatalf("Error registrations by ID: %v", err)
			}
		}
	})

	b.Run("RegistrationByName", func(b *testing.B) {
		b.N = 100
		for i := 0; i < b.N; i++ {
			name := "test" + strconv.Itoa(i*10)
			_, err := db.RegistrationByName(name)
			if err != nil {
				b.Fatalf("Error registrations by name: %v", err)
			}
		}
	})

	b.Run("DeleteRegistrationByName", func(b *testing.B) {
		b.N = 1000
		for i := 0; i < b.N; i++ {
			name := "test" + strconv.Itoa(i)
			err := db.DeleteRegistrationByName(name)
			if err != nil {
				b.Fatalf("Error delete registrations by name: %v", err)
			}
		}
	})

	db.CloseSession()
}
