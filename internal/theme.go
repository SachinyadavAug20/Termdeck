package internal

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme represents a complete color scheme and styling configuration for Termdeck.
type Theme struct {
	ID        string
	Name      string
	Accent    string
	Secondary string
	Success   string
	Warning   string
	Comment   string
	Muted     string
	DimTrack  string
	Laser     string

	// Precomputed Lipgloss Styles
	H1Style                 lipgloss.Style
	CodeSpanStyle           lipgloss.Style
	NotesTitleStyle         lipgloss.Style
	TableHeaderStyle        lipgloss.Style
	TableBorderStyle        lipgloss.Style
	TableCellStyle          lipgloss.Style
	HelpBoxStyle            lipgloss.Style
	HelpTitleStyle          lipgloss.Style
	HelpHeaderStyle         lipgloss.Style
	HelpKeyStyle            lipgloss.Style
	HelpDescStyle           lipgloss.Style
	ProgressLineFilledStyle lipgloss.Style
	ProgressLineDimStyle    lipgloss.Style
	LaserPointerStyle       lipgloss.Style
	SyntaxKeyword           lipgloss.Style
	SyntaxString            lipgloss.Style
	SyntaxComment           lipgloss.Style
	SyntaxNumber            lipgloss.Style
	SyntaxType              lipgloss.Style
	SyntaxPlain             lipgloss.Style
}

func buildTheme(id, name, accent, secondary, success, warning, comment, muted, dimTrack, laser string) Theme {
	return Theme{
		ID:        id,
		Name:      name,
		Accent:    accent,
		Secondary: secondary,
		Success:   success,
		Warning:   warning,
		Comment:   comment,
		Muted:     muted,
		DimTrack:  dimTrack,
		Laser:     laser,

		H1Style:                 lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accent)).Underline(true),
		CodeSpanStyle:           lipgloss.NewStyle().Foreground(lipgloss.Color(accent)).Background(lipgloss.Color(dimTrack)).Padding(0, 1),
		NotesTitleStyle:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accent)),
		TableHeaderStyle:        lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accent)).Padding(0, 1),
		TableBorderStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color(muted)),
		TableCellStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Padding(0, 1),
		HelpBoxStyle:            lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(accent)).Foreground(lipgloss.Color("252")).Padding(1, 2),
		HelpTitleStyle:          lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(accent)),
		HelpHeaderStyle:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(secondary)),
		HelpKeyStyle:            lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(success)),
		HelpDescStyle:           lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		ProgressLineFilledStyle: lipgloss.NewStyle().Foreground(lipgloss.Color(accent)),
		ProgressLineDimStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color(dimTrack)),
		LaserPointerStyle:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(laser)),
		SyntaxKeyword:           lipgloss.NewStyle().Foreground(lipgloss.Color(accent)).Bold(true),
		SyntaxString:            lipgloss.NewStyle().Foreground(lipgloss.Color(success)),
		SyntaxComment:           lipgloss.NewStyle().Foreground(lipgloss.Color(comment)).Italic(true),
		SyntaxNumber:            lipgloss.NewStyle().Foreground(lipgloss.Color(warning)),
		SyntaxType:              lipgloss.NewStyle().Foreground(lipgloss.Color(secondary)),
		SyntaxPlain:             lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
	}
}

// Built-in curated theme registry
var registry = []Theme{
	buildTheme("termdeck", "Termdeck Pink", "212", "75", "114", "208", "240", "240", "236", "#FF2A55"),
	buildTheme("tokyo-night", "Tokyo Night", "#7aa2f7", "#7dcfff", "#9ece6a", "#ff9e64", "#565f89", "#565f89", "#24283b", "#f7768e"),
	buildTheme("dracula", "Dracula", "#bd93f9", "#8be9fd", "#50fa7b", "#ffb86c", "#6272a4", "#6272a4", "#282a36", "#ff5555"),
	buildTheme("catppuccin", "Catppuccin Mocha", "#cba6f7", "#89dceb", "#a6e3a1", "#fab387", "#6c7086", "#585b70", "#313244", "#f38ba8"),
	buildTheme("nord", "Nord", "#88c0d0", "#81a1c1", "#a3be8c", "#ebcb8b", "#4c566a", "#4c566a", "#2e3440", "#bf616a"),
	buildTheme("gruvbox", "Gruvbox Dark", "#fe8019", "#83a598", "#b8bb26", "#fabd2f", "#928374", "#665c54", "#282828", "#fb4934"),
	buildTheme("monokai", "Monokai", "#66d9ef", "#ae81ff", "#a6e22e", "#fd971f", "#75715e", "#75715e", "#272822", "#f92672"),
	buildTheme("solarized", "Solarized Dark", "#268bd2", "#2aa198", "#859900", "#cb4b16", "#586e75", "#073642", "#002b36", "#dc322f"),
	buildTheme("cyberpunk", "Cyberpunk", "#00ffff", "#ff0055", "#00ff66", "#ffff00", "#625470", "#3d2d4c", "#1a1025", "#ff0055"),
}

var currentTheme = registry[0]

// AvailableThemes returns a copy of all registered themes.
func AvailableThemes() []Theme {
	cp := make([]Theme, len(registry))
	copy(cp, registry)
	return cp
}

// ThemeIDs returns the string IDs of all registered themes.
func ThemeIDs() []string {
	ids := make([]string, len(registry))
	for i, t := range registry {
		ids[i] = t.ID
	}
	return ids
}

// CurrentTheme returns the currently active theme.
func CurrentTheme() Theme {
	return currentTheme
}

// SetCurrentTheme resolves and sets the active theme by name or hex code.
func SetCurrentTheme(name string) Theme {
	th := ResolveTheme(name)
	currentTheme = th
	return th
}

// ResolveTheme finds a matching theme from candidates, falling back to the default Termdeck theme.
func ResolveTheme(names ...string) Theme {
	for _, raw := range names {
		raw = strings.ToLower(strings.TrimSpace(raw))
		if raw == "" {
			continue
		}
		// Direct custom hex color support (e.g. #3b82f6)
		if strings.HasPrefix(raw, "#") && (len(raw) == 4 || len(raw) == 7) {
			return buildTheme("custom", "Custom ("+raw+")", raw, "75", "114", "208", "240", "240", "236", raw)
		}
		// Normalization aliases
		switch raw {
		case "pink", "default":
			raw = "termdeck"
		case "tokyonight":
			raw = "tokyo-night"
		case "catppuccin-mocha", "mocha":
			raw = "catppuccin"
		case "gruvbox-dark":
			raw = "gruvbox"
		case "solarized-dark":
			raw = "solarized"
		case "neon":
			raw = "cyberpunk"
		}
		for _, t := range registry {
			if t.ID == raw {
				return t
			}
		}
	}
	return registry[0]
}

// NextTheme returns the theme immediately following the specified theme in the registry cycle.
func NextTheme(currentID string) Theme {
	normalized := strings.ToLower(strings.TrimSpace(currentID))
	switch normalized {
	case "pink", "default":
		normalized = "termdeck"
	case "tokyonight":
		normalized = "tokyo-night"
	case "catppuccin-mocha", "mocha":
		normalized = "catppuccin"
	case "gruvbox-dark":
		normalized = "gruvbox"
	case "solarized-dark":
		normalized = "solarized"
	case "neon":
		normalized = "cyberpunk"
	}

	curIdx := 0
	for i, t := range registry {
		if t.ID == normalized {
			curIdx = i
			break
		}
	}
	nextIdx := (curIdx + 1) % len(registry)
	return registry[nextIdx]
}
