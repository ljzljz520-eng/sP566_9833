package main

import (
	"errors"
	"fmt"
	"strings"

	"pharmacycounter"
	"pharmacycounter/config"
	"pharmacycounter/report"
)

type commandResult struct {
	Text   string
	ExitOK bool
}

func runCommand(app *pharmacy.Pharmacy, settings config.Settings) (commandResult, error) {
	if app == nil {
		return commandResult{}, errors.New("pharmacy is unavailable")
	}
	switch settings.Command {
	case "summary":
		return commandResult{Text: app.Summary(), ExitOK: true}, nil
	case "call":
		ticket, err := app.Call(settings.Ticket, settings.Window)
		if err != nil {
			return commandResult{}, err
		}
		return commandResult{Text: report.RenderTicket(ticket), ExitOK: true}, nil
	case "complete":
		ticket, err := app.Complete(settings.Ticket, settings.Operator)
		if err != nil {
			return commandResult{}, err
		}
		return commandResult{Text: report.RenderTicket(ticket), ExitOK: true}, nil
	case "search":
		result, err := app.Search(settings.Ticket)
		if err != nil {
			return commandResult{}, err
		}
		return commandResult{Text: report.RenderSearch(result), ExitOK: true}, nil
	case "register":
		return commandResult{}, errors.New("register requires the service API; use pharmacycounter.Pharmacy.Register")
	default:
		return commandResult{}, fmt.Errorf("unsupported command %s", settings.Command)
	}
}

func splitArguments(value string) []string {
	parts := strings.Fields(value)
	return append([]string{}, parts...)
}

func summarizeResult(result commandResult) string {
	if !result.ExitOK {
		return ""
	}
	return strings.TrimSpace(result.Text)
}
