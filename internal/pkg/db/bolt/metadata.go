/*******************************************************************************
 * Copyright 2017 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package bolt

import (
	"bytes"

	"github.com/coreos/bbolt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

/* -----------------------Schedule Event ------------------------*/
func (bc *BoltClient) UpdateScheduleEvent(se models.ScheduleEvent) error {
	se.Modified = db.MakeTimestamp()
	bse := boltScheduleEvent{ScheduleEvent: se}
	return bc.update(db.ScheduleEvent, bse, bse.Id)
}

func (bc *BoltClient) AddScheduleEvent(se *models.ScheduleEvent) error {

	// Check if the name exist
	var dummy models.ScheduleEvent
	err := bc.GetScheduleEventByName(&dummy, se.Name)
	if err == nil {
		return db.ErrNotUnique
	}
	// Check if the name exist
	var a models.Addressable
	err = bc.GetAddressableByName(&a, se.Addressable.Name)
	if err != nil {
		return db.ErrNotFound
	}

	ts := db.MakeTimestamp()
	se.Created = ts
	se.Modified = ts
	se.Id = bson.NewObjectId()

	bse := boltScheduleEvent{ScheduleEvent: *se}
	return bc.add(db.ScheduleEvent, bse, bse.Id)
}

func (bc *BoltClient) GetAllScheduleEvents(se *[]models.ScheduleEvent) error {
	return bc.getScheduleEventsBy(se, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetScheduleEventByName(se *models.ScheduleEvent, n string) error {
	bse := boltScheduleEvent{ScheduleEvent: *se}
	err := bc.getByName(&bse, db.ScheduleEvent, n)
	*se = bse.ScheduleEvent
	return err
}

func (bc *BoltClient) GetScheduleEventById(se *models.ScheduleEvent, id string) error {
	bse := boltScheduleEvent{ScheduleEvent: *se}
	err := bc.getById(&bse, db.ScheduleEvent, id)
	*se = bse.ScheduleEvent
	return err
}

func (bc *BoltClient) getScheduleEventsBy(ses *[]models.ScheduleEvent, fn func(encoded []byte) bool) error {
	bse := boltScheduleEvent{}
	*ses = []models.ScheduleEvent{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.ScheduleEvent))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &bse)
				if err != nil {
					return err
				}
				*ses = append(*ses, bse.ScheduleEvent)
			}
			return nil
		})
		return err
	})
	return err
}

func (bc *BoltClient) GetScheduleEventsByScheduleName(se *[]models.ScheduleEvent, n string) error {
	return bc.getScheduleEventsBy(se, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "schedule").ToString()
		if value == n {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetScheduleEventsByAddressableId(se *[]models.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return bc.getScheduleEventsBy(se, func(encoded []byte) bool {
			value := jsoniter.Get(encoded, "addressableId").ToString()
			if value == id {
				return true
			}
			return false
		})
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetScheduleEventsByServiceName(se *[]models.ScheduleEvent, n string) error {
	return bc.getScheduleEventsBy(se, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "service").ToString()
		if value == n {
			return true
		}
		return false
	})
}

func (bc *BoltClient) DeleteScheduleEventById(id string) error {
	return bc.deleteById(id, db.ScheduleEvent)
}

//  --------------------------Schedule ---------------------------*/
func (bc *BoltClient) GetAllSchedules(sch *[]models.Schedule) error {
	return bc.getSchedulesBy(sch, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetScheduleByName(sch *models.Schedule, n string) error {
	return bc.getByName(sch, db.Schedule, n)
}

func (bc *BoltClient) GetScheduleById(sch *models.Schedule, id string) error {
	return bc.getById(sch, db.Schedule, id)
}

func (bc *BoltClient) AddSchedule(sch *models.Schedule) error {
	// Check if the name exist
	var dummy models.Schedule
	err := bc.GetScheduleByName(&dummy, sch.Name)
	if err == nil {
		return db.ErrNotUnique
	}

	sch.Created = db.MakeTimestamp()
	sch.Id = bson.NewObjectId()

	return bc.add(db.Schedule, sch, sch.Id)
}

func (bc *BoltClient) UpdateSchedule(sch models.Schedule) error {
	sch.Modified = db.MakeTimestamp()
	return bc.update(db.Schedule, sch, sch.Id)
}

func (bc *BoltClient) DeleteScheduleById(id string) error {
	return bc.deleteById(id, db.Schedule)
}

func (bc *BoltClient) getSchedulesBy(schs *[]models.Schedule, fn func(encoded []byte) bool) error {
	sch := models.Schedule{}
	*schs = []models.Schedule{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Schedule))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &sch)
				if err != nil {
					return err
				}
				*schs = append(*schs, sch)
			}
			return nil
		})
		return err
	})
	return err
}

/* ----------------------Device Report --------------------------*/
func (bc *BoltClient) GetAllDeviceReports(dr *[]models.DeviceReport) error {
	return bc.getDeviceReportsBy(dr, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetDeviceReportByName(dr *models.DeviceReport, n string) error {
	return bc.getByName(dr, db.DeviceReport, n)
}

func (bc *BoltClient) GetDeviceReportByDeviceName(dr *[]models.DeviceReport, n string) error {
	return bc.getDeviceReportsBy(dr, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "device").ToString()
		if value == n {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetDeviceReportById(dr *models.DeviceReport, id string) error {
	return bc.getById(dr, db.DeviceReport, id)
}

func (bc *BoltClient) GetDeviceReportsByScheduleEventName(dr *[]models.DeviceReport, n string) error {
	return bc.getDeviceReportsBy(dr, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "event").ToString()
		if value == n {
			return true
		}
		return false
	})
}

func (bc *BoltClient) getDeviceReportsBy(drs *[]models.DeviceReport, fn func(encoded []byte) bool) error {
	dr := models.DeviceReport{}
	*drs = []models.DeviceReport{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceReport))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &dr)
				if err != nil {
					return err
				}
				*drs = append(*drs, dr)
			}
			return nil
		})
		return err
	})
	return err
}

func (bc *BoltClient) AddDeviceReport(dr *models.DeviceReport) error {
	// Check if the name exist
	var dummy models.DeviceReport
	err := bc.GetDeviceReportByName(&dummy, dr.Name)
	if err == nil {
		return db.ErrNotUnique
	}

	dr.Created = db.MakeTimestamp()
	dr.Id = bson.NewObjectId()

	return bc.add(db.DeviceReport, dr, dr.Id)
}

func (bc *BoltClient) UpdateDeviceReport(dr *models.DeviceReport) error {
	dr.Modified = db.MakeTimestamp()
	return bc.update(db.DeviceReport, dr, dr.Id)
}

func (bc *BoltClient) DeleteDeviceReportById(id string) error {
	return bc.deleteById(id, db.DeviceReport)
}

/* ----------------------------- Device ---------------------------------- */
func (bc *BoltClient) AddDevice(d *models.Device) error {
	// Check if the name exist
	var dummy models.Device
	err := bc.GetDeviceByName(&dummy, d.Name)
	if err == nil {
		return db.ErrNotUnique
	}

	d.Created = db.MakeTimestamp()
	d.Id = bson.NewObjectId()

	bd := boltDevice{Device: *d}
	return bc.add(db.Device, bd, d.Id)
}

func (bc *BoltClient) UpdateDevice(d models.Device) error {
	d.Modified = db.MakeTimestamp()
	bd := boltDevice{Device: d}
	return bc.update(db.Device, bd, bd.Id)
}

func (bc *BoltClient) DeleteDeviceById(id string) error {
	return bc.deleteById(id, db.Device)
}

func (bc *BoltClient) GetAllDevices(d *[]models.Device) error {
	return bc.getDevicesBy(d, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetDeviceById(d *models.Device, id string) error {
	bd := boltDevice{Device: *d}
	err := bc.getById(&bd, db.Device, id)
	*d = bd.Device
	return err
}

func (bc *BoltClient) GetDeviceByName(d *models.Device, n string) error {
	bd := boltDevice{Device: *d}
	err := bc.getByName(&bd, db.Device, n)
	*d = bd.Device
	return err
}

func (bc *BoltClient) GetDevicesByServiceId(d *[]models.Device, sid string) error {
	if bson.IsObjectIdHex(sid) {
		return bc.getDevicesBy(d, func(encoded []byte) bool {
			value := jsoniter.Get(encoded, "serviceId").ToString()
			if value == sid {
				return true
			}
			return false
		})
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetDevicesByAddressableId(d *[]models.Device, aid string) error {
	if bson.IsObjectIdHex(aid) {
		return bc.getDevicesBy(d, func(encoded []byte) bool {
			value := jsoniter.Get(encoded, "addressableId").ToString()
			if value == aid {
				return true
			}
			return false
		})
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetDevicesByProfileId(d *[]models.Device, pid string) error {
	if bson.IsObjectIdHex(pid) {
		return bc.getDevicesBy(d, func(encoded []byte) bool {
			value := jsoniter.Get(encoded, "profileId").ToString()
			if value == pid {
				return true
			}
			return false
		})
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetDevicesWithLabel(d *[]models.Device, label string) error {
	return bc.getDevicesBy(d, func(encoded []byte) bool {
		labels := jsoniter.Get(encoded, "labels").GetInterface().([]interface{})
		for _, value := range labels {
			if label == value.(string) {
				return true
			}
		}
		return false
	})
}

func (bc *BoltClient) getDevicesBy(ds *[]models.Device, fn func(encoded []byte) bool) error {
	bd := boltDevice{}
	*ds = []models.Device{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Device))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &bd)
				if err != nil {
					return err
				}
				*ds = append(*ds, bd.Device)
			}
			return nil
		})
		return err
	})
	return err
}

/* -----------------------------Device Profile -----------------------------*/
func (bc *BoltClient) GetDeviceProfileById(dp *models.DeviceProfile, id string) error {
	bdp := boltDeviceProfile{DeviceProfile: *dp}
	err := bc.getById(&bdp, db.DeviceProfile, id)
	*dp = bdp.DeviceProfile
	return err
}

func (bc *BoltClient) GetAllDeviceProfiles(dp *[]models.DeviceProfile) error {
	return bc.getDeviceProfilesBy(dp, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetDeviceProfilesByModel(dp *[]models.DeviceProfile, model string) error {
	return bc.getDeviceProfilesBy(dp, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "model").ToString()
		if value == model {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetDeviceProfilesWithLabel(dp *[]models.DeviceProfile, label string) error {
	return bc.getDeviceProfilesBy(dp, func(encoded []byte) bool {
		labels := jsoniter.Get(encoded, "labels").GetInterface().([]interface{})
		for _, value := range labels {
			if label == value.(string) {
				return true
			}
		}
		return false
	})
}

func (bc *BoltClient) GetDeviceProfilesByManufacturerModel(dp *[]models.DeviceProfile, man string, mod string) error {
	return bc.getDeviceProfilesBy(dp, func(encoded []byte) bool {
		valueMod := jsoniter.Get(encoded, "model").ToString()
		valueMan := jsoniter.Get(encoded, "manufacturer").ToString()
		if valueMod == mod && valueMan == man {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetDeviceProfilesByManufacturer(dp *[]models.DeviceProfile, man string) error {
	return bc.getDeviceProfilesBy(dp, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "manufacturer").ToString()
		if value == man {
			return true
		}
		return false
	})
}

func (bc *BoltClient) getDeviceProfilesBy(dps *[]models.DeviceProfile, fn func(encoded []byte) bool) error {
	bdp := boltDeviceProfile{}
	*dps = []models.DeviceProfile{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceProfile))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &bdp)
				if err != nil {
					return err
				}
				*dps = append(*dps, bdp.DeviceProfile)
			}
			return nil
		})
		return err
	})
	return err
}

func (bc *BoltClient) GetDeviceProfileByName(dp *models.DeviceProfile, n string) error {
	bdp := boltDeviceProfile{DeviceProfile: *dp}
	err := bc.getByName(&bdp, db.DeviceProfile, n)
	*dp = bdp.DeviceProfile
	return err
}

func (bc *BoltClient) AddDeviceProfile(dp *models.DeviceProfile) error {
	// Check if the name exist
	var dummy models.DeviceProfile
	err := bc.GetDeviceProfileByName(&dummy, dp.Name)
	if err == nil {
		return db.ErrNotUnique
	}

	for i := 0; i < len(dp.Commands); i++ {
		if errs := bc.AddCommand(&dp.Commands[i]); errs != nil {
			return errs
		}
	}

	dp.Created = db.MakeTimestamp()
	dp.Id = bson.NewObjectId()

	bdp := boltDeviceProfile{DeviceProfile: *dp}
	return bc.add(db.DeviceProfile, bdp, dp.Id)
}

func (bc *BoltClient) UpdateDeviceProfile(dp *models.DeviceProfile) error {
	dp.Modified = db.MakeTimestamp()
	bdp := boltDeviceProfile{DeviceProfile: *dp}
	return bc.update(db.DeviceProfile, bdp, bdp.Id)
}

// Get the device profiles that are currently using the command
func (bc *BoltClient) GetDeviceProfilesUsingCommand(dp *[]models.DeviceProfile, c models.Command) error {
	return bc.getDeviceProfilesBy(dp, func(encoded []byte) bool {
		commands := jsoniter.Get(encoded, "commands").GetInterface().([]interface{})
		for _, value := range commands {
			if c.Id.Hex() == value.(string) {
				return true
			}
		}
		return false
	})
}

func (bc *BoltClient) DeleteDeviceProfileById(id string) error {
	return bc.deleteById(id, db.DeviceProfile)
}

//  -----------------------------------Addressable --------------------------*/
func (bc *BoltClient) UpdateAddressable(ra *models.Addressable, r *models.Addressable) error {

	if ra.Name != "" {
		r.Name = ra.Name
	}
	if ra.Protocol != "" {
		r.Protocol = ra.Protocol
	}
	if ra.Address != "" {
		r.Address = ra.Address
	}
	if ra.Port != int(0) {
		r.Port = ra.Port
	}
	if ra.Path != "" {
		r.Path = ra.Path
	}
	if ra.Publisher != "" {
		r.Publisher = ra.Publisher
	}
	if ra.User != "" {
		r.User = ra.User
	}
	if ra.Password != "" {
		r.Password = ra.Password
	}
	if ra.Topic != "" {
		r.Topic = ra.Topic
	}
	r.Modified = db.MakeTimestamp()

	return bc.update(db.Addressable, r, r.Id)
}

func (bc *BoltClient) GetAddressables(a *[]models.Addressable) error {
	return bc.getAddressablesBy(a, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) getAddressablesBy(as *[]models.Addressable, fn func(encoded []byte) bool) error {
	a := models.Addressable{}
	*as = []models.Addressable{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Addressable))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &a)
				if err != nil {
					return err
				}
				*as = append(*as, a)
			}
			return nil
		})
		return err
	})
	return err
}

func (bc *BoltClient) GetAddressableById(a *models.Addressable, id string) error {
	return bc.getById(a, db.Addressable, id)
}

func (bc *BoltClient) AddAddressable(a *models.Addressable) (bson.ObjectId, error) {
	// Check if the name exist
	var dummy models.Addressable
	err := bc.GetAddressableByName(&dummy, a.Name)
	if err == nil {
		return dummy.Id, db.ErrNotUnique
	}

	a.Created = db.MakeTimestamp()
	a.Id = bson.NewObjectId()

	err = bc.add(db.Addressable, a, a.Id)
	return a.Id, err
}

func (bc *BoltClient) GetAddressableByName(a *models.Addressable, name string) error {
	return bc.getByName(&a, db.Addressable, name)
}

func (bc *BoltClient) GetAddressablesByTopic(a *[]models.Addressable, topic string) error {
	return bc.getAddressablesBy(a, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "topic").ToString()
		if value == topic {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetAddressablesByPort(a *[]models.Addressable, port int) error {
	return bc.getAddressablesBy(a, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "port").ToInt()
		if value == port {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetAddressablesByPublisher(a *[]models.Addressable, publisher string) error {
	return bc.getAddressablesBy(a, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "publisher").ToString()
		if value == publisher {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetAddressablesByAddress(a *[]models.Addressable, address string) error {
	return bc.getAddressablesBy(a, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "address").ToString()
		if value == address {
			return true
		}
		return false
	})
}

func (bc *BoltClient) DeleteAddressableById(id string) error {
	return bc.deleteById(id, db.Addressable)
}

/* ----------------------------- Device Service ----------------------------------*/
func (bc *BoltClient) GetDeviceServiceByName(ds *models.DeviceService, n string) error {
	bds := boltDeviceService{DeviceService: *ds}
	err := bc.getByName(&bds, db.DeviceService, n)
	*ds = bds.DeviceService
	return err
}

func (bc *BoltClient) GetDeviceServiceById(ds *models.DeviceService, id string) error {
	bds := boltDeviceService{DeviceService: *ds}
	err := bc.getById(&bds, db.DeviceService, id)
	*ds = bds.DeviceService
	return err
}

func (bc *BoltClient) GetAllDeviceServices(ds *[]models.DeviceService) error {
	return bc.getDeviceServicesBy(ds, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetDeviceServicesByAddressableId(ds *[]models.DeviceService, id string) error {
	if bson.IsObjectIdHex(id) {
		return bc.getDeviceServicesBy(ds, func(encoded []byte) bool {
			value := jsoniter.Get(encoded, "addressableId").ToString()
			if value == id {
				return true
			}
			return false
		})
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetDeviceServicesWithLabel(ds *[]models.DeviceService, label string) error {
	return bc.getDeviceServicesBy(ds, func(encoded []byte) bool {
		labels := jsoniter.Get(encoded, "labels").GetInterface().([]interface{})
		for _, value := range labels {
			if label == value.(string) {
				return true
			}
		}
		return false
	})
}

func (bc *BoltClient) getDeviceServicesBy(dss *[]models.DeviceService, fn func(encoded []byte) bool) error {
	bds := boltDeviceService{}
	*dss = []models.DeviceService{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceService))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &bds)
				if err != nil {
					return err
				}
				*dss = append(*dss, bds.DeviceService)
			}
			return nil
		})
		return err
	})
	return err
}

func (bc *BoltClient) AddDeviceService(ds *models.DeviceService) error {
	// Check if the name exist
	var dummy models.DeviceService
	err := bc.GetDeviceServiceByName(&dummy, ds.Name)
	if err == nil {
		return db.ErrNotUnique
	}

	ds.Created = db.MakeTimestamp()
	ds.Id = bson.NewObjectId()

	bds := boltDeviceService{DeviceService: *ds}
	return bc.add(db.DeviceService, bds, bds.Id)
}

func (bc *BoltClient) UpdateDeviceService(ds models.DeviceService) error {
	ds.Modified = db.MakeTimestamp()
	bds := boltDeviceService{DeviceService: ds}
	return bc.update(db.DeviceService, bds, bds.Id)
}

func (bc *BoltClient) DeleteDeviceServiceById(id string) error {
	return bc.deleteById(id, db.DeviceService)
}

//  ----------------------Provision Watcher -----------------------------*/
func (bc *BoltClient) GetAllProvisionWatchers(pw *[]models.ProvisionWatcher) error {
	return bc.getProvisionWatchersBy(pw, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error {
	bpw := boltProvisionWatcher{ProvisionWatcher: *pw}
	err := bc.getByName(&bpw, db.ProvisionWatcher, n)
	*pw = bpw.ProvisionWatcher
	return err
}

func (bc *BoltClient) GetProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error {
	return bc.getProvisionWatchersBy(pw, func(encoded []byte) bool {
		identifier := jsoniter.Get(encoded, "identifiers").ToString()
		keyvalue := "\"" + k + "\"" + ":" + "\"" + v + "\""
		if bytes.Contains([]byte(identifier), []byte(keyvalue)) {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return bc.getProvisionWatchersBy(pw, func(encoded []byte) bool {
			value := jsoniter.Get(encoded, "serviceId").ToString()
			if value == id {
				return true
			}
			return false
		})
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetProvisionWatchersByProfileId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return bc.getProvisionWatchersBy(pw, func(encoded []byte) bool {
			value := jsoniter.Get(encoded, "profileId").ToString()
			if value == id {
				return true
			}
			return false
		})
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetProvisionWatcherById(pw *models.ProvisionWatcher, id string) error {
	bpw := boltProvisionWatcher{ProvisionWatcher: *pw}
	err := bc.getById(&bpw, db.ProvisionWatcher, id)
	*pw = bpw.ProvisionWatcher
	return err
}

func (bc *BoltClient) getProvisionWatchersBy(pws *[]models.ProvisionWatcher, fn func(encoded []byte) bool) error {
	bpw := boltProvisionWatcher{}
	*pws = []models.ProvisionWatcher{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.ProvisionWatcher))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &bpw)
				if err != nil {
					return err
				}
				*pws = append(*pws, bpw.ProvisionWatcher)
			}
			return nil
		})
		return err
	})
	return err
}

func (bc *BoltClient) AddProvisionWatcher(pw *models.ProvisionWatcher) error {
	// Check if the name exist
	var dummy models.ProvisionWatcher
	err := bc.GetProvisionWatcherByName(&dummy, pw.Name)
	if err == nil {
		return db.ErrNotUnique
	}

	pw.Created = db.MakeTimestamp()
	pw.Id = bson.NewObjectId()

	bpw := boltProvisionWatcher{ProvisionWatcher: *pw}
	return bc.add(db.ProvisionWatcher, bpw, pw.Id)
}

func (bc *BoltClient) UpdateProvisionWatcher(pw models.ProvisionWatcher) error {
	pw.Modified = db.MakeTimestamp()
	bpw := boltProvisionWatcher{ProvisionWatcher: pw}
	return bc.update(db.ProvisionWatcher, bpw, pw.Id)
}

func (bc *BoltClient) DeleteProvisionWatcherById(id string) error {
	return bc.deleteById(id, db.ProvisionWatcher)
}

//  ------------------------Command -------------------------------------*/
func (bc *BoltClient) GetAllCommands(c *[]models.Command) error {
	return bc.getCommandsBy(c, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) getCommandsBy(cs *[]models.Command, fn func(encoded []byte) bool) error {
	c := models.Command{}
	*cs = []models.Command{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Command))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &c)
				if err != nil {
					return err
				}
				*cs = append(*cs, c)
			}
			return nil
		})
		return err
	})
	return err
}

func (bc *BoltClient) GetCommandById(c *models.Command, id string) error {
	return bc.getById(c, db.Command, id)
}

func (bc *BoltClient) GetCommandByName(c *[]models.Command, name string) error {
	return bc.getCommandsBy(c, func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "name").ToString()
		if value == name {
			return true
		}
		return false
	})
}

func (bc *BoltClient) AddCommand(c *models.Command) error {
	c.Created = db.MakeTimestamp()
	c.Id = bson.NewObjectId()

	return bc.add(db.Command, c, c.Id)
}

func (bc *BoltClient) UpdateCommand(c *models.Command, r *models.Command) error {

	// Update the fields
	if c.Name != "" {
		r.Name = c.Name
	}
	// TODO check for Get and Put Equality

	if (c.Get.String() != models.Get{}.String()) {
		r.Get = c.Get
	}
	if (c.Put.String() != models.Put{}.String()) {
		r.Put = c.Put
	}
	if c.Origin != 0 {
		r.Origin = c.Origin
	}
	c.Modified = db.MakeTimestamp()

	return bc.update(db.Command, r, r.Id)
}

func (bc *BoltClient) DeleteCommandById(id string) error {
	return bc.deleteById(id, db.Command)
}
