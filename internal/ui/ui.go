// Package ui
package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Couleurs
	colorPrimary = lipgloss.Color("#7C3AED") // violet
	colorSuccess = lipgloss.Color("#10B981") // vert
	colorWarning = lipgloss.Color("#F59E0B") // orange
	colorError   = lipgloss.Color("#EF4444") // rouge
	colorMuted   = lipgloss.Color("#6B7280") // gris
	colorText    = lipgloss.Color("#F9FAFB") // blanc cassé

	// Styles de base
	Bold  = lipgloss.NewStyle().Bold(true)
	Muted = lipgloss.NewStyle().Foreground(colorMuted)

	// Badge coloré (ex: "front" / "back")
	BadgeFront = lipgloss.NewStyle().
			Background(lipgloss.Color("#1D4ED8")).
			Foreground(colorText).
			Padding(0, 1).
			Bold(true)

	BadgeBack = lipgloss.NewStyle().
			Background(lipgloss.Color("#065F46")).
			Foreground(colorText).
			Padding(0, 1).
			Bold(true)

	// Bloc titre principal
	Title = lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(colorPrimary).
		MarginBottom(1)

	// Messages
	Success = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	Warning = lipgloss.NewStyle().Foreground(colorWarning).Bold(true)
	Error   = lipgloss.NewStyle().Foreground(colorError).Bold(true)

	// Carte projet (pour tdev list)
	ProjectCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 2).
			MarginBottom(1)
)

// Helpers

func PrintSuccess(msg string) string {
	return Success.Render("✓ ") + msg
}

func PrintError(msg string) string {
	return Error.Render("✗ ") + msg
}

func PrintWarning(msg string) string {
	return Warning.Render("⚠ ") + msg
}

func PrintInfo(label, value string) string {
	return Muted.Render(label+": ") + value
}
