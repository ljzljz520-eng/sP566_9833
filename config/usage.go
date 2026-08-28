package config

import (
	"fmt"
	"strings"
)

type CommandSpec struct {
	Name        string
	Description string
	NeedsTicket bool
	NeedsWindow bool
	NeedsAgent  bool
}

func CommandSpecs() []CommandSpec {
	return []CommandSpec{
		{Name: "summary", Description: "show pending, called, and completed lists"},
		{Name: "call", Description: "move a pending ticket to a service window", NeedsTicket: true, NeedsWindow: true},
		{Name: "complete", Description: "record medication pickup", NeedsTicket: true, NeedsAgent: true},
		{Name: "search", Description: "find a patient, prescription, or ticket", NeedsTicket: true},
		{Name: "register", Description: "reserve a ticket through the service API"},
	}
}

func Usage() string {
	lines := []string{"pharmacy [flags]", "commands:"}
	for _, spec := range CommandSpecs() {
		lines = append(lines, fmt.Sprintf("  %-9s %s", spec.Name, spec.Description))
	}
	lines = append(lines, "flags: -db path -window name -operator name -ticket number")
	return strings.Join(lines, "\n")
}

func FindCommand(name string) (CommandSpec, bool) {
	for _, spec := range CommandSpecs() {
		if spec.Name == name {
			return spec, true
		}
	}
	return CommandSpec{}, false
}

func (s Settings) Normalized() Settings {
	result := s
	result.DatabasePath = strings.TrimSpace(result.DatabasePath)
	result.Operator = strings.TrimSpace(result.Operator)
	result.Window = strings.TrimSpace(result.Window)
	result.Command = strings.ToLower(strings.TrimSpace(result.Command))
	result.Ticket = strings.ToUpper(strings.TrimSpace(result.Ticket))
	return result
}

func (s Settings) Requirements() CommandSpec {
	spec, _ := FindCommand(strings.ToLower(strings.TrimSpace(s.Command)))
	return spec
}

func (s Settings) MissingRequirements() []string {
	normalized := s.Normalized()
	spec := normalized.Requirements()
	missing := make([]string, 0)
	if spec.NeedsTicket && normalized.Ticket == "" {
		missing = append(missing, "ticket")
	}
	if spec.NeedsWindow && normalized.Window == "" {
		missing = append(missing, "window")
	}
	if spec.NeedsAgent && normalized.Operator == "" {
		missing = append(missing, "operator")
	}
	return missing
}

func (s Settings) IsReadOnly() bool {
	command := strings.ToLower(strings.TrimSpace(s.Command))
	return command == "summary" || command == "search"
}

func (s Settings) DatabaseLabel() string {
	if strings.TrimSpace(s.DatabasePath) == "" {
		return "pharmacy.db"
	}
	return strings.TrimSpace(s.DatabasePath)
}
