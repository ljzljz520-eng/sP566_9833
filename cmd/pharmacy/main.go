package main

import (
	"fmt"
	"os"
	"strings"

	"pharmacycounter"
	"pharmacycounter/config"
)

func main() {
	settings, err := config.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	app, err := pharmacy.Open(settings.DatabasePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer app.Close()
	if err := execute(app, settings); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute(app *pharmacy.Pharmacy, settings config.Settings) error {
	result, err := runCommand(app, settings)
	if err != nil {
		return err
	}
	if result.Text != "" {
		fmt.Println(result.Text)
	}
	return nil
}

func renderSearch(result interface{}) string { return fmt.Sprint(result) }

func cleanLines(value string) []string {
	lines := strings.Split(value, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}
