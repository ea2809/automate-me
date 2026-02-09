package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Title    lipgloss.Style
	Filter   lipgloss.Style
	Dim      lipgloss.Style
	Selected lipgloss.Style
	Group    lipgloss.Style
	Footer   lipgloss.Style
	Running  lipgloss.Style
	Loading  lipgloss.Style
}

func DefaultTheme() Theme {
	accent := colorEnv("AUTOMATE_ME_THEME_ACCENT", "42")
	accentDark := colorEnv("AUTOMATE_ME_THEME_ACCENT_DARK", "22")
	accentLight := colorEnv("AUTOMATE_ME_THEME_ACCENT_LIGHT", "120")
	text := colorEnv("AUTOMATE_ME_THEME_TEXT", "15")
	muted := colorEnv("AUTOMATE_ME_THEME_MUTED", "243")
	muted2 := colorEnv("AUTOMATE_ME_THEME_MUTED_2", "240")

	return Theme{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accentLight)),
		Filter:   lipgloss.NewStyle().Foreground(lipgloss.Color(accent)),
		Dim:      lipgloss.NewStyle().Foreground(lipgloss.Color(muted)),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(text)).Background(lipgloss.Color(accentDark)),
		Group:    lipgloss.NewStyle().Foreground(lipgloss.Color(accentLight)),
		Footer:   lipgloss.NewStyle().Foreground(lipgloss.Color(muted2)),
		Running:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accentLight)),
		Loading:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accent)),
	}
}

func colorEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
