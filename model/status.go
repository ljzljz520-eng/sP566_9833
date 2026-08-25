package model

import "fmt"

const (
	StatusPending   = "pending"
	StatusCalled    = "called"
	StatusCompleted = "completed"
)

type StatusTransition struct {
	From string
	To   string
}

func (s StatusTransition) Valid() bool {
	if s.From == StatusPending && s.To == StatusCalled {
		return true
	}
	if s.From == StatusCalled && s.To == StatusCompleted {
		return true
	}
	return false
}

func ValidateTransition(from, to string) error {
	if !(StatusTransition{From: from, To: to}).Valid() {
		return fmt.Errorf("cannot move ticket from %s to %s", from, to)
	}
	return nil
}

func StatusDescription(status string) string {
	switch status {
	case StatusPending:
		return "waiting to be called"
	case StatusCalled:
		return "ready at a service window"
	case StatusCompleted:
		return "dispensed and recorded"
	default:
		return "unknown"
	}
}
