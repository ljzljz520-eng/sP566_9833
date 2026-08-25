package queue

import (
	"testing"

	"pharmacycounter/model"
)

func TestQueueAllocationAndCompletion(t *testing.T) {
	q := New()
	first, err := q.Allocate("p", "r", false)
	if err != nil || first.Number != "P001" {
		t.Fatalf("first=%#v err=%v", first, err)
	}
	called, err := q.Claim("P001", "window-1")
	if err != nil || !called.IsCalled() {
		t.Fatalf("called=%#v err=%v", called, err)
	}
	finished, err := q.Complete("P001", "tech")
	if err != nil || !finished.IsCompleted() || q.Contains("P001") == false {
		t.Fatalf("finished=%#v err=%v", finished, err)
	}
	if _, err := q.Complete("P001", "tech"); err == nil {
		t.Fatal("duplicate completion accepted")
	}
}

func TestQueueRestoreUsesTicketOrder(t *testing.T) {
	q := FromSnapshot(model.QueueSnapshot{Next: 1}, []model.PharmacyTicket{{Number: "P002", PatientID: "p", PrescriptionID: "r", Status: model.StatusPending, CreatedOrder: 2}, {Number: "P001", PatientID: "p", PrescriptionID: "r", Status: model.StatusCompleted, CreatedOrder: 1}})
	if got := q.Snapshot().Next; got != 3 {
		t.Fatalf("next=%d", got)
	}
	if len(q.List(model.StatusPending)) != 1 || len(q.List(model.StatusCompleted)) != 1 {
		t.Fatal("restore lost status")
	}
}
