package persistence

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"pharmacycounter/model"
)

type TicketFilter struct {
	Status       string
	PatientID    string
	Prescription string
	Window       string
}

func (f TicketFilter) Matches(ticket model.PharmacyTicket) bool {
	if f.Status != "" && ticket.Status != f.Status {
		return false
	}
	if f.PatientID != "" && ticket.PatientID != f.PatientID {
		return false
	}
	if f.Prescription != "" && ticket.PrescriptionID != f.Prescription {
		return false
	}
	if f.Window != "" && ticket.CalledBy != f.Window && ticket.CompletedBy != f.Window {
		return false
	}
	return true
}

func (s *Store) FilterTickets(filter TicketFilter) ([]model.PharmacyTicket, error) {
	items, err := s.Tickets()
	if err != nil {
		return nil, err
	}
	result := make([]model.PharmacyTicket, 0, len(items))
	for _, item := range items {
		if filter.Matches(item) {
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *Store) PatientHistory(patientID string) ([]model.PharmacyTicket, error) {
	if strings.TrimSpace(patientID) == "" {
		return nil, errors.New("patient id is required")
	}
	items, err := s.FilterTickets(TicketFilter{PatientID: patientID})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedOrder != items[j].CreatedOrder {
			return items[i].CreatedOrder < items[j].CreatedOrder
		}
		return items[i].Number < items[j].Number
	})
	return items, nil
}

func (s *Store) StatusCounts() (map[string]int, error) {
	items, err := s.Tickets()
	if err != nil {
		return nil, err
	}
	return model.StatusCounts(items), nil
}

func (s *Store) ExportState() ([]byte, error) {
	patients, err := s.Patients()
	if err != nil {
		return nil, err
	}
	prescriptions, err := s.Prescriptions()
	if err != nil {
		return nil, err
	}
	tickets, err := s.Tickets()
	if err != nil {
		return nil, err
	}
	claims, err := s.Claims()
	if err != nil {
		return nil, err
	}
	snapshot, err := s.Snapshot()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	state := struct {
		Patients      []model.PatientRecord     `json:"patients"`
		Prescriptions []model.PrescriptionOrder `json:"prescriptions"`
		Tickets       []model.PharmacyTicket    `json:"tickets"`
		Claims        []model.ClaimEvent        `json:"claims"`
		Snapshot      model.QueueSnapshot       `json:"snapshot"`
	}{patients, prescriptions, tickets, claims, snapshot}
	return json.MarshalIndent(state, "", "  ")
}

func (s *Store) HighestTicketOrder() (int, error) {
	items, err := s.Tickets()
	if err != nil {
		return 0, err
	}
	highest := 0
	for _, item := range items {
		if item.CreatedOrder > highest {
			highest = item.CreatedOrder
		}
	}
	return highest, nil
}

func (s *Store) ClaimCount(number string) (int, error) {
	if strings.TrimSpace(number) == "" {
		return 0, errors.New("ticket number is required")
	}
	items, err := s.Claims()
	if err != nil {
		return 0, err
	}
	count := 0
	for _, event := range items {
		if event.Ticket == number && event.Action == "claim" && event.Successful {
			count++
		}
	}
	return count, nil
}
