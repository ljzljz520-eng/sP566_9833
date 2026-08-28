package config

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type Settings struct {
	DatabasePath string
	Operator     string
	Window       string
	Command      string
	Ticket       string
}

func Default() Settings {
	return Settings{DatabasePath: "pharmacy.db", Operator: "desk", Window: "window-1", Command: "summary"}
}

func Parse(args []string) (Settings, error) {
	settings := Default()
	flags := flag.NewFlagSet("pharmacy", flag.ContinueOnError)
	flags.SetOutput(nil)
	flags.StringVar(&settings.DatabasePath, "db", settings.DatabasePath, "bbolt database path")
	flags.StringVar(&settings.Operator, "operator", settings.Operator, "operator name")
	flags.StringVar(&settings.Window, "window", settings.Window, "service window")
	flags.StringVar(&settings.Command, "command", settings.Command, "operation")
	flags.StringVar(&settings.Ticket, "ticket", settings.Ticket, "ticket number")
	if err := flags.Parse(args); err != nil {
		return Settings{}, err
	}
	settings = settings.Normalized()
	if err := settings.Validate(); err != nil {
		return Settings{}, err
	}
	return settings, nil
}

func (s Settings) Validate() error {
	if strings.TrimSpace(s.DatabasePath) == "" {
		return errors.New("database path is required")
	}
	if strings.TrimSpace(s.Operator) == "" {
		return errors.New("operator is required")
	}
	if strings.TrimSpace(s.Window) == "" {
		return errors.New("window is required")
	}
	if _, ok := FindCommand(s.Command); !ok {
		return fmt.Errorf("unsupported command %q", s.Command)
	}
	if missing := s.MissingRequirements(); len(missing) > 0 {
		return fmt.Errorf("%s is required for this command", strings.Join(missing, ", "))
	}
	return nil
}
