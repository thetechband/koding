package koding

import (
	"fmt"
	"koding/db/models"
	"koding/kites/kloud/contexthelper/session"
	"koding/kites/kloud/eventer"
	"koding/kites/kloud/klient"
	"koding/kites/kloud/kloud"
	"koding/kites/kloud/machinestate"
	"koding/kites/kloud/plans"
	"time"

	"golang.org/x/net/context"

	"github.com/koding/logging"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var callCount int

// Machine represents a single MongodDB document that represents a Koding
// Provider from the jMachines collection.
type Machine struct {
	Id          bson.ObjectId `bson:"_id" json:"-"`
	Label       string        `bson:"label"`
	Domain      string        `bson:"domain"`
	QueryString string        `bson:"queryString"`
	IpAddress   string        `bson:"ipAddress"`
	Assignee    struct {
		InProgress      bool      `bson:"inProgress"`
		AssignedAt      time.Time `bson:"assignedAt"`
		KlientMissingAt time.Time `bson:"klientMissingAt"`
	} `bson:"assignee"`
	Status struct {
		State      string    `bson:"state"`
		Reason     string    `bson:"reason"`
		ModifiedAt time.Time `bson:"modifiedAt"`
	} `bson:"status"`
	Provider   string    `bson:"provider"`
	Credential string    `bson:"credential"`
	CreatedAt  time.Time `bson:"createdAt"`
	Meta       struct {
		AlwaysOn     bool   `bson:"alwaysOn"`
		InstanceId   string `structs:"instanceId" bson:"instanceId"`
		InstanceType string `structs:"instance_type" bson:"instance_type"`
		InstanceName string `structs:"instanceName" bson:"instanceName"`
		Region       string `structs:"region" bson:"region"`
		StorageSize  int    `structs:"storage_size" bson:"storage_size"`
		SourceAmi    string `structs:"source_ami" bson:"source_ami"`
		SnapshotId   string `structs:"snapshotId" bson:"snapshotId"`
	} `bson:"meta"`
	Users  []models.Permissions `bson:"users"`
	Groups []models.Permissions `bson:"groups"`

	// internal fields, not availabile in MongoDB schema
	Username string                 `bson:"-"`
	User     *models.User           `bson:"-"`
	Payment  *plans.PaymentResponse `bson:"-"`
	Checker  plans.Checker          `bson:"-"`
	Session  *session.Session       `bson:"-"`
	Log      logging.Logger         `bson:"-"`
	locker   kloud.Locker           `bson:"-"`

	// cleanFuncs are a list of functions that are called when after a method
	// is finished
	cleanFuncs []func()
}

// runCleanupFunctions calls all cleanup functions and set the
// list to nil. Once called any other call will not have any
// effect.
func (m *Machine) runCleanupFunctions() {
	if m.cleanFuncs == nil {
		return
	}

	for _, fn := range m.cleanFuncs {
		fn()
	}

	m.cleanFuncs = nil
}

func (m *Machine) State() machinestate.State {
	return machinestate.States[m.Status.State]
}

func (m *Machine) PublicIpAddress() string {
	return m.IpAddress
}

// push pushes the given message to the eventer
func (m *Machine) push(msg string, percentage int, state machinestate.State) {
	if m.Session.Eventer != nil {
		m.Session.Eventer.Push(&eventer.Event{
			Message:    msg,
			Percentage: percentage,
			Status:     state,
		})
	}
}

// switchAWSRegion switches to the given AWS region. This should be only used when
// you know what to do, otherwiese never, never change the region of a machine.
func (m *Machine) switchAWSRegion(region string) error {
	m.Meta.InstanceId = "" // we neglect any previous instanceId
	m.QueryString = ""     //
	m.Meta.Region = "us-east-1"

	client, err := m.Session.AWSClients.Region("us-east-1")
	if err != nil {
		return err
	}
	m.Session.AWSClient.Client = client

	return m.Session.DB.Run("jMachines", func(c *mgo.Collection) error {
		return c.UpdateId(
			m.Id,
			bson.M{"$set": bson.M{
				"meta.instanceId": "",
				"queryString":     "",
				"meta.region":     "us-east-1",
			}},
		)
	})
}

// markAsNotInitialized marks the machine as NotInitialized by cleaning up all
// necessary fields and marking the VM as notinitialized so the User can build
// it again.
func (m *Machine) markAsNotInitialized() error {
	m.Log.Warning("Instance is not available. Marking it as NotInitialized")
	if err := m.Session.DB.Run("jMachines", func(c *mgo.Collection) error {
		return c.UpdateId(
			m.Id,
			bson.M{"$set": bson.M{
				"ipAddress":         "",
				"queryString":       "",
				"meta.instanceType": "",
				"meta.instanceName": "",
				"meta.instanceId":   "",
				"status.state":      machinestate.NotInitialized.String(),
				"status.modifiedAt": time.Now().UTC(),
				"status.reason":     "Machine is marked as NotInitialized",
			}},
		)
	}); err != nil {
		return err
	}

	m.IpAddress = ""
	m.QueryString = ""
	m.Meta.InstanceType = ""
	m.Meta.InstanceName = ""
	m.Meta.InstanceId = ""

	// so any State() method can return the correct status
	m.Status.State = machinestate.NotInitialized.String()
	return nil
}

func (m *Machine) markAsStopped() error {
	return m.markAsStoppedWithReason("Machine is stopped")
}

func (m *Machine) markAsStoppedWithReason(reason string) error {
	m.Log.Debug("Marking instance as stopped")
	if err := m.Session.DB.Run("jMachines", func(c *mgo.Collection) error {
		return c.UpdateId(
			m.Id,
			bson.M{"$set": bson.M{
				"ipAddress":         "",
				"status.state":      machinestate.Stopped.String(),
				"status.modifiedAt": time.Now().UTC(),
				"status.reason":     reason,
			}},
		)
	}); err != nil {
		return err
	}

	// so any State() method returns the correct status
	m.Status.State = machinestate.Stopped.String()
	m.IpAddress = ""
	return nil
}

func (m *Machine) updateStorageSize(size int) error {
	return m.Session.DB.Run("jMachines", func(c *mgo.Collection) error {
		return c.UpdateId(
			m.Id,
			bson.M{"$set": bson.M{"meta.storage_size": size}},
		)
	})
}

func (m *Machine) isKlientReady() bool {
	m.Log.Debug("All finished, testing for klient connection IP [%s]", m.IpAddress)
	klientRef, err := klient.NewWithTimeout(m.Session.Kite, m.QueryString, time.Minute*5)
	if err != nil {
		m.Log.Warning("Connecting to remote Klient instance err: %s", err)
		return false
	}
	defer klientRef.Close()

	m.Log.Debug("Sending a ping message")
	if err := klientRef.Ping(); err != nil {
		m.Log.Debug("Sending a ping message err: %s", err)
		return false
	}

	return true
}

// Lock performs a Lock on this Machine
func (m *Machine) Lock() error {
	if !m.Id.Valid() {
		return kloud.NewError(kloud.ErrMachineIdMissing)
	}

	if m.locker == nil {
		return fmt.Errorf("Machine '%s' missing Locker", m.Id.Hex())
	}

	return m.locker.Lock(m.Id.Hex())
}

// Unlock performs an Unlock on this Machine instance
func (m *Machine) Unlock() error {
	if !m.Id.Valid() {
		return kloud.NewError(kloud.ErrMachineIdMissing)
	}

	if m.locker == nil {
		return fmt.Errorf("Machine '%s' missing Locker", m.Id.Hex())
	}

	// Unlock does not return an error
	m.locker.Unlock(m.Id.Hex())
	return nil
}

// klientIsNotMissing will unset the `assignee.klientMissingAt` value
// from the database, only if the Machine.Assignee.KlientMissingAt value
// has data. Therefor it is safe to call as frequently.
func (m *Machine) klientIsNotMissing() error {
	if m.Assignee.KlientMissingAt.IsZero() {
		return nil
	}

	m.Log.Debug("Clearing assignee.klientMissingAt")

	return m.Session.DB.Run("jMachines", func(c *mgo.Collection) error {
		return c.UpdateId(
			m.Id,
			bson.M{"$unset": bson.M{"assignee.klientMissingAt": ""}},
		)
	})
}

// stopIfKlientIsMissing will stop the current Machine X minutes after
// the `assignee.klientMissingAt` value. If the value does not exist in
// the databse, it will write it and return.
//
// Therefor, this method is expected be called as often as needed,
// and will shutdown the Machine if klient has been missing for too long.
func (m *Machine) stopIfKlientIsMissing(ctx context.Context) error {

	// If this is the first time Klient has been found missing,
	// set the missingat time and return
	if m.Assignee.KlientMissingAt.IsZero() {
		m.Log.Debug("Klient has been reported missing, recording this as the first time it went missing")

		return m.Session.DB.Run("jMachines", func(c *mgo.Collection) error {
			return c.UpdateId(
				m.Id,
				bson.M{"$set": bson.M{"assignee.klientMissingAt": time.Now().UTC()}},
			)
		})
	}

	// If the klient has been missing less than X minutes, don't stop
	if time.Since(m.Assignee.KlientMissingAt) < time.Minute*5 {
		return nil
	}

	// lock so it doesn't interfere with others.
	err := m.Lock()
	defer m.Unlock()
	if err != nil {
		return err
	}

	callCount++

	// Hasta la vista, baby!
	m.Log.Info("======> STOP started (missing klient) <======")
	if err := m.Stop(ctx); err != nil {
		m.Log.Info("======> STOP aborted (missing klient: %s) <======", err)
		return err
	}
	m.Log.Info("======> STOP finished (missing klient) <======")

	// Clear the klientMissingAt field, or we risk Stopping the user's
	// machine next time they run it, without waiting the proper X minute
	// timeout.
	m.klientIsNotMissing()

	return nil
}
