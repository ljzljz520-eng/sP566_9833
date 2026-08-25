package service

import (
	"errors"
	"sort"

	"pharmacycounter/audit"
	"pharmacycounter/model"
	"pharmacycounter/persistence"
	"pharmacycounter/queue"
)

func Restore(store *persistence.Store) (*Counter, error) {
	if store == nil {
		return nil, errors.New("store is required")
	}
	return New(store)
}

func RebuildSnapshot(store *persistence.Store) (model.QueueSnapshot, error) {
	tickets, err := store.Tickets()
	if err != nil {
		return model.QueueSnapshot{}, err
	}
	snapshot := model.QueueSnapshot{Pending: []string{}, Called: []string{}, Finished: []string{}, Next: 1}
	for _, ticket := range tickets {
		if ticket.CreatedOrder >= snapshot.Next {
			snapshot.Next = ticket.CreatedOrder + 1
		}
		switch ticket.Status {
		case model.StatusPending:
			snapshot.Pending = append(snapshot.Pending, ticket.Number)
		case model.StatusCalled:
			snapshot.Called = append(snapshot.Called, ticket.Number)
		case model.StatusCompleted:
			snapshot.Finished = append(snapshot.Finished, ticket.Number)
		}
	}
	sort.Strings(snapshot.Pending)
	sort.Strings(snapshot.Called)
	sort.Strings(snapshot.Finished)
	return snapshot, nil
}

func RestoreWithFallback(store *persistence.Store) (*Counter, error) {
	snapshot, err := store.Snapshot()
	if err != nil && !errors.Is(err, persistence.ErrNotFound) {
		return nil, err
	}
	if errors.Is(err, persistence.ErrNotFound) {
		snapshot, err = RebuildSnapshot(store)
		if err != nil {
			return nil, err
		}
		if err := store.SaveSnapshot(snapshot); err != nil {
			return nil, err
		}
	}
	return queueCounter(store, snapshot)
}

func queueCounter(store *persistence.Store, snapshot model.QueueSnapshot) (*Counter, error) {
	tickets, err := store.Tickets()
	if err != nil {
		return nil, err
	}
	return &Counter{store: store, queue: queue.FromSnapshot(snapshot, tickets), audit: auditLoggerFromStore(store), seq: nextSequence(store)}, nil
}

func auditLoggerFromStore(store *persistence.Store) *audit.Logger {
	logger := audit.New()
	items, err := store.Claims()
	if err != nil {
		return logger
	}
	for _, item := range items {
		_ = logger.Record(item)
	}
	return logger
}

func nextSequence(store *persistence.Store) int {
	items, err := store.Claims()
	if err != nil {
		return 1
	}
	max := 0
	for _, item := range items {
		if item.Sequence > max {
			max = item.Sequence
		}
	}
	return max + 1
}
