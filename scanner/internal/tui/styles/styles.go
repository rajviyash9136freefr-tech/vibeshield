package styles

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

var (
	// HasNoColor checks standard NO_COLOR environment variable.
	HasNoColor = os.Getenv("NO_COLOR") != ""

	// Brand & Palette Tokens
	ColorBrand    = lipgloss.Color("#818CF8") // Indigo
	ColorSubtle   = lipgloss.Color("#71717A") // Muted zinc
	ColorSuccess  = lipgloss.Color("#34D399") // Emerald
	ColorCritical = lipgloss.Color("#F87171") // Red
	ColorHigh     = lipgloss.Color("#FB923C") // Orange
	ColorMedium   = lipgloss.Color("#FBBF24") // Amber
	ColorLow      = lipgloss.Color("#60A5FA") // Blue
	ColorInfo     = lipgloss.Color("#A1A1AA") // Slate

	// Typography & Layout Styles
	BannerBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBrand).
			Padding(0, 1).
			MarginBottom(1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBrand)

	DimStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	PromptStyle = lipgloss.NewStyle().
			Foreground(ColorBrand).
			Bold(true)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorBrand).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorSubtle)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorBrand).
			Padding(0, 1)

	// Severity Badges
	BadgeCritical = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorCritical).
			Padding(0, 1)

	BadgeHigh = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorHigh).
			Padding(0, 1)

	BadgeMedium = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#000000")).
			Background(ColorMedium).
			Padding(0, 1)

	BadgeLow = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorLow).
			Padding(0, 1)

	BadgeClean = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorSuccess).
			Padding(0, 1)

	// Selected Row Style
	SelectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#27272A")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	// Diff Styles
	DiffMinusStyle = lipgloss.NewStyle().
			Foreground(ColorCritical)

	DiffPlusStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)
)

// SeverityBadge returns a styled badge according to severity.
func SeverityBadge(sev string) string {
	if HasNoColor {
		return "[" + sev + "]"
	}
	switch sev {
	case "critical":
		return BadgeCritical.Render("CRITICAL")
	case "high":
		return BadgeHigh.Render("HIGH")
	case "medium":
		return BadgeMedium.Render("MEDIUM")
	case "low":
		return BadgeLow.Render("LOW")
	default:
		return DimStyle.Render("INFO")
	}
}

// SeverityDot returns a color bullet for the severity.
func SeverityDot(sev string) string {
	if HasNoColor {
		return "•"
	}
	switch sev {
	case "critical":
		return lipgloss.NewStyle().Foreground(ColorCritical).Render("🔴")
	case "high":
		return lipgloss.NewStyle().Foreground(ColorHigh).Render("🟠")
	case "medium":
		return lipgloss.NewStyle().Foreground(ColorMedium).Render("🟡")
	case "low":
		return lipgloss.NewStyle().Foreground(ColorLow).Render("🔵")
	default:
		return lipgloss.NewStyle().Foreground(ColorInfo).Render("⚪")
	}
}
