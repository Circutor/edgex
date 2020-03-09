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

	bolt "github.com/coreos/bbolt"
	"github.com/edgexfoundry/edgex-go/internal/pkg/db"
	"github.com/edgexfoundry/edgex-go/pkg/models"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

/* ----------------------------- Device ---------------------------------- */
func (bc *BoltClient) AddDevice(d models.Device) (string, error) {
	// Check if the name exist
	_, err := bc.GetDeviceByName(d.Name)
	if err == nil {
		return "", db.ErrNotUnique
	}

	d.Id = uuid.New().String()
	d.Created = db.MakeTimestamp()
	d.Modified = d.Created

	bd := boltDevice{Device: d}
	return d.Id, bc.add(db.Device, bd, d.Id)
}

func (bc *BoltClient) UpdateDevice(d models.Device) error {
	d.Modified = db.MakeTimestamp()
	bd := boltDevice{Device: d}
	return bc.update(db.Device, bd, bd.Id)
}

func (bc *BoltClient) DeleteDeviceById(id string) error {
	return bc.deleteById(id, db.Device)
}

func (bc *BoltClient) GetAllDevices() ([]models.Device, error) {
	return bc.getDevicesBy(func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetDeviceById(id string) (models.Device, error) {
	bd := boltDevice{}
	err := bc.getById(&bd, db.Device, id)
	d := bd.Device
	return d, err
}

func (bc *BoltClient) GetDeviceByName(name string) (models.Device, error) {
	bd := boltDevice{}
	err := bc.getByName(&bd, db.Device, name)
	d := bd.Device
	return d, err
}

func (bc *BoltClient) GetDevicesByServiceId(sid string) ([]models.Device, error) {
	// Check if this device service exists
	err := bc.checkId(db.DeviceService, sid)
	if err != nil {
		return []models.Device{}, err
	}

	return bc.getDevicesBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "serviceId").ToString()
		if value == sid {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetDevicesByAddressableId(aid string) ([]models.Device, error) {
	// Check if this addressable exists
	err := bc.checkId(db.Addressable, aid)
	if err != nil {
		return []models.Device{}, err
	}

	return bc.getDevicesBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "addressableId").ToString()
		if value == aid {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetDevicesByProfileId(pid string) ([]models.Device, error) {
	// Check if this device profile exists
	err := bc.checkId(db.DeviceProfile, pid)
	if err != nil {
		return []models.Device{}, err
	}

	return bc.getDevicesBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "profileId").ToString()
		if value == pid {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetDevicesWithLabel(label string) ([]models.Device, error) {
	return bc.getDevicesBy(func(encoded []byte) bool {
		labels := jsoniter.Get(encoded, "labels").GetInterface().([]interface{})
		for _, value := range labels {
			if label == value.(string) {
				return true
			}
		}
		return false
	})
}

func (bc *BoltClient) getDevicesBy(fn func(encoded []byte) bool) ([]models.Device, error) {
	bd := boltDevice{}
	ds := []models.Device{}
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
				ds = append(ds, bd.Device)
			}
			return nil
		})
		return err
	})
	return ds, err
}

/* -----------------------------Device Profile -----------------------------*/
func (bc *BoltClient) GetDeviceProfileById(id string) (models.DeviceProfile, error) {
	bdp := boltDeviceProfile{}
	err := bc.getById(&bdp, db.DeviceProfile, id)
	dp := bdp.DeviceProfile
	return dp, err
}

func (bc *BoltClient) GetAllDeviceProfiles() ([]models.DeviceProfile, error) {
	return bc.getDeviceProfilesBy(func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetDeviceProfilesByModel(model string) ([]models.DeviceProfile, error) {
	return bc.getDeviceProfilesBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "model").ToString()
		if value == model {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetDeviceProfilesWithLabel(label string) ([]models.DeviceProfile, error) {
	return bc.getDeviceProfilesBy(func(encoded []byte) bool {
		labels := jsoniter.Get(encoded, "labels").GetInterface().([]interface{})
		for _, value := range labels {
			if label == value.(string) {
				return true
			}
		}
		return false
	})
}

func (bc *BoltClient) GetDeviceProfilesByManufacturerModel(man string, mod string) ([]models.DeviceProfile, error) {
	return bc.getDeviceProfilesBy(func(encoded []byte) bool {
		valueMod := jsoniter.Get(encoded, "model").ToString()
		valueMan := jsoniter.Get(encoded, "manufacturer").ToString()
		if valueMod == mod && valueMan == man {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetDeviceProfilesByManufacturer(man string) ([]models.DeviceProfile, error) {
	return bc.getDeviceProfilesBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "manufacturer").ToString()
		if value == man {
			return true
		}
		return false
	})
}

func (bc *BoltClient) getDeviceProfilesBy(fn func(encoded []byte) bool) ([]models.DeviceProfile, error) {
	bdp := boltDeviceProfile{}
	dps := []models.DeviceProfile{}
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
				dps = append(dps, bdp.DeviceProfile)
			}
			return nil
		})
		return err
	})
	return dps, err
}

func (bc *BoltClient) GetDeviceProfileByName(name string) (models.DeviceProfile, error) {
	bdp := boltDeviceProfile{}
	err := bc.getByName(&bdp, db.DeviceProfile, name)
	dp := bdp.DeviceProfile
	return dp, err
}

func (bc *BoltClient) AddDeviceProfile(dp models.DeviceProfile) (string, error) {
	// Check if the name exist
	_, err := bc.GetDeviceProfileByName(dp.Name)
	if err == nil {
		return "", db.ErrNotUnique
	}

	for i := 0; i < len(dp.Commands); i++ {
		if newId, errs := bc.AddCommand(dp.Commands[i]); errs != nil {
			return "", errs
		} else {
			dp.Commands[i].Id = newId
		}
	}

	dp.Id = uuid.New().String()
	dp.Created = db.MakeTimestamp()
	dp.Modified = dp.Created

	bdp := boltDeviceProfile{DeviceProfile: dp}
	return dp.Id, bc.add(db.DeviceProfile, bdp, dp.Id)
}

func (bc *BoltClient) UpdateDeviceProfile(dp models.DeviceProfile) error {
	dp.Modified = db.MakeTimestamp()
	bdp := boltDeviceProfile{DeviceProfile: dp}
	return bc.update(db.DeviceProfile, bdp, bdp.Id)
}

// Get the device profiles that are currently using the command
func (bc *BoltClient) GetDeviceProfilesByCommandId(id string) ([]models.DeviceProfile, error) {
	// Check if this command exists
	err := bc.checkId(db.Command, id)
	if err != nil {
		return []models.DeviceProfile{}, err
	}

	return bc.getDeviceProfilesBy(func(encoded []byte) bool {
		commands := jsoniter.Get(encoded, "commands").GetInterface().([]interface{})
		for _, value := range commands {
			if id == value.(string) {
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
func (bc *BoltClient) UpdateAddressable(a models.Addressable) error {
	a.Modified = db.MakeTimestamp()

	return bc.update(db.Addressable, a, a.Id)
}

func (bc *BoltClient) GetAddressables() ([]models.Addressable, error) {
	return bc.getAddressablesBy(func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) getAddressablesBy(fn func(encoded []byte) bool) ([]models.Addressable, error) {
	a := models.Addressable{}
	as := []models.Addressable{}
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
				as = append(as, a)
			}
			return nil
		})
		return err
	})
	return as, err
}

func (bc *BoltClient) GetAddressableById(id string) (models.Addressable, error) {
	var a models.Addressable
	err := bc.getById(&a, db.Addressable, id)
	return a, err
}

func (bc *BoltClient) AddAddressable(a models.Addressable) (string, error) {
	// Check if the name exist
	var dummy models.Addressable
	err := bc.getByName(&dummy, db.Addressable, a.Name)
	if err == nil {
		return dummy.Id, db.ErrNotUnique
	}

	a.Id = uuid.New().String()
	a.Created = db.MakeTimestamp()
	a.Modified = a.Created

	err = bc.add(db.Addressable, a, a.Id)
	return a.Id, err
}

func (bc *BoltClient) GetAddressableByName(name string) (models.Addressable, error) {
	var a models.Addressable
	err := bc.getByName(&a, db.Addressable, name)
	return a, err
}

func (bc *BoltClient) GetAddressablesByTopic(topic string) ([]models.Addressable, error) {
	return bc.getAddressablesBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "topic").ToString()
		if value == topic {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetAddressablesByPort(port int) ([]models.Addressable, error) {
	return bc.getAddressablesBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "port").ToInt()
		if value == port {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetAddressablesByPublisher(publisher string) ([]models.Addressable, error) {
	return bc.getAddressablesBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "publisher").ToString()
		if value == publisher {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetAddressablesByAddress(address string) ([]models.Addressable, error) {
	return bc.getAddressablesBy(func(encoded []byte) bool {
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
func (bc *BoltClient) GetDeviceServiceByName(name string) (models.DeviceService, error) {
	bds := boltDeviceService{}
	err := bc.getByName(&bds, db.DeviceService, name)
	ds := bds.DeviceService
	return ds, err
}

func (bc *BoltClient) GetDeviceServiceById(id string) (models.DeviceService, error) {
	bds := boltDeviceService{}
	err := bc.getById(&bds, db.DeviceService, id)
	ds := bds.DeviceService
	return ds, err
}

func (bc *BoltClient) GetAllDeviceServices() ([]models.DeviceService, error) {
	return bc.getDeviceServicesBy(func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetDeviceServicesByAddressableId(id string) ([]models.DeviceService, error) {
	// Check if this addressable exists
	err := bc.checkId(db.Addressable, id)
	if err != nil {
		return []models.DeviceService{}, err
	}

	return bc.getDeviceServicesBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "addressableId").ToString()
		if value == id {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetDeviceServicesWithLabel(label string) ([]models.DeviceService, error) {
	return bc.getDeviceServicesBy(func(encoded []byte) bool {
		labels := jsoniter.Get(encoded, "labels").GetInterface().([]interface{})
		for _, value := range labels {
			if label == value.(string) {
				return true
			}
		}
		return false
	})
}

func (bc *BoltClient) getDeviceServicesBy(fn func(encoded []byte) bool) ([]models.DeviceService, error) {
	bds := boltDeviceService{}
	dss := []models.DeviceService{}
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
				dss = append(dss, bds.DeviceService)
			}
			return nil
		})
		return err
	})
	return dss, err
}

func (bc *BoltClient) AddDeviceService(ds models.DeviceService) (string, error) {
	// Check if the name exist
	_, err := bc.GetDeviceServiceByName(ds.Name)
	if err == nil {
		return "", db.ErrNotUnique
	}

	ds.Id = uuid.New().String()
	ds.Created = db.MakeTimestamp()
	ds.Modified = ds.Created

	bds := boltDeviceService{DeviceService: ds}
	return ds.Id, bc.add(db.DeviceService, bds, bds.Id)
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
func (bc *BoltClient) GetAllProvisionWatchers() ([]models.ProvisionWatcher, error) {
	return bc.getProvisionWatchersBy(func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) GetProvisionWatcherByName(name string) (models.ProvisionWatcher, error) {
	bpw := boltProvisionWatcher{}
	err := bc.getByName(&bpw, db.ProvisionWatcher, name)
	pw := bpw.ProvisionWatcher
	return pw, err
}

func (bc *BoltClient) GetProvisionWatchersByIdentifier(k string, v string) ([]models.ProvisionWatcher, error) {
	return bc.getProvisionWatchersBy(func(encoded []byte) bool {
		identifier := jsoniter.Get(encoded, "identifiers").ToString()
		keyvalue := "\"" + k + "\"" + ":" + "\"" + v + "\""
		if bytes.Contains([]byte(identifier), []byte(keyvalue)) {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetProvisionWatchersByServiceId(id string) ([]models.ProvisionWatcher, error) {
	// Check if this device service exists
	err := bc.checkId(db.DeviceService, id)
	if err != nil {
		return []models.ProvisionWatcher{}, err
	}

	return bc.getProvisionWatchersBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "serviceId").ToString()
		if value == id {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetProvisionWatchersByProfileId(id string) ([]models.ProvisionWatcher, error) {
	// Check if this device profile exists
	err := bc.checkId(db.DeviceProfile, id)
	if err != nil {
		return []models.ProvisionWatcher{}, err
	}

	return bc.getProvisionWatchersBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "profileId").ToString()
		if value == id {
			return true
		}
		return false
	})
}

func (bc *BoltClient) GetProvisionWatcherById(id string) (models.ProvisionWatcher, error) {
	bpw := boltProvisionWatcher{}
	err := bc.getById(&bpw, db.ProvisionWatcher, id)
	pw := bpw.ProvisionWatcher
	return pw, err
}

func (bc *BoltClient) getProvisionWatchersBy(fn func(encoded []byte) bool) ([]models.ProvisionWatcher, error) {
	bpw := boltProvisionWatcher{}
	pws := []models.ProvisionWatcher{}
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
				pws = append(pws, bpw.ProvisionWatcher)
			}
			return nil
		})
		return err
	})
	return pws, err
}

func (bc *BoltClient) AddProvisionWatcher(pw models.ProvisionWatcher) (string, error) {
	// Check if the name exist
	_, err := bc.GetProvisionWatcherByName(pw.Name)
	if err == nil {
		return "", db.ErrNotUnique
	}

	pw.Id = uuid.New().String()
	pw.Created = db.MakeTimestamp()
	pw.Modified = pw.Created

	bpw := boltProvisionWatcher{ProvisionWatcher: pw}
	return pw.Id, bc.add(db.ProvisionWatcher, bpw, pw.Id)
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
func (bc *BoltClient) GetAllCommands() ([]models.Command, error) {
	return bc.getCommandsBy(func(encoded []byte) bool {
		return true
	})
}

func (bc *BoltClient) getCommandsBy(fn func(encoded []byte) bool) ([]models.Command, error) {
	c := models.Command{}
	cs := []models.Command{}
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
				cs = append(cs, c)
			}
			return nil
		})
		return err
	})
	return cs, err
}

func (bc *BoltClient) GetCommandById(id string) (models.Command, error) {
	var c models.Command
	err := bc.getById(&c, db.Command, id)
	return c, err
}

func (bc *BoltClient) GetCommandByName(name string) ([]models.Command, error) {
	return bc.getCommandsBy(func(encoded []byte) bool {
		value := jsoniter.Get(encoded, "name").ToString()
		if value == name {
			return true
		}
		return false
	})
}

func (bc *BoltClient) AddCommand(c models.Command) (string, error) {
	c.Id = uuid.New().String()
	c.Created = db.MakeTimestamp()
	c.Modified = c.Created

	return c.Id, bc.add(db.Command, c, c.Id)
}

func (bc *BoltClient) UpdateCommand(c models.Command) error {
	c.Modified = db.MakeTimestamp()
	return bc.update(db.Command, c, c.Id)
}

func (bc *BoltClient) DeleteCommandById(id string) error {
	return bc.deleteById(id, db.Command)
}

// Scrub all metadata
func (bc *BoltClient) ScrubMetadata() error {
	err := bc.scrubAll(db.Addressable)
	if err != nil {
		return err
	}
	err = bc.scrubAll(db.DeviceService)
	if err != nil {
		return err
	}
	err = bc.scrubAll(db.DeviceProfile)
	if err != nil {
		return err
	}
	err = bc.scrubAll(db.Device)
	if err != nil {
		return err
	}
	err = bc.scrubAll(db.Command)
	if err != nil {
		return err
	}
	err = bc.scrubAll(db.ProvisionWatcher)
	if err != nil {
		return err
	}

	return nil
}
