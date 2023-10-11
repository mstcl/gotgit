package main

// Render gotgit with bubbletea

import (
	"fmt"
	// "io"
	"os"
	"slices"
	"strconv"
	// "strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

var (
	paginationStyle = list.DefaultStyles().PaginationStyle.
			PaddingLeft(2)
	helpStyle = list.DefaultStyles().HelpStyle.
			PaddingLeft(2).
			PaddingBottom(1)
	// selectedItemStyle = lipgloss.NewStyle().
	// 			PaddingLeft(2).
	// 			Foreground(lipgloss.Color("1"))
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("1")).
			Padding(0, 1).
			MarginLeft(0)
	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginLeft(1)
	// dotStyle = spinnerStyle.Copy().
	// 		MarginLeft(1).
	// 		Border(lipgloss.NormalBorder(), false, false, false, true).
	// 		BorderForeground(lipgloss.Color("8"))
	// helpStyle = spinnerStyle.Copy().
	// 		Margin(1, 0)
	// repoStyle = spinnerStyle.Copy().UnsetForeground().
	// 		Foreground(lipgloss.Color("16")).
	// 		MarginLeft(1).
	// 		PaddingLeft(1)
	repoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7")).
			PaddingLeft(2).
			PaddingRight(1)
	altStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))
	addStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))
	modStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5"))
	delStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))
	othStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("4"))
	delimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Background(lipgloss.Color("0"))
	appStyle = lipgloss.NewStyle().Width(100).
			Margin(3, 2)
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type resultMsg struct {
	repo repo
}

func gatherInfo(ctr int, symb string) string {
	if ctr != 0 {
		return fmt.Sprintf("%s%s", symb, strconv.Itoa(ctr))
	}
	return ""
}

func (r resultMsg) String() item {
	var info string
	info += addStyle.Render(gatherInfo(r.repo.uChanges.added, "+"))
	info += modStyle.Render(gatherInfo(r.repo.uChanges.modified, "~"))
	info += delStyle.Render(gatherInfo(r.repo.uChanges.deleted, "-"))
	info += othStyle.Render(gatherInfo(r.repo.uChanges.others, "!"))

	desc := fmt.Sprintf("%s %s",
		altStyle.Render("Σ"+strconv.Itoa(r.repo.cmtCnt)),
		info)
	return item{title: r.repo.name, desc: desc}
}

type model struct {
	list                   list.Model
	choice                 string
	spinner                spinner.Model
	results                []resultMsg
	quitting, ascr, loaded bool
	size, ctr              int
}

func newModel(l list.Model, size int) model {
	s := spinner.New()
	s.Style = spinnerStyle
	return model{
		size:    size,
		list:    l,
		ctr:     0,
		loaded:  false,
		spinner: s,
		results: make([]resultMsg, size),
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case resultMsg:
		m.list.InsertItem(0, msg.String())
		m.ctr++
		if m.ctr == len(m.results) {
			m.loaded = true
			m.list.NewStatusMessage("")
		} else {
			m.list.NewStatusMessage(m.spinner.View() + spinnerStyle.Render(fmt.Sprintf("Loading %s/%s", strconv.Itoa(m.ctr), strconv.Itoa(m.size))))
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s string
	s += m.list.View()
	// if m.loaded == false {
	// 	s += m.spinner.View() +
	// 		spinnerStyle.Render("Loading repository "+
	// 			fmt.Sprintf("%s/%s",
	// 				strconv.Itoa(m.ctr),
	// 				strconv.Itoa(m.size)))
	// }
	// for _, res := range m.results {
	// 	s += res.String() + "\n"
	// }
	// if !m.quitting {
	// 	s += helpStyle.Render("f: switch modes • q: exit\n")
	// }
	// if m.quitting {
	// 	s += "\n"
	// }
	return appStyle.Render(s) + "\n"
}

// type itemDelegate struct{}
//
// func (d itemDelegate) Height() int {
// 	return 1
// }
// func (d itemDelegate) Spacing() int {
// 	return 0
// }
// func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
// 	return nil
// }
// func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
// 	i, ok := listItem.(item)
// 	if !ok {
// 		return
// 	}
// 	str := fmt.Sprintf("%s", i.title)
// 	fn := func(s ...string) string {
// 		return repoStyle.Render("░" + strings.Join(s, " "))
// 	}
// 	if index == m.Index() {
// 		fn = func(s ...string) string {
// 			return selectedItemStyle.Render("█") + lipgloss.NewStyle().
// 				Foreground(lipgloss.Color("0")).
// 				Background(lipgloss.Color("1")).Render(strings.Join(s, " "))
// 		}
// 	}
// 	fmt.Fprint(w, fn(str))
// }

// Program starts here:
// A Go activity is started to process in unprocessed repos sequentially
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
	items := []list.Item{}
	d := list.NewDefaultDelegate()
	d.SetSpacing(0)
	d.Styles.DimmedTitle = d.Styles.DimmedTitle.Foreground(lipgloss.Color("8"))
	d.Styles.DimmedDesc = d.Styles.DimmedDesc.Foreground(lipgloss.Color("8"))
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("0")).
		Padding(0, 1).
		MarginLeft(1).
		UnsetBorderLeft()
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		UnsetBorderLeft().
		MarginLeft(1)
	l := list.New(items, d, 0, 30)
	l.Title = "Repository browser"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := newModel(l, len(dirsSorted))

	// TODO: option for fullscreen
	p := tea.NewProgram(m, tea.WithAltScreen())

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
