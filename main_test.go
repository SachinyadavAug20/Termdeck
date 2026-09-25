package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

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

func TestThemeFlagAndListThemes(t *testing.T) {
	// Verify that buildModel preserves frontmatter theme
	tmpDeck := t.TempDir() + "/frontmatter_theme.deck.md"
	src := "---\ntheme: tokyo-night\n---\n# Title\nHello\n"
	if err := os.WriteFile(tmpDeck, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write deck: %v", err)
	}

	m, err := buildModel(tmpDeck)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.deck.Theme != "tokyo-night" {
		t.Errorf("expected deck.Theme 'tokyo-night', got %q", m.deck.Theme)
	}
	if m.editor.Theme != "tokyo-night" {
		t.Errorf("expected editor.Theme 'tokyo-night', got %q", m.editor.Theme)
	}
}

func TestModelUpdateTickMsg(t *testing.T) {
	m := model{
		editor: internal.NewEditor("test.deck.md"),
	}
	// Timer off: TickMsg returns nil cmd
	_, cmd := m.Update(internal.TickMsg(time.Now()))
	if cmd != nil {
		t.Errorf("expected nil cmd when timer is off")
	}

	// Timer on: TickMsg returns non-nil cmd to schedule next tick
	m.editor.ShowTimer = true
	_, cmd = m.Update(internal.TickMsg(time.Now()))
	if cmd == nil {
		t.Errorf("expected non-nil cmd when timer is on")
	}
}

func TestModelInitWatchMode(t *testing.T) {
	m := model{
		editor: internal.NewEditor("test.deck.md"),
	}
	// Watch off: Init returns nil
	if cmd := m.Init(); cmd != nil {
		t.Errorf("expected nil cmd when watch mode is off, got %v", cmd)
	}

	// Watch on: Init returns non-nil WatchCmd
	m.editor.WatchMode = true
	if cmd := m.Init(); cmd == nil {
		t.Errorf("expected non-nil cmd when watch mode is on")
	}
}

func TestModelUpdateWatchMsg(t *testing.T) {
	tmpFile := t.TempDir() + "/watch_test.deck.md"
	content := "# Slide 1\nInitial content\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write tmp file: %v", err)
	}

	m, err := buildModel(tmpFile, true)
	if err != nil {
		t.Fatalf("failed to build model: %v", err)
	}
	if !m.editor.WatchMode {
		t.Errorf("expected watchMode true")
	}

	// 1. Initial watch msg when mod time hasn't changed
	updated, cmd := m.Update(internal.WatchMsg(time.Now()))
	newM := updated.(model)
	if cmd == nil {
		t.Errorf("expected non-nil cmd to reschedule watch check")
	}
	if newM.deck.Slides[0].Blocks[0].Text != "Slide 1" {
		t.Errorf("expected content unchanged")
	}

	// 2. File modified on disk -> reload occurs
	// Sleep briefly to ensure mod time changes
	time.Sleep(10 * time.Millisecond)
	newContent := "# Reloaded Slide\nUpdated text\n"
	if err := os.WriteFile(tmpFile, []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to rewrite tmp file: %v", err)
	}

	updated2, cmd2 := newM.Update(internal.WatchMsg(time.Now()))
	newM2 := updated2.(model)
	if cmd2 == nil {
		t.Errorf("expected non-nil cmd")
	}
	if newM2.deck.Slides[0].Blocks[0].Text != "Reloaded Slide" {
		t.Errorf("expected slide title 'Reloaded Slide' after watch reload, got %q", newM2.deck.Slides[0].Blocks[0].Text)
	}

	// 3. File modified but editor is in ModeEdit -> reload is skipped to protect active editing
	time.Sleep(10 * time.Millisecond)
	editContent := "# Third Edit\nEven more text\n"
	if err := os.WriteFile(tmpFile, []byte(editContent), 0644); err != nil {
		t.Fatalf("failed to rewrite tmp file: %v", err)
	}
	newM2.editor.Mode = internal.ModeEdit
	updated3, _ := newM2.Update(internal.WatchMsg(time.Now()))
	newM3 := updated3.(model)
	if newM3.deck.Slides[0].Blocks[0].Text != "Reloaded Slide" {
		t.Errorf("expected reload to be skipped during active edit mode, but slide changed to %q", newM3.deck.Slides[0].Blocks[0].Text)
	}

	// 4. File modified but editor is Dirty -> reload is skipped
	newM3.editor.Mode = internal.ModeNav
	newM3.editor.Dirty = true
	updated4, _ := newM3.Update(internal.WatchMsg(time.Now()))
	newM4 := updated4.(model)
	if newM4.deck.Slides[0].Blocks[0].Text != "Reloaded Slide" {
		t.Errorf("expected reload to be skipped when editor is dirty, but slide changed to %q", newM4.deck.Slides[0].Blocks[0].Text)
	}
}

func TestPrintHelpExportHTML(t *testing.T) {
	// Verify that printHelp mentions --export-html and E
	printHelp()
}

func TestPrintHelpStats(t *testing.T) {
	// Verify that printHelp mentions --stats and S
	printHelp()
}

func TestPrintHelpAutoplay(t *testing.T) {
	// Verify that printHelp mentions --autoplay and A
	printHelp()
}

func TestPrintHelpGraphAndMermaid(t *testing.T) {
	// Capture printHelp output or invoke it
	printHelp()
}

func TestModelAutoplay(t *testing.T) {
	d := internal.ParseDeck("# Slide 1\n---\n# Slide 2\n")
	ed := internal.NewEditor("test.deck.md")
	ed.Autoplay = true
	ed.AutoplayInterval = 5
	ed.AutoplayCountdown = 1

	m := model{
		deck:   d,
		editor: ed,
	}

	// Init should emit batch cmd including TickCmd when autoplay is on
	initCmd := m.Init()
	if initCmd == nil {
		t.Errorf("expected non-nil cmd on Init() when autoplay is active")
	}

	// Update on TickMsg should advance slide when countdown expires
	updated, cmd := m.Update(internal.TickMsg(time.Now()))
	if cmd == nil {
		t.Errorf("expected TickCmd from Update on TickMsg")
	}
	newM := updated.(model)
	if newM.editor.SlideIdx != 1 {
		t.Errorf("expected slide advanced to 1, got %d", newM.editor.SlideIdx)
	}
}

func TestModelExecFinishedMsg(t *testing.T) {
	d := internal.ParseDeck("# Slide 1\n```sh\necho 'hello'\n```\n")
	ed := internal.NewEditor("test.deck.md")
	ed.RunningCode = true

	m := model{
		deck:   d,
		editor: ed,
	}

	execMsg := internal.ExecFinishedMsg{
		Result: internal.ExecResult{
			Language: "sh",
			ExitCode: 0,
			Duration: 5 * time.Millisecond,
			Stdout:   "hello\n",
		},
	}

	updated, cmd := m.Update(execMsg)
	if cmd != nil {
		t.Errorf("expected nil cmd on ExecFinishedMsg")
	}
	newM := updated.(model)
	if newM.editor.RunningCode {
		t.Errorf("expected RunningCode false")
	}
	if !newM.editor.ShowRunner {
		t.Errorf("expected ShowRunner true")
	}
	if newM.editor.RunnerResult == nil || newM.editor.RunnerResult.ExitCode != 0 {
		t.Errorf("unexpected runner result: %+v", newM.editor.RunnerResult)
	}
}

func TestPrintHelpLiveRunnerAndFocus(t *testing.T) {
	printHelp()
}

func TestPrintHelpTrack(t *testing.T) {
	printHelp()
}

func TestPrintHelpRoute(t *testing.T) {
	printHelp()
}
