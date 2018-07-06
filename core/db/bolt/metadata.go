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
	"errors"

	"github.com/coreos/bbolt"
	"github.com/edgexfoundry/edgex-go/core/db"
	"github.com/edgexfoundry/edgex-go/core/domain/models"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

/* -----------------------Schedule Event ------------------------*/
func (b *BoltClient) UpdateScheduleEvent(se models.ScheduleEvent) error {

	return nil
}

func (b *BoltClient) AddScheduleEvent(se *models.ScheduleEvent) error {
	return nil
}

func (b *BoltClient) GetAllScheduleEvents(se *[]models.ScheduleEvent) error {
	return nil
}

func (b *BoltClient) GetScheduleEventByName(se *models.ScheduleEvent, n string) error {
	return nil
}

func (b *BoltClient) GetScheduleEventById(se *models.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		err := errors.New("mgoGetScheduleEventById Invalid Object ID " + id)
		return err
	}
}

func (b *BoltClient) GetScheduleEventsByScheduleName(se *[]models.ScheduleEvent, n string) error {
	return nil
}

func (b *BoltClient) GetScheduleEventsByAddressableId(se *[]models.ScheduleEvent, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		err := errors.New("mgoGetScheduleEventsByAddressableId Invalid Object ID" + id)
		return err
	}
}

func (b *BoltClient) GetScheduleEventsByServiceName(se *[]models.ScheduleEvent, n string) error {
	return nil
}

func (b *BoltClient) GetScheduleEvent(se *models.ScheduleEvent, q bson.M) error {

	return nil
}

func (b *BoltClient) GetScheduleEvents(se *[]models.ScheduleEvent, q bson.M) error {

	return nil
}

func (b *BoltClient) DeleteScheduleEvent(se models.ScheduleEvent) error {
	return nil
}

//  --------------------------Schedule ---------------------------*/
func (b *BoltClient) GetAllSchedules(s *[]models.Schedule) error {
	return nil
}

func (b *BoltClient) GetScheduleByName(s *models.Schedule, n string) error {
	return nil
}

func (b *BoltClient) GetScheduleById(s *models.Schedule, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		err := errors.New("mgoGetScheduleById Invalid Object ID " + id)
		return err
	}
}

func (b *BoltClient) AddSchedule(sch *models.Schedule) error {

	return nil
}

func (b *BoltClient) UpdateSchedule(sch models.Schedule) error {

	return nil
}

func (b *BoltClient) DeleteSchedule(s models.Schedule) error {
	return nil
}

func (b *BoltClient) GetSchedule(sch *models.Schedule, q bson.M) error {

	return nil
}

func (b *BoltClient) GetSchedules(sch *[]models.Schedule, q bson.M) error {

	return nil
}

/* ----------------------Device Report --------------------------*/
func (b *BoltClient) GetAllDeviceReports(d *[]models.DeviceReport) error {
	return nil
}

func (b *BoltClient) GetDeviceReportByName(d *models.DeviceReport, n string) error {
	return nil
}

func (b *BoltClient) GetDeviceReportByDeviceName(d *[]models.DeviceReport, n string) error {
	return nil
}

func (b *BoltClient) GetDeviceReportById(d *models.DeviceReport, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		err := errors.New("mgoGetDeviceReportById Invalid Object ID " + id)
		return err
	}
}

func (b *BoltClient) GetDeviceReportsByScheduleEventName(d *[]models.DeviceReport, n string) error {
	return nil
}

func (b *BoltClient) GetDeviceReports(d *[]models.DeviceReport, q bson.M) error {

	return nil
}

func (b *BoltClient) GetDeviceReport(d *models.DeviceReport, q bson.M) error {

	return nil
}

func (b *BoltClient) AddDeviceReport(d *models.DeviceReport) error {

	return nil
}

func (b *BoltClient) UpdateDeviceReport(dr *models.DeviceReport) error {

	return nil
}

func (b *BoltClient) DeleteDeviceReport(dr models.DeviceReport) error {
	return nil
}

/* ----------------------------- Device ---------------------------------- */
func (b *BoltClient) AddDevice(d *models.Device) error {

	// check if the name exist (Device names must be unique)
	var dumy models.Device
	//err := b.getByName(&dumy, db.Device, d.Name)
	err := b.GetDeviceByName(&dumy, d.Name)
	if err == nil {
		return db.ErrNotUnique
	}

	ts := db.MakeTimestamp()
	d.Created = ts
	d.Modified = ts
	d.Id = bson.NewObjectId()

	// Wrap the device in boltDevice
	bd := boltDevice{Device: *d}
	return b.add(db.Device, bd, d.Id)
}

func (b *BoltClient) UpdateDevice(d models.Device) error {
	d.Modified = db.MakeTimestamp()
	bd := boltDevice{Device: d}
	return b.update(db.Device, bd, bd.Id)
}

func (b *BoltClient) DeleteDevice(d models.Device) error {
	return b.deleteById(d.Id.Hex(), db.Device)
}

func (b *BoltClient) GetAllDevices(d *[]models.Device) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Device))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*d = []models.Device{}
		err := b.ForEach(func(id, encoded []byte) error {
			var bd boltDevice
			err := json.Unmarshal(encoded, &bd)
			if err != nil {
				return err
			}
			*d = append(*d, bd.Device)
			return nil
		})
		return err
	})
	return err
}

func (b *BoltClient) GetDeviceById(d *models.Device, id string) error {
	if bson.IsObjectIdHex(id) {
		bd := boltDevice{Device: *d}
		err := b.getById(&bd, db.Device, id)
		*d = bd.Device
		return err
	} else {
		err := errors.New("boltGetDeviceById Invalid Object ID " + id)
		return err
	}
}

func (b *BoltClient) GetDeviceByName(d *models.Device, n string) error {
	bd := boltDevice{Device: *d}
	err := b.getByName(&bd, db.Device, n)
	*d = bd.Device
	return err
}

func (b *BoltClient) GetDevicesByServiceId(d *[]models.Device, sid string) error {
	if bson.IsObjectIdHex(sid) {
		return b.GetDevicesBy(d, "serviceId", sid)
	} else {
		err := errors.New("mgoGetDevicesByServiceId Invalid Object ID " + sid)
		return err
	}
}

func (b *BoltClient) GetDevicesByAddressableId(d *[]models.Device, aid string) error {
	if bson.IsObjectIdHex(aid) {
		return b.GetDevicesBy(d, "addressableId", aid)
	} else {
		err := errors.New("mgoGetDevicesByAddressableId Invalid Object ID " + aid)
		return err
	}
}

func (b *BoltClient) GetDevicesByProfileId(d *[]models.Device, pid string) error {
	if bson.IsObjectIdHex(pid) {
		return b.GetDevicesBy(d, "profileId", pid)
	} else {
		err := errors.New("mgoGetDevicesByProfileId Invalid Object ID " + pid)
		return err
	}
}

func (b *BoltClient) GetDevicesWithLabel(d *[]models.Device, l string) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Device))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*d = []models.Device{}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var bd boltDevice
			err := json.Unmarshal(encoded, &bd)
			if err != nil {
				return err
			}
			for i := 0; i < len(bd.Labels); i++ {
				value := jsoniter.Get(encoded, "labels", i).ToString()
				if value == l {
					*d = append(*d, bd.Device)
				}
			}
			return nil
		})
		return err
	})
	return err

}
func (b *BoltClient) GetDevicesBy(d *[]models.Device, tag string, filter string) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Device))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*d = []models.Device{}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var bd boltDevice
			value := jsoniter.Get(encoded, tag).ToString()
			if value == filter {
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
func (b *BoltClient) GetDeviceProfileById(dp *models.DeviceProfile, id string) error {
	if bson.IsObjectIdHex(id) {
		bdp := boltDeviceProfile{DeviceProfile: *dp}
		err := b.getById(&bdp, db.DeviceProfile, id)
		*dp = bdp.DeviceProfile
		return err
	} else {
		err := errors.New("boltGetDeviceProfileById Invalid Object ID " + id)
		return err
	}
}

func (b *BoltClient) GetAllDeviceProfiles(dp *[]models.DeviceProfile) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceProfile))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*dp = []models.DeviceProfile{}
		err := b.ForEach(func(id, encoded []byte) error {
			var bdp boltDeviceProfile
			err := json.Unmarshal(encoded, &bdp)
			if err != nil {
				return err
			}
			*dp = append(*dp, bdp.DeviceProfile)
			return nil
		})
		return err
	})
	return err
}

func (b *BoltClient) GetDeviceProfilesByModel(dp *[]models.DeviceProfile, model string) error {
	return b.GetDeviceProfilesBy(dp, "model", model)
}
func (b *BoltClient) GetDeviceProfilesBy(dp *[]models.DeviceProfile, tag string, filter string) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceProfile))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*dp = []models.DeviceProfile{}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var bdp boltDeviceProfile
			value := jsoniter.Get(encoded, tag).ToString()
			if value == filter {
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

func (b *BoltClient) GetDeviceProfilesWithLabel(dp *[]models.DeviceProfile, l string) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceProfile))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*dp = []models.DeviceProfile{}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var bdp boltDeviceProfile
			err := json.Unmarshal(encoded, &bdp)
			if err != nil {
				return err
			}
			for i := 0; i < len(bdp.Labels); i++ {
				value := jsoniter.Get(encoded, "labels", i).ToString()
				if value == l {
					*dp = append(*dp, bdp.DeviceProfile)
				}
			}
			return nil
		})
		return err
	})
	return err
}

func (b *BoltClient) GetDeviceProfilesByManufacturerModel(dp *[]models.DeviceProfile, man string, mod string) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceProfile))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*dp = []models.DeviceProfile{}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var bdp boltDeviceProfile
			valuemod := jsoniter.Get(encoded, "model").ToString()
			valueman := jsoniter.Get(encoded, "manufacturer").ToString()
			if valuemod == mod && valueman == man {
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

func (b *BoltClient) GetDeviceProfilesByManufacturer(dp *[]models.DeviceProfile, man string) error {
	return b.GetDeviceProfilesBy(dp, "manufacturer", man)
}
func (b *BoltClient) GetDeviceProfileByName(dp *models.DeviceProfile, n string) error {
	bdp := boltDeviceProfile{DeviceProfile: *dp}
	err := b.getByName(&bdp, db.DeviceProfile, n)
	*dp = bdp.DeviceProfile
	return err
}

func (b *BoltClient) AddDeviceProfile(dp *models.DeviceProfile) error {
	// check if the name exist
	var dumy models.DeviceProfile
	//err := b.getByName(&dumy, db.DeviceProfile, dp.Name)
	err := b.GetDeviceProfileByName(&dumy, dp.Name)
	if err == nil {
		return db.ErrNotUnique
	}

	for i := 0; i < len(dp.Commands); i++ {
		if errs := b.AddCommand(&dp.Commands[i]); errs != nil {

			return errs
		}
	}
	dp.Created = db.MakeTimestamp()
	dp.Id = bson.NewObjectId()
	bdp := boltDeviceProfile{DeviceProfile: *dp}
	return b.add(db.DeviceProfile, bdp, dp.Id)
}

func (b *BoltClient) UpdateDeviceProfile(dp *models.DeviceProfile) error {
	dp.Modified = db.MakeTimestamp()
	bdp := boltDeviceProfile{DeviceProfile: *dp}
	return b.update(db.DeviceProfile, bdp, bdp.Id)
}

// Get the device profiles that are currently using the command
func (b *BoltClient) GetDeviceProfilesUsingCommand(dp *[]models.DeviceProfile, c models.Command) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceProfile))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*dp = []models.DeviceProfile{}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var bdp boltDeviceProfile
			err := json.Unmarshal(encoded, &bdp)
			if err != nil {
				return err
			}
			for i := 0; i < len(bdp.Commands); i++ {
				value := jsoniter.Get(encoded, "commands", i).ToString()
				if value == c.Id.Hex() {
					*dp = append(*dp, bdp.DeviceProfile)
				}
			}
			return nil
		})
		return err
	})
	return err
}

func (b *BoltClient) DeleteDeviceProfile(dp models.DeviceProfile) error {
	return b.deleteById(dp.Id.Hex(), db.DeviceProfile)
}

//  -----------------------------------Addressable --------------------------*/
func (b *BoltClient) UpdateAddressable(ra *models.Addressable, r *models.Addressable) error {

	res := b.GetAddressableByName(ra, ra.Name)
	if res == nil {
		return db.ErrNotUnique
	}
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

	return b.update(db.Addressable, r, r.Id)
}

func (b *BoltClient) GetAddressables(d *[]models.Addressable) error {

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Addressable))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*d = []models.Addressable{}
		err := b.ForEach(func(id, encoded []byte) error {
			var a models.Addressable
			err := json.Unmarshal(encoded, &a)
			if err != nil {
				return err
			}
			*d = append(*d, a)
			return nil
		})
		return err
	})
	return err
}

func (b *BoltClient) GetAddressablesByTag(d *[]models.Addressable, tag string, filter interface{}) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Addressable))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		*d = []models.Addressable{}
		err := b.ForEach(func(id, encoded []byte) error {
			var a models.Addressable
			switch filter.(type) {
			case int:
				value := jsoniter.Get(encoded, tag).ToInt()
				if value == filter {
					err := json.Unmarshal(encoded, &a)
					if err != nil {
						return err
					}
					*d = append(*d, a)
				}
			case string:
				value := jsoniter.Get(encoded, tag).ToString()
				if value == filter {
					err := json.Unmarshal(encoded, &a)
					if err != nil {
						return err
					}
					*d = append(*d, a)
				}
			default:
			}
			return nil
		})
		return err
	})
	return err
}

func (b *BoltClient) GetAddressableById(a *models.Addressable, id string) error {
	if bson.IsObjectIdHex(id) {
		return b.getById(a, db.Addressable, id)
	} else {
		err := errors.New("boltGetAddressableById Invalid Object ID " + id)
		return err
	}
}

func (b *BoltClient) AddAddressable(a *models.Addressable) (bson.ObjectId, error) {
	// check if the name exist
	var dumy models.Addressable
	err := b.getByName(&dumy, db.Addressable, a.Name)
	if err == nil {
		return a.Id, db.ErrNotUnique
	}

	ts := db.MakeTimestamp()
	a.Created = ts
	a.Id = bson.NewObjectId()
	err = b.add(db.Addressable, a, a.Id)
	return a.Id, err
}

func (b *BoltClient) GetAddressableByName(a *models.Addressable, n string) error {
	return b.getByName(&a, db.Addressable, n)
}

func (b *BoltClient) GetAddressablesByTopic(a *[]models.Addressable, t string) error {
	return b.GetAddressablesByTag(a, "topic", t)
}

func (b *BoltClient) GetAddressablesByPort(a *[]models.Addressable, p int) error {
	return b.GetAddressablesByTag(a, "port", p)
}

func (b *BoltClient) GetAddressablesByPublisher(a *[]models.Addressable, p string) error {
	return b.GetAddressablesByTag(a, "publisher", p)
}

func (b *BoltClient) GetAddressablesByAddress(a *[]models.Addressable, add string) error {
	return b.GetAddressablesByTag(a, "address", add)
}

func (b *BoltClient) DeleteAddressable(a models.Addressable) error {
	return b.deleteById(a.Id.Hex(), db.Addressable)
}

/* ----------------------------- Device Service ----------------------------------*/
func (b *BoltClient) GetDeviceServiceByName(d *models.DeviceService, n string) error {
	bds := boltDeviceService{DeviceService: *d}
	err := b.getByName(&bds, db.DeviceService, n)
	*d = bds.DeviceService
	return err
}

func (b *BoltClient) GetDeviceServiceById(d *models.DeviceService, id string) error {
	if bson.IsObjectIdHex(id) {
		bds := boltDeviceService{DeviceService: *d}
		err := b.getById(&bds, db.DeviceService, id)
		*d = bds.DeviceService
		return err
	} else {
		err := errors.New("boltGetDeviceServiceByName Invalid Object ID " + id)
		return err
	}
}

func (b *BoltClient) GetAllDeviceServices(d *[]models.DeviceService) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceService))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*d = []models.DeviceService{}
		err := b.ForEach(func(id, encoded []byte) error {
			var bds boltDeviceService
			err := json.Unmarshal(encoded, &bds)
			if err != nil {
				return err
			}
			*d = append(*d, bds.DeviceService)
			return nil
		})
		return err
	})
	return err
}

func (b *BoltClient) GetDeviceServicesByAddressableId(d *[]models.DeviceService, ide string) error {
	if bson.IsObjectIdHex(ide) {
		err := b.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(db.DeviceService))
			if b == nil {
				return db.ErrUnsupportedDatabase
			}
			*d = []models.DeviceService{}
			json := jsoniter.ConfigCompatibleWithStandardLibrary
			err := b.ForEach(func(id, encoded []byte) error {
				var bds boltDeviceService

				value := jsoniter.Get(encoded, "addressableId").ToString()
				if value == ide {
					err := json.Unmarshal(encoded, &bds)
					if err != nil {
						return err
					}
					*d = append(*d, bds.DeviceService)
				}
				return nil
			})
			return err
		})
		return err

	} else {
		err := errors.New("boltGetDeviceServicesByAddressableId Invalid Object ID " + ide)
		return err
	}
}

func (b *BoltClient) GetDeviceServicesWithLabel(d *[]models.DeviceService, l string) error {
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.DeviceService))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*d = []models.DeviceService{}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var bds boltDeviceService
			err := json.Unmarshal(encoded, &bds)
			if err != nil {
				return err
			}
			for i := 0; i < len(bds.Labels); i++ {
				value := jsoniter.Get(encoded, "labels", i).ToString()
				if value == l {
					*d = append(*d, bds.DeviceService)
				}
			}
			return nil
		})
		return err
	})
	return err

}

func (b *BoltClient) AddDeviceService(d *models.DeviceService) error {
	dummy := models.DeviceService{}
	err := b.GetDeviceServiceByName(&dummy, d.Name)
	if err == nil {
		return db.ErrNotUnique
	}

	d.Created = db.MakeTimestamp()
	d.Id = bson.NewObjectId()
	bds := boltDeviceService{DeviceService: *d}
	err = b.add(db.DeviceService, bds, bds.Id)
	return err
}

func (b *BoltClient) UpdateDeviceService(d models.DeviceService) error {

	d.Modified = db.MakeTimestamp()
	bds := boltDeviceService{DeviceService: d}
	return b.update(db.DeviceService, bds, bds.Id)
}

func (b *BoltClient) DeleteDeviceService(ds models.DeviceService) error {
	return b.deleteById(ds.Id.Hex(), db.DeviceService)
}

//  ----------------------Provision Watcher -----------------------------*/
func (b *BoltClient) GetAllProvisionWatchers(pw *[]models.ProvisionWatcher) error {
	return nil
}

func (b *BoltClient) GetProvisionWatcherByName(pw *models.ProvisionWatcher, n string) error {
	return nil
}

func (b *BoltClient) GetProvisionWatchersByIdentifier(pw *[]models.ProvisionWatcher, k string, v string) error {
	return nil
}

func (b *BoltClient) GetProvisionWatchersByServiceId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		return errors.New("mgoGetProvisionWatchersByServiceId Invalid Object ID " + id)
	}
}

func (b *BoltClient) GetProvisionWatchersByProfileId(pw *[]models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		err := errors.New("mgoGetProvisionWatcherByProfileId Invalid Object ID " + id)
		return err
	}
}

func (b *BoltClient) GetProvisionWatcherById(pw *models.ProvisionWatcher, id string) error {
	if bson.IsObjectIdHex(id) {
		return nil
	} else {
		err := errors.New("mgoGetProvisionWatcherById Invalid Object ID " + id)
		return err
	}
}

func (b *BoltClient) GetProvisionWatcher(pw *models.ProvisionWatcher, q bson.M) error {

	return nil
}

func (b *BoltClient) GetProvisionWatchers(pw *[]models.ProvisionWatcher, q bson.M) error {

	return nil
}

func (b *BoltClient) AddProvisionWatcher(pw *models.ProvisionWatcher) error {

	return nil
}

func (b *BoltClient) UpdateProvisionWatcher(pw models.ProvisionWatcher) error {

	return nil
}

func (b *BoltClient) DeleteProvisionWatcher(pw models.ProvisionWatcher) error {

	return nil
}

//  ------------------------Command -------------------------------------*/
func (b *BoltClient) GetAllCommands(c *[]models.Command) error {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Command))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*c = []models.Command{}
		err := b.ForEach(func(id, encoded []byte) error {
			var a models.Command
			err := json.Unmarshal(encoded, &a)
			if err != nil {
				return err
			}
			*c = append(*c, a)
			return nil
		})
		return err
	})
	return err
}

func (b *BoltClient) GetCommandById(c *models.Command, id string) error {
	if bson.IsObjectIdHex(id) {
		return b.getById(c, db.Command, id)
	} else {
		return errors.New("boltGetCommandById Invalid Object ID " + id)
	}
}

func (b *BoltClient) GetCommandByName(c *[]models.Command, n string) error {
	// Don't use getByName, can be various commands with same name
	err := b.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Command))
		if b == nil {
			return db.ErrUnsupportedDatabase
		}
		*c = []models.Command{}
		json := jsoniter.ConfigCompatibleWithStandardLibrary
		err := b.ForEach(func(id, encoded []byte) error {
			var a models.Command
			value := jsoniter.Get(encoded, "name").ToString()
			if value == n {
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

func (b *BoltClient) AddCommand(c *models.Command) error {
	c.Created = db.MakeTimestamp()
	c.Id = bson.NewObjectId()
	return b.add(db.Command, c, c.Id)
}

// Update command uses the ID of the command for identification
func (b *BoltClient) UpdateCommand(c *models.Command, r *models.Command) error {

	// Check if the command has a valid ID
	if len(c.Id.Hex()) == 0 || !bson.IsObjectIdHex(c.Id.Hex()) {
		return errors.New("ID required for updating a command")
	}

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

	return b.update(db.Command, r, r.Id)
}

// Delete the command by ID
// Check if the command is still in use by device profiles
func (b *BoltClient) DeleteCommandById(id string) error {
	if !bson.IsObjectIdHex(id) {
		return db.ErrInvalidObjectId
	}
	return b.deleteById(id, db.Command)
}
