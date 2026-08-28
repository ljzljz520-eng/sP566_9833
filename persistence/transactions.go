package persistence

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.etcd.io/bbolt"
	"pharmacycounter/model"
)

type IntakeTransaction struct {
	Patient      model.PatientRecord
	Prescription model.PrescriptionOrder
	Ticket       model.PharmacyTicket
	Claim        model.ClaimEvent
	Snapshot     model.QueueSnapshot
}

func (s *Store) SaveIntake(transaction IntakeTransaction) error {
	if err := transaction.Patient.Validate(); err != nil {
		return err
	}
	if err := transaction.Prescription.Validate(); err != nil {
		return err
	}
	if err := transaction.Ticket.Validate(); err != nil {
		return err
	}
	if err := transaction.Claim.Validate(); err != nil {
		return err
	}
	if transaction.Patient.ID != transaction.Prescription.PatientID || transaction.Ticket.PatientID != transaction.Patient.ID {
		return errors.New("intake entities do not belong together")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		entries := []struct {
			bucket []byte
			key    string
			value  any
		}{
			{patientsBucket, transaction.Patient.ID, transaction.Patient},
			{prescriptionsBucket, transaction.Prescription.ID, transaction.Prescription},
			{ticketsBucket, transaction.Ticket.Number, transaction.Ticket},
			{claimsBucket, transaction.Claim.ID, transaction.Claim},
			{snapshotBucket, "current", transaction.Snapshot},
		}
		for _, entry := range entries {
			data, err := json.Marshal(entry.value)
			if err != nil {
				return fmt.Errorf("encode %s: %w", entry.key, err)
			}
			if err := tx.Bucket(entry.bucket).Put([]byte(entry.key), data); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) UpdateTicket(ticket model.PharmacyTicket) error {
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

func (s *Store) ReplaceSnapshot(snapshot model.QueueSnapshot) error {
	if snapshot.Next < 1 {
		return errors.New("snapshot sequence must be positive")
	}
	snapshot.Pending = append([]string{}, snapshot.Pending...)
	snapshot.Called = append([]string{}, snapshot.Called...)
	snapshot.Finished = append([]string{}, snapshot.Finished...)
	return s.SaveSnapshot(snapshot)
}

func (s *Store) DeleteTicket(number string) error {
	if number == "" {
		return errors.New("ticket number is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error { return tx.Bucket(ticketsBucket).Delete([]byte(number)) })
}

func (s *Store) SaveTicketAndClaim(ticket model.PharmacyTicket, event model.ClaimEvent, snapshot model.QueueSnapshot) error {
	if err := ticket.Validate(); err != nil {
		return err
	}
	if err := event.Validate(); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return errors.New("store is closed")
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, entry := range []struct {
			bucket []byte
			key    string
			value  any
		}{{ticketsBucket, ticket.Number, ticket}, {claimsBucket, event.ID, event}, {snapshotBucket, "current", snapshot}} {
			data, err := json.Marshal(entry.value)
			if err != nil {
				return err
			}
			if err := tx.Bucket(entry.bucket).Put([]byte(entry.key), data); err != nil {
				return err
			}
		}
		return nil
	})
}
