package report

import (
	"fmt"
	"sort"
	"strings"

	"pharmacycounter/model"
	"pharmacycounter/service"
)

type Summary struct {
	Pending   []model.PharmacyTicket
	Called    []model.PharmacyTicket
	Completed []model.PharmacyTicket
}

func Build(counter *service.Counter) Summary {
	return Summary{Pending: counter.Pending(), Called: counter.Called(), Completed: counter.Completed()}
}

func Render(summary Summary) string {
	sections := []struct {
		name  string
		items []model.PharmacyTicket
	}{
		{"待取药", summary.Pending},
		{"已叫号", summary.Called},
		{"已完成", summary.Completed},
	}
	var lines []string
	for _, section := range sections {
		lines = append(lines, fmt.Sprintf("%s (%d)", section.name, len(section.items)))
		if len(section.items) == 0 {
			lines = append(lines, "  -")
		} else {
			for _, item := range section.items {
				lines = append(lines, formatTicket(item))
			}
		}
	}
	return strings.Join(lines, "\n")
}

func formatTicket(ticket model.PharmacyTicket) string {
	parts := []string{ticket.Number, ticket.PatientID, ticket.PrescriptionID, ticket.Status}
	if ticket.CalledBy != "" {
		parts = append(parts, "window="+ticket.CalledBy)
	}
	if ticket.CompletedBy != "" {
		parts = append(parts, "operator="+ticket.CompletedBy)
	}
	return "  " + strings.Join(parts, " | ")
}

func RenderTicket(ticket model.PharmacyTicket) string { return formatTicket(ticket) }

func RenderSearch(result service.SearchResult) string {
	lines := []string{"搜索结果"}
	for _, patient := range result.Patients {
		lines = append(lines, fmt.Sprintf("患者 %s %s (%s)", patient.ID, patient.Name, patient.Priority))
	}
	for _, order := range result.Prescriptions {
		lines = append(lines, fmt.Sprintf("处方 %s %s %s", order.ID, order.PatientID, order.Medication))
	}
	for _, ticket := range result.Tickets {
		lines = append(lines, RenderTicket(ticket))
	}
	return strings.Join(lines, "\n")
}

func CompletionOrder(summary Summary) []string {
	items := append([]model.PharmacyTicket{}, summary.Completed...)
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedOrder < items[j].CreatedOrder })
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.Number)
	}
	return result
}

func CountByStatus(summary Summary) map[string]int {
	return map[string]int{model.StatusPending: len(summary.Pending), model.StatusCalled: len(summary.Called), model.StatusCompleted: len(summary.Completed)}
}
