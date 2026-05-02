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

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E5E7EB")).
			Background(lipgloss.Color("#1E1B4B")).
			Padding(0, 2)

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	statusActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#10B981"))

	statusInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#374151"))
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
		leftWidth := msg.Width * 2 / 5
		if leftWidth < 30 {
			leftWidth = 30
		}
		m.langList.SetSize(leftWidth, h)
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

func (m Model) envDetailsView() string {
	selected, ok := m.langList.SelectedItem().(item)
	if !ok || m.app.Config == nil {
		return ""
	}

	langName := selected.title
	var lang *config.Language
	for i := range m.app.Config.Languages {
		if m.app.Config.Languages[i].Name == langName {
			lang = &m.app.Config.Languages[i]
			break
		}
	}
	if lang == nil {
		return ""
	}

	versions := m.app.Versions[langName]
	activeInfo, hasActive := m.app.State.ActiveVersions[lang.SymlinkName]

	// Find the active version's binary path
	var activeBinPath string
	for _, v := range versions {
		if v.IsActive {
			activeBinPath = v.BinaryPath
			break
		}
	}

	s := detailTitleStyle.Render("ENV DETAILS") + "\n"
	s += subtleStyle.Render(strings.Repeat("─", 20)) + "\n\n"

	// Status
	if hasActive {
		s += labelStyle.Render("Status    ") + statusActiveStyle.Render("● Active") + "\n"
		s += labelStyle.Render("Version   ") + valueStyle.Render(activeInfo.Version) + "\n"
		if activeBinPath != "" {
			s += labelStyle.Render("Binary    ") + valueStyle.Render(activeBinPath) + "\n"
		}
	} else {
		s += labelStyle.Render("Status    ") + statusInactiveStyle.Render("● Inactive") + "\n"
		s += labelStyle.Render("Version   ") + subtleStyle.Render("—") + "\n"
	}

	// Home paths and detected versions
	s += labelStyle.Render("Homes     ") + valueStyle.Render(fmt.Sprintf("%d configured", len(lang.HomePaths))) + "\n"
	s += labelStyle.Render("Detected  ") + valueStyle.Render(fmt.Sprintf("%d versions", len(versions))) + "\n"

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
	s := headerStyle.Render("⚡ poly-switch") + "\n"
	s += separatorStyle.Render(strings.Repeat("─", m.width)) + "\n\n"
	s += fmt.Sprintf("  %s Scanning for installed versions...\n", m.spinner.View())
	s += subtleStyle.Render("\n  q  quit")
	return s
}

func (m Model) langListView() string {
	// Header
	s := headerStyle.Render("⚡ poly-switch") + "\n"
	s += separatorStyle.Render(strings.Repeat("─", m.width)) + "\n"

	// Two-column layout: language list | environment details
	leftCol := m.langList.View()
	rightCol := m.envDetailsView()
	sep := "  " + separatorStyle.Render("│") + "  "
	s += lipgloss.JoinHorizontal(lipgloss.Top, leftCol, sep, rightCol) + "\n"

	// PATH warning below the columns
	s += m.pathWarning()
	if !m.binInPath {
		s += "\n"
	}

	// Footer
	s += subtleStyle.Render("  ↑/↓ navigate  enter select  q/esc quit")
	return s
}

func (m Model) versionListView() string {
	s := headerStyle.Render(fmt.Sprintf("⚡ poly-switch / %s", m.selectedLang)) + "\n"
	s += separatorStyle.Render(strings.Repeat("─", m.width)) + "\n"
	s += m.versionList.View()
	s += "\n" + subtleStyle.Render("  ↑/↓ navigate  enter select  esc back  q quit")
	return s
}

func (m Model) applyingView() string {
	s := headerStyle.Render(fmt.Sprintf("⚡ poly-switch / %s", m.selectedLang)) + "\n"
	s += separatorStyle.Render(strings.Repeat("─", m.width)) + "\n\n"
	s += fmt.Sprintf("  %s Switching to %s %s...\n",
		m.spinner.View(), m.selectedLang, m.selectedVerLabel)
	return s
}

func (m Model) doneView() string {
	s := headerStyle.Render("⚡ poly-switch") + "\n"
	s += separatorStyle.Render(strings.Repeat("─", m.width)) + "\n\n"
	if m.err != nil {
		s += fmt.Sprintf("  %s Failed to switch: %v\n\n  Press any key to quit.\n",
			errorStyle.Render("ERROR"), m.err)
		return s
	}
	s += fmt.Sprintf("  %s Switched to %s %s\n",
		activeStyle.Render("✓"), m.selectedLang, m.selectedVerLabel)
	s += m.pathWarning()
	s += "\n  Press any key to quit.\n"
	return s
}
