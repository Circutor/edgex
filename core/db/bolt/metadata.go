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
	"github.com/coreos/bbolt"
	"github.com/edgexfoundry/edgex-go/core/db"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

/* -----------------------Schedule Event ------------------------*/
func (bc *BoltClient) UpdateScheduleEvent(se models.ScheduleEvent) error {
	return nil
}

func (bc *BoltClient) AddScheduleEvent(se *models.ScheduleEvent) error {
	return nil
}

func (bc *BoltClient) GetAllScheduleEvents(se *[]models.ScheduleEvent) error {
	return nil
}

func (bc *BoltClient) GetScheduleEventByName(se *models.ScheduleEvent, n string) error {
	return nil
}

func (bc *BoltClient) GetScheduleEventById(se *models.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetScheduleEventsByScheduleName(se *[]models.ScheduleEvent, n string) error {
	return nil
}

func (bc *BoltClient) GetScheduleEventsByAddressableId(se *[]models.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetScheduleEventsByServiceName(se *[]models.ScheduleEvent, n string) error {
	return nil
}

func (bc *BoltClient) GetScheduleEvent(se *models.ScheduleEvent, q bson.M) error {
	return nil
}

func (bc *BoltClient) GetScheduleEvents(se *[]models.ScheduleEvent, q bson.M) error {
	return nil
}

func (bc *BoltClient) DeleteScheduleEventById(id string) error {
	return nil
}

//  --------------------------Schedule ---------------------------*/
func (bc *BoltClient) GetAllSchedules(s *[]models.Schedule) error {
	return nil
}

func (bc *BoltClient) GetScheduleByName(s *models.Schedule, n string) error {
	return nil
}

func (bc *BoltClient) GetScheduleById(s *models.Schedule, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) AddSchedule(sch *models.Schedule) error {
	return nil
}

func (bc *BoltClient) UpdateSchedule(sch models.Schedule) error {
	return nil
}

func (bc *BoltClient) DeleteScheduleById(id string) error {
	return nil
}

func (bc *BoltClient) GetSchedule(sch *models.Schedule, q bson.M) error {
	return nil
}

func (bc *BoltClient) GetSchedules(sch *[]models.Schedule, q bson.M) error {
	return nil
}

/* ----------------------Device Report --------------------------*/
func (bc *BoltClient) GetAllDeviceReports(d *[]models.DeviceReport) error {
	return nil
}

func (bc *BoltClient) GetDeviceReportByName(d *models.DeviceReport, n string) error {
	return nil
}

func (bc *BoltClient) GetDeviceReportByDeviceName(d *[]models.DeviceReport, n string) error {
	return nil
}

func (bc *BoltClient) GetDeviceReportById(d *models.DeviceReport, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetDeviceReportsByScheduleEventName(d *[]models.DeviceReport, n string) error {
	return nil
}

func (bc *BoltClient) GetDeviceReports(d *[]models.DeviceReport, q bson.M) error {
	return nil
}

func (bc *BoltClient) GetDeviceReport(d *models.DeviceReport, q bson.M) error {
	return nil
}

func (bc *BoltClient) AddDeviceReport(d *models.DeviceReport) error {
	return nil
}

func (bc *BoltClient) UpdateDeviceReport(dr *models.DeviceReport) error {
	return nil
}

func (bc *BoltClient) DeleteDeviceReportById(id string) error {
	return nil
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

func (bc *BoltClient) getDevicesBy(d *[]models.Device, fn func(encoded []byte) bool) error {
	bd := boltDevice{}
	*d = []models.Device{}
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
				*d = append(*d, bd.Device)
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

func (bc *BoltClient) getDeviceProfilesBy(dp *[]models.DeviceProfile, fn func(encoded []byte) bool) error {
	bdp := boltDeviceProfile{}
	*dp = []models.DeviceProfile{}
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
				*dp = append(*dp, bdp.DeviceProfile)
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

func (bc *BoltClient) getDeviceServicesBy(ds *[]models.DeviceService, fn func(encoded []byte) bool) error {
	bds := boltDeviceService{}
	*ds = []models.DeviceService{}
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
				*ds = append(*ds, bds.DeviceService)
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
	return nil
}

func (bc *BoltClient) GetProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error {
	return nil
}

func (bc *BoltClient) GetProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error {
	return nil
}

func (bc *BoltClient) GetProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetProvisionWatchersByProfileId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetProvisionWatcherById(pw *models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		return db.ErrInvalidObjectId
	}
}

func (bc *BoltClient) GetProvisionWatcher(pw *models.ProvisionWatcher, q bson.M) error {
	return nil
}

func (bc *BoltClient) GetProvisionWatchers(pw *[]models.ProvisionWatcher, q bson.M) error {
	return nil
}

func (bc *BoltClient) AddProvisionWatcher(pw *models.ProvisionWatcher) error {
	return nil
}

func (bc *BoltClient) UpdateProvisionWatcher(pw models.ProvisionWatcher) error {
	return nil
}

func (bc *BoltClient) DeleteProvisionWatcherById(id string) error {
	return nil
}

//  ------------------------Command -------------------------------------*/
func (bc *BoltClient) GetAllCommands(c *[]models.Command) error {
	return bc.getCommandsBy(c, func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) getCommandsBy(c *[]models.Command, fn func(encoded []byte) bool) error {
	a := models.Command{}
	*c = []models.Command{}
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Command))
		if b == nil {
			return nil
		}
		err := b.ForEach(func(id, encoded []byte) error {
			if fn(encoded) == true {
				err := json.Unmarshal(encoded, &a)
				if err != nil {
					return err
				}
				*c = append(*c, a)
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
