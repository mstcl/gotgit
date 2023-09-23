package main

// A simple example that shows how to send messages to a Bubble Tea program
// from outside the program using Program.Send(Msg).

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var (
	spinnerStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle            = helpStyle.Copy().UnsetMargins()
	durationStyle       = dotStyle.Copy()
	appStyle            = lipgloss.NewStyle().Margin(1, 2, 0, 2)
)

type resultMsg struct {
	repo repo
}

func (r resultMsg) String() string {
	if r.repo.cmtCnt == 0 {
		return dotStyle.Render(strings.Repeat(".", 30))
	}
	return fmt.Sprintf("%s %s", r.repo.name, durationStyle.Render(strconv.Itoa(r.repo.cmtCnt)))
}

type model struct {
	spinner  spinner.Model
	results  []resultMsg
	quitting bool
}

func newModel(size int) model {
	s := spinner.New()
	s.Style = spinnerStyle
	return model{
		spinner: s,
		results: make([]resultMsg, size),
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case resultMsg:
		m.results = append(m.results[1:], msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() string {
	var s string

	if m.quitting {
		s += "Goodbye!"
	} else {
		s += m.spinner.View() + " Getting repositories..."
	}

	s += "\n\n"

	for _, res := range m.results {
		s += res.String() + "\n"
	}

	if !m.quitting {
		s += helpStyle.Render("Press any key to exit")
	}

	if m.quitting {
		s += "\n"
	}

	return appStyle.Render(s)
}

func main() {
	dirs := getGitDirs()
	var dirsSorted []string
	seen := make([]plumbing.Hash, 0)
	for idx := 0; idx < len(dirs)-1; idx++ {
		r, err := git.PlainOpen(dirs[idx])
		if err != nil {
			fmt.Printf("%s", err)
		}
		ref := getRef(r)
		if slices.Contains(seen, ref.Hash()) {
			continue
		}
		seen = append(seen, ref.Hash())
		dirsSorted = append(dirsSorted, dirs[idx])
	}

	p := tea.NewProgram(newModel(len(dirsSorted)))

	go func(repos []repo) {
		for idx := 0; idx < len(repos); idx++ {
			r := processRepo(repos[idx])
			p.Send(resultMsg{repo: r})
		}
	}(getRepos(dirsSorted))

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
