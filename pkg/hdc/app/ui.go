package app

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

type model struct {
	scanPaths  []string
	objectData map[string]Object
	objectList []string         // points to string within objectData
	cursor     int              // points to index within objectList
	selected   map[int]struct{} // points to index within objectList
	log        io.Writer
}

func (m *model) scan() {
	// TODO support all paths
	*m = newModel(m.scanPaths, m.log)
	f := os.DirFS(m.scanPaths[0])
	traverse(f, m.scanPaths[0], m)
}

func newModel(scanPaths []string, log io.Writer) model {
	return model{
		scanPaths:  scanPaths,
		objectData: make(map[string]Object),
		selected:   make(map[int]struct{}),
		log:        log,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "s":
			m.scan()

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.objectList)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	s := "Processed objects:\n\n"

	for i, canPath := range m.objectList {

		// Is the cursor pointing at this object?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this object selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		obj := m.objectData[canPath]
		s += fmt.Sprintf("%s [%s] %s - %s\n", cursor, checked, canPath, obj.Path)
	}

	s += helpStyle("\n up/down/j/k : navigate - s: scan - q: quit\n")

	return s
}
