package app

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func Run() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: janitor <path> [<path>...]")
		os.Exit(1)
	}

	log, err := tea.LogToFile("janitor.log", "")
	perr(err)
	defer log.Close()

	p := tea.NewProgram(newModel(os.Args[1:], log), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Fprintf(log, "ERROR there's been an error: %v - shutting down", err)
		os.Exit(1)
	}
	fmt.Fprintln(log, "INF closing")
}
