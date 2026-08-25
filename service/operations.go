package service

import (
	"errors"
	"sort"
	"strings"

	"pharmacycounter/model"
	"pharmacycounter/persistence"
	"pharmacycounter/queue"
)

type Dashboard struct {
	Counts        map[string]int
	NextPending   model.PharmacyTicket
	HasNext       bool
	OpenWindows   map[string]int
	TotalPatients int
}

type PatientHistory struct {
	Patient model.PatientRecord
	Tickets []model.PharmacyTicket
	Orders  []model.PrescriptionOrder
}

func (c *Counter) Dashboard() (Dashboard, error) {
	if c == nil || c.store == nil {
		return Dashboard{}, errors.New("counter is unavailable")
	}
	counts, err := c.store.StatusCounts()
	if err != nil {
		return Dashboard{}, err
	}
	next, nextErr := c.queue.NextPending()
	if nextErr != nil && !errors.Is(nextErr, queue.ErrTicketMissing) {
		return Dashboard{}, nextErr
	}
	patients, err := c.store.Patients()
	if err != nil {
		return Dashboard{}, err
	}
	windows := make(map[string]int)
	for _, ticket := range c.Called() {
		if ticket.CalledBy != "" {
			windows[ticket.CalledBy]++
		}
	}
	return Dashboard{Counts: counts, NextPending: next, HasNext: nextErr == nil, OpenWindows: windows, TotalPatients: len(patients)}, nil
}

func (c *Counter) History(patientID string) (PatientHistory, error) {
	if c == nil || c.store == nil {
		return PatientHistory{}, errors.New("counter is unavailable")
	}
	patient, err := c.store.Patient(strings.TrimSpace(patientID))
	if err != nil {
		return PatientHistory{}, err
	}
	tickets, err := c.store.PatientHistory(patient.ID)
	if err != nil {
		return PatientHistory{}, err
	}
	orders, err := c.store.Prescriptions()
	if err != nil {
		return PatientHistory{}, err
	}
	filtered := make([]model.PrescriptionOrder, 0)
	for _, order := range orders {
		if order.PatientID == patient.ID {
			filtered = append(filtered, order)
		}
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })
	return PatientHistory{Patient: patient, Tickets: tickets, Orders: filtered}, nil
}

func (c *Counter) Reconcile() error {
	if c == nil || c.queue == nil {
		return errors.New("counter is unavailable")
	}
	if err := c.queue.Validate(); err != nil {
		return err
	}
	tickets, err := c.store.Tickets()
	if err != nil {
		return err
	}
	if err := model.QueueInvariant(tickets); err != nil {
		return err
	}
	return c.store.ReplaceSnapshot(c.queue.Snapshot())
}

func (c *Counter) ClaimAudit(number string) (int, error) {
	if c == nil || c.store == nil {
		return 0, errors.New("counter is unavailable")
	}
	return c.store.ClaimCount(number)
}

func (c *Counter) WindowAssignments(window string) ([]queue.Assignment, error) {
	if err := queue.ValidateWindow(window); err != nil {
		return nil, err
	}
	return c.queue.Assignments(window), nil
}

func (c *Counter) Ticket(number string) (model.PharmacyTicket, error) {
	if c == nil || c.store == nil {
		return model.PharmacyTicket{}, errors.New("counter is unavailable")
	}
	return c.store.Ticket(strings.TrimSpace(number))
}

func (c *Counter) ExportState() ([]byte, error) {
	if c == nil || c.store == nil {
		return nil, errors.New("counter is unavailable")
	}
	return c.store.ExportState()
}

func (c *Counter) EnsureReady() error {
	if c == nil || c.store == nil || c.queue == nil {
		return errors.New("counter is unavailable")
	}
	if _, err := c.store.Snapshot(); err != nil && !errors.Is(err, persistence.ErrNotFound) {
		return err
	}
	return c.Reconcile()
}
