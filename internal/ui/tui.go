package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ea2809/automate-me/internal/app"
	"github.com/ea2809/automate-me/internal/core"
)

type taskModel struct {
	tasks   []core.TaskRecord
	filter  string
	cursor  int
	width   int
	height  int
	choice  *core.TaskRecord
	theme   Theme
	refresh bool
}

var SelectTask = func(tasks []core.TaskRecord, state app.SelectionState) (core.TaskRecord, app.SelectionState, error) {
	model := taskModel{
		tasks:  tasks,
		filter: state.Filter,
		cursor: state.Cursor,
		theme:  DefaultTheme(),
	}
	program := tea.NewProgram(model)
	result, err := program.Run()
	if err != nil {
		return core.TaskRecord{}, state, err
	}
	finalModel := result.(taskModel)
	if finalModel.refresh {
		return core.TaskRecord{}, app.SelectionState{Filter: finalModel.filter, Cursor: finalModel.cursor}, app.ErrRefresh
	}
	if finalModel.choice == nil {
		return core.TaskRecord{}, state, ErrUserCanceled
	}
	return *finalModel.choice, app.SelectionState{Filter: finalModel.filter, Cursor: finalModel.cursor}, nil
}

func (m taskModel) Init() tea.Cmd {
	return nil
}

func (m taskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered())-1 {
				m.cursor++
			}
		case "enter":
			items := m.filtered()
			if len(items) > 0 {
				choice := items[m.cursor]
				m.choice = &choice
				return m, tea.Quit
			}
		case "r":
			m.refresh = true
			return m, tea.Quit
		case "backspace", "ctrl+h":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.cursor = 0
			}
		default:
			if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
				m.filter += string(msg.Runes)
				m.cursor = 0
			}
		}
	}
	m.clampCursor()
	return m, nil
}

func (m taskModel) View() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render("Automate-Me"))
	b.WriteString("\n")
	b.WriteString(m.theme.Filter.Render(fmt.Sprintf("Filter: %s", m.filter)))
	b.WriteString("\n\n")

	items := m.filtered()
	if len(items) == 0 {
		b.WriteString(m.theme.Dim.Render("No tasks match."))
		b.WriteString("\n")
		return b.String()
	}

	start, end := visibleRange(len(items), m.cursor, m.maxRows())
	if start > 0 || end < len(items) {
		b.WriteString(m.theme.Dim.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(items))))
		b.WriteString("\n")
	}
	for i := start; i < end; i++ {
		item := items[i]
		line := m.formatTaskLine(item)
		if i == m.cursor {
			b.WriteString(m.theme.Selected.Render(" > " + line))
		} else {
			b.WriteString("   " + line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Footer.Render("Enter: run  Esc: cancel  R: refresh  ↑/↓: move  Type: filter"))
	b.WriteString("\n")
	return b.String()
}

func (m taskModel) filtered() []core.TaskRecord {
	if strings.TrimSpace(m.filter) == "" {
		return m.tasks
	}
	needle := strings.ToLower(m.filter)
	var out []core.TaskRecord
	for _, task := range m.tasks {
		haystack := strings.ToLower(strings.Join([]string{
			core.TaskID(task.PluginID, task.Task.Name),
			task.Task.Title,
			task.Task.Description,
			task.Task.Group,
			task.PluginID,
			task.PluginTitle,
		}, " "))
		if strings.Contains(haystack, needle) {
			out = append(out, task)
		}
	}
	return out
}

func (m *taskModel) clampCursor() {
	items := m.filtered()
	if len(items) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(items) {
		m.cursor = len(items) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m taskModel) maxRows() int {
	if m.height > 0 {
		rows := m.height - 6
		if rows > 0 {
			return rows
		}
	}
	return 12
}

func visibleRange(total, cursor, maxRows int) (int, int) {
	if total <= maxRows {
		return 0, total
	}
	start := cursor - maxRows/2
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > total {
		end = total
		start = end - maxRows
	}
	return start, end
}

func (m taskModel) formatTaskLine(task core.TaskRecord) string {
	group := task.Task.Group
	if group == "" {
		group = "General"
	}
	title := task.Task.Title
	desc := task.Task.Description
	if desc != "" {
		return fmt.Sprintf("%s %s - %s %s",
			m.theme.Group.Render("["+group+"]"),
			title,
			m.theme.Dim.Render(desc),
			m.theme.Dim.Render("("+core.TaskID(task.PluginID, task.Task.Name)+")"),
		)
	}
	return fmt.Sprintf("%s %s %s",
		m.theme.Group.Render("["+group+"]"),
		title,
		m.theme.Dim.Render("("+core.TaskID(task.PluginID, task.Task.Name)+")"),
	)
}
