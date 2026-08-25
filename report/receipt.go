package report

import (
	"fmt"
	"strings"

	"pharmacycounter/model"
)

type Receipt struct {
	TicketNumber string
	PatientName  string
	Medication   string
	Instructions []string
}

func MakeReceipt(patient model.PatientRecord, order model.PrescriptionOrder, ticket model.PharmacyTicket) Receipt {
	return Receipt{TicketNumber: ticket.Number, PatientName: patient.Name, Medication: order.Medication, Instructions: append([]string{}, order.Instructions...)}
}

func (r Receipt) Render() string {
	lines := []string{fmt.Sprintf("取药凭条 %s", r.TicketNumber), fmt.Sprintf("患者: %s", r.PatientName), fmt.Sprintf("药品: %s", r.Medication)}
	if len(r.Instructions) == 0 {
		lines = append(lines, "用法: 请按处方说明服用")
	} else {
		lines = append(lines, "用法: "+strings.Join(r.Instructions, "; "))
	}
	return strings.Join(lines, "\n")
}

func (r Receipt) Lines() []string { return strings.Split(r.Render(), "\n") }
