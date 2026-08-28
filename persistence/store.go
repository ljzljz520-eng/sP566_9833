package persistence

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

	"go.etcd.io/bbolt"
	"pharmacycounter/model"
)

var (
	ErrNotFound         = errors.New("record not found")
	patientsBucket      = []byte("patients")
	prescriptionsBucket = []byte("prescriptions")
	ticketsBucket       = []byte("tickets")
	claimsBucket        = []byte("claims")
	snapshotBucket      = []byte("snapshots")
)

type Store struct {
	db *bbolt.DB
	mu sync.RWMutex
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("database path is empty")
	}
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	store := &Store{db: db}
	if err := store.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range [][]byte{patientsBucket, prescriptionsBucket, ticketsBucket, claimsBucket, snapshotBucket} {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func encode(value any) ([]byte, error) { return json.Marshal(value) }

func put(db *bbolt.DB, bucket []byte, key string, value any) error {
	data, err := encode(value)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(bucket).Put([]byte(key), data) })
}

func get(db *bbolt.DB, bucket []byte, key string, target any) error {
	return db.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucket).Get([]byte(key))
		if value == nil {
			return ErrNotFound
		}
		return json.Unmarshal(value, target)
	})
}

func (s *Store) SavePatient(patient model.PatientRecord) error {
	if err := patient.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return put(s.db, patientsBucket, patient.ID, patient)
}

func (s *Store) Patient(id string) (model.PatientRecord, error) {
	var patient model.PatientRecord
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return patient, errors.New("store is closed")
	}
	err := get(s.db, patientsBucket, id, &patient)
	return patient, err
}

func (s *Store) SavePrescription(order model.PrescriptionOrder) error {
	if err := order.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return put(s.db, prescriptionsBucket, order.ID, order)
}

func (s *Store) Prescription(id string) (model.PrescriptionOrder, error) {
	var order model.PrescriptionOrder
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return order, errors.New("store is closed")
	}
	err := get(s.db, prescriptionsBucket, id, &order)
	return order, err
}

func (s *Store) SaveTicket(ticket model.PharmacyTicket) error {
	if err := ticket.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return put(s.db, ticketsBucket, ticket.Number, ticket)
}

func (s *Store) Ticket(number string) (model.PharmacyTicket, error) {
	var ticket model.PharmacyTicket
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return ticket, errors.New("store is closed")
	}
	err := get(s.db, ticketsBucket, number, &ticket)
	return ticket, err
}

func (s *Store) SaveClaim(event model.ClaimEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return put(s.db, claimsBucket, event.ID, event)
}

func (s *Store) Claim(id string) (model.ClaimEvent, error) {
	var event model.ClaimEvent
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return event, errors.New("store is closed")
	}
	err := get(s.db, claimsBucket, id, &event)
	return event, err
}

func (s *Store) SaveSnapshot(snapshot model.QueueSnapshot) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return put(s.db, snapshotBucket, "current", snapshot)
}

func (s *Store) Snapshot() (model.QueueSnapshot, error) {
	var snapshot model.QueueSnapshot
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return snapshot, errors.New("store is closed")
	}
	err := get(s.db, snapshotBucket, "current", &snapshot)
	return snapshot, err
}

func list[T any](db *bbolt.DB, bucket []byte) ([]T, error) {
	items := make([]T, 0)
	err := db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucket).ForEach(func(_, value []byte) error {
			if value == nil {
				return nil
			}
			var item T
			if err := json.Unmarshal(value, &item); err != nil {
				return err
			}
			items = append(items, item)
			return nil
		})
	})
	return items, err
}

func (s *Store) Tickets() ([]model.PharmacyTicket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	items, err := list[model.PharmacyTicket](s.db, ticketsBucket)
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedOrder < items[j].CreatedOrder })
	return items, err
}

func (s *Store) Claims() ([]model.ClaimEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	items, err := list[model.ClaimEvent](s.db, claimsBucket)
	sort.Slice(items, func(i, j int) bool { return items[i].Sequence < items[j].Sequence })
	return items, err
}

func (s *Store) Patients() ([]model.PatientRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	items, err := list[model.PatientRecord](s.db, patientsBucket)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, err
}

func (s *Store) Prescriptions() ([]model.PrescriptionOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return nil, errors.New("store is closed")
	}
	items, err := list[model.PrescriptionOrder](s.db, prescriptionsBucket)
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, err
}

func DatabaseExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
