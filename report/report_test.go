package report

import (
	"strings"
	"testing"

	"pharmacycounter/model"
)

func TestRenderIncludesAllLists(t *testing.T) {
	value := Render(Summary{Pending: []model.PharmacyTicket{{Number: "P001", PatientID: "p", PrescriptionID: "r", Status: model.StatusPending, CreatedOrder: 1}}})
	for _, part := range []string{"待取药", "已叫号", "已完成", "P001"} {
		if !strings.Contains(value, part) {
			t.Fatalf("missing %s in %s", part, value)
		}
	}
}
