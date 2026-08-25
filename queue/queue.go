package queue

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"pharmacycounter/model"
)

var (
	ErrTicketMissing = errors.New("ticket is not in the queue")
	ErrAlreadyCalled = errors.New("ticket has already been called")
	ErrNotCalled     = errors.New("ticket has not been called")
	ErrWindowMissing = errors.New("service window is required")
)

type Queue struct {
	mu         sync.RWMutex
	pending    map[string]model.PharmacyTicket
	called     map[string]model.PharmacyTicket
	finished   map[string]model.PharmacyTicket
	next       int
	claimPause func()
}

func New() *Queue {
	return &Queue{pending: make(map[string]model.PharmacyTicket), called: make(map[string]model.PharmacyTicket), finished: make(map[string]model.PharmacyTicket), next: 1}
}

func FromSnapshot(snapshot model.QueueSnapshot, tickets []model.PharmacyTicket) *Queue {
	q := New()
	q.next = snapshot.Next
	if q.next < 1 {
		q.next = 1
	}
	for _, ticket := range tickets {
		switch ticket.Status {
		case model.StatusPending:
			q.pending[ticket.Number] = ticket
		case model.StatusCalled:
			q.called[ticket.Number] = ticket
		case model.StatusCompleted:
			q.finished[ticket.Number] = ticket
		}
		if ticket.CreatedOrder >= q.next {
			q.next = ticket.CreatedOrder + 1
		}
	}
	return q
}

func (q *Queue) Allocate(patientID, prescriptionID string, urgent bool) (model.PharmacyTicket, error) {
	if patientID == "" || prescriptionID == "" {
		return model.PharmacyTicket{}, errors.New("patient and prescription are required")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	number := "P" + fmt.Sprintf("%03d", q.next)
	order := q.next
	q.next++
	ticket := model.PharmacyTicket{Number: number, PatientID: patientID, PrescriptionID: prescriptionID, Status: model.StatusPending, CreatedOrder: order}
	if urgent {
		ticket.CreatedOrder = order
	}
	q.pending[number] = ticket
	return ticket, nil
}

func (q *Queue) Claim(number, window string) (model.PharmacyTicket, error) {
	if window == "" {
		return model.PharmacyTicket{}, ErrWindowMissing
	}
	q.mu.RLock()
	ticket, ok := q.pending[number]
	q.mu.RUnlock()
	if !ok {
		return model.PharmacyTicket{}, ErrTicketMissing
	}
	if q.claimPause != nil {
		q.claimPause()
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if err := model.ValidateTransition(ticket.Status, model.StatusCalled); err != nil {
		return model.PharmacyTicket{}, err
	}
	ticket.Status = model.StatusCalled
	ticket.CalledBy = window
	q.called[number] = ticket
	delete(q.pending, number)
	return ticket, nil
}

func (q *Queue) SetClaimPause(pause func()) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.claimPause = pause
}

func (q *Queue) Complete(number, operator string) (model.PharmacyTicket, error) {
	if operator == "" {
		return model.PharmacyTicket{}, errors.New("operator is required")
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	ticket, ok := q.called[number]
	if !ok {
		return model.PharmacyTicket{}, ErrNotCalled
	}
	ticket.Status = model.StatusCompleted
	ticket.CompletedBy = operator
	delete(q.called, number)
	delete(q.pending, number)
	q.finished[number] = ticket
	return ticket, nil
}

func (q *Queue) Get(number string) (model.PharmacyTicket, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if ticket, ok := q.pending[number]; ok {
		return ticket, nil
	}
	if ticket, ok := q.called[number]; ok {
		return ticket, nil
	}
	if ticket, ok := q.finished[number]; ok {
		return ticket, nil
	}
	return model.PharmacyTicket{}, ErrTicketMissing
}

func (q *Queue) Snapshot() model.QueueSnapshot {
	q.mu.RLock()
	defer q.mu.RUnlock()
	snapshot := model.QueueSnapshot{Pending: make([]string, 0), Called: make([]string, 0), Finished: make([]string, 0), Next: q.next}
	for number := range q.pending {
		snapshot.Pending = append(snapshot.Pending, number)
	}
	for number := range q.called {
		snapshot.Called = append(snapshot.Called, number)
	}
	for number := range q.finished {
		snapshot.Finished = append(snapshot.Finished, number)
	}
	sort.Strings(snapshot.Pending)
	sort.Strings(snapshot.Called)
	sort.Strings(snapshot.Finished)
	return snapshot
}

func (q *Queue) List(status string) []model.PharmacyTicket {
	q.mu.RLock()
	defer q.mu.RUnlock()
	var source map[string]model.PharmacyTicket
	switch status {
	case model.StatusPending:
		source = q.pending
	case model.StatusCalled:
		source = q.called
	case model.StatusCompleted:
		source = q.finished
	default:
		return []model.PharmacyTicket{}
	}
	result := make([]model.PharmacyTicket, 0, len(source))
	for _, ticket := range source {
		result = append(result, ticket)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedOrder < result[j].CreatedOrder })
	return result
}

func (q *Queue) Counts() map[string]int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return map[string]int{model.StatusPending: len(q.pending), model.StatusCalled: len(q.called), model.StatusCompleted: len(q.finished)}
}

func ParseNumber(number string) (int, error) {
	if len(number) < 2 || number[0] != 'P' {
		return 0, errors.New("ticket number must start with P")
	}
	value, err := strconv.Atoi(number[1:])
	if err != nil || value < 1 {
		return 0, errors.New("ticket number has invalid digits")
	}
	return value, nil
}

func (q *Queue) Contains(number string) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	_, pending := q.pending[number]
	_, called := q.called[number]
	_, finished := q.finished[number]
	return pending || called || finished
}
