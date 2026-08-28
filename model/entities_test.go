package model

import "testing"

func TestEntityValidation(t *testing.T) {
	if err := (PatientRecord{ID: "p", Name: "A", Priority: "normal"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (PrescriptionOrder{ID: "r", PatientID: "p", Medication: "M"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (PharmacyTicket{Number: "P001", PatientID: "p", PrescriptionID: "r", Status: StatusPending, CreatedOrder: 1}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTransition(StatusPending, StatusCompleted); err == nil {
		t.Fatal("invalid transition accepted")
	}
	if NormalizePriority("high") != "urgent" {
		t.Fatal("priority normalization failed")
	}
}
