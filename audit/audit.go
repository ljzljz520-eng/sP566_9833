package audit

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"pharmacycounter/model"
)

var ErrUnsupportedAction = errors.New("unsupported audit action")

type Logger struct {
	mu     sync.RWMutex
	events []model.ClaimEvent
}

func New() *Logger { return &Logger{events: make([]model.ClaimEvent, 0)} }

func (l *Logger) Record(event model.ClaimEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}
	if !supported(event.Action) {
		return fmt.Errorf("%w: %s", ErrUnsupportedAction, event.Action)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, event)
	return nil
}

func supported(action string) bool {
	switch action {
	case "claim", "complete", "register":
		return true
	default:
		return false
	}
}

func (l *Logger) Events() []model.ClaimEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()
	result := make([]model.ClaimEvent, len(l.events))
	copy(result, l.events)
	return result
}

func (l *Logger) LastForTicket(number string) (model.ClaimEvent, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for i := len(l.events) - 1; i >= 0; i-- {
		if l.events[i].Ticket == number {
			return l.events[i], true
		}
	}
	return model.ClaimEvent{}, false
}

func Describe(event model.ClaimEvent) string {
	state := "failed"
	if event.Successful {
		state = "successful"
	}
	return strings.Join([]string{event.Action, event.Ticket, event.Window, state}, " | ")
}
