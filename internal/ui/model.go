package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"poly-switch/internal/config"
	"poly-switch/internal/core"
)

type state int

const (
	stateLoading state = iota
	stateLangSelect
	stateVersionSelect
	stateApplying
	stateDone
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	activeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444"))

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	warningStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F59E0B"))
)

// item represents a selectable list entry in the TUI.
type item struct {
	title       string
	description string
	path        string // version directory path, used internally for version items
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

// Model implements the Bubble Tea state machine:
// Loading -> LanguageSelect -> VersionSelect -> Applying -> Done
type Model struct {
	app              *core.App
	state            state
	langList         list.Model
	versionList      list.Model
	spinner          spinner.Model
	selectedLang     string
	selectedVerPath  string
	selectedVerLabel string
	err              error
	width            int
	height           int
	binInPath        bool
}

func NewModel(app *core.App) Model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
	s.Spinner = spinner.Dot

	langItems := make([]list.Item, len(app.Config.Languages))
	for i, lang := range app.Config.Languages {
		langItems[i] = item{
			title:       lang.Name,
			description: fmt.Sprintf("Switch %s version", lang.Name),
		}
	}

	langList := list.New(langItems, list.NewDefaultDelegate(), 0, 0)
	langList.Title = "Select Language"
	langList.Styles.Title = titleStyle
	langList.SetShowStatusBar(false)

	versionList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	versionList.Title = "Select Version"
	versionList.Styles.Title = titleStyle
	versionList.SetShowStatusBar(false)

	return Model{
		app:         app,
		state:       stateLoading,
		spinner:     s,
		langList:    langList,
		versionList: versionList,
		binInPath:   checkBinInPath(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadVersions())
}

func (m Model) loadVersions() tea.Cmd {
	return func() tea.Msg {
		m.app.DetectAll()
		return versionsLoadedMsg{}
	}
}

type versionsLoadedMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h := msg.Height - 6
		if h < 1 {
			h = 1
		}
		m.langList.SetSize(msg.Width, h)
		m.versionList.SetSize(msg.Width, h)
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case versionsLoadedMsg:
		m.state = stateLangSelect
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case switchDoneMsg:
		m.state = stateDone
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateLoading, stateApplying:
		return m, nil

	case stateLangSelect:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, tea.Quit
		case "enter":
			selected, ok := m.langList.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			m.selectedLang = selected.title
			var cmd tea.Cmd
			m, cmd = m.buildVersionList()
			return m, cmd
		default:
			var cmd tea.Cmd
			m.langList, cmd = m.langList.Update(msg)
			return m, cmd
		}

	case stateVersionSelect:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateLangSelect
			return m, nil
		case "enter":
			selected, ok := m.versionList.SelectedItem().(item)
			if !ok {
				return m, nil
			}
			m.selectedVerLabel = selected.title
			m.selectedVerPath = selected.path
			m.state = stateApplying
			return m, tea.Batch(m.spinner.Tick, m.applySwitch())
		default:
			var cmd tea.Cmd
			m.versionList, cmd = m.versionList.Update(msg)
			return m, cmd
		}

	case stateDone:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) buildVersionList() (Model, tea.Cmd) {
	versions := m.app.Versions[m.selectedLang]
	items := make([]list.Item, 0, len(versions))

	for _, v := range versions {
		desc := ""
		if v.IsActive {
			desc = "currently active"
		}
		items = append(items, item{
			title:       v.Label,
			description: desc,
			path:        v.Path,
		})
	}

	if len(items) == 0 {
		items = append(items, item{
			title:       "No versions found",
			description: "Press esc to go back",
		})
	}

	m.versionList.SetItems(items)
	m.state = stateVersionSelect
	return m, nil
}

func (m Model) applySwitch() tea.Cmd {
	return func() tea.Msg {
		err := m.app.ApplySwitch(m.selectedLang, m.selectedVerPath)
		return switchDoneMsg{err: err}
	}
}

type switchDoneMsg struct {
	err error
}

// checkBinInPath reports whether ~/.poly-switch/bin is present in the user's PATH.
func checkBinInPath() bool {
	binDir, err := config.BinDir()
	if err != nil {
		return true
	}
	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		abs, err := filepath.Abs(os.ExpandEnv(dir))
		if err != nil {
			continue
		}
		if abs == binDir {
			return true
		}
	}
	return false
}

func shellRcFile() string {
	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return "~/.zshrc"
	case strings.Contains(shell, "bash"):
		return "~/.bashrc"
	default:
		return "~/.profile"
	}
}

func (m Model) pathWarning() string {
	if m.binInPath {
		return ""
	}
	s := "\n  " + warningStyle.Render("! ") + subtleStyle.Render("~/.poly-switch/bin is not in your PATH") + "\n"
	s += subtleStyle.Render("    Add to " + shellRcFile() + ":") + "\n"
	s += subtleStyle.Render(`    export PATH="$HOME/.poly-switch/bin:$PATH"`)
	return s
}

func (m Model) View() string {
	switch m.state {
	case stateLoading:
		return m.loadingView()
	case stateLangSelect:
		return m.langListView()
	case stateVersionSelect:
		return m.versionListView()
	case stateApplying:
		return m.applyingView()
	case stateDone:
		return m.doneView()
	default:
		return "unknown state"
	}
}

func (m Model) loadingView() string {
	s := fmt.Sprintf("\n  %s Scanning for installed versions...\n", m.spinner.View())
	s += subtleStyle.Render("\n  q  quit")
	return s
}

func (m Model) langListView() string {
	s := titleStyle.Render("poly-switch") + "\n"
	s += m.pathWarning()
	if !m.binInPath {
		s += "\n"
	}
	s += m.langList.View()
	s += "\n" + subtleStyle.Render("  up/down navigate  enter select  q/esc quit")
	return s
}

func (m Model) versionListView() string {
	s := titleStyle.Render(fmt.Sprintf("poly-switch / %s", m.selectedLang)) + "\n"
	s += m.versionList.View()
	s += "\n" + subtleStyle.Render("  ↑/↓ navigate  enter select  esc back  q quit")
	return s
}

func (m Model) applyingView() string {
	return fmt.Sprintf("\n  %s Switching to %s %s...\n",
		m.spinner.View(), m.selectedLang, m.selectedVerLabel)
}

func (m Model) doneView() string {
	if m.err != nil {
		return fmt.Sprintf("\n  %s Failed to switch: %v\n\n  Press any key to quit.\n",
			errorStyle.Render("ERROR"), m.err)
	}
	s := fmt.Sprintf("\n  %s Switched to %s %s\n",
		activeStyle.Render("OK"), m.selectedLang, m.selectedVerLabel)
	s += m.pathWarning()
	s += "\n  Press any key to quit.\n"
	return s
}
