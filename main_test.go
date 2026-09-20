package main

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"deck/internal"
	tea "github.com/charmbracelet/bubbletea"
)

var reANSI = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return reANSI.ReplaceAllString(s, "")
}

func TestModelInit(t *testing.T) {
	m := model{}
	if cmd := m.Init(); cmd != nil {
		t.Errorf("expected Init to return nil cmd, got %v", cmd)
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	d := internal.Deck{
		Slides: []internal.Slide{
			{Blocks: []internal.Block{{Kind: internal.BlockParagraph, Text: "Test"}}},
		},
	}
	ed := internal.NewEditor("test.deck.md")
	m := model{
		deck:   d,
		editor: ed,
	}

	// First WindowSizeMsg should record dimensions and trigger ClearScreen
	msg := tea.WindowSizeMsg{Width: 100, Height: 40}
	updated, cmd := m.Update(msg)
	newModel, ok := updated.(model)
	if !ok {
		t.Fatalf("expected updated model of type model")
	}
	if newModel.width != 100 || newModel.height != 40 {
		t.Errorf("expected dimensions 100x40, got %dx%d", newModel.width, newModel.height)
	}
	if !newModel.resized {
		t.Errorf("expected resized to be true")
	}
	if cmd == nil {
		t.Errorf("expected ClearScreen cmd on first resize")
	}

	// Second WindowSizeMsg should not trigger ClearScreen again
	msg2 := tea.WindowSizeMsg{Width: 120, Height: 50}
	updated2, cmd2 := newModel.Update(msg2)
	newModel2 := updated2.(model)
	if newModel2.width != 120 || newModel2.height != 50 {
		t.Errorf("expected dimensions 120x50, got %dx%d", newModel2.width, newModel2.height)
	}
	if cmd2 != nil {
		t.Errorf("expected nil cmd on subsequent resize, got %v", cmd2)
	}
}

func TestModelUpdateKeyMsg(t *testing.T) {
	d := internal.Deck{
		Slides: []internal.Slide{
			{
				Blocks: []internal.Block{
					{Kind: internal.BlockHeading, Level: 1, Text: "Title"},
					{Kind: internal.BlockParagraph, Text: "Para"},
				},
			},
		},
	}
	ed := internal.NewEditor("test.deck.md")
	m := model{
		deck:   d,
		editor: ed,
	}

	// Down arrow key
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(keyMsg)
	newModel := updated.(model)
	if newModel.editor.BlockIdx != 1 {
		t.Errorf("expected editor blockIdx 1 after KeyDown, got %d", newModel.editor.BlockIdx)
	}
}

func TestModelView(t *testing.T) {
	d := internal.Deck{
		Slides: []internal.Slide{
			{Blocks: []internal.Block{{Kind: internal.BlockHeading, Level: 1, Text: "My Slide"}}},
		},
	}
	ed := internal.NewEditor("test.deck.md")
	m := model{
		deck:   d,
		editor: ed,
		width:  80,
		height: 24,
	}

	cleanView := stripANSI(m.View())
	if !strings.Contains(cleanView, "My Slide") {
		t.Errorf("expected view to contain 'My Slide', got %q", cleanView)
	}
	if !strings.Contains(cleanView, "slide 1/1") {
		t.Errorf("expected view to contain status bar with 'slide 1/1', got %q", cleanView)
	}
}

func TestModelUpdateQuitKey(t *testing.T) {
	d := internal.Deck{
		Slides: []internal.Slide{
			{Blocks: []internal.Block{{Kind: internal.BlockParagraph, Text: "Test"}}},
		},
	}
	ed := internal.NewEditor("test.deck.md")
	m := model{
		deck:   d,
		editor: ed,
	}

	quitKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(quitKey)
	if cmd == nil {
		t.Errorf("expected tea.Quit cmd when pressing 'q', got nil")
	}
}

func TestBuildModel(t *testing.T) {
	// 1. Non-existent file
	_, err := buildModel("non_existent_deck_file_98765.md")
	if err == nil {
		t.Errorf("expected error for non-existent file, got nil")
	}

	// 2. File with no slides
	tmpEmpty := t.TempDir() + "/empty.deck.md"
	if err := os.WriteFile(tmpEmpty, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write empty deck: %v", err)
	}
	_, err = buildModel(tmpEmpty)
	if err == nil || !strings.Contains(err.Error(), "no slides found") {
		t.Errorf("expected 'no slides found' error, got %v", err)
	}

	// 3. Valid deck file
	tmpValid := t.TempDir() + "/valid.deck.md"
	content := "---\ntitle: Sample\n---\n# First Slide\nHello World\n"
	if err := os.WriteFile(tmpValid, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write valid deck: %v", err)
	}
	m, err := buildModel(tmpValid)
	if err != nil {
		t.Fatalf("unexpected error building valid model: %v", err)
	}
	if len(m.deck.Slides) != 1 {
		t.Errorf("expected 1 slide, got %d", len(m.deck.Slides))
	}
	if m.editor.FilePath != tmpValid {
		t.Errorf("expected editor file path %q, got %q", tmpValid, m.editor.FilePath)
	}
}

func TestPrintHelp(t *testing.T) {
	// Verify printHelp runs without panic
	printHelp()
}
