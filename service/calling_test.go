package service

import (
	"testing"

	"pharmacycounter/model"
)

func TestCounterCallRequiresPendingTicket(t *testing.T) {
	counter := testCounter(t)
	if _, err := counter.Call("P999", "window-1"); err == nil {
		t.Fatal("missing ticket called")
	}
	result, err := counter.Register(model.PatientRecord{ID: "p", Name: "A", Priority: "normal"}, model.PrescriptionOrder{ID: "r", PatientID: "p", Medication: "M"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := counter.Call(result.Ticket.Number, ""); err == nil {
		t.Fatal("blank window accepted")
	}
}
