package viewer

import (
	"github.com/TotallyNotLost/gotes/cmd"
	"github.com/TotallyNotLost/gotes/formatter"
	glist "github.com/TotallyNotLost/gotes/list"
	"github.com/TotallyNotLost/gotes/storage"
	"github.com/TotallyNotLost/gotes/tabs"
	"github.com/TotallyNotLost/gotes/tags"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"strconv"
)

var (
	viewerStyle             = lipgloss.NewStyle().Padding(0, 2)
	helpStyle               = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingTop(1).Render
	minWidthForRelated      = 100
	relatedViewWidthPercent = 0.4
	relatedViewMaxWidth     = 40
)

type mode struct {
	id     int
	keyMap keyMap
}

func (m mode) Equal(m2 mode) bool {
	if m.id == m2.id {
		return true
	}
	return false
}

var normal = mode{
	id: 0,
	keyMap: keyMap{
		Back:           key.NewBinding(key.WithKeys("backspace", "q"), key.WithHelp("q", "back")),
		Edit:           key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		ToggleMarkdown: key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "toggle markdown")),
		RelatedMode:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "related")),
	},
}

var related = mode{
	id: 1,
	keyMap: keyMap{
		View:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "view")),
		NormalMode: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "normal")),
	},
}

func New(storage *storage.Storage) Model {
	var d list.DefaultDelegate
	d = list.NewDefaultDelegate()
	l := list.New([]list.Item{}, d, 0, 0)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.Title = "Related"

	mdFormatter := formatter.NewMarkdownFormatter(storage)
	tbs := tabs.New()
	tbs.SetFormatter(mdFormatter)
	return Model{
		tabs:                tbs,
		relatedList:         l,
		relatedListDelegate: d,
		renderMarkdown:      true,
		mode:                normal,
		markdownFormatter:   mdFormatter,
		storage:             storage,
		help:                help.New(),
	}
}

type Model struct {
	tabs                tabs.Model
	revisions           []storage.Entry
	relatedList         list.Model
	relatedListDelegate list.DefaultDelegate
	lastActiveRevision  string
	lastActiveTab       int
	renderMarkdown      bool
	mode                mode
	markdownFormatter   formatter.MarkdownFormatter
	width               int
	storage             *storage.Storage
	help                help.Model
}

func (m *Model) SetHeight(height int) {
	m.tabs.SetHeight(height - lipgloss.Height(m.tagsView()) - lipgloss.Height(m.helpView()))
	m.relatedList.SetHeight(height)
}

func (m *Model) SetWidth(width int) {
	m.width = width
	relatedViewWidth := min(relatedViewMaxWidth, int(float64(width)*relatedViewWidthPercent))
	if width > minWidthForRelated {
		width -= relatedViewWidth
	}
	paddingWidth := lipgloss.Width(viewerStyle.Render(""))
	m.tabs.SetWidth(width - paddingWidth)
	m.markdownFormatter.SetWidth(width - paddingWidth)
	m.relatedList.SetWidth(relatedViewWidth)
}

func (m *Model) SetRevisions(revisions []storage.Entry) {
	m.revisions = revisions
	tabs := lo.Map(m.revisions, func(revision storage.Entry, i int) tabs.Tab {
		title := "Revision HEAD~" + strconv.Itoa(i)
		if i == 0 {
			title += " (latest)"
		}
		return tabs.NewTab(title, revision.Text())
	})
	m.tabs.SetTabs(tabs)
	m.tabs.SetEntryId(m.getActiveRevision().Id())
	m.tabs.AdjustHeight()
	m.mode = normal
	m.updateRelatedList()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.mode.keyMap.Back):
			return m, cmd.Back
		case key.Matches(msg, m.mode.keyMap.Edit):
			return m, cmd.EditEntry(m.getActiveRevision())
		case key.Matches(msg, m.mode.keyMap.View):
			entry := m.relatedList.SelectedItem().(glist.Item).Entry()
			return m, cmd.ViewEntry(entry)
		case key.Matches(msg, m.mode.keyMap.ToggleMarkdown):
			m.renderMarkdown = !m.renderMarkdown
			if m.renderMarkdown {
				m.tabs.SetFormatter(m.markdownFormatter)
			} else {
				m.tabs.SetFormatter(formatter.Default)
			}
		case key.Matches(msg, m.mode.keyMap.NormalMode):
			m.mode = normal
		case key.Matches(msg, m.mode.keyMap.RelatedMode):
			m.mode = related
		}
	}

	var cmd, rlCmd tea.Cmd
	m.tabs, cmd = m.tabs.Update(msg)
	if m.mode.Equal(related) {
		m.relatedList, rlCmd = m.relatedList.Update(msg)
	}

	m.updateRelatedList()

	return m, tea.Batch(cmd, rlCmd)
}

func (m Model) View() string {
	body := m.tabs.View()
	viewer := lipgloss.JoinVertical(lipgloss.Left, m.tagsView(), body, m.helpView())

	if m.width < minWidthForRelated {
		return lipgloss.JoinHorizontal(lipgloss.Top, viewerStyle.Render(viewer))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, viewerStyle.Render(viewer), m.relatedList.View())
}

func (m Model) tagsView() string {
	revision := m.getActiveRevision()
	return lipgloss.JoinHorizontal(lipgloss.Center, tags.RenderTags(revision.Tags()))
}

func (m *Model) updateRelatedList() {
	var d list.DefaultDelegate
	d = m.relatedListDelegate

	if m.mode.Equal(normal) {
		// Change colors
		d.Styles.SelectedTitle = d.Styles.NormalTitle
		d.Styles.SelectedDesc = d.Styles.NormalDesc
	}

	m.relatedList.SetDelegate(d)

	if m.lastActiveRevision == m.getActiveRevision().Id() && m.lastActiveTab == m.tabs.ActiveTab {
		// Already up to date
		return
	}

	m.lastActiveRevision = m.getActiveRevision().Id()
	m.lastActiveTab = m.tabs.ActiveTab

	entries := m.storage.GetRelatedTo(m.getActiveRevision())

	m.relatedList.SetItems(lo.Map(entries, func(entry storage.Entry, index int) list.Item {
		return glist.EntryToItem(entry)
	}))
}

func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{
		m.mode.keyMap.Back,
		m.mode.keyMap.Edit,
		m.mode.keyMap.View,
		m.mode.keyMap.ToggleMarkdown,
		m.mode.keyMap.NormalMode,
		m.mode.keyMap.RelatedMode,
	}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{append(m.ShortHelp(), m.tabs.ShortHelp()...)}
}

func (m Model) getActiveRevision() storage.Entry {
	if m.tabs.ActiveTab >= len(m.revisions) {
		return storage.Entry{}
	}
	return m.revisions[m.tabs.ActiveTab]
}

func (m Model) helpView() string {
	return helpStyle(m.help.View(m))
}

type keyMap struct {
	Back           key.Binding
	Edit           key.Binding
	View           key.Binding
	ToggleMarkdown key.Binding
	NormalMode     key.Binding
	RelatedMode    key.Binding
}
