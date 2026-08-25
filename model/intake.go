package model

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidIntake = errors.New("invalid pharmacy intake packet")

type IntakePacket struct {
	Patient      PatientRecord
	Prescription PrescriptionOrder
	Ticket       PharmacyTicket
	Receipt      ReceiptRecord
}

type ReceiptRecord struct {
	TicketNumber string
	PatientName  string
	Medication   string
	Instructions []string
}

type DispenseInstruction struct {
	Label    string
	Value    string
	Required bool
}

func NewIntakePacket(patient PatientRecord, order PrescriptionOrder, ticket PharmacyTicket) (IntakePacket, error) {
	if err := patient.Validate(); err != nil {
		return IntakePacket{}, err
	}
	if err := order.Validate(); err != nil {
		return IntakePacket{}, err
	}
	if err := ticket.Validate(); err != nil {
		return IntakePacket{}, err
	}
	if patient.ID != order.PatientID || order.ID != ticket.PrescriptionID || patient.ID != ticket.PatientID {
		return IntakePacket{}, ErrInvalidIntake
	}
	receipt := ReceiptRecord{TicketNumber: ticket.Number, PatientName: PatientDisplayName(patient), Medication: MedicationSummary(order), Instructions: append([]string{}, order.Instructions...)}
	return IntakePacket{Patient: patient, Prescription: order, Ticket: ticket, Receipt: receipt}, nil
}

func (p IntakePacket) Validate() error {
	if err := p.Patient.Validate(); err != nil {
		return err
	}
	if err := p.Prescription.Validate(); err != nil {
		return err
	}
	if err := p.Ticket.Validate(); err != nil {
		return err
	}
	if p.Patient.ID != p.Prescription.PatientID || p.Ticket.PatientID != p.Patient.ID {
		return ErrInvalidIntake
	}
	return nil
}

func (p IntakePacket) ReceiptLines() []string {
	lines := []string{fmt.Sprintf("取药凭条 %s", p.Receipt.TicketNumber), fmt.Sprintf("患者: %s", p.Receipt.PatientName), fmt.Sprintf("药品: %s", p.Receipt.Medication)}
	if len(p.Receipt.Instructions) == 0 {
		return append(lines, "用法: 请按处方说明服用")
	}
	return append(lines, "用法: "+strings.Join(p.Receipt.Instructions, "; "))
}

func (p IntakePacket) ReceiptText() string { return strings.Join(p.ReceiptLines(), "\n") }

func (p IntakePacket) InstructionChecklist() []DispenseInstruction {
	checklist := []DispenseInstruction{{Label: "patient", Value: p.Patient.ID, Required: true}, {Label: "ticket", Value: p.Ticket.Number, Required: true}, {Label: "medication", Value: p.Prescription.Medication, Required: true}}
	for _, instruction := range p.Prescription.Instructions {
		value := strings.TrimSpace(instruction)
		if value != "" {
			checklist = append(checklist, DispenseInstruction{Label: "instruction", Value: value, Required: false})
		}
	}
	return checklist
}

func (p IntakePacket) HasInstruction(fragment string) bool {
	needle := strings.ToLower(strings.TrimSpace(fragment))
	if needle == "" {
		return false
	}
	for _, instruction := range p.Prescription.Instructions {
		if strings.Contains(strings.ToLower(instruction), needle) {
			return true
		}
	}
	return false
}
