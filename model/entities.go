package model

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidPatient      = errors.New("invalid patient record")
	ErrInvalidPrescription = errors.New("invalid prescription order")
	ErrInvalidTicket       = errors.New("invalid pharmacy ticket")
	ErrInvalidClaim        = errors.New("invalid claim event")
)

type PatientRecord struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Priority string `json:"priority"`
}

type PrescriptionOrder struct {
	ID           string   `json:"id"`
	PatientID    string   `json:"patient_id"`
	Medication   string   `json:"medication"`
	Dosage       string   `json:"dosage"`
	Refills      int      `json:"refills"`
	Instructions []string `json:"instructions"`
}

type PharmacyTicket struct {
	Number         string `json:"number"`
	PatientID      string `json:"patient_id"`
	PrescriptionID string `json:"prescription_id"`
	Status         string `json:"status"`
	CreatedOrder   int    `json:"created_order"`
	CalledBy       string `json:"called_by"`
	CompletedBy    string `json:"completed_by"`
}

type ClaimEvent struct {
	ID         string `json:"id"`
	Ticket     string `json:"ticket"`
	Window     string `json:"window"`
	Sequence   int    `json:"sequence"`
	Action     string `json:"action"`
	Successful bool   `json:"successful"`
}

type QueueSnapshot struct {
	Pending  []string `json:"pending"`
	Called   []string `json:"called"`
	Finished []string `json:"finished"`
	Next     int      `json:"next"`
}

func (p PatientRecord) Validate() error {
	if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.Name) == "" {
		return ErrInvalidPatient
	}
	if p.Priority != "normal" && p.Priority != "urgent" {
		return fmt.Errorf("%w: priority must be normal or urgent", ErrInvalidPatient)
	}
	return nil
}

func (p PrescriptionOrder) Validate() error {
	if strings.TrimSpace(p.ID) == "" || strings.TrimSpace(p.PatientID) == "" || strings.TrimSpace(p.Medication) == "" {
		return ErrInvalidPrescription
	}
	if p.Refills < 0 || p.Refills > 12 {
		return fmt.Errorf("%w: refills outside supported range", ErrInvalidPrescription)
	}
	return nil
}

func (t PharmacyTicket) Validate() error {
	if strings.TrimSpace(t.Number) == "" || strings.TrimSpace(t.PatientID) == "" || strings.TrimSpace(t.PrescriptionID) == "" {
		return ErrInvalidTicket
	}
	allowed := map[string]bool{"pending": true, "called": true, "completed": true}
	if !allowed[t.Status] {
		return fmt.Errorf("%w: unsupported status", ErrInvalidTicket)
	}
	if t.CreatedOrder < 1 {
		return fmt.Errorf("%w: order must be positive", ErrInvalidTicket)
	}
	return nil
}

func (e ClaimEvent) Validate() error {
	if e.ID == "" || e.Ticket == "" || e.Window == "" || e.Action == "" {
		return ErrInvalidClaim
	}
	if e.Sequence < 1 {
		return fmt.Errorf("%w: sequence must be positive", ErrInvalidClaim)
	}
	return nil
}

func (t PharmacyTicket) IsPending() bool   { return t.Status == "pending" }
func (t PharmacyTicket) IsCalled() bool    { return t.Status == "called" }
func (t PharmacyTicket) IsCompleted() bool { return t.Status == "completed" }

func NormalizePriority(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "urgent" || v == "high" {
		return "urgent"
	}
	return "normal"
}

func TicketLabel(number string, status string) string {
	return fmt.Sprintf("%s [%s]", strings.ToUpper(strings.TrimSpace(number)), status)
}
