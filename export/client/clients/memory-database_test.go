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

}

func TestMemoryDB(t *testing.T) {
	memory := &memDB{}
	testDB(t, memory)
}

func BenchmarkAddRegistrationMongoDB(b *testing.B) {

	b.StopTimer()
	config := DBConfiguration{
		DbType:       MONGO,
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}

	// Create the mongo client
	mc, err := newMongoClient(config)
	if err != nil {
		b.Fatalf("Could not connect with mongodb: %v", err)
	}
	b.N = 100
	registration := export.Registration{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		registration.Name = "test" + strconv.Itoa(i)
		id, err := mc.AddRegistration(&registration)
		if err != nil {
			b.Fatalf("Error adding registration %v: %v", id, err)
		}
	}

	b.StopTimer()
	mc.CloseSession()
	b.StartTimer()
}

func BenchmarkGetRegistrationbyNameMongoDB(b *testing.B) {

	b.StopTimer()
	config := DBConfiguration{
		DbType:       MONGO,
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}

	// Create the mongo client
	mc, err := newMongoClient(config)
	if err != nil {
		b.Fatalf("Could not connect with mongodb: %v", err)
	}
	b.N = 100

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		name := "test" + strconv.Itoa(i)
		_, err := mc.RegistrationByName(name)
		if err != nil {
			b.Fatalf("Error getting registration test: %v", err)
		}
	}

	b.StopTimer()
	mc.CloseSession()
	b.StartTimer()
}

func BenchmarkDeleteRegistrationbyNameMongoDB(b *testing.B) {

	b.StopTimer()
	config := DBConfiguration{
		DbType:       MONGO,
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}
	// Create the mongo client
	mc, err := newMongoClient(config)
	if err != nil {
		b.Fatalf("Could not connect with mongodb: %v", err)
	}
	b.N = 100

	b.StartTimer()
	var name string
	for i := 0; i < b.N; i++ {
		name = "test" + strconv.Itoa(i)
		err := mc.DeleteRegistrationByName(name)
		if err != nil {
			b.Fatalf("Error deleting registration %v: %v", name, err)
		}
	}

	b.StopTimer()
	mc.CloseSession()
	b.StartTimer()
}

func BenchmarkAddRegistrationBoltDB(b *testing.B) {

	b.StopTimer()
	config := DBConfiguration{
		DbType:       BOLT,
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}
	// Create the bolt client
	bolt, err := newBoltClient(config)
	if err != nil {
		b.Fatalf("Could not connect with boltdb: %v", err)
	}
	b.N = 100
	registration := export.Registration{}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		registration.Name = "test" + strconv.Itoa(i)
		id, err := bolt.AddRegistration(&registration)
		if err != nil {
			b.Fatalf("Error adding registration %v: %v", id, err)
		}

	}

	b.StopTimer()
	bolt.CloseSession()
	b.StartTimer()
}
func BenchmarkGetRegistrationbyNameBoltDB(b *testing.B) {

	b.StopTimer()
	config := DBConfiguration{
		DbType:       BOLT,
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}
	// Create the bolt client
	bolt, err := newBoltClient(config)
	if err != nil {
		b.Fatalf("Could not connect with boltdb: %v", err)
	}
	b.N = 100

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		name := "test" + strconv.Itoa(i)
		_, err := bolt.RegistrationByName(name)
		if err != nil {
			b.Fatalf("Error getting registration test: %v", err)
		}
	}

	b.StopTimer()
	bolt.CloseSession()
	b.StartTimer()
}

func BenchmarkDeleteRegistrationbyNameBoltDB(b *testing.B) {

	b.StopTimer()
	config := DBConfiguration{
		DbType:       BOLT,
		Host:         "0.0.0.0",
		Port:         27017,
		DatabaseName: "coredata",
		Timeout:      1000,
	}
	// Create the bolt client
	bolt, err := newBoltClient(config)
	if err != nil {
		b.Fatalf("Could not connect with boltdb: %v", err)
	}
	b.N = 100

	b.StartTimer()
	var name string
	for i := 0; i < b.N; i++ {
		name = "test" + strconv.Itoa(i)
		err := bolt.DeleteRegistrationByName(name)
		if err != nil {
			b.Fatalf("Error deleting registration %v: %v", name, err)
		}
	}

	b.StopTimer()
	bolt.CloseSession()
	b.StartTimer()
}
