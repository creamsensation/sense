package sense

import "github.com/charmbracelet/lipgloss"

var (
	BlueColor    = lipgloss.NewStyle().Foreground(lipgloss.Color("#60a5fa"))
	EmeraldColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#34d399"))
	WhiteColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
)

var (
	Divider = WhiteColor.Render("------------------------------")
)
