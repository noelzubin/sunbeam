package tui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pomdtr/sunbeam/internal/fzf"
)

type FilterItem interface {
	FilterValue() string
	Render(width int, selected bool) string
	ID() string
}

type Filter struct {
	minIndex      int
	Width, Height int
	Query         string
	Less          func(i, j FilterItem) bool
	EmptyText     string

	items    []FilterItem
	filtered []FilterItem

	DrawLines bool
	cursor    int
}

func NewFilter(items ...FilterItem) Filter {
	return Filter{
		items:    items,
		filtered: items,
	}
}

func (f *Filter) ResetSelection() {
	f.minIndex = 0
	f.cursor = 0
}

func (f *Filter) SetSize(width, height int) {
	f.Width = width
	f.Height = height
}

func (f Filter) Selection() FilterItem {
	if f.cursor >= len(f.filtered) || f.cursor < 0 {
		return nil
	}
	return f.filtered[f.cursor]
}

func (f *Filter) SetItems(items ...FilterItem) {
	f.items = items
	f.filtered = items

	if f.cursor < 0 {
		f.cursor = 0
	}

	if f.cursor >= len(f.filtered) {
		f.cursor = len(f.filtered) - 1
	}
}

func (f *Filter) FilterItems(query string) {
	f.Query = query
	values := make([]string, len(f.items))
	for i, choice := range f.items {
		values[i] = choice.FilterValue()
	}
	// If the search field is empty, let's not display the matches
	// (none), but rather display all possible choices.
	if query == "" {
		f.filtered = f.items
	} else {
		f.filtered = make([]FilterItem, 0)
		for i := 0; i < len(f.items); i++ {
			filterValue := f.items[i].FilterValue()
			score := fzf.Score(filterValue, query)
			if score > 0 {
				f.filtered = append(f.filtered, f.items[i])
			}
		}

		sort.SliceStable(f.filtered, func(i, j int) bool {
			return fzf.Score(f.filtered[i].FilterValue(), query) > fzf.Score(f.filtered[j].FilterValue(), query)
		})
	}

	if f.cursor >= len(f.filtered) {
		f.cursor = len(f.filtered) - 1
	}
}

func (f *Filter) Select(id string) {
	for i, item := range f.filtered {
		if item.ID() == id {
			f.cursor = i
		}
	}
}

func (m Filter) Init() tea.Cmd { return nil }

func (m Filter) View() string {
	itemWidth := m.Width - 2
	rows := make([]string, 0)

	if len(m.filtered) == 0 {
		var emptyText string
		if m.EmptyText != "" {
			emptyText = m.EmptyText
		} else if len(m.items) > 0 && m.Query != "" {
			emptyText = "No matches"
		} else {
			emptyText = "No Items"
		}

		emptyText = lipgloss.NewStyle().Faint(true).Render(emptyText)
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, emptyText)
	}

	index := m.minIndex
	nbVisibleItems := m.Height
	if m.DrawLines {
		nbVisibleItems = nbVisibleItems/2 + nbVisibleItems%2
	}

	for nbVisibleItems > 0 && index < len(m.filtered) {
		item := m.filtered[index]
		itemView := item.Render(itemWidth, index == m.cursor)
		rows = append(rows, itemView)

		index++
		nbVisibleItems--

		if m.DrawLines && index < len(m.filtered) && nbVisibleItems > 0 {
			separator := strings.Repeat("─", itemWidth)
			separator = lipgloss.NewStyle().Faint(true).Render(separator)
			rows = append(rows, separator)
		}
	}

	if len(rows) == 0 {
		return ""
	}

	filteredView := lipgloss.JoinVertical(lipgloss.Left, rows...)
	filteredView = lipgloss.NewStyle().Padding(0, 1).Render(filteredView)
	return lipgloss.Place(m.Width, m.Height, lipgloss.Left, lipgloss.Top, filteredView)
}

func (f Filter) Update(msg tea.Msg) (Filter, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "ctrl+j", "ctrl+n":
			f.CursorDown()
		case "up", "ctrl+k", "ctrl+p":
			f.CursorUp()
		case "ctrl+u":
			shift := min(f.nbVisibleItems(), f.cursor)
			for i := 0; i < shift; i++ {
				f.CursorUp()
			}
		case "ctrl+d":
			shift := min(f.nbVisibleItems(), len(f.filtered)-f.cursor-1)
			for i := 0; i < shift; i++ {
				f.CursorDown()
			}
		}
	}

	return f, nil
}

func (m Filter) itemHeight() int {
	if m.DrawLines {
		return 2
	}
	return 1
}

func (m *Filter) CursorUp() {
	if m.cursor > 0 {
		m.cursor = m.cursor - 1
		if m.cursor < m.minIndex {
			m.minIndex = m.cursor
		}
	} else {
		m.cursor = len(m.filtered) - 1
		m.minIndex = max(0, m.cursor-m.nbVisibleItems()+1)
	}
}

func (m Filter) nbVisibleItems() int {
	return m.Height/m.itemHeight() + m.Height%m.itemHeight()
}

func (m *Filter) CursorDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor += 1
		if m.cursor >= m.minIndex+m.nbVisibleItems() {
			m.minIndex += 1
		}
	} else {
		m.cursor = 0
		m.minIndex = 0
	}
}
