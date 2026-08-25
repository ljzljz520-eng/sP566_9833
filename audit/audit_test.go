package audit

import (
	"testing"

	"pharmacycounter/model"
)

func TestLoggerRecordsSupportedActions(t *testing.T) {
	logger := New()
	event := model.ClaimEvent{ID: "e", Ticket: "P001", Window: "window-1", Sequence: 1, Action: "claim", Successful: true}
	if err := logger.Record(event); err != nil {
		t.Fatal(err)
	}
	if _, ok := logger.LastForTicket("P001"); !ok || len(logger.Events()) != 1 {
		t.Fatal("event missing")
	}
	if err := logger.Record(model.ClaimEvent{ID: "x", Ticket: "P001", Window: "window-1", Sequence: 2, Action: "unknown", Successful: true}); err == nil {
		t.Fatal("unknown action accepted")
	}
}
