package persistence

import (
	"path/filepath"
	"testing"

	"pharmacycounter/model"
)

func TestStoreRoundTripLists(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.SavePatient(model.PatientRecord{ID: "p", Name: "A", Priority: "normal"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SavePrescription(model.PrescriptionOrder{ID: "r", PatientID: "p", Medication: "M"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveTicket(model.PharmacyTicket{Number: "P001", PatientID: "p", PrescriptionID: "r", Status: model.StatusPending, CreatedOrder: 1}); err != nil {
		t.Fatal(err)
	}
	patients, err := store.Patients()
	if err != nil || len(patients) != 1 {
		t.Fatalf("patients=%v err=%v", patients, err)
	}
	orders, err := store.Prescriptions()
	if err != nil || len(orders) != 1 {
		t.Fatalf("orders=%v err=%v", orders, err)
	}
	tickets, err := store.Tickets()
	if err != nil || len(tickets) != 1 {
		t.Fatalf("tickets=%v err=%v", tickets, err)
	}
}
