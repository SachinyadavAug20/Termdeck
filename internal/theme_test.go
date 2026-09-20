package internal

import (
	"strings"
	"testing"
)

func TestAvailableThemes(t *testing.T) {
	themes := AvailableThemes()
	if len(themes) < 9 {
		t.Fatalf("expected at least 9 curated themes, got %d", len(themes))
	}

	seen := make(map[string]bool)
	for _, th := range themes {
		if th.ID == "" {
			t.Errorf("theme ID must not be empty: %+v", th)
		}
		if seen[th.ID] {
			t.Errorf("duplicate theme ID: %s", th.ID)
		}
		seen[th.ID] = true

		if th.Accent == "" || th.Name == "" {
			t.Errorf("theme %s missing required Accent or Name", th.ID)
		}
	}

	ids := ThemeIDs()
	if len(ids) != len(themes) {
		t.Errorf("ThemeIDs length %d does not match AvailableThemes %d", len(ids), len(themes))
	}
}

func TestResolveTheme(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"dracula", "dracula"},
		{"DRACULA", "dracula"},
		{"tokyo-night", "tokyo-night"},
		{"tokyonight", "tokyo-night"},
		{"catppuccin", "catppuccin"},
		{"catppuccin-mocha", "catppuccin"},
		{"mocha", "catppuccin"},
		{"nord", "nord"},
		{"gruvbox", "gruvbox"},
		{"gruvbox-dark", "gruvbox"},
		{"monokai", "monokai"},
		{"solarized", "solarized"},
		{"solarized-dark", "solarized"},
		{"cyberpunk", "cyberpunk"},
		{"neon", "cyberpunk"},
		{"pink", "termdeck"},
		{"default", "termdeck"},
		{"termdeck", "termdeck"},
		{"non_existent_theme", "termdeck"}, // Fallback
		{"", "termdeck"},                   // Empty fallback
	}

	for _, tc := range tests {
		th := ResolveTheme(tc.input)
		if th.ID != tc.expected {
			t.Errorf("ResolveTheme(%q) = %s, expected %s", tc.input, th.ID, tc.expected)
		}
	}
}

func TestCustomHexTheme(t *testing.T) {
	customHex := "#ff5722"
	th := ResolveTheme(customHex)
	if th.ID != "custom" {
		t.Errorf("expected ID 'custom', got %q", th.ID)
	}
	if th.Accent != customHex {
		t.Errorf("expected accent %q, got %q", customHex, th.Accent)
	}
}

func TestNextThemeCycle(t *testing.T) {
	th1 := ResolveTheme("termdeck")
	next1 := NextTheme(th1.ID)
	if next1.ID == th1.ID {
		t.Errorf("NextTheme should advance to next theme, got same %s", next1.ID)
	}

	// Cycling len(registry) times should return to the original theme
	cur := th1.ID
	count := len(AvailableThemes())
	for i := 0; i < count; i++ {
		cur = NextTheme(cur).ID
	}
	if cur != th1.ID {
		t.Errorf("cycling %d times did not return to original theme: got %s, expected %s", count, cur, th1.ID)
	}
}

func TestThemeAppliedInView(t *testing.T) {
	d := Deck{
		Theme: "tokyo-night",
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Tokyo Night Slide"},
					{Kind: BlockParagraph, Text: "Some `inline code` text"},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")
	ed.Theme = "tokyo-night"

	out := View(d, ed, 80, 24)
	clean := stripANSI(out)
	if !strings.Contains(clean, "Tokyo Night Slide") {
		t.Errorf("expected view to contain slide content, got %q", clean)
	}

	// Check active theme in Theme Engine
	cur := CurrentTheme()
	if cur.ID != "tokyo-night" {
		t.Errorf("expected currentTheme to be 'tokyo-night', got %s", cur.ID)
	}
}

func TestThemeFrontmatterRoundTrip(t *testing.T) {
	deckSrc := "---\ntitle: My Deck\ntheme: dracula\n---\n\n# Slide 1\nHello\n"
	parsed := ParseDeck(deckSrc)
	if parsed.Theme != "dracula" {
		t.Errorf("expected parsed.Theme 'dracula', got %q", parsed.Theme)
	}

	serialized := SerializeDeck(parsed)
	if !strings.Contains(serialized, "theme: dracula") {
		t.Errorf("expected serialized deck to contain 'theme: dracula', got:\n%s", serialized)
	}

	reparsed := ParseDeck(serialized)
	if reparsed.Theme != "dracula" {
		t.Errorf("expected reparsed theme 'dracula', got %q", reparsed.Theme)
	}
}
