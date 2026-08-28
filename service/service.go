package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"pharmacycounter/audit"
	"pharmacycounter/model"
	"pharmacycounter/persistence"
	"pharmacycounter/queue"
)

var (
	ErrSearchEmpty = errors.New("search query is empty")
	ErrNoMatches   = errors.New("no matching pharmacy records")
)

type IntakeResult struct {
	Patient model.PatientRecord
	Order   model.PrescriptionOrder
	Ticket  model.PharmacyTicket
	Receipt string
}

type SearchResult struct {
	Patients      []model.PatientRecord
	Prescriptions []model.PrescriptionOrder
	Tickets       []model.PharmacyTicket
}

type Counter struct {
	store *persistence.Store
	queue *queue.Queue
	audit *audit.Logger
	mu    sync.Mutex
	seq   int
}

func New(store *persistence.Store) (*Counter, error) {
	if store == nil {
		return nil, errors.New("store is required")
	}
	snapshot, err := store.Snapshot()
	if err != nil && !errors.Is(err, persistence.ErrNotFound) {
		return nil, err
	}
	tickets, err := store.Tickets()
	if err != nil {
		return nil, err
	}
	return &Counter{store: store, queue: queue.FromSnapshot(snapshot, tickets), audit: audit.New(), seq: 1}, nil
}

func (c *Counter) Register(patient model.PatientRecord, order model.PrescriptionOrder) (IntakeResult, error) {
	if patient.Priority == "" {
		patient.Priority = model.NormalizePriority(patient.Priority)
	}
	if order.PatientID == "" {
		order.PatientID = patient.ID
	}
	if err := patient.Validate(); err != nil {
		return IntakeResult{}, err
	}
	if strings.TrimSpace(patient.Phone) != "" {
		if err := model.ValidatePatientContact(patient); err != nil {
			return IntakeResult{}, err
		}
	}
	if err := order.Validate(); err != nil {
		return IntakeResult{}, err
	}
	if order.PatientID != patient.ID {
		return IntakeResult{}, errors.New("prescription belongs to another patient")
	}
	ticket, err := c.queue.Allocate(patient.ID, order.ID, patient.Priority == "urgent")
	if err != nil {
		return IntakeResult{}, err
	}
	c.mu.Lock()
	event := model.ClaimEvent{ID: c.eventID(), Ticket: ticket.Number, Window: "intake", Sequence: c.seq, Action: "register", Successful: true}
	c.seq++
	c.mu.Unlock()
	if err := c.audit.Record(event); err != nil {
		return IntakeResult{}, err
	}
	if err := c.store.SaveIntake(persistence.IntakeTransaction{Patient: patient, Prescription: order, Ticket: ticket, Claim: event, Snapshot: c.queue.Snapshot()}); err != nil {
		return IntakeResult{}, err
	}
	packet, err := model.NewIntakePacket(patient, order, ticket)
	if err != nil {
		return IntakeResult{}, err
	}
	return IntakeResult{Patient: patient, Order: order, Ticket: ticket, Receipt: fmt.Sprintf("%s for %s: %s", packet.Ticket.Number, packet.Receipt.PatientName, packet.Prescription.ID)}, nil
}

func (c *Counter) Search(query string) (SearchResult, error) {
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return SearchResult{}, ErrSearchEmpty
	}
	patients, err := c.allPatients()
	if err != nil {
		return SearchResult{}, err
	}
	orders, err := c.allPrescriptions()
	if err != nil {
		return SearchResult{}, err
	}
	tickets, err := c.store.Tickets()
	if err != nil {
		return SearchResult{}, err
	}
	result := SearchResult{Patients: make([]model.PatientRecord, 0), Prescriptions: make([]model.PrescriptionOrder, 0), Tickets: make([]model.PharmacyTicket, 0)}
	for _, patient := range patients {
		if containsAny(needle, patient.ID, patient.Name, patient.Phone) {
			result.Patients = append(result.Patients, patient)
		}
	}
	for _, order := range orders {
		if containsAny(needle, order.ID, order.PatientID, order.Medication, order.Dosage) {
			result.Prescriptions = append(result.Prescriptions, order)
		}
	}
	for _, ticket := range tickets {
		if containsAny(needle, ticket.Number, ticket.PatientID, ticket.PrescriptionID, ticket.Status) {
			result.Tickets = append(result.Tickets, ticket)
		}
	}
	if len(result.Patients) == 0 && len(result.Prescriptions) == 0 && len(result.Tickets) == 0 {
		return result, ErrNoMatches
	}
	return result, nil
}

func containsAny(needle string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), needle) {
			return true
		}
	}
	return false
}

func (c *Counter) Call(number, window string) (model.PharmacyTicket, error) {
	ticket, err := c.queue.Claim(number, window)
	if err != nil {
		return model.PharmacyTicket{}, err
	}
	event := c.newEvent(number, window, "claim", true)
	if err := c.audit.Record(event); err != nil {
		return model.PharmacyTicket{}, err
	}
	if err := c.store.SaveTicketAndClaim(ticket, event, c.queue.Snapshot()); err != nil {
		return model.PharmacyTicket{}, err
	}
	return ticket, nil
}

func (c *Counter) Complete(number, operator string) (model.PharmacyTicket, error) {
	ticket, err := c.queue.Complete(number, operator)
	if err != nil {
		return model.PharmacyTicket{}, err
	}
	event := c.newEvent(number, operator, "complete", true)
	if err := c.audit.Record(event); err != nil {
		return model.PharmacyTicket{}, err
	}
	if err := c.store.SaveTicketAndClaim(ticket, event, c.queue.Snapshot()); err != nil {
		return model.PharmacyTicket{}, err
	}
	return ticket, nil
}

func (c *Counter) Pending() []model.PharmacyTicket   { return c.queue.List(model.StatusPending) }
func (c *Counter) Called() []model.PharmacyTicket    { return c.queue.List(model.StatusCalled) }
func (c *Counter) Completed() []model.PharmacyTicket { return c.queue.List(model.StatusCompleted) }
func (c *Counter) Snapshot() model.QueueSnapshot     { return c.queue.Snapshot() }
func (c *Counter) Events() []model.ClaimEvent        { return c.audit.Events() }

func (c *Counter) SetClaimPause(pause func()) {
	if c == nil || c.queue == nil {
		return
	}
	c.queue.SetClaimPause(pause)
}

func (c *Counter) persistSnapshot() error { return c.store.SaveSnapshot(c.queue.Snapshot()) }

func (c *Counter) newEvent(ticket, window, action string, successful bool) model.ClaimEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	event := model.ClaimEvent{ID: c.eventID(), Ticket: ticket, Window: window, Sequence: c.seq, Action: action, Successful: successful}
	c.seq++
	return event
}

func (c *Counter) eventID() string { return fmt.Sprintf("event-%03d", c.seq) }

func (c *Counter) allPatients() ([]model.PatientRecord, error) {
	return collectPatients(c.store)
}

func (c *Counter) allPrescriptions() ([]model.PrescriptionOrder, error) {
	return collectPrescriptions(c.store)
}

func collectPatients(store *persistence.Store) ([]model.PatientRecord, error) {
	return store.Patients()
}

func collectPrescriptions(store *persistence.Store) ([]model.PrescriptionOrder, error) {
	return store.Prescriptions()
}
