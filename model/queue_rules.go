package model

import (
	"errors"
	"sort"
	"strings"
)

var (
	ErrInvalidContact = errors.New("patient contact is incomplete")
	ErrInvalidWindow  = errors.New("service window name is invalid")
)

type QueuePolicy struct {
	NormalLabel string
	UrgentLabel string
	MaxWaiting  int
}

func DefaultQueuePolicy() QueuePolicy {
	return QueuePolicy{NormalLabel: "normal", UrgentLabel: "urgent", MaxWaiting: 200}
}

func (p QueuePolicy) Validate() error {
	if strings.TrimSpace(p.NormalLabel) == "" || strings.TrimSpace(p.UrgentLabel) == "" {
		return errors.New("queue labels are required")
	}
	if p.NormalLabel == p.UrgentLabel {
		return errors.New("queue labels must differ")
	}
	if p.MaxWaiting < 1 || p.MaxWaiting > 10000 {
		return errors.New("queue capacity is outside supported range")
	}
	return nil
}

func ValidatePatientContact(patient PatientRecord) error {
	if strings.TrimSpace(patient.Phone) == "" {
		return ErrInvalidContact
	}
	digits := 0
	for _, r := range patient.Phone {
		if r >= '0' && r <= '9' {
			digits++
		}
	}
	if digits < 7 {
		return ErrInvalidContact
	}
	return nil
}

func ValidateWindow(window string) error {
	value := strings.TrimSpace(window)
	if value == "" || len(value) > 40 {
		return ErrInvalidWindow
	}
	for _, r := range value {
		if r == ' ' || r == '\t' || r == '\n' {
			return ErrInvalidWindow
		}
	}
	return nil
}

func PatientDisplayName(patient PatientRecord) string {
	if strings.TrimSpace(patient.Name) == "" {
		return patient.ID
	}
	return strings.TrimSpace(patient.Name)
}

func MedicationSummary(order PrescriptionOrder) string {
	parts := []string{strings.TrimSpace(order.Medication)}
	if strings.TrimSpace(order.Dosage) != "" {
		parts = append(parts, strings.TrimSpace(order.Dosage))
	}
	if order.Refills > 0 {
		parts = append(parts, "refills="+itoa(order.Refills))
	}
	return strings.Join(parts, " ")
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := make([]byte, 0, 4)
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	for left, right := 0, len(digits)-1; left < right; left, right = left+1, right-1 {
		digits[left], digits[right] = digits[right], digits[left]
	}
	if negative {
		return "-" + string(digits)
	}
	return string(digits)
}

func SortTicketsByPriority(tickets []PharmacyTicket, patients map[string]PatientRecord) []PharmacyTicket {
	ordered := append([]PharmacyTicket{}, tickets...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left := priorityRank(patients[ordered[i].PatientID].Priority)
		right := priorityRank(patients[ordered[j].PatientID].Priority)
		if left != right {
			return left < right
		}
		if ordered[i].CreatedOrder != ordered[j].CreatedOrder {
			return ordered[i].CreatedOrder < ordered[j].CreatedOrder
		}
		return ordered[i].Number < ordered[j].Number
	})
	return ordered
}

func priorityRank(priority string) int {
	if NormalizePriority(priority) == "urgent" {
		return 0
	}
	return 1
}

func StatusCounts(tickets []PharmacyTicket) map[string]int {
	counts := map[string]int{StatusPending: 0, StatusCalled: 0, StatusCompleted: 0}
	for _, ticket := range tickets {
		if _, known := counts[ticket.Status]; known {
			counts[ticket.Status]++
		} else {
			counts[ticket.Status] = 1
		}
	}
	return counts
}

func TicketNumbers(tickets []PharmacyTicket) []string {
	numbers := make([]string, 0, len(tickets))
	for _, ticket := range tickets {
		numbers = append(numbers, ticket.Number)
	}
	sort.Strings(numbers)
	return numbers
}

func QueueInvariant(tickets []PharmacyTicket) error {
	seen := make(map[string]bool, len(tickets))
	for _, ticket := range tickets {
		if seen[ticket.Number] {
			return errors.New("duplicate ticket number")
		}
		seen[ticket.Number] = true
		if err := ticket.Validate(); err != nil {
			return err
		}
	}
	return nil
}
