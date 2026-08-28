package service

import (
	"testing"

	"pharmacycounter/model"
)

func TestCounterCompleteRequiresCall(t *testing.T) {
	counter := testCounter(t)
	result, err := counter.Register(model.PatientRecord{ID: "p", Name: "A", Priority: "normal"}, model.PrescriptionOrder{ID: "r", PatientID: "p", Medication: "M"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := counter.Complete(result.Ticket.Number, "tech"); err == nil {
		t.Fatal("pending ticket completed")
	}
	if _, err := counter.Call(result.Ticket.Number, "window-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := counter.Complete(result.Ticket.Number, "tech"); err != nil {
		t.Fatal(err)
	}
}
