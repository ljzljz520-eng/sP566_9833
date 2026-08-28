package service

import (
	"errors"
	"sort"
	"strings"

	"pharmacycounter/model"
	"pharmacycounter/persistence"
)

type SearchQuery struct {
	Text      string
	Status    string
	PatientID string
	Window    string
	Urgent    bool
}

func ParseSearchQuery(value string) (SearchQuery, error) {
	query := SearchQuery{}
	for _, token := range strings.Fields(value) {
		parts := strings.SplitN(token, ":", 2)
		if len(parts) == 1 {
			if query.Text != "" {
				query.Text += " "
			}
			query.Text += parts[0]
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		if val == "" {
			return SearchQuery{}, errors.New("search filter value is empty")
		}
		switch key {
		case "status":
			if val != model.StatusPending && val != model.StatusCalled && val != model.StatusCompleted {
				return SearchQuery{}, errors.New("unsupported ticket status filter")
			}
			query.Status = val
		case "patient":
			query.PatientID = val
		case "window":
			query.Window = val
		case "urgent":
			if val != "true" && val != "false" {
				return SearchQuery{}, errors.New("urgent filter must be true or false")
			}
			query.Urgent = val == "true"
		default:
			return SearchQuery{}, errors.New("unsupported search filter")
		}
	}
	if strings.TrimSpace(query.Text) == "" && query.Status == "" && query.PatientID == "" && query.Window == "" && !query.Urgent {
		return SearchQuery{}, ErrSearchEmpty
	}
	query.Text = strings.ToLower(strings.TrimSpace(query.Text))
	return query, nil
}

func (q SearchQuery) MatchesPatient(patient model.PatientRecord) bool {
	if q.PatientID != "" && patient.ID != q.PatientID {
		return false
	}
	if q.Text == "" {
		return true
	}
	return containsAny(q.Text, patient.ID, patient.Name, patient.Phone, patient.Priority)
}

func (q SearchQuery) MatchesOrder(order model.PrescriptionOrder) bool {
	if q.PatientID != "" && order.PatientID != q.PatientID {
		return false
	}
	if q.Text == "" {
		return true
	}
	return containsAny(q.Text, order.ID, order.PatientID, order.Medication, order.Dosage)
}

func (q SearchQuery) MatchesTicket(ticket model.PharmacyTicket, urgentPatients map[string]bool) bool {
	if q.Status != "" && ticket.Status != q.Status {
		return false
	}
	if q.PatientID != "" && ticket.PatientID != q.PatientID {
		return false
	}
	if q.Window != "" && ticket.CalledBy != q.Window && ticket.CompletedBy != q.Window {
		return false
	}
	if q.Urgent && !urgentPatients[ticket.PatientID] {
		return false
	}
	if q.Text == "" {
		return true
	}
	return containsAny(q.Text, ticket.Number, ticket.PatientID, ticket.PrescriptionID, ticket.Status, ticket.CalledBy, ticket.CompletedBy)
}

func sortSearchResult(result SearchResult) SearchResult {
	sort.Slice(result.Patients, func(i, j int) bool { return result.Patients[i].ID < result.Patients[j].ID })
	sort.Slice(result.Prescriptions, func(i, j int) bool { return result.Prescriptions[i].ID < result.Prescriptions[j].ID })
	sort.Slice(result.Tickets, func(i, j int) bool {
		if result.Tickets[i].CreatedOrder != result.Tickets[j].CreatedOrder {
			return result.Tickets[i].CreatedOrder < result.Tickets[j].CreatedOrder
		}
		return result.Tickets[i].Number < result.Tickets[j].Number
	})
	return result
}

func (c *Counter) AdvancedSearch(value string) (SearchResult, error) {
	query, err := ParseSearchQuery(value)
	if err != nil {
		return SearchResult{}, err
	}
	patients, err := c.store.Patients()
	if err != nil {
		return SearchResult{}, err
	}
	orders, err := c.store.Prescriptions()
	if err != nil {
		return SearchResult{}, err
	}
	tickets, err := c.store.Tickets()
	if err != nil {
		return SearchResult{}, err
	}
	urgentPatients := make(map[string]bool)
	for _, patient := range patients {
		if patient.Priority == "urgent" {
			urgentPatients[patient.ID] = true
		}
	}
	result := SearchResult{Patients: make([]model.PatientRecord, 0), Prescriptions: make([]model.PrescriptionOrder, 0), Tickets: make([]model.PharmacyTicket, 0)}
	for _, patient := range patients {
		if query.MatchesPatient(patient) {
			result.Patients = append(result.Patients, patient)
		}
	}
	for _, order := range orders {
		if query.MatchesOrder(order) {
			result.Prescriptions = append(result.Prescriptions, order)
		}
	}
	for _, ticket := range tickets {
		if query.MatchesTicket(ticket, urgentPatients) {
			result.Tickets = append(result.Tickets, ticket)
		}
	}
	result = sortSearchResult(result)
	if len(result.Patients) == 0 && len(result.Prescriptions) == 0 && len(result.Tickets) == 0 {
		return result, ErrNoMatches
	}
	return result, nil
}

func (c *Counter) SearchByStatus(status string) ([]model.PharmacyTicket, error) {
	if status != model.StatusPending && status != model.StatusCalled && status != model.StatusCompleted {
		return nil, errors.New("unsupported status")
	}
	return c.store.FilterTickets(persistence.TicketFilter{Status: status})
}
