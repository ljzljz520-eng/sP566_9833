package pharmacy

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"pharmacycounter/model"
	"pharmacycounter/persistence"
)

func openTestPharmacy(t *testing.T) *Pharmacy {
	t.Helper()
	app, err := Open(filepath.Join(t.TempDir(), "counter.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = app.Close() })
	return app
}

func registerSample(t *testing.T, app *Pharmacy) {
	t.Helper()
	if _, err := app.Register("patient-001", "李明", "13800000001", "normal", "rx-001", "阿莫西林", "500mg", []string{"每日三次", "饭后服用"}); err != nil {
		t.Fatal(err)
	}
}

func TestWorkflowRegistrationAndSearch(t *testing.T) {
	app := openTestPharmacy(t)
	receipt, err := app.Register("patient-001", "李明", "13800000001", "normal", "rx-001", "阿莫西林", "500mg", []string{"每日三次"})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.TicketNumber != "P001" {
		t.Fatalf("unexpected ticket %s", receipt.TicketNumber)
	}
	result, err := app.Search("阿莫")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Patients) != 0 || len(result.Prescriptions) != 1 || len(result.Tickets) != 0 {
		t.Fatalf("unexpected search result %#v", result)
	}
	byPatient, err := app.Search("patient-001")
	if err != nil || len(byPatient.Patients) != 1 || len(byPatient.Tickets) != 1 {
		t.Fatalf("patient search result %#v err=%v", byPatient, err)
	}
}

func TestWorkflowCallingMovesTicket(t *testing.T) {
	app := openTestPharmacy(t)
	registerSample(t, app)
	ticket, err := app.Call("P001", "window-1")
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Status != model.StatusCalled || len(app.PendingNumbers()) != 0 {
		t.Fatalf("call state %#v pending=%v", ticket, app.PendingNumbers())
	}
	if len(app.CalledNumbers()) != 1 || app.CalledNumbers()[0] != "P001" {
		t.Fatalf("called=%v", app.CalledNumbers())
	}
}

func TestWorkflowCompletionPersistsOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "counter.db")
	app, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	registerSample(t, app)
	if _, err := app.Call("P001", "window-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Complete("P001", "tech-1"); err != nil {
		t.Fatal(err)
	}
	if err := app.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if got := reopened.CompletedNumbers(); len(got) != 1 || got[0] != "P001" {
		t.Fatalf("completed order %v", got)
	}
}

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reopen.db")
	store, err := persistence.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	patient := model.PatientRecord{ID: "patient-001", Name: "周宁", Priority: "normal"}
	order := model.PrescriptionOrder{ID: "rx-001", PatientID: patient.ID, Medication: "维生素", Dosage: "1片"}
	ticket := model.PharmacyTicket{Number: "P001", PatientID: patient.ID, PrescriptionID: order.ID, Status: model.StatusCompleted, CreatedOrder: 1, CompletedBy: "tech-1"}
	event := model.ClaimEvent{ID: "event-001", Ticket: "P001", Window: "tech-1", Sequence: 1, Action: "complete", Successful: true}
	for _, save := range []func() error{func() error { return store.SavePatient(patient) }, func() error { return store.SavePrescription(order) }, func() error { return store.SaveTicket(ticket) }, func() error { return store.SaveClaim(event) }, func() error { return store.SaveSnapshot(model.QueueSnapshot{Finished: []string{"P001"}, Next: 2}) }} {
		if err := save(); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := persistence.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if got, err := reopened.Patient("patient-001"); err != nil || got.Name != "周宁" {
		t.Fatalf("patient %#v err=%v", got, err)
	}
	if got, err := reopened.Ticket("P001"); err != nil || got.Status != model.StatusCompleted {
		t.Fatalf("ticket %#v err=%v", got, err)
	}
	if got, err := reopened.Snapshot(); err != nil || len(got.Finished) != 1 {
		t.Fatalf("snapshot %#v err=%v", got, err)
	}
}

func TestPharmacyNumberClaimedOnce(t *testing.T) {
	app := openTestPharmacy(t)
	registerSample(t, app)
	gate := make(chan struct{})
	ready := make(chan struct{}, 2)
	app.Counter.SetClaimPause(func() { ready <- struct{}{}; <-gate })
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for _, window := range []string{"window-1", "window-2"} {
		wait.Add(1)
		go func(window string) { defer wait.Done(); <-start; _, err := app.Call("P001", window); results <- err }(window)
	}
	close(start)
	<-ready
	<-ready
	close(gate)
	wait.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("number P001 claimed %d times", successes)
	}
	if len(app.PendingNumbers()) != 0 {
		t.Fatalf("pending numbers still contain claimed ticket: %v", app.PendingNumbers())
	}
}

func TestClosedStoreReturnsError(t *testing.T) {
	store, err := persistence.Open(filepath.Join(t.TempDir(), "closed.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Ticket("P001"); err == nil || errors.Is(err, persistence.ErrNotFound) {
		t.Fatalf("expected closed error, got %v", err)
	}
}
