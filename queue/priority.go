package queue

import (
	"errors"
	"sort"
	"strings"

	"pharmacycounter/model"
)

type Assignment struct {
	Window   string
	Ticket   model.PharmacyTicket
	Priority string
}

func ValidateWindow(window string) error { return model.ValidateWindow(window) }

func (q *Queue) NextPending() (model.PharmacyTicket, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if len(q.pending) == 0 {
		return model.PharmacyTicket{}, ErrTicketMissing
	}
	items := make([]model.PharmacyTicket, 0, len(q.pending))
	for _, item := range q.pending {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedOrder != items[j].CreatedOrder {
			return items[i].CreatedOrder < items[j].CreatedOrder
		}
		return items[i].Number < items[j].Number
	})
	return items[0], nil
}

func (q *Queue) PendingByPatient(patientID string) []model.PharmacyTicket {
	q.mu.RLock()
	defer q.mu.RUnlock()
	result := make([]model.PharmacyTicket, 0)
	for _, ticket := range q.pending {
		if ticket.PatientID == patientID {
			result = append(result, ticket)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedOrder < result[j].CreatedOrder })
	return result
}

func (q *Queue) Assignable(number, window string) (Assignment, error) {
	if err := ValidateWindow(window); err != nil {
		return Assignment{}, err
	}
	q.mu.RLock()
	ticket, ok := q.pending[number]
	q.mu.RUnlock()
	if !ok {
		return Assignment{}, ErrTicketMissing
	}
	return Assignment{Window: strings.TrimSpace(window), Ticket: ticket, Priority: "normal"}, nil
}

func (q *Queue) Validate() error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	all := make([]model.PharmacyTicket, 0, len(q.pending)+len(q.called)+len(q.finished))
	for _, ticket := range q.pending {
		all = append(all, ticket)
	}
	for _, ticket := range q.called {
		all = append(all, ticket)
	}
	for _, ticket := range q.finished {
		all = append(all, ticket)
	}
	if err := model.QueueInvariant(all); err != nil {
		return err
	}
	if q.next < 1 {
		return errors.New("queue sequence is not positive")
	}
	return nil
}

func (q *Queue) Remove(number string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.pending[number]; ok {
		delete(q.pending, number)
		return nil
	}
	if _, ok := q.called[number]; ok {
		delete(q.called, number)
		return nil
	}
	if _, ok := q.finished[number]; ok {
		delete(q.finished, number)
		return nil
	}
	return ErrTicketMissing
}

func (q *Queue) Assignments(window string) []Assignment {
	if ValidateWindow(window) != nil {
		return []Assignment{}
	}
	items := q.List(model.StatusCalled)
	result := make([]Assignment, 0, len(items))
	for _, ticket := range items {
		if ticket.CalledBy == window {
			result = append(result, Assignment{Window: window, Ticket: ticket, Priority: "normal"})
		}
	}
	return result
}
