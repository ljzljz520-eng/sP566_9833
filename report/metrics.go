package report

import (
	"fmt"
	"sort"
	"strings"

	"pharmacycounter/model"
)

func TicketTable(tickets []model.PharmacyTicket) string {
	ordered := append([]model.PharmacyTicket{}, tickets...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].CreatedOrder < ordered[j].CreatedOrder })
	rows := []string{"number\tpatient\tprescription\tstatus\torder"}
	for _, ticket := range ordered {
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%d", ticket.Number, ticket.PatientID, ticket.PrescriptionID, ticket.Status, ticket.CreatedOrder))
	}
	return strings.Join(rows, "\n")
}

func StatusLabel(status string) string { return model.StatusDescription(status) }

func CompareOrder(first, second model.PharmacyTicket) int {
	if first.CreatedOrder < second.CreatedOrder {
		return -1
	}
	if first.CreatedOrder > second.CreatedOrder {
		return 1
	}
	return strings.Compare(first.Number, second.Number)
}
