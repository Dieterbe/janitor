package app

import (
	"fmt"
	"io"
	"os"

	"github.com/Dieterbe/sandbox/homedirclean/pkg/hdc"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render
)

type model struct {
	scanPaths     []string
	rootDirPrints []hdc.DirPrint // corresponding to each scanpath. Not sure yet if we'll need this
	allDirPrints  map[string]hdc.DirPrint
	pairSims      []hdc.PairSim
	cursor        int              // points to index within pairSims
	selected      map[int]struct{} // points to index within pairSims
	log           io.Writer
}

func (m *model) scan() {
	// TODO support all paths
	*m = newModel(m.scanPaths, m.log)
	f := os.DirFS(m.scanPaths[0])
	root, all, err := WalkFS(f, m.scanPaths[0], hdc.Sha256FingerPrint, m.log)
	perr(err)
	m.rootDirPrints = []hdc.DirPrint{root}
	m.allDirPrints = all
	m.pairSims = hdc.GetPairSims(all)
}

func newModel(scanPaths []string, log io.Writer) model {
	return model{
		scanPaths:    scanPaths,
		allDirPrints: make(map[string]hdc.DirPrint),
		selected:     make(map[int]struct{}),
		log:          log,
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
			if m.cursor < len(m.pairSims)-1 {
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
	s := "Processed directories:\n\n"

	for i, ps := range m.pairSims {

		// Is the cursor pointing at this Pairesim?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this PairSim selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] Path1: %s\nPath2: %s\nSimilarity: %s\n\n", cursor, checked, ps.Path1, ps.Path2, ps.Sim)
	}

	s += helpStyle("\n up/down/j/k : navigate - s: scan - q: quit\n")

	return s
}
