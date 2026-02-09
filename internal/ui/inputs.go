package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ea2809/automate-me/internal/app"
	"github.com/ea2809/automate-me/internal/core"
	"golang.org/x/term"
)

var ErrUserCanceled = app.ErrUserCanceled

func PromptInputs(inputs []core.InputSpec) (map[string]any, error) {
	return PromptInputsWithDefaults(inputs, nil)
}

func PromptInputsWithDefaults(inputs []core.InputSpec, defaults map[string]any) (map[string]any, error) {
	values := make(map[string]any)
	reader := bufio.NewReader(os.Stdin)

	for _, input := range inputs {
		if defaults != nil {
			if override, ok := defaults[input.Name]; ok {
				input.Default = override
			}
		}
		value, err := promptOneInput(reader, input)
		if err != nil {
			return nil, err
		}
		values[input.Name] = value
	}

	return values, nil
}

func promptOneInput(reader *bufio.Reader, input core.InputSpec) (any, error) {
	prompt := input.Prompt
	if prompt == "" {
		prompt = input.Name
	}
	if input.Type == "enum" && len(input.Choices) > 0 {
		choice, err := selectEnum(prompt, input.Choices, input.Default)
		if err != nil {
			return nil, err
		}
		if input.Required && choice == "" {
			return nil, fmt.Errorf("missing required input: %s", input.Name)
		}
		return choice, nil
	}
	if input.Type == "enum" && len(input.Choices) == 0 && input.Required {
		return nil, fmt.Errorf("no choices available for %s", input.Name)
	}
	if input.Type == "multienum" && len(input.Choices) > 0 {
		choices, err := selectMultiEnum(prompt, input.Choices, input.Default)
		if err != nil {
			return nil, err
		}
		if input.Required && len(choices) == 0 {
			return nil, fmt.Errorf("missing required input: %s", input.Name)
		}
		return choices, nil
	}
	if input.Type == "multienum" && len(input.Choices) == 0 && input.Required {
		return nil, fmt.Errorf("no choices available for %s", input.Name)
	}
	for {
		fmt.Print(formatPrompt(prompt, input))
		line, err := readInput(reader, input.Secret)
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" && input.Default != nil {
			return input.Default, nil
		}
		if line == "" && !input.Required {
			return nil, nil
		}
		value, err := parseInputValue(input, line)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		return value, nil
	}
}

func readInput(reader *bufio.Reader, secret bool) (string, error) {
	if !secret {
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return line, nil
	}
	fmt.Print(" ")
	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func formatPrompt(prompt string, input core.InputSpec) string {
	var extras []string
	if len(input.Choices) > 0 {
		extras = append(extras, "choices: "+strings.Join(input.Choices, ","))
	}
	if input.Default != nil {
		extras = append(extras, fmt.Sprintf("default: %v", input.Default))
	}
	if len(extras) > 0 {
		return fmt.Sprintf("%s (%s): ", prompt, strings.Join(extras, "; "))
	}
	return fmt.Sprintf("%s: ", prompt)
}

func parseInputValue(input core.InputSpec, raw string) (any, error) {
	switch input.Type {
	case "string", "path":
		return raw, nil
	case "int":
		value, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid int for %s", input.Name)
		}
		return value, nil
	case "float":
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float for %s", input.Name)
		}
		return value, nil
	case "bool":
		switch strings.ToLower(raw) {
		case "true", "t", "yes", "y", "1":
			return true, nil
		case "false", "f", "no", "n", "0":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid bool for %s", input.Name)
		}
	case "enum":
		return parseEnum(input, raw)
	case "multienum":
		parts := splitCSV(raw)
		for _, part := range parts {
			if !contains(input.Choices, part) {
				return nil, fmt.Errorf("invalid choice %q for %s", part, input.Name)
			}
		}
		return parts, nil
	default:
		return nil, fmt.Errorf("unsupported input type %q for %s", input.Type, input.Name)
	}
}

func parseEnum(input core.InputSpec, raw string) (string, error) {
	if len(input.Choices) == 0 {
		return raw, nil
	}
	if !contains(input.Choices, raw) {
		return "", fmt.Errorf("invalid choice %q for %s", raw, input.Name)
	}
	return raw, nil
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	var out []string
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

type enumModel struct {
	title    string
	choices  []string
	cursor   int
	picked   string
	canceled bool
	filter   string
	height   int
	theme    Theme
}

func selectEnum(prompt string, choices []string, def any) (string, error) {
	model := enumModel{
		title:   prompt,
		choices: choices,
		cursor:  defaultIndex(choices, def),
		theme:   DefaultTheme(),
	}
	program := tea.NewProgram(model)
	result, err := program.Run()
	if err != nil {
		return "", err
	}
	final := result.(enumModel)
	if final.canceled {
		return "", ErrUserCanceled
	}
	if final.picked == "" && len(choices) > 0 {
		return choices[final.cursor], nil
	}
	return final.picked, nil
}

func (m enumModel) Init() tea.Cmd { return nil }

func (m enumModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(filteredIndices(m.choices, m.filter))-1 {
				m.cursor++
			}
		case "backspace", "ctrl+h":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.cursor = 0
			}
		case "enter":
			indices := filteredIndices(m.choices, m.filter)
			if len(indices) > 0 {
				m.picked = m.choices[indices[m.cursor]]
			}
			return m, tea.Quit
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

func (m enumModel) View() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render(m.title))
	b.WriteString("\n")
	b.WriteString(m.theme.Filter.Render("Filter: " + m.filter))
	b.WriteString("\n\n")
	indices := filteredIndices(m.choices, m.filter)
	start, end := visibleRangeInput(len(indices), m.cursor, m.maxRows())
	if start > 0 || end < len(indices) {
		b.WriteString(m.theme.Dim.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(indices))))
		b.WriteString("\n")
	}
	for i := start; i < end; i++ {
		idx := indices[i]
		choice := m.choices[idx]
		if i == m.cursor {
			b.WriteString(m.theme.Selected.Render(" > " + choice))
			b.WriteString("\n")
			continue
		}
		b.WriteString("   " + choice)
		b.WriteString("\n")
	}
	if len(indices) == 0 {
		b.WriteString(m.theme.Dim.Render("No matches."))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(m.theme.Footer.Render("Enter: select  Esc: cancel  ↑/↓: move  Type: filter"))
	b.WriteString("\n")
	return b.String()
}

func (m *enumModel) clampCursor() {
	indices := filteredIndices(m.choices, m.filter)
	if len(indices) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(indices) {
		m.cursor = len(indices) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m enumModel) maxRows() int {
	if m.height > 0 {
		rows := m.height - 6
		if rows > 0 {
			return rows
		}
	}
	return 10
}

type multiEnumModel struct {
	title    string
	choices  []string
	selected map[int]bool
	cursor   int
	canceled bool
	filter   string
	height   int
	theme    Theme
}

func selectMultiEnum(prompt string, choices []string, def any) ([]string, error) {
	model := multiEnumModel{
		title:    prompt,
		choices:  choices,
		selected: defaultSelection(choices, def),
		theme:    DefaultTheme(),
	}
	program := tea.NewProgram(model)
	result, err := program.Run()
	if err != nil {
		return nil, err
	}
	final := result.(multiEnumModel)
	if final.canceled {
		return nil, ErrUserCanceled
	}
	var out []string
	for i, choice := range final.choices {
		if final.selected[i] {
			out = append(out, choice)
		}
	}
	return out, nil
}

func (m multiEnumModel) Init() tea.Cmd { return nil }

func (m multiEnumModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(filteredIndices(m.choices, m.filter))-1 {
				m.cursor++
			}
		case " ":
			indices := filteredIndices(m.choices, m.filter)
			if len(indices) > 0 {
				idx := indices[m.cursor]
				m.selected[idx] = !m.selected[idx]
			}
		case "backspace", "ctrl+h":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.cursor = 0
			}
		case "enter":
			return m, tea.Quit
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

func (m multiEnumModel) View() string {
	var b strings.Builder
	b.WriteString(m.theme.Title.Render(m.title))
	b.WriteString("\n")
	b.WriteString(m.theme.Filter.Render("Filter: " + m.filter))
	b.WriteString("\n\n")
	indices := filteredIndices(m.choices, m.filter)
	start, end := visibleRangeInput(len(indices), m.cursor, m.maxRows())
	if start > 0 || end < len(indices) {
		b.WriteString(m.theme.Dim.Render(fmt.Sprintf("Showing %d-%d of %d", start+1, end, len(indices))))
		b.WriteString("\n")
	}
	for i := start; i < end; i++ {
		idx := indices[i]
		choice := m.choices[idx]
		mark := " "
		if m.selected[idx] {
			mark = "x"
		}
		if i == m.cursor {
			b.WriteString(m.theme.Selected.Render(fmt.Sprintf(" > [%s] %s", mark, choice)))
			b.WriteString("\n")
			continue
		}
		b.WriteString(fmt.Sprintf("   [%s] %s\n", mark, choice))
	}
	if len(indices) == 0 {
		b.WriteString(m.theme.Dim.Render("No matches."))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(m.theme.Footer.Render("Space: toggle  Enter: confirm  Esc: cancel  ↑/↓: move  Type: filter"))
	b.WriteString("\n")
	return b.String()
}

func (m *multiEnumModel) clampCursor() {
	indices := filteredIndices(m.choices, m.filter)
	if len(indices) == 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(indices) {
		m.cursor = len(indices) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m multiEnumModel) maxRows() int {
	if m.height > 0 {
		rows := m.height - 6
		if rows > 0 {
			return rows
		}
	}
	return 10
}

func defaultIndex(choices []string, def any) int {
	value, ok := def.(string)
	if !ok || value == "" {
		return 0
	}
	for i, choice := range choices {
		if choice == value {
			return i
		}
	}
	return 0
}

func defaultSelection(choices []string, def any) map[int]bool {
	selected := make(map[int]bool)
	list, ok := def.([]string)
	if !ok {
		return selected
	}
	for i, choice := range choices {
		for _, value := range list {
			if choice == value {
				selected[i] = true
			}
		}
	}
	return selected
}

func filteredIndices(choices []string, filter string) []int {
	if strings.TrimSpace(filter) == "" {
		indices := make([]int, len(choices))
		for i := range choices {
			indices[i] = i
		}
		return indices
	}
	needle := strings.ToLower(filter)
	var out []int
	for i, choice := range choices {
		if strings.Contains(strings.ToLower(choice), needle) {
			out = append(out, i)
		}
	}
	return out
}

func visibleRangeInput(total, cursor, maxRows int) (int, int) {
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
