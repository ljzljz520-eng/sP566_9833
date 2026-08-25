package service

import (
	"path/filepath"
	"testing"

	"pharmacycounter/model"
	"pharmacycounter/persistence"
)

func testCounter(t *testing.T) *Counter {
	t.Helper()
	store, err := persistence.Open(filepath.Join(t.TempDir(), "service.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	counter, err := New(store)
	if err != nil {
		t.Fatal(err)
	}
	return counter
}

func TestCounterRegisterValidatesOwnership(t *testing.T) {
	counter := testCounter(t)
	_, err := counter.Register(model.PatientRecord{ID: "p", Name: "A", Priority: "normal"}, model.PrescriptionOrder{ID: "r", PatientID: "other", Medication: "M"})
	if err == nil {
		t.Fatal("mismatched order accepted")
	}
}
