package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"pharmacycounter/model"
)

type ExportDocument struct {
	Title      string         `json:"title"`
	Generated  string         `json:"generated"`
	Rows       []ExportRow    `json:"rows"`
	Counts     map[string]int `json:"counts"`
	Completion []string       `json:"completion_order"`
}

type ExportRow struct {
	Number       string `json:"number"`
	Patient      string `json:"patient"`
	Prescription string `json:"prescription"`
	Status       string `json:"status"`
	Window       string `json:"window,omitempty"`
	Order        int    `json:"order"`
}

func BuildExport(summary Summary) ExportDocument {
	all := make([]model.PharmacyTicket, 0, len(summary.Pending)+len(summary.Called)+len(summary.Completed))
	all = append(all, summary.Pending...)
	all = append(all, summary.Called...)
	all = append(all, summary.Completed...)
	sort.Slice(all, func(i, j int) bool { return CompareOrder(all[i], all[j]) < 0 })
	rows := make([]ExportRow, 0, len(all))
	for _, ticket := range all {
		window := ticket.CalledBy
		if window == "" {
			window = ticket.CompletedBy
		}
		rows = append(rows, ExportRow{Number: ticket.Number, Patient: ticket.PatientID, Prescription: ticket.PrescriptionID, Status: ticket.Status, Window: window, Order: ticket.CreatedOrder})
	}
	return ExportDocument{Title: "Pharmacy pickup queue", Generated: "deterministic", Rows: rows, Counts: CountByStatus(summary), Completion: CompletionOrder(summary)}
}

func (d ExportDocument) JSON() ([]byte, error) { return json.MarshalIndent(d, "", "  ") }

func (d ExportDocument) CSV() (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write([]string{"number", "patient", "prescription", "status", "window", "order"}); err != nil {
		return "", err
	}
	for _, row := range d.Rows {
		if err := writer.Write([]string{row.Number, row.Patient, row.Prescription, row.Status, row.Window, fmt.Sprint(row.Order)}); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func (d ExportDocument) ByStatus(status string) []ExportRow {
	rows := make([]ExportRow, 0)
	for _, row := range d.Rows {
		if row.Status == status {
			rows = append(rows, row)
		}
	}
	return rows
}

func RenderQueueColumns(summary Summary) []string {
	return []string{TicketTable(summary.Pending), TicketTable(summary.Called), TicketTable(summary.Completed)}
}

func RenderReceipt(receipt Receipt) string {
	lines := []string{fmt.Sprintf("取药凭条 %s", receipt.TicketNumber), fmt.Sprintf("患者: %s", receipt.PatientName), fmt.Sprintf("药品: %s", receipt.Medication)}
	if len(receipt.Instructions) == 0 {
		lines = append(lines, "用法: 请按处方说明服用")
	} else {
		lines = append(lines, "用法: "+strings.Join(receipt.Instructions, "; "))
	}
	return strings.Join(lines, "\n")
}

func StatusRows(summary Summary, status string) []model.PharmacyTicket {
	var source []model.PharmacyTicket
	switch status {
	case model.StatusPending:
		source = summary.Pending
	case model.StatusCalled:
		source = summary.Called
	case model.StatusCompleted:
		source = summary.Completed
	default:
		return []model.PharmacyTicket{}
	}
	return append([]model.PharmacyTicket{}, source...)
}
