package internal

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCalculateStats(t *testing.T) {
	// Empty deck
	emptyDeck := Deck{}
	emptyStats := CalculateStats(&emptyDeck, 0)
	if emptyStats.TotalSlides != 0 {
		t.Errorf("expected 0 slides, got %d", emptyStats.TotalSlides)
	}

	markdown := `---
title: Stats Test Deck
theme: dracula
---

# Architecture Overview

We are modernizing the distributed backend service.

` + "```go\nfunc main() {\n\tprintln(\"hello\")\n}\n```" + `

---

## Migration Progress

| Service | Status | Version |
|---|---|---|
| Auth | Migrated | v2.1 |
| Payment | In Progress | v1.9 |

> [!TIP]
> Keep database transactions short.

- [x] Database schema migrated
- [x] Redis cluster deployed
- [ ] API gateway switched

---

## Media & Visuals

![Architecture](diagram.png)

***

::notes
Remember to emphasize backward compatibility and zero downtime.
`

	deck := ParseDeck(markdown)
	stats := CalculateStats(&deck, 1)

	if stats.TotalSlides != 3 {
		t.Fatalf("expected 3 slides, got %d", stats.TotalSlides)
	}
	if stats.CurrentSlideIdx != 1 {
		t.Errorf("expected active slide 1, got %d", stats.CurrentSlideIdx)
	}
	if stats.CurrentSlideTitle != "Migration Progress" {
		t.Errorf("expected title 'Migration Progress', got %q", stats.CurrentSlideTitle)
	}
	if stats.CodeBlocks != 1 {
		t.Errorf("expected 1 code block, got %d", stats.CodeBlocks)
	}
	if stats.CodeLines != 3 {
		t.Errorf("expected 3 code lines, got %d", stats.CodeLines)
	}
	if stats.TableBlocks != 1 {
		t.Errorf("expected 1 table block, got %d", stats.TableBlocks)
	}
	if stats.CalloutBlocks != 1 {
		t.Errorf("expected 1 callout block, got %d", stats.CalloutBlocks)
	}
	if stats.ImageBlocks != 1 {
		t.Errorf("expected 1 image block, got %d", stats.ImageBlocks)
	}
	if stats.DividerBlocks != 1 {
		t.Errorf("expected 1 divider block, got %d", stats.DividerBlocks)
	}
	if stats.TaskTotal != 3 {
		t.Errorf("expected 3 tasks, got %d", stats.TaskTotal)
	}
	if stats.TaskCompleted != 2 {
		t.Errorf("expected 2 completed tasks, got %d", stats.TaskCompleted)
	}
	expectedPercent := (2.0 / 3.0) * 100.0
	if stats.TaskPercent < expectedPercent-0.1 || stats.TaskPercent > expectedPercent+0.1 {
		t.Errorf("expected ~66.7%% task percent, got %f", stats.TaskPercent)
	}
	if stats.TotalWords == 0 {
		t.Errorf("expected non-zero total words")
	}
	if stats.NotesWords == 0 {
		t.Errorf("expected non-zero notes words, got %d", stats.NotesWords)
	}
	if stats.EstDurationMin < 0 || stats.EstDurationSec < 0 {
		t.Errorf("invalid duration: %d min %d sec", stats.EstDurationMin, stats.EstDurationSec)
	}

	// Boundary clamping on active slide index
	outStats := CalculateStats(&deck, 999)
	if outStats.CurrentSlideIdx != 2 {
		t.Errorf("expected clamped slide index 2, got %d", outStats.CurrentSlideIdx)
	}
	negStats := CalculateStats(&deck, -5)
	if negStats.CurrentSlideIdx != 0 {
		t.Errorf("expected clamped slide index 0, got %d", negStats.CurrentSlideIdx)
	}
}

func TestRenderProgressBar(t *testing.T) {
	bar0 := RenderProgressBar(0, 10, "#00ff00", "#333333")
	if !strings.Contains(bar0, "[") || !strings.Contains(bar0, "]") {
		t.Errorf("expected brackets in progress bar: %s", bar0)
	}

	bar100 := RenderProgressBar(100, 10, "#00ff00", "#333333")
	if !strings.Contains(bar100, "█") {
		t.Errorf("expected filled blocks in 100%% bar: %s", bar100)
	}

	bar50 := RenderProgressBar(50, 10, "#00ff00", "#333333")
	if !strings.Contains(bar50, "█") || !strings.Contains(bar50, "░") {
		t.Errorf("expected both filled and empty blocks in 50%% bar: %s", bar50)
	}

	// Clamping
	barNeg := RenderProgressBar(-20, 0, "#00ff00", "#333333")
	if !strings.Contains(barNeg, "[") {
		t.Errorf("expected valid bar on negative percent: %s", barNeg)
	}
	barOver := RenderProgressBar(150, 10, "#00ff00", "#333333")
	if !strings.Contains(barOver, "█") {
		t.Errorf("expected filled blocks on over 100%% bar: %s", barOver)
	}
}

func TestFormatStatsCLI(t *testing.T) {
	stats := DeckStats{
		TotalSlides:       4,
		TotalBlocks:       12,
		VisibleBlocks:     11,
		TotalWords:        260,
		NotesWords:        40,
		EstDurationMin:    2,
		EstDurationSec:    18,
		CodeBlocks:        2,
		CodeLines:         15,
		TableBlocks:       1,
		CalloutBlocks:     1,
		ImageBlocks:       1,
		DividerBlocks:     1,
		TaskTotal:         5,
		TaskCompleted:     4,
		TaskPercent:       80.0,
		CurrentSlideIdx:   0,
		CurrentSlideTitle: "Intro",
	}

	theme := ResolveTheme("dracula")
	cliOutput := FormatStatsCLI(stats, theme)

	if !strings.Contains(cliOutput, "Presentation Statistics & Deck Metrics") {
		t.Errorf("expected title in CLI stats: %s", cliOutput)
	}
	if !strings.Contains(cliOutput, "DECK OVERVIEW") {
		t.Errorf("expected DECK OVERVIEW in CLI stats: %s", cliOutput)
	}
	if !strings.Contains(cliOutput, "TECHNICAL DENSITY") {
		t.Errorf("expected TECHNICAL DENSITY in CLI stats: %s", cliOutput)
	}
	if !strings.Contains(cliOutput, "SPRINT TASK COMPLETION") {
		t.Errorf("expected SPRINT TASK COMPLETION in CLI stats: %s", cliOutput)
	}
	if !strings.Contains(cliOutput, "80.0%") {
		t.Errorf("expected 80.0%% in CLI stats: %s", cliOutput)
	}
	if !strings.Contains(cliOutput, "FOCUSED SLIDE") {
		t.Errorf("expected FOCUSED SLIDE in CLI stats: %s", cliOutput)
	}

	// No tasks case
	statsNoTasks := DeckStats{TotalSlides: 1}
	cliNoTasks := FormatStatsCLI(statsNoTasks, theme)
	if !strings.Contains(cliNoTasks, "no task checklists found") {
		t.Errorf("expected 'no task checklists found', got: %s", cliNoTasks)
	}
}

func TestEditorStatsModalAndKeyNav(t *testing.T) {
	d := ParseDeck("# Test Deck\n\nContent")
	ed := NewEditor("test.deck.md")

	// Press S to open stats modal
	ed.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")}, &d)
	if !ed.ShowStats {
		t.Fatalf("expected ShowStats to be true after pressing 'S'")
	}
	if !strings.Contains(ed.Message, "statistics") {
		t.Errorf("expected statistics status message, got: %q", ed.Message)
	}

	// Press Esc to dismiss
	ed.HandleKey(tea.KeyMsg{Type: tea.KeyEsc}, &d)
	if ed.ShowStats {
		t.Fatalf("expected ShowStats to be false after pressing Esc")
	}

	// Press S again to open
	ed.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")}, &d)
	if !ed.ShowStats {
		t.Fatalf("expected ShowStats to be true after second 'S'")
	}

	// Press q to dismiss while in ShowStats
	ed.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}, &d)
	if ed.ShowStats {
		t.Fatalf("expected ShowStats to be false after pressing 'q'")
	}

	// Press S then S to toggle off
	ed.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")}, &d)
	ed.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("S")}, &d)
	if ed.ShowStats {
		t.Fatalf("expected ShowStats to be false after toggling 'S'")
	}
}

func TestRenderStatsModalView(t *testing.T) {
	d := ParseDeck("# Modern System Architecture\n\n- [x] Task 1\n- [ ] Task 2\n\n```go\nfunc Run() {}\n```")
	ed := NewEditor("test.deck.md")
	ed.ShowStats = true

	modal := renderStatsModal(d, ed, 80, 24)
	if !strings.Contains(modal, "Presentation Statistics") {
		t.Errorf("expected title in modal: %s", modal)
	}
	if !strings.Contains(modal, "DECK METRICS") {
		t.Errorf("expected DECK METRICS in modal: %s", modal)
	}
	if !strings.Contains(modal, "TECHNICAL DENSITY") {
		t.Errorf("expected TECHNICAL DENSITY in modal: %s", modal)
	}
	if !strings.Contains(modal, "SPRINT TASK COMPLETION") {
		t.Errorf("expected SPRINT TASK COMPLETION in modal: %s", modal)
	}

	// Check top-level View() renders modal when ShowStats is true
	fullView := View(d, ed, 80, 24)
	if !strings.Contains(fullView, "Presentation Statistics") {
		t.Errorf("expected View() to render stats modal when ShowStats is true")
	}
}
