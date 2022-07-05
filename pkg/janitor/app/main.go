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

	fd, err := tea.LogToFile("janitor.log", "")
	perr(err)
	defer fd.Close()

	p := tea.NewProgram(newModel(os.Args[1:], fd), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Fprintf(fd, "ERROR there's been an error: %v - shutting down", err)
		os.Exit(1)
	}
	fmt.Fprintln(fd, "INF closing")
}
