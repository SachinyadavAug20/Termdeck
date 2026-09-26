package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func sampleDeck() Deck {
	return Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Slide 1"},
					{Kind: BlockParagraph, Text: "First paragraph"},
					{Kind: BlockImage, Src: "sample.png"},
				},
			},
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 2, Text: "Slide 2"},
				},
			},
		},
	}
}

func TestEditorNavigation(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	if ed.BlockIdx != 0 {
		t.Fatalf("expected initial blockIdx 0, got %d", ed.BlockIdx)
	}

	ed.MoveDown(&d)
	if ed.BlockIdx != 1 {
		t.Fatalf("expected blockIdx 1 after MoveDown, got %d", ed.BlockIdx)
	}

	ed.MoveDown(&d)
	if ed.BlockIdx != 2 {
		t.Fatalf("expected blockIdx 2 after MoveDown, got %d", ed.BlockIdx)
	}

	// Should not go past last block
	ed.MoveDown(&d)
	if ed.BlockIdx != 2 {
		t.Fatalf("expected blockIdx to stay at 2, got %d", ed.BlockIdx)
	}

	ed.MoveUp(&d)
	if ed.BlockIdx != 1 {
		t.Fatalf("expected blockIdx 1 after MoveUp, got %d", ed.BlockIdx)
	}
}

func TestEditorBlockEditing(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	// Navigate to paragraph
	ed.MoveDown(&d)
	ed.EnterEdit(&d)

	if ed.Mode != ModeEdit {
		t.Fatalf("expected ModeEdit, got %v", ed.Mode)
	}
	if ed.Draft != "First paragraph" {
		t.Fatalf("expected draft 'First paragraph', got %q", ed.Draft)
	}

	// Edit draft and commit
	ed.Draft = "Updated paragraph"
	ed.ExitEdit(&d)

	if ed.Mode != ModeNav {
		t.Fatalf("expected ModeNav after exit, got %v", ed.Mode)
	}
	if d.Slides[0].Blocks[1].Text != "Updated paragraph" {
		t.Fatalf("expected block text updated, got %q", d.Slides[0].Blocks[1].Text)
	}
	if !ed.Dirty {
		t.Errorf("expected dirty flag to be true")
	}

	// Test image editing
	ed.MoveDown(&d)
	ed.EnterEdit(&d)
	if ed.Draft != "sample.png" {
		t.Fatalf("expected image draft 'sample.png', got %q", ed.Draft)
	}
	ed.Draft = "new_image.png"
	ed.ExitEdit(&d)
	if d.Slides[0].Blocks[2].Src != "new_image.png" {
		t.Fatalf("expected image src 'new_image.png', got %q", d.Slides[0].Blocks[2].Src)
	}
}

func TestEditorAddDeleteBlock(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	initialBlocks := len(d.Slides[0].Blocks)
	ed.AddBlock(&d)

	if len(d.Slides[0].Blocks) != initialBlocks+1 {
		t.Fatalf("expected %d blocks after AddBlock, got %d", initialBlocks+1, len(d.Slides[0].Blocks))
	}

	ed.DeleteBlock(&d)
	if len(d.Slides[0].Blocks) != initialBlocks {
		t.Fatalf("expected %d blocks after DeleteBlock, got %d", initialBlocks, len(d.Slides[0].Blocks))
	}
}

func TestEditorUndoRedo(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	originalText := d.Slides[0].Blocks[0].Text
	ed.EnterEdit(&d)
	ed.Draft = "Modified Title"
	ed.ExitEdit(&d)

	if d.Slides[0].Blocks[0].Text != "Modified Title" {
		t.Fatalf("expected 'Modified Title', got %q", d.Slides[0].Blocks[0].Text)
	}

	if len(ed.UndoStack) == 0 {
		t.Fatalf("expected non-empty undo stack")
	}

	// Undo
	currentState := SerializeDeck(d)
	ed.RedoStack = append(ed.RedoStack, currentState)
	lastUndo := ed.UndoStack[len(ed.UndoStack)-1]
	ed.UndoStack = ed.UndoStack[:len(ed.UndoStack)-1]
	d = ParseDeck(lastUndo)

	if d.Slides[0].Blocks[0].Text != originalText {
		t.Fatalf("expected undone text %q, got %q", originalText, d.Slides[0].Blocks[0].Text)
	}
}

func TestEditorToggleAlign(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	// Initial default should cycle from center to right
	ed.ToggleAlign(&d)
	if d.Slides[0].Align != AlignRight {
		t.Errorf("expected align right after first toggle, got %q", d.Slides[0].Align)
	}

	// Cycle to left
	ed.ToggleAlign(&d)
	if d.Slides[0].Align != AlignLeft {
		t.Errorf("expected align left after second toggle, got %q", d.Slides[0].Align)
	}

	// Cycle back to center
	ed.ToggleAlign(&d)
	if d.Slides[0].Align != AlignCenter {
		t.Errorf("expected align center after third toggle, got %q", d.Slides[0].Align)
	}
}

func sendTestKey(ed *Editor, d *Deck, keyStr string) tea.Cmd {
	var msg tea.KeyMsg
	switch keyStr {
	case "ctrl+c":
		msg = tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+n":
		msg = tea.KeyMsg{Type: tea.KeyCtrlN}
	case "ctrl+d":
		msg = tea.KeyMsg{Type: tea.KeyCtrlD}
	case "ctrl+k":
		msg = tea.KeyMsg{Type: tea.KeyCtrlK}
	case "ctrl+j":
		msg = tea.KeyMsg{Type: tea.KeyCtrlJ}
	case "ctrl+s":
		msg = tea.KeyMsg{Type: tea.KeyCtrlS}
	case "ctrl+r":
		msg = tea.KeyMsg{Type: tea.KeyCtrlR}
	case "ctrl+a":
		msg = tea.KeyMsg{Type: tea.KeyCtrlA}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "backspace":
		msg = tea.KeyMsg{Type: tea.KeyBackspace}
	case "delete":
		msg = tea.KeyMsg{Type: tea.KeyDelete}
	case "left":
		msg = tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		msg = tea.KeyMsg{Type: tea.KeyRight}
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "home":
		msg = tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		msg = tea.KeyMsg{Type: tea.KeyEnd}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	case "pgup":
		msg = tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		msg = tea.KeyMsg{Type: tea.KeyPgDown}
	case "space":
		msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keyStr)}
	}
	return ed.HandleKey(msg, d)
}

func TestEditorHandleKeyNav(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	// 1. Quit keys
	if cmd := sendTestKey(&ed, &d, "q"); cmd == nil {
		t.Errorf("expected tea.Quit on 'q'")
	}
	if cmd := sendTestKey(&ed, &d, "ctrl+c"); cmd == nil {
		t.Errorf("expected tea.Quit on 'ctrl+c'")
	}

	// 2. Slide navigation
	sendTestKey(&ed, &d, "right")
	if ed.SlideIdx != 1 {
		t.Errorf("expected slideIdx 1 after 'right', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "left")
	if ed.SlideIdx != 0 {
		t.Errorf("expected slideIdx 0 after 'left', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "l")
	if ed.SlideIdx != 1 {
		t.Errorf("expected slideIdx 1 after 'l', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "h")
	if ed.SlideIdx != 0 {
		t.Errorf("expected slideIdx 0 after 'h', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, " ")
	if ed.SlideIdx != 1 {
		t.Errorf("expected slideIdx 1 after ' ', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "backspace")
	if ed.SlideIdx != 0 {
		t.Errorf("expected slideIdx 0 after 'backspace', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "pgdown")
	if ed.SlideIdx != 1 {
		t.Errorf("expected slideIdx 1 after 'pgdown', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "pgup")
	if ed.SlideIdx != 0 {
		t.Errorf("expected slideIdx 0 after 'pgup', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "enter")
	if ed.SlideIdx != 1 {
		t.Errorf("expected slideIdx 1 after 'enter', got %d", ed.SlideIdx)
	}

	// 3. Jump to first / last slide
	sendTestKey(&ed, &d, "g")
	if ed.SlideIdx != 0 {
		t.Errorf("expected slideIdx 0 after 'g', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "G")
	if ed.SlideIdx != 1 {
		t.Errorf("expected slideIdx 1 after 'G', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "g")

	// 4. Block movement
	sendTestKey(&ed, &d, "j")
	if ed.BlockIdx != 1 {
		t.Errorf("expected blockIdx 1 after 'j', got %d", ed.BlockIdx)
	}
	sendTestKey(&ed, &d, "k")
	if ed.BlockIdx != 0 {
		t.Errorf("expected blockIdx 0 after 'k', got %d", ed.BlockIdx)
	}
	sendTestKey(&ed, &d, "down")
	if ed.BlockIdx != 1 {
		t.Errorf("expected blockIdx 1 after 'down', got %d", ed.BlockIdx)
	}
	sendTestKey(&ed, &d, "up")
	if ed.BlockIdx != 0 {
		t.Errorf("expected blockIdx 0 after 'up', got %d", ed.BlockIdx)
	}

	// 5. Block reordering
	sendTestKey(&ed, &d, "ctrl+j")
	if ed.BlockIdx != 1 {
		t.Errorf("expected blockIdx 1 after 'ctrl+j', got %d", ed.BlockIdx)
	}
	sendTestKey(&ed, &d, "ctrl+k")
	if ed.BlockIdx != 0 {
		t.Errorf("expected blockIdx 0 after 'ctrl+k', got %d", ed.BlockIdx)
	}

	// 6. Block addition / deletion
	initialBlocks := len(d.Slides[0].Blocks)
	sendTestKey(&ed, &d, "ctrl+n")
	if len(d.Slides[0].Blocks) != initialBlocks+1 {
		t.Errorf("expected %d blocks after 'ctrl+n', got %d", initialBlocks+1, len(d.Slides[0].Blocks))
	}
	sendTestKey(&ed, &d, "ctrl+d")
	if len(d.Slides[0].Blocks) != initialBlocks {
		t.Errorf("expected %d blocks after 'ctrl+d', got %d", initialBlocks, len(d.Slides[0].Blocks))
	}

	// 7. Slide addition / deletion
	initialSlides := len(d.Slides)
	sendTestKey(&ed, &d, "ctrl+N")
	if len(d.Slides) != initialSlides+1 {
		t.Errorf("expected %d slides after 'ctrl+N', got %d", initialSlides+1, len(d.Slides))
	}
	sendTestKey(&ed, &d, "ctrl+D")
	if len(d.Slides) != initialSlides {
		t.Errorf("expected %d slides after 'ctrl+D', got %d", initialSlides, len(d.Slides))
	}

	// 8. Alignment toggling via keys
	ed.SlideIdx = 0
	ed.BlockIdx = 0
	sendTestKey(&ed, &d, "tab")
	if d.Slides[0].Align != AlignRight {
		t.Errorf("expected slide align right after 'tab', got %s", d.Slides[0].Align)
	}
	sendTestKey(&ed, &d, "ctrl+a")
	if d.Slides[0].Align != AlignLeft {
		t.Errorf("expected slide align left after 'ctrl+a', got %s", d.Slides[0].Align)
	}

	// 9. Undo and redo
	sendTestKey(&ed, &d, "u")
	if ed.Message != "undo" {
		t.Errorf("expected message 'undo', got %q", ed.Message)
	}
	sendTestKey(&ed, &d, "ctrl+r")
	if ed.Message != "redo" {
		t.Errorf("expected message 'redo', got %q", ed.Message)
	}

	// 10. Message clearing via esc
	ed.Message = "some status"
	sendTestKey(&ed, &d, "esc")
	if ed.Message != "" {
		t.Errorf("expected message cleared after 'esc', got %q", ed.Message)
	}

	// 11. Image opening via 'p'
	ed.SlideIdx = 0
	ed.BlockIdx = 2 // sample.png
	sendTestKey(&ed, &d, "p")
	if !strings.Contains(ed.Message, "image not found") && !strings.Contains(ed.Message, "opened") {
		t.Errorf("expected image open message on 'p', got %q", ed.Message)
	}

	// 12. Entering edit mode
	ed.SlideIdx = 0
	ed.BlockIdx = 0
	for _, k := range []string{"i", "a", "I"} {
		ed.Mode = ModeNav
		sendTestKey(&ed, &d, k)
		if ed.Mode != ModeEdit {
			t.Errorf("expected ModeEdit after pressing %q, got %v", k, ed.Mode)
		}
	}

	// 13. Unhandled key does nothing
	ed.Mode = ModeNav
	cmd := sendTestKey(&ed, &d, "unmapped_key")
	if cmd != nil {
		t.Errorf("expected nil cmd for unmapped key, got %v", cmd)
	}
}

func TestEditorHandleKeyEdit(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	// Enter edit mode on paragraph
	ed.BlockIdx = 1
	ed.EnterEdit(&d)
	if ed.Draft != "First paragraph" {
		t.Fatalf("expected initial draft 'First paragraph', got %q", ed.Draft)
	}

	// 1. Character typing
	sendTestKey(&ed, &d, "!")
	if ed.Draft != "First paragraph!" {
		t.Errorf("expected draft with '!', got %q", ed.Draft)
	}

	// 2. Cursor navigation
	sendTestKey(&ed, &d, "left")
	if ed.CursorCol != len("First paragraph!")-1 {
		t.Errorf("expected cursor column decremented, got %d", ed.CursorCol)
	}
	sendTestKey(&ed, &d, "right")
	if ed.CursorCol != len("First paragraph!") {
		t.Errorf("expected cursor column incremented, got %d", ed.CursorCol)
	}
	sendTestKey(&ed, &d, "home")
	if ed.CursorCol != 0 {
		t.Errorf("expected cursor column 0 on home, got %d", ed.CursorCol)
	}
	sendTestKey(&ed, &d, "end")
	if ed.CursorCol != len(ed.Draft) {
		t.Errorf("expected cursor column at end on end, got %d", ed.CursorCol)
	}
	sendTestKey(&ed, &d, "0")
	if ed.CursorCol != 0 {
		t.Errorf("expected cursor column 0 on '0', got %d", ed.CursorCol)
	}
	sendTestKey(&ed, &d, "$")
	if ed.CursorCol != len(ed.Draft) {
		t.Errorf("expected cursor column at end on '$', got %d", ed.CursorCol)
	}

	// 3. Deletion keys
	// Backspace at end
	sendTestKey(&ed, &d, "backspace")
	if ed.Draft != "First paragraph" {
		t.Errorf("expected draft after backspace 'First paragraph', got %q", ed.Draft)
	}
	// Delete
	sendTestKey(&ed, &d, "home")
	sendTestKey(&ed, &d, "delete")
	if ed.Draft != "irst paragraph" {
		t.Errorf("expected draft after delete 'irst paragraph', got %q", ed.Draft)
	}

	// 4. Cancel edit with esc
	sendTestKey(&ed, &d, "esc")
	if ed.Mode != ModeNav {
		t.Errorf("expected ModeNav after esc, got %v", ed.Mode)
	}
	// Text should remain original "First paragraph"
	if d.Slides[0].Blocks[1].Text != "First paragraph" {
		t.Errorf("expected block text unchanged on cancel, got %q", d.Slides[0].Blocks[1].Text)
	}

	// 5. Cancel edit with ctrl+c
	ed.EnterEdit(&d)
	sendTestKey(&ed, &d, "X")
	sendTestKey(&ed, &d, "ctrl+c")
	if ed.Mode != ModeNav {
		t.Errorf("expected ModeNav after ctrl+c, got %v", ed.Mode)
	}

	// 6. Confirm edit with enter
	ed.EnterEdit(&d)
	ed.Draft = "Brand new text"
	sendTestKey(&ed, &d, "enter")
	if ed.Mode != ModeNav {
		t.Errorf("expected ModeNav after enter, got %v", ed.Mode)
	}
	if d.Slides[0].Blocks[1].Text != "Brand new text" {
		t.Errorf("expected block text updated to 'Brand new text', got %q", d.Slides[0].Blocks[1].Text)
	}
}

func TestEditorSaveAndError(t *testing.T) {
	d := sampleDeck()
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "saved.deck.md")

	ed := NewEditor(validPath)
	ed.Dirty = true
	sendTestKey(&ed, &d, "ctrl+s")

	if ed.Dirty {
		t.Errorf("expected dirty flag false after save")
	}
	if ed.Message != "saved" {
		t.Errorf("expected message 'saved', got %q", ed.Message)
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if !strings.Contains(string(content), "Slide 1") {
		t.Errorf("expected saved file to contain 'Slide 1'")
	}

	// Save with unwritable path to test error branch
	invalidEd := NewEditor("/non_existent_folder_xyz/file.md")
	invalidEd.Save(d)
	if !strings.Contains(invalidEd.Message, "save error") {
		t.Errorf("expected 'save error' message, got %q", invalidEd.Message)
	}
}

func TestEditorEdgeCases(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockCode, Lines: []string{"fmt.Println(1)", "fmt.Println(2)"}},
					{Kind: BlockList, Text: "- Item"},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")

	// EnterEdit and ExitEdit on Code block
	ed.EnterEdit(&d)
	if !strings.Contains(ed.Draft, "fmt.Println") {
		t.Errorf("expected code draft, got %q", ed.Draft)
	}
	ed.Draft = "fmt.Println(99)"
	ed.ExitEdit(&d)
	if len(d.Slides[0].Blocks[0].Lines) != 1 || d.Slides[0].Blocks[0].Lines[0] != "fmt.Println(99)" {
		t.Errorf("expected updated code block lines, got %v", d.Slides[0].Blocks[0].Lines)
	}

	// EnterEdit on List block
	ed.BlockIdx = 1
	ed.EnterEdit(&d)
	if ed.Draft != "- Item" {
		t.Errorf("expected list draft, got %q", ed.Draft)
	}
	ed.CancelEdit()

	// Bounds checks: MoveBlockUp at index 0
	ed.BlockIdx = 0
	ed.MoveBlockUp(&d)
	if ed.BlockIdx != 0 {
		t.Errorf("expected blockIdx 0, got %d", ed.BlockIdx)
	}

	// MoveBlockDown at last block
	ed.BlockIdx = 1
	ed.MoveBlockDown(&d)
	if ed.BlockIdx != 1 {
		t.Errorf("expected blockIdx 1, got %d", ed.BlockIdx)
	}

	// DeleteBlock when only 1 block remains
	ed.DeleteBlock(&d) // deletes block 1
	if len(d.Slides[0].Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(d.Slides[0].Blocks))
	}
	ed.DeleteBlock(&d) // attempt to delete last block
	if ed.Message != "can't delete last block" {
		t.Errorf("expected 'can't delete last block', got %q", ed.Message)
	}

	// DeleteSlide when only 1 slide remains
	ed.DeleteSlide(&d)
	if ed.Message != "can't delete last slide" {
		t.Errorf("expected 'can't delete last slide', got %q", ed.Message)
	}

	// Undo and Redo with empty stacks
	emptyEd := NewEditor("empty.deck.md")
	emptyEd.UndoStack = nil
	emptyEd.RedoStack = nil
	sendTestKey(&emptyEd, &d, "u")
	if emptyEd.Message == "undo" {
		t.Errorf("expected no undo with empty stack")
	}
	sendTestKey(&emptyEd, &d, "ctrl+r")
	if emptyEd.Message == "redo" {
		t.Errorf("expected no redo with empty stack")
	}

	// currentBlock out of bounds
	emptyEd.SlideIdx = 99
	if blk := emptyEd.currentBlock(&d); blk != nil {
		t.Errorf("expected nil currentBlock for out-of-range slideIdx")
	}
	emptyEd.SlideIdx = 0
	emptyEd.BlockIdx = 99
	if blk := emptyEd.currentBlock(&d); blk != nil {
		t.Errorf("expected nil currentBlock for out-of-range blockIdx")
	}

	// AddBlock / MoveBlockUp on out of bounds slide
	emptyEd.SlideIdx = 99
	emptyEd.AddBlock(&d)
	emptyEd.MoveBlockUp(&d)
	emptyEd.MoveBlockDown(&d)
	emptyEd.ToggleAlign(&d)
}

func TestAutoSaveFeatures(t *testing.T) {
	tmpDir := t.TempDir()
	deckPath := filepath.Join(tmpDir, "presentation.deck.md")
	initialContent := "---\ntitle: AutoSave Test\n---\n# Slide 1\nHello World\n"
	if err := os.WriteFile(deckPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	d := ParseDeck(initialContent)
	ed := NewEditor(deckPath)

	// 1. ToggleAlign should auto-save immediately to disk when file exists
	ed.ToggleAlign(&d)
	if ed.Dirty {
		t.Errorf("expected dirty flag to be false after auto-save on ToggleAlign")
	}
	if !strings.Contains(ed.Message, "(saved)") {
		t.Errorf("expected message to indicate '(saved)', got %q", ed.Message)
	}
	diskContent, err := os.ReadFile(deckPath)
	if err != nil {
		t.Fatalf("failed to read disk file: %v", err)
	}
	if !strings.Contains(string(diskContent), "::align right") {
		t.Errorf("expected disk file to contain '::align right' after ToggleAlign, got:\n%s", string(diskContent))
	}

	// 2. handleEdit 'enter' should auto-save committed text to disk
	ed.BlockIdx = 1 // "Hello World"
	ed.EnterEdit(&d)
	ed.Draft = "Updated via Live Edit"
	sendTestKey(&ed, &d, "enter")
	diskContentAfterEdit, _ := os.ReadFile(deckPath)
	if !strings.Contains(string(diskContentAfterEdit), "Updated via Live Edit") {
		t.Errorf("expected disk file to contain 'Updated via Live Edit' after enter, got:\n%s", string(diskContentAfterEdit))
	}

	// 3. handleNav 'q' should auto-save if dirty
	ed.AddBlock(&d) // marks dirty
	if !ed.Dirty {
		t.Fatalf("expected dirty flag true after AddBlock")
	}
	cmd := sendTestKey(&ed, &d, "q")
	if cmd == nil {
		t.Errorf("expected tea.Quit command on 'q'")
	}
	if ed.Dirty {
		t.Errorf("expected dirty flag to be false after auto-save on quit")
	}
	diskContentAfterQuit, _ := os.ReadFile(deckPath)
	if !strings.Contains(string(diskContentAfterQuit), "new block") {
		t.Errorf("expected disk file to contain 'new block' after auto-save on quit, got:\n%s", string(diskContentAfterQuit))
	}
}

func TestSpeakerNotesEditor(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Slide with Notes"},
					{Kind: BlockParagraph, Text: "Visible body"},
					{Kind: BlockDirective, Directive: "::notes", Lines: []string{"Secret speaker note"}},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")

	if ed.ShowNotes {
		t.Errorf("expected ShowNotes to be false initially")
	}

	// Move down to visible body (BlockIdx: 1)
	ed.MoveDown(&d)
	if ed.BlockIdx != 1 {
		t.Fatalf("expected blockIdx 1, got %d", ed.BlockIdx)
	}

	// Move down again - should NOT navigate to ::notes (BlockIdx: 2)
	ed.MoveDown(&d)
	if ed.BlockIdx != 1 {
		t.Fatalf("expected blockIdx to stay at last visible block 1, got %d", ed.BlockIdx)
	}

	// Toggle notes open with 'n'
	sendTestKey(&ed, &d, "n")
	if !ed.ShowNotes {
		t.Errorf("expected ShowNotes to be true after pressing 'n'")
	}
	if !strings.Contains(ed.Message, "notes open") {
		t.Errorf("expected notes open message, got %q", ed.Message)
	}

	// Toggle notes closed with 'n'
	sendTestKey(&ed, &d, "n")
	if ed.ShowNotes {
		t.Errorf("expected ShowNotes to be false after pressing 'n' second time")
	}
	if !strings.Contains(ed.Message, "notes closed") {
		t.Errorf("expected notes closed message, got %q", ed.Message)
	}
}

func TestEditorHelpModal(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	// Help modal toggle
	if ed.ShowHelp {
		t.Errorf("expected ShowHelp false initially")
	}
	sendTestKey(&ed, &d, "?")
	if !ed.ShowHelp {
		t.Errorf("expected ShowHelp true after '?'")
	}
	// 'esc' closes help
	sendTestKey(&ed, &d, "esc")
	if ed.ShowHelp {
		t.Errorf("expected ShowHelp false after 'esc'")
	}
	// 'f1' toggles help
	sendTestKey(&ed, &d, "f1")
	if !ed.ShowHelp {
		t.Errorf("expected ShowHelp true after 'f1'")
	}
	sendTestKey(&ed, &d, "?")
	if ed.ShowHelp {
		t.Errorf("expected ShowHelp false after second '?'")
	}
}

func TestEditorCycleTheme(t *testing.T) {
	d := sampleDeck()
	tmpFile := t.TempDir() + "/theme_test.deck.md"
	ed := NewEditor(tmpFile)

	initialTheme := ed.Theme
	if initialTheme == "" {
		initialTheme = "termdeck"
	}

	// Press 't' to cycle theme
	sendTestKey(&ed, &d, "t")
	if ed.Theme == initialTheme || ed.Theme == "" {
		t.Errorf("expected theme to change after 't', got %q", ed.Theme)
	}
	if !strings.Contains(ed.Message, "theme:") {
		t.Errorf("expected theme notification message, got %q", ed.Message)
	}
	if d.Theme != ed.Theme {
		t.Errorf("expected deck.Theme %q to match editor.Theme %q", d.Theme, ed.Theme)
	}

	// Press 'T' (shift+t) to cycle again
	prevTheme := ed.Theme
	sendTestKey(&ed, &d, "T")
	if ed.Theme == prevTheme {
		t.Errorf("expected theme to change after 'T', got %q", ed.Theme)
	}

	// Press 'f2' to cycle again
	prevTheme2 := ed.Theme
	sendTestKey(&ed, &d, "f2")
	if ed.Theme == prevTheme2 {
		t.Errorf("expected theme to change after 'f2', got %q", ed.Theme)
	}

	// Press 'ctrl+t' to cycle again
	prevTheme3 := ed.Theme
	sendTestKey(&ed, &d, "ctrl+t")
	if ed.Theme == prevTheme3 {
		t.Errorf("expected theme to change after 'ctrl+t', got %q", ed.Theme)
	}
}

func TestEditorCycleThemeInEditMode(t *testing.T) {
	d := sampleDeck()
	tmpFile := t.TempDir() + "/theme_edit_test.deck.md"
	ed := NewEditor(tmpFile)

	// Enter edit mode
	sendTestKey(&ed, &d, "i")
	if ed.Mode != ModeEdit {
		t.Fatalf("expected ModeEdit, got %v", ed.Mode)
	}

	initialTheme := ed.Theme
	if initialTheme == "" {
		initialTheme = "termdeck"
	}

	// In edit mode, press ctrl+t to cycle theme
	sendTestKey(&ed, &d, "ctrl+t")
	if ed.Theme == initialTheme || ed.Theme == "" {
		t.Errorf("expected theme to change in edit mode after ctrl+t, got %q", ed.Theme)
	}
	if ed.Mode != ModeEdit {
		t.Errorf("expected to stay in ModeEdit after ctrl+t, got %v", ed.Mode)
	}

	// Press f2 to cycle again in edit mode
	prevTheme := ed.Theme
	sendTestKey(&ed, &d, "f2")
	if ed.Theme == prevTheme {
		t.Errorf("expected theme to change in edit mode after f2, got %q", ed.Theme)
	}

	// Confirm editStatus renders the theme
	es := editStatus(ed, 80)
	if !strings.Contains(es, "^t theme") {
		t.Errorf("expected editStatus to contain '^t theme': %q", es)
	}
}

func TestEditorZenMode(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	if ed.ZenMode {
		t.Fatalf("expected ZenMode initially false, got true")
	}

	sendTestKey(&ed, &d, "z")
	if !ed.ZenMode {
		t.Errorf("expected ZenMode true after pressing 'z'")
	}

	sendTestKey(&ed, &d, "z")
	if ed.ZenMode {
		t.Errorf("expected ZenMode false after pressing 'z' again")
	}
}

func TestEditorQuickJumpPrompt(t *testing.T) {
	d := sampleDeck()
	ed := NewEditor("test.deck.md")

	// 1. Pressing '/' enters ModePrompt
	sendTestKey(&ed, &d, "/")
	if ed.Mode != ModePrompt {
		t.Fatalf("expected ModePrompt after '/', got %v", ed.Mode)
	}

	// 2. Esc cancels ModePrompt
	sendTestKey(&ed, &d, "esc")
	if ed.Mode != ModeNav {
		t.Fatalf("expected ModeNav after esc, got %v", ed.Mode)
	}

	// 3. Numeric jump: type '2' and enter
	sendTestKey(&ed, &d, "/")
	sendTestKey(&ed, &d, "2")
	if ed.Draft != "2" {
		t.Errorf("expected draft '2', got %q", ed.Draft)
	}
	sendTestKey(&ed, &d, "enter")
	if ed.Mode != ModeNav {
		t.Errorf("expected ModeNav after enter")
	}
	if ed.SlideIdx != 1 {
		t.Errorf("expected slideIdx 1 after jumping to slide 2, got %d", ed.SlideIdx)
	}

	// 4. Clamping out-of-range number
	sendTestKey(&ed, &d, "/")
	sendTestKey(&ed, &d, "9")
	sendTestKey(&ed, &d, "9")
	sendTestKey(&ed, &d, "enter")
	if ed.SlideIdx != 1 {
		t.Errorf("expected slideIdx clamped to 1 (last slide), got %d", ed.SlideIdx)
	}

	// 5. Backspace in prompt
	sendTestKey(&ed, &d, "/")
	sendTestKey(&ed, &d, "1")
	sendTestKey(&ed, &d, "5")
	sendTestKey(&ed, &d, "backspace")
	if ed.Draft != "1" {
		t.Errorf("expected draft '1' after backspace, got %q", ed.Draft)
	}
	sendTestKey(&ed, &d, "enter")
	if ed.SlideIdx != 0 {
		t.Errorf("expected slideIdx 0 after jump to slide 1, got %d", ed.SlideIdx)
	}

	// 6. Title search jump: type "Slide 2"
	sendTestKey(&ed, &d, "/")
	for _, ch := range "Slide 2" {
		sendTestKey(&ed, &d, string(ch))
	}
	sendTestKey(&ed, &d, "enter")
	if ed.SlideIdx != 1 {
		t.Errorf("expected jump to slide 2 via search, got %d", ed.SlideIdx)
	}

	// 7. Search no match
	sendTestKey(&ed, &d, "/")
	for _, ch := range "nonexistent" {
		sendTestKey(&ed, &d, string(ch))
	}
	sendTestKey(&ed, &d, "enter")
	if !strings.Contains(ed.Message, "no slide matching") {
		t.Errorf("expected no match message, got %q", ed.Message)
	}

	// 8. Empty query enter
	sendTestKey(&ed, &d, "/")
	sendTestKey(&ed, &d, "enter")
	if ed.Mode != ModeNav {
		t.Errorf("expected ModeNav after empty enter")
	}
}

func TestEditorToggleTask(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Sprint Goals"},
					{Kind: BlockList, Text: "- [ ] Setup CI pipeline"},
					{Kind: BlockList, Text: "* Regular bullet"},
				},
			},
		},
	}
	tmpFile := t.TempDir() + "/tasks.deck.md"
	ed := NewEditor(tmpFile)

	// 1. Move to block 0 (Heading) and press 'x' -> should do nothing
	sendTestKey(&ed, &d, "x")
	if d.Slides[0].Blocks[0].Text != "Sprint Goals" {
		t.Errorf("expected heading unchanged after 'x'")
	}

	// 2. Move to block 1 (- [ ] Setup CI pipeline) and press 'x'
	sendTestKey(&ed, &d, "j")
	if ed.BlockIdx != 1 {
		t.Fatalf("expected BlockIdx 1, got %d", ed.BlockIdx)
	}
	sendTestKey(&ed, &d, "x")
	if d.Slides[0].Blocks[1].Text != "- [x] Setup CI pipeline" {
		t.Errorf("expected task marked [x], got %q", d.Slides[0].Blocks[1].Text)
	}
	if ed.Message != "task: complete" {
		t.Errorf("expected message 'task: complete', got %q", ed.Message)
	}

	// 3. Press 'x' again -> should toggle back to [ ]
	sendTestKey(&ed, &d, "x")
	if d.Slides[0].Blocks[1].Text != "- [ ] Setup CI pipeline" {
		t.Errorf("expected task marked [ ], got %q", d.Slides[0].Blocks[1].Text)
	}
	if ed.Message != "task: pending" {
		t.Errorf("expected message 'task: pending', got %q", ed.Message)
	}

	// 4. Test undo on task toggle
	sendTestKey(&ed, &d, "x")
	if d.Slides[0].Blocks[1].Text != "- [x] Setup CI pipeline" {
		t.Errorf("expected task [x], got %q", d.Slides[0].Blocks[1].Text)
	}
	sendTestKey(&ed, &d, "u")
	if d.Slides[0].Blocks[1].Text != "- [ ] Setup CI pipeline" {
		t.Errorf("expected task restored to [ ] via undo, got %q", d.Slides[0].Blocks[1].Text)
	}

	// 5. Move to block 2 (* Regular bullet) and press 'x' -> converts to task
	sendTestKey(&ed, &d, "j")
	sendTestKey(&ed, &d, "x")
	if d.Slides[0].Blocks[2].Text != "* [x] Regular bullet" {
		t.Errorf("expected converted task, got %q", d.Slides[0].Blocks[2].Text)
	}
}

func TestEditorToggleLineNumbers(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockCode, Lang: "go", Lines: []string{"func main() {}"}},
				},
			},
		},
	}
	ed := NewEditor("")
	if ed.ShowLineNumbers {
		t.Errorf("expected ShowLineNumbers to start false")
	}

	sendTestKey(&ed, &d, "L")
	if !ed.ShowLineNumbers {
		t.Errorf("expected ShowLineNumbers to be true after pressing 'L'")
	}
	if ed.Message != "line numbers: on" {
		t.Errorf("expected message 'line numbers: on', got %q", ed.Message)
	}

	sendTestKey(&ed, &d, "L")
	if ed.ShowLineNumbers {
		t.Errorf("expected ShowLineNumbers to be false after second 'L'")
	}
	if ed.Message != "line numbers: off" {
		t.Errorf("expected message 'line numbers: off', got %q", ed.Message)
	}
}

func TestEditorToggleTimer(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Timer Slide"}}},
		},
	}
	ed := NewEditor("")
	if ed.ShowTimer {
		t.Errorf("expected ShowTimer to start false")
	}

	// Press 'c' to turn on timer
	cmd := ed.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}, &d)
	if !ed.ShowTimer {
		t.Errorf("expected ShowTimer to be true after 'c'")
	}
	if ed.TimerStart.IsZero() {
		t.Errorf("expected TimerStart to be non-zero")
	}
	if cmd == nil {
		t.Errorf("expected non-nil TickCmd after starting timer")
	}

	// Press 'c' again to turn off timer
	cmd = ed.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")}, &d)
	if ed.ShowTimer {
		t.Errorf("expected ShowTimer to be false after second 'c'")
	}
	if cmd != nil {
		t.Errorf("expected nil cmd when turning timer off")
	}

	// Press 'C' to reset timer
	prevStart := ed.TimerStart
	time.Sleep(10 * time.Millisecond)
	cmd = ed.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("C")}, &d)
	if !ed.ShowTimer {
		t.Errorf("expected ShowTimer to be true after 'C'")
	}
	if !ed.TimerStart.After(prevStart) {
		t.Errorf("expected TimerStart to be reset to newer time")
	}
	if cmd == nil {
		t.Errorf("expected non-nil TickCmd after resetting timer")
	}
}

func TestEditorReload(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "reload.deck.md")
	initialContent := "---\ntitle: Deck 1\n---\n# Slide 1\nHello\n"
	if err := os.WriteFile(tmpFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write tmp file: %v", err)
	}

	d := ParseDeck(initialContent)
	ed := NewEditor(tmpFile)

	// Modify file externally on disk
	updatedContent := "---\ntitle: Deck 2\n---\n# Slide 1 Updated\nNew Content\n---\n# Slide 2\nSecond slide\n"
	if err := os.WriteFile(tmpFile, []byte(updatedContent), 0644); err != nil {
		t.Fatalf("failed to update tmp file: %v", err)
	}

	// Trigger reload
	err := ed.Reload(&d)
	if err != nil {
		t.Fatalf("unexpected reload error: %v", err)
	}
	if len(d.Slides) != 2 {
		t.Errorf("expected 2 slides after reload, got %d", len(d.Slides))
	}
	if d.Slides[0].Blocks[0].Text != "Slide 1 Updated" {
		t.Errorf("expected updated title after reload, got %q", d.Slides[0].Blocks[0].Text)
	}
	if ed.Message != "reloaded from disk" {
		t.Errorf("expected message 'reloaded from disk', got %q", ed.Message)
	}

	// Error handling: missing file
	edNoFile := NewEditor("")
	if err := edNoFile.Reload(&d); err == nil {
		t.Errorf("expected error when reloading with empty FilePath")
	}

	// Error handling: file with 0 slides
	emptyFile := filepath.Join(t.TempDir(), "empty.deck.md")
	_ = os.WriteFile(emptyFile, []byte(""), 0644)
	edEmpty := NewEditor(emptyFile)
	if err := edEmpty.Reload(&d); err == nil {
		t.Errorf("expected error when reloading empty file")
	}
}

func TestEditorReloadKeyNav(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "reload_key.deck.md")
	content := "# Original Slide\nBody\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write tmp file: %v", err)
	}

	d := ParseDeck(content)
	ed := NewEditor(tmpFile)

	// External change
	_ = os.WriteFile(tmpFile, []byte("# Changed Slide\nBody\n"), 0644)

	// Press 'r'
	sendTestKey(&ed, &d, "r")
	if d.Slides[0].Blocks[0].Text != "Changed Slide" {
		t.Errorf("expected slide reloaded on 'r', got %q", d.Slides[0].Blocks[0].Text)
	}
	if ed.Message != "reloaded from disk" {
		t.Errorf("expected message 'reloaded from disk', got %q", ed.Message)
	}
}

func TestEditorSlideOverview(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 1"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 2"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 3"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 4"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 5"}}},
		},
	}
	ed := NewEditor("test.deck.md")
	ed.SlideIdx = 1

	// 1. Press 'o' to open overview
	sendTestKey(&ed, &d, "o")
	if !ed.ShowOverview {
		t.Errorf("expected ShowOverview true on 'o'")
	}
	if ed.OverviewCursor != 1 {
		t.Errorf("expected OverviewCursor initialized to current SlideIdx 1, got %d", ed.OverviewCursor)
	}

	// 2. Navigate right / l
	sendTestKey(&ed, &d, "right")
	if ed.OverviewCursor != 2 {
		t.Errorf("expected OverviewCursor 2 after right arrow, got %d", ed.OverviewCursor)
	}
	sendTestKey(&ed, &d, "l")
	if ed.OverviewCursor != 3 {
		t.Errorf("expected OverviewCursor 3 after 'l', got %d", ed.OverviewCursor)
	}

	// 3. Navigate left / h
	sendTestKey(&ed, &d, "left")
	if ed.OverviewCursor != 2 {
		t.Errorf("expected OverviewCursor 2 after left arrow, got %d", ed.OverviewCursor)
	}
	sendTestKey(&ed, &d, "h")
	if ed.OverviewCursor != 1 {
		t.Errorf("expected OverviewCursor 1 after 'h', got %d", ed.OverviewCursor)
	}

	// 4. Navigate down / j (3 cols by default)
	// Cursor 1 + 3 = 4 (slide 5)
	sendTestKey(&ed, &d, "down")
	if ed.OverviewCursor != 4 {
		t.Errorf("expected OverviewCursor 4 after down arrow, got %d", ed.OverviewCursor)
	}

	// Down at bottom clamps
	sendTestKey(&ed, &d, "j")
	if ed.OverviewCursor != 4 {
		t.Errorf("expected OverviewCursor clamped at 4, got %d", ed.OverviewCursor)
	}

	// 5. Navigate up / k (Cursor 4 - 3 = 1)
	sendTestKey(&ed, &d, "up")
	if ed.OverviewCursor != 1 {
		t.Errorf("expected OverviewCursor 1 after up arrow, got %d", ed.OverviewCursor)
	}

	// 6. Navigate g / G
	sendTestKey(&ed, &d, "G")
	if ed.OverviewCursor != 4 {
		t.Errorf("expected OverviewCursor 4 after G, got %d", ed.OverviewCursor)
	}
	sendTestKey(&ed, &d, "g")
	if ed.OverviewCursor != 0 {
		t.Errorf("expected OverviewCursor 0 after g, got %d", ed.OverviewCursor)
	}

	// 7. Jump on Enter
	ed.OverviewCursor = 3 // Slide 4
	sendTestKey(&ed, &d, "enter")
	if ed.ShowOverview {
		t.Errorf("expected ShowOverview false after Enter")
	}
	if ed.SlideIdx != 3 {
		t.Errorf("expected SlideIdx 3 after jump, got %d", ed.SlideIdx)
	}
	if !strings.Contains(ed.Message, "jumped to slide 4/5") {
		t.Errorf("expected jump message, got %q", ed.Message)
	}

	// 8. Open and cancel with Esc
	sendTestKey(&ed, &d, "O")
	if !ed.ShowOverview {
		t.Errorf("expected ShowOverview true after 'O'")
	}
	sendTestKey(&ed, &d, "esc")
	if ed.ShowOverview {
		t.Errorf("expected ShowOverview false after esc")
	}

	// 9. Open and cancel with 'o' toggle
	sendTestKey(&ed, &d, "o")
	sendTestKey(&ed, &d, "o")
	if ed.ShowOverview {
		t.Errorf("expected ShowOverview false after toggling 'o' again")
	}

	// 10. Close overview with 'q'
	sendTestKey(&ed, &d, "o")
	cmd := sendTestKey(&ed, &d, "q")
	if ed.ShowOverview {
		t.Errorf("expected ShowOverview false after 'q'")
	}
	if cmd != nil {
		t.Errorf("expected nil cmd from 'q' in overview modal (not quit application)")
	}
}

func TestEditorYankAndBlankScreen(t *testing.T) {
	// 1. Test OSC52Copy formatting
	osc := OSC52Copy("hello world")
	if !strings.HasPrefix(osc, "\x1b]52;c;") || !strings.HasSuffix(osc, "\x07") {
		t.Errorf("unexpected OSC 52 format: %q", osc)
	}

	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Slide Title"},
					{Kind: BlockCode, Lang: "go", Lines: []string{"func main() {", `    println("copied")`, "}"}},
					{Kind: BlockTable, Lines: []string{"| Col 1 | Col 2 |", "|---|---|", "| A | B |"}},
					{Kind: BlockCallout, Callout: "tip", Lines: []string{"Tip line 1", "Tip line 2"}},
					{Kind: BlockParagraph, Text: "A paragraph text"},
					{Kind: BlockParagraph, Text: ""}, // Empty block
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")

	// 2. Yank Code block
	ed.BlockIdx = 1
	yanked, err := ed.YankBlock(&d)
	if err != nil {
		t.Fatalf("unexpected error yanking code block: %v", err)
	}
	if !strings.Contains(yanked, "func main()") || !strings.Contains(yanked, "println(\"copied\")") {
		t.Errorf("unexpected yanked code content: %q", yanked)
	}

	// 3. Key press 'y' on Code block
	cmd := sendTestKey(&ed, &d, "y")
	if cmd == nil {
		t.Errorf("expected non-nil tea.Cmd emitting OSC 52 sequence on 'y'")
	}
	if !strings.Contains(ed.Message, "yanked 3 code lines") {
		t.Errorf("expected message 'yanked 3 code lines', got %q", ed.Message)
	}

	// 4. Yank Table block
	ed.BlockIdx = 2
	yankedTable, err := ed.YankBlock(&d)
	if err != nil || !strings.Contains(yankedTable, "| Col 1 | Col 2 |") {
		t.Errorf("unexpected yanked table: %q", yankedTable)
	}

	// 5. Yank Callout block
	ed.BlockIdx = 3
	yankedCallout, err := ed.YankBlock(&d)
	if err != nil || !strings.Contains(yankedCallout, "Tip line 1") {
		t.Errorf("unexpected yanked callout: %q", yankedCallout)
	}

	// 6. Yank Paragraph block
	ed.BlockIdx = 4
	sendTestKey(&ed, &d, "y")
	if !strings.Contains(ed.Message, "yanked 16 chars") {
		t.Errorf("expected message 'yanked 16 chars', got %q", ed.Message)
	}

	// 7. Yank Empty block -> error
	ed.BlockIdx = 5
	_, err = ed.YankBlock(&d)
	if err == nil {
		t.Errorf("expected error yanking empty block")
	}

	// 8. Blank screen toggle with 'b'
	sendTestKey(&ed, &d, "b")
	if !ed.ScreenBlank {
		t.Errorf("expected ScreenBlank true after pressing 'b'")
	}
	if !strings.Contains(ed.Message, "screen blanked") {
		t.Errorf("expected 'screen blanked' message, got %q", ed.Message)
	}

	// 9. Any key resumes screen
	sendTestKey(&ed, &d, "space")
	if ed.ScreenBlank {
		t.Errorf("expected ScreenBlank false after pressing any key to resume")
	}
	if ed.Message != "screen resumed" {
		t.Errorf("expected 'screen resumed' message, got %q", ed.Message)
	}

	// 10. Blank screen with 'B'
	sendTestKey(&ed, &d, "B")
	if !ed.ScreenBlank {
		t.Errorf("expected ScreenBlank true after pressing 'B'")
	}
	sendTestKey(&ed, &d, "enter")
	if ed.ScreenBlank {
		t.Errorf("expected ScreenBlank false after pressing enter to resume")
	}
}

func TestEditorExportHTMLKeyNav(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "export_key.deck.md")
	content := "# Original Slide\nBody\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write tmp file: %v", err)
	}

	d := ParseDeck(content)
	ed := NewEditor(tmpFile)

	// Press 'E'
	sendTestKey(&ed, &d, "E")
	if ed.Message != "exported to export_key.html" {
		t.Errorf("expected status 'exported to export_key.html', got %q", ed.Message)
	}
	expectedHTML := filepath.Join(tmpDir, "export_key.html")
	if _, err := os.Stat(expectedHTML); err != nil {
		t.Errorf("expected exported html file %q to exist", expectedHTML)
	}
}

func TestEditorAutoplay(t *testing.T) {
	d := ParseDeck("# Slide 1\nContent 1\n---\n# Slide 2\nContent 2\n---\n# Slide 3\nContent 3\n")
	ed := NewEditor("test.deck.md")

	// 1. Initial state
	if ed.Autoplay {
		t.Fatalf("expected Autoplay initially false")
	}

	// 2. Press 'A' to toggle autoplay on
	cmd := sendTestKey(&ed, &d, "A")
	if !ed.Autoplay {
		t.Fatalf("expected Autoplay true after pressing 'A'")
	}
	if cmd == nil {
		t.Errorf("expected non-nil TickCmd when enabling autoplay")
	}
	if !strings.Contains(ed.Message, "autoplay: on") {
		t.Errorf("expected 'autoplay: on' message, got %q", ed.Message)
	}
	if ed.AutoplayCountdown != ed.AutoplayInterval {
		t.Errorf("expected countdown %d, got %d", ed.AutoplayInterval, ed.AutoplayCountdown)
	}

	// 3. Tick countdown without expiring
	ed.AutoplayCountdown = 3
	advanced := ed.TickAutoplay(&d)
	if advanced {
		t.Errorf("expected slide not advanced when countdown > 1")
	}
	if ed.AutoplayCountdown != 2 {
		t.Errorf("expected countdown 2, got %d", ed.AutoplayCountdown)
	}

	// 4. Tick countdown expiring -> advance slide
	ed.AutoplayCountdown = 1
	ed.SlideIdx = 0
	advanced = ed.TickAutoplay(&d)
	if !advanced {
		t.Errorf("expected slide advanced when countdown expired")
	}
	if ed.SlideIdx != 1 {
		t.Errorf("expected slide index 1, got %d", ed.SlideIdx)
	}
	if ed.AutoplayCountdown != ed.AutoplayInterval {
		t.Errorf("expected countdown reset to interval, got %d", ed.AutoplayCountdown)
	}

	// 5. Expiring on last slide loops to start
	ed.SlideIdx = 2
	ed.AutoplayCountdown = 1
	advanced = ed.TickAutoplay(&d)
	if !advanced {
		t.Errorf("expected slide to loop to start")
	}
	if ed.SlideIdx != 0 {
		t.Errorf("expected slide index 0 after loop, got %d", ed.SlideIdx)
	}

	// 6. Manual navigation resets countdown
	ed.AutoplayCountdown = 1
	sendTestKey(&ed, &d, "right")
	if ed.AutoplayCountdown != ed.AutoplayInterval {
		t.Errorf("expected countdown reset after manual 'right', got %d", ed.AutoplayCountdown)
	}

	ed.AutoplayCountdown = 1
	sendTestKey(&ed, &d, "left")
	if ed.AutoplayCountdown != ed.AutoplayInterval {
		t.Errorf("expected countdown reset after manual 'left', got %d", ed.AutoplayCountdown)
	}

	// 7. Toggle with custom interval
	ed.Autoplay = false
	ed.ToggleAutoplay(10)
	if !ed.Autoplay || ed.AutoplayInterval != 10 {
		t.Errorf("expected autoplay 10s interval, got %d", ed.AutoplayInterval)
	}

	// 8. Press 'A' to toggle off
	cmdOff := sendTestKey(&ed, &d, "A")
	if ed.Autoplay {
		t.Errorf("expected Autoplay false after pressing 'A' again")
	}
	if cmdOff != nil {
		t.Errorf("expected nil cmd when disabling autoplay")
	}
	if ed.Message != "autoplay: off" {
		t.Errorf("expected 'autoplay: off' message, got %q", ed.Message)
	}

	// 9. TickAutoplay when disabled or nil deck returns false
	if ed.TickAutoplay(&d) {
		t.Errorf("expected false when autoplay disabled")
	}
	if ed.TickAutoplay(nil) {
		t.Errorf("expected false when deck is nil")
	}
}

func TestEditorBranchAndGraphNavigation(t *testing.T) {
	src := `---
title: Branch Navigation Deck
---

# Architecture Overview
Choose a module to explore:

::branch [1] Storage Engine -> storage
::branch [2] Network Transport -> network

---

::id storage
# Storage Engine
::next conclusion
LSM tree design.

---

::id network
# Network Transport
::next conclusion
gRPC and multiplexing.

---

::id conclusion
::prev storage
# Conclusion
Summary of all modules.`

	d := ParseDeck(src)
	ed := NewEditor("")

	// 1. Initial state
	if ed.SlideIdx != 0 {
		t.Fatalf("expected initial slide 0, got %d", ed.SlideIdx)
	}

	// 2. Press '1' to follow branch 1
	sendTestKey(&ed, &d, "1")
	if ed.SlideIdx != 1 {
		t.Errorf("expected jump to slide 1 (storage), got %d", ed.SlideIdx)
	}
	if len(ed.History) != 1 || ed.History[0] != 0 {
		t.Errorf("expected history [0], got %+v", ed.History)
	}
	if !strings.Contains(ed.Message, "Storage Engine") {
		t.Errorf("expected branch message, got %q", ed.Message)
	}

	// 3. Slide 1 has ::next conclusion -> pressing 'right' should jump to slide 3
	sendTestKey(&ed, &d, "right")
	if ed.SlideIdx != 3 {
		t.Errorf("expected jump to slide 3 (conclusion) via ::next, got %d", ed.SlideIdx)
	}
	if len(ed.History) != 2 || ed.History[1] != 1 {
		t.Errorf("expected history [0, 1], got %+v", ed.History)
	}

	// 4. Press 'backspace' to pop history -> returns to slide 1
	sendTestKey(&ed, &d, "backspace")
	if ed.SlideIdx != 1 {
		t.Errorf("expected return to slide 1 via backspace, got %d", ed.SlideIdx)
	}
	if len(ed.History) != 1 {
		t.Errorf("expected history len 1, got %+v", ed.History)
	}

	// 5. Press 'H' to open history modal and '1' to rewind -> returns to slide 0
	sendTestKey(&ed, &d, "H")
	sendTestKey(&ed, &d, "1")
	if ed.SlideIdx != 0 {
		t.Errorf("expected return to slide 0 via H -> 1, got %d", ed.SlideIdx)
	}
	if len(ed.History) != 0 {
		t.Errorf("expected empty history, got %+v", ed.History)
	}

	// 6. Enter on focused branch block:
	// Slide 0 has: block 0 (heading), block 1 (paragraph), block 2 (branch 1), block 3 (branch 2)
	ed.BlockIdx = 3
	sendTestKey(&ed, &d, "enter")
	if ed.SlideIdx != 2 {
		t.Errorf("expected jump to slide 2 (network) via enter on branch block, got %d", ed.SlideIdx)
	}
	if len(ed.History) != 1 || ed.History[0] != 0 {
		t.Errorf("expected history [0], got %+v", ed.History)
	}

	// 7. Backtrack again
	sendTestKey(&ed, &d, "backspace")
	if ed.SlideIdx != 0 {
		t.Errorf("expected return to slide 0, got %d", ed.SlideIdx)
	}

	// 8. Test Graph Map Modal ('M')
	sendTestKey(&ed, &d, "M")
	if !ed.ShowGraphMap {
		t.Errorf("expected ShowGraphMap true after pressing 'M'")
	}
	if ed.GraphMapCursor != 0 {
		t.Errorf("expected GraphMapCursor 0, got %d", ed.GraphMapCursor)
	}

	// Navigate cursor in graph map
	sendTestKey(&ed, &d, "j")
	if ed.GraphMapCursor != 1 {
		t.Errorf("expected GraphMapCursor 1 after 'j', got %d", ed.GraphMapCursor)
	}
	sendTestKey(&ed, &d, "down")
	if ed.GraphMapCursor != 2 {
		t.Errorf("expected GraphMapCursor 2 after 'down', got %d", ed.GraphMapCursor)
	}
	sendTestKey(&ed, &d, "k")
	if ed.GraphMapCursor != 1 {
		t.Errorf("expected GraphMapCursor 1 after 'k', got %d", ed.GraphMapCursor)
	}

	// Press 'enter' in graph map to jump
	sendTestKey(&ed, &d, "enter")
	if ed.ShowGraphMap {
		t.Errorf("expected ShowGraphMap false after enter")
	}
	if ed.SlideIdx != 1 {
		t.Errorf("expected jump to slide 1, got %d", ed.SlideIdx)
	}

	// Toggle graph map on and dismiss with esc
	sendTestKey(&ed, &d, "M")
	if !ed.ShowGraphMap {
		t.Errorf("expected ShowGraphMap true")
	}
	sendTestKey(&ed, &d, "esc")
	if ed.ShowGraphMap {
		t.Errorf("expected ShowGraphMap false after esc")
	}

	// 9. Edge cases
	if ed.FollowBranch("non-existent", &d) {
		t.Errorf("expected false for non-existent branch target")
	}
	if ed.FollowBranch("storage", nil) {
		t.Errorf("expected false for nil deck")
	}
	edEmpty := NewEditor("")
	if edEmpty.BackHistory(&d) {
		t.Errorf("expected false when history is empty")
	}
	if edEmpty.BackHistory(nil) {
		t.Errorf("expected false for nil deck")
	}
}

func TestEditorFocusModeAndLiveRunner(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Slide With Code"},
					{Kind: BlockCode, Lang: "sh", Text: "echo 'running from editor'"},
					{Kind: BlockList, Text: "- [ ] A task to toggle"},
				},
			},
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Slide Without Code"},
					{Kind: BlockParagraph, Text: "Just text"},
				},
			},
		},
	}

	ed := NewEditor("")

	// 1. Test RunFocusedCode on Slide 1
	ed.SlideIdx = 0
	ed.BlockIdx = 1 // Focused on BlockCode
	cmd := ed.RunFocusedCode(&d)
	if cmd == nil {
		t.Fatalf("expected non-nil cmd from RunFocusedCode")
	}
	if !ed.RunningCode {
		t.Fatalf("expected RunningCode true")
	}
	// Execute cmd
	msg := cmd()
	execMsg, ok := msg.(ExecFinishedMsg)
	if !ok {
		t.Fatalf("expected ExecFinishedMsg, got %T", msg)
	}
	if !strings.Contains(execMsg.Result.Stdout, "running from editor") {
		t.Fatalf("unexpected stdout: %s", execMsg.Result.Stdout)
	}

	// 2. Test RunFocusedCode on slide without code
	ed.SlideIdx = 1
	ed.BlockIdx = 0
	nilCmd := ed.RunFocusedCode(&d)
	if nilCmd != nil {
		t.Fatalf("expected nil cmd when no code block exists")
	}
	if !strings.Contains(ed.Message, "no code block") {
		t.Fatalf("unexpected message: %s", ed.Message)
	}

	// 3. Test keyboard shortcuts: 'X' and 'ctrl+x'
	ed.SlideIdx = 0
	ed.BlockIdx = 1
	cmdX := sendTestKey(&ed, &d, "X")
	if cmdX == nil {
		t.Fatalf("expected non-nil cmd on 'X'")
	}

	cmdCtrlX := sendTestKey(&ed, &d, "ctrl+x")
	if cmdCtrlX == nil {
		t.Fatalf("expected non-nil cmd on 'ctrl+x'")
	}

	// 4. Test 'x' key: on BlockCode it runs code, on BlockList it toggles task
	ed.BlockIdx = 1 // Code block
	cmdXKey := sendTestKey(&ed, &d, "x")
	if cmdXKey == nil {
		t.Fatalf("expected 'x' on BlockCode to return exec cmd")
	}

	ed.BlockIdx = 2 // Task list item
	cmdTask := sendTestKey(&ed, &d, "x")
	if cmdTask != nil {
		t.Fatalf("expected 'x' on task item to return nil cmd")
	}
	if !strings.Contains(d.Slides[0].Blocks[2].Text, "[x]") {
		t.Fatalf("expected task toggled to [x], got: %s", d.Slides[0].Blocks[2].Text)
	}

	// 5. Test Runner result display and dismissal
	ed.ShowRunner = true
	ed.RunnerResult = &execMsg.Result

	// Pressing 'y' when runner is visible yanks runner output
	yankCmd := sendTestKey(&ed, &d, "y")
	if yankCmd == nil {
		t.Fatalf("expected non-nil yankCmd")
	}
	if !strings.Contains(ed.Message, "yanked") {
		t.Fatalf("expected yank message, got: %s", ed.Message)
	}

	// Pressing 'esc' dismisses runner
	sendTestKey(&ed, &d, "esc")
	if ed.ShowRunner {
		t.Fatalf("expected ShowRunner false after esc")
	}
	if ed.Message != "runner closed" {
		t.Fatalf("unexpected message: %s", ed.Message)
	}

	// DismissRunner method directly
	ed.ShowRunner = true
	ed.DismissRunner()
	if ed.ShowRunner {
		t.Fatalf("expected ShowRunner false after DismissRunner()")
	}

	// Switching slides clears ShowRunner
	ed.ShowRunner = true
	sendTestKey(&ed, &d, "right")
	if ed.ShowRunner {
		t.Fatalf("expected ShowRunner false after right")
	}

	// 6. Test Focus Mode ('f' / 'F')
	ed.SlideIdx = 0
	ed.BlockIdx = 1
	sendTestKey(&ed, &d, "f")
	if !ed.FocusMode {
		t.Fatalf("expected FocusMode true after 'f'")
	}

	// Test scrolling inside FocusMode
	sendTestKey(&ed, &d, "j")
	if ed.FocusScroll != 1 {
		t.Fatalf("expected FocusScroll 1 after 'j', got %d", ed.FocusScroll)
	}
	sendTestKey(&ed, &d, "k")
	if ed.FocusScroll != 0 {
		t.Fatalf("expected FocusScroll 0 after 'k', got %d", ed.FocusScroll)
	}
	sendTestKey(&ed, &d, "G")
	if ed.FocusScroll != 100 {
		t.Fatalf("expected FocusScroll 100 after 'G', got %d", ed.FocusScroll)
	}
	sendTestKey(&ed, &d, "g")
	if ed.FocusScroll != 0 {
		t.Fatalf("expected FocusScroll 0 after 'g', got %d", ed.FocusScroll)
	}

	// Toggle line numbers and theme in FocusMode
	sendTestKey(&ed, &d, "L")
	if !ed.ShowLineNumbers {
		t.Fatalf("expected ShowLineNumbers true")
	}
	sendTestKey(&ed, &d, "t")
	if ed.Theme == "" {
		t.Fatalf("expected theme cycled in FocusMode")
	}

	// Run code from FocusMode
	focusRunCmd := sendTestKey(&ed, &d, "X")
	if focusRunCmd == nil {
		t.Fatalf("expected non-nil cmd from 'X' in FocusMode")
	}

	// Exit FocusMode via 'f' or 'esc'
	sendTestKey(&ed, &d, "f")
	if ed.FocusMode {
		t.Fatalf("expected FocusMode false after 'f'")
	}

	ed.ToggleFocusMode(&d)
	if !ed.FocusMode {
		t.Fatalf("expected FocusMode true after ToggleFocusMode")
	}
	sendTestKey(&ed, &d, "esc")
	if ed.FocusMode {
		t.Fatalf("expected FocusMode false after 'esc'")
	}
}

func TestEditorAudienceTracksAndModal(t *testing.T) {
	src := `---
title: Tracks Test
---

::id s1
::tags arch
# Slide 1 (Arch)

---

::id s2
::tags demo
# Slide 2 (Demo)

---

::id s3
::tags arch,demo
# Slide 3 (Both)

---

::id s4
# Slide 4 (Untagged)
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")

	// SelectTrack for "demo": current slide is s1 (arch), so it should auto-hop to s2 (demo)
	ed.SelectTrack("demo", &d)
	if ed.ActiveTrack != "demo" {
		t.Errorf("expected ActiveTrack 'demo', got %q", ed.ActiveTrack)
	}
	if ed.SlideIdx != 1 {
		t.Errorf("expected SlideIdx 1 (slide 2), got %d", ed.SlideIdx)
	}

	// NextTrackSlide hops to slide 3 (index 2) which also has tag "demo"
	if !ed.NextTrackSlide(&d) {
		t.Errorf("expected NextTrackSlide to succeed")
	}
	if ed.SlideIdx != 2 {
		t.Errorf("expected SlideIdx 2 (slide 3), got %d", ed.SlideIdx)
	}

	// NextTrackSlide wraps around back to slide 2 (index 1)
	if !ed.NextTrackSlide(&d) {
		t.Errorf("expected NextTrackSlide wrap-around to succeed")
	}
	if ed.SlideIdx != 1 {
		t.Errorf("expected SlideIdx 1 (slide 2), got %d", ed.SlideIdx)
	}

	// PrevTrackSlide steps backward to slide 3 (index 2)
	if !ed.PrevTrackSlide(&d) {
		t.Errorf("expected PrevTrackSlide to succeed")
	}
	if ed.SlideIdx != 2 {
		t.Errorf("expected SlideIdx 2 (slide 3), got %d", ed.SlideIdx)
	}

	// SelectTrack("all") clears active track
	ed.SelectTrack("all", &d)
	if ed.ActiveTrack != "" {
		t.Errorf("expected ActiveTrack to be cleared by 'all'")
	}

	// When ActiveTrack is empty, NextTrackSlide and PrevTrackSlide behave as linear slide advances
	ed.SlideIdx = 0
	if !ed.NextTrackSlide(&d) || ed.SlideIdx != 1 {
		t.Errorf("expected NextTrackSlide without active track to advance to 1")
	}
	if !ed.PrevTrackSlide(&d) || ed.SlideIdx != 0 {
		t.Errorf("expected PrevTrackSlide without active track to return to 0")
	}

	// Test nil deck handling
	var nilDeck *Deck
	if ed.NextTrackSlide(nilDeck) || ed.PrevTrackSlide(nilDeck) {
		t.Errorf("expected nil deck to return false")
	}

	// Test 'K' key to toggle Track Modal
	sendTestKey(&ed, &d, "K")
	if !ed.ShowTrackModal {
		t.Fatalf("expected ShowTrackModal true after 'K'")
	}

	// Test cursor movement in Track Modal
	sendTestKey(&ed, &d, "j")
	if ed.TrackCursor != 1 {
		t.Errorf("expected TrackCursor 1 after 'j', got %d", ed.TrackCursor)
	}
	sendTestKey(&ed, &d, "k")
	if ed.TrackCursor != 0 {
		t.Errorf("expected TrackCursor 0 after 'k', got %d", ed.TrackCursor)
	}
	sendTestKey(&ed, &d, "G")
	if ed.TrackCursor == 0 {
		t.Errorf("expected TrackCursor > 0 after 'G'")
	}
	sendTestKey(&ed, &d, "g")
	if ed.TrackCursor != 0 {
		t.Errorf("expected TrackCursor 0 after 'g', got %d", ed.TrackCursor)
	}

	// Quick select with number keys: '1' selects first tag ("arch")
	sendTestKey(&ed, &d, "1")
	if ed.ShowTrackModal {
		t.Errorf("expected modal to dismiss after '1'")
	}
	if ed.ActiveTrack != "arch" {
		t.Errorf("expected ActiveTrack 'arch' after '1', got %q", ed.ActiveTrack)
	}

	// Open modal again and select "0" (all slides)
	sendTestKey(&ed, &d, "K")
	sendTestKey(&ed, &d, "0")
	if ed.ShowTrackModal || ed.ActiveTrack != "" {
		t.Errorf("expected modal dismissed and ActiveTrack empty after '0'")
	}

	// Open modal, move cursor down and press Enter
	sendTestKey(&ed, &d, "K")
	sendTestKey(&ed, &d, "j")
	sendTestKey(&ed, &d, "enter")
	if ed.ShowTrackModal || ed.ActiveTrack == "" {
		t.Errorf("expected modal dismissed and active track applied after enter")
	}

	// Open modal and dismiss via esc
	sendTestKey(&ed, &d, "K")
	sendTestKey(&ed, &d, "esc")
	if ed.ShowTrackModal {
		t.Errorf("expected modal dismissed after esc")
	}

	// Test '[' and ']' keys in nav mode
	ed.ActiveTrack = "demo"
	ed.SlideIdx = 1
	sendTestKey(&ed, &d, "]")
	if ed.SlideIdx != 2 {
		t.Errorf("expected SlideIdx 2 after ']', got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "[")
	if ed.SlideIdx != 1 {
		t.Errorf("expected SlideIdx 1 after '[', got %d", ed.SlideIdx)
	}
}

func TestRunFocusedCodeInColumns(t *testing.T) {
	src := `---
title: Columns Code Run
---

# Slide 1

:::columns
### Col A
Some intro
:::col
### Col B
::code lang=bash
echo "inside column"
::code
:::
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")
	ed.SlideIdx = 0
	ed.BlockIdx = 1 // the BlockColumns block

	cmd := ed.RunFocusedCode(&d)
	if cmd == nil {
		t.Fatalf("expected non-nil cmd when executing code inside column")
	}
	if !ed.RunningCode {
		t.Errorf("expected RunningCode to be true")
	}
}

func TestEditorRoutesAndModal(t *testing.T) {
	src := `---
title: Routes Editor Test
routes:
  quick: intro -> outro
  thorough: intro -> middle -> outro
---

::id intro
# Intro

---

::id middle
# Middle

::branch [1] Quick Exit -> outro

---

::id outro
# Outro
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")

	// 1. Direct method calls
	ed.SelectRoute("thorough", &d)
	if ed.ActiveRoute != "thorough" || ed.RouteStep != 0 || ed.SlideIdx != 0 {
		t.Fatalf("expected route thorough at step 0, slide 0; got route=%q step=%d slide=%d", ed.ActiveRoute, ed.RouteStep, ed.SlideIdx)
	}

	// Step forward
	if !ed.NextRouteSlide(&d) || ed.RouteStep != 1 || ed.SlideIdx != 1 {
		t.Fatalf("expected NextRouteSlide to step 1 slide 1, got step=%d slide=%d", ed.RouteStep, ed.SlideIdx)
	}
	if !ed.NextRouteSlide(&d) || ed.RouteStep != 2 || ed.SlideIdx != 2 {
		t.Fatalf("expected NextRouteSlide to step 2 slide 2, got step=%d slide=%d", ed.RouteStep, ed.SlideIdx)
	}
	// At end of route, NextRouteSlide should return false
	if ed.NextRouteSlide(&d) {
		t.Errorf("expected NextRouteSlide to return false at end of route")
	}

	// Step backwards
	if !ed.PrevRouteSlide(&d) || ed.RouteStep != 1 || ed.SlideIdx != 1 {
		t.Fatalf("expected PrevRouteSlide to step 1 slide 1, got step=%d slide=%d", ed.RouteStep, ed.SlideIdx)
	}
	if !ed.PrevRouteSlide(&d) || ed.RouteStep != 0 || ed.SlideIdx != 0 {
		t.Fatalf("expected PrevRouteSlide to step 0 slide 0, got step=%d slide=%d", ed.RouteStep, ed.SlideIdx)
	}
	// At start of route, PrevRouteSlide should return false
	if ed.PrevRouteSlide(&d) {
		t.Errorf("expected PrevRouteSlide to return false at start of route")
	}

	// Clear route
	ed.SelectRoute("", &d)
	if ed.ActiveRoute != "" || ed.RouteStep != 0 {
		t.Errorf("expected route cleared, got %q", ed.ActiveRoute)
	}

	// Unknown route
	ed.SelectRoute("nonexistent", &d)
	if ed.ActiveRoute != "" {
		t.Errorf("expected nonexistent route to be rejected")
	}

	// 2. Route Modal Keyboard Dispatch
	// Press 'P' to open route modal
	sendTestKey(&ed, &d, "P")
	if !ed.ShowRouteModal {
		t.Fatalf("expected ShowRouteModal true after 'P'")
	}

	// j / k navigation in modal
	sendTestKey(&ed, &d, "j")
	if ed.RouteCursor != 1 {
		t.Errorf("expected RouteCursor 1 after 'j', got %d", ed.RouteCursor)
	}
	sendTestKey(&ed, &d, "k")
	if ed.RouteCursor != 0 {
		t.Errorf("expected RouteCursor 0 after 'k', got %d", ed.RouteCursor)
	}

	// Press '1' to activate first route ("quick")
	sendTestKey(&ed, &d, "1")
	if ed.ShowRouteModal || ed.ActiveRoute != "quick" {
		t.Errorf("expected modal closed and active route quick after '1', got route=%q", ed.ActiveRoute)
	}

	// Open modal, press '0' to clear route
	sendTestKey(&ed, &d, "P")
	sendTestKey(&ed, &d, "0")
	if ed.ShowRouteModal || ed.ActiveRoute != "" {
		t.Errorf("expected modal closed and active route empty after '0'")
	}

	// Open modal, navigate with down, press enter to select route
	sendTestKey(&ed, &d, "P")
	sendTestKey(&ed, &d, "down")
	sendTestKey(&ed, &d, "enter")
	if ed.ShowRouteModal || ed.ActiveRoute == "" {
		t.Errorf("expected modal closed and route selected after enter")
	}

	// Dismiss via esc
	sendTestKey(&ed, &d, "P")
	sendTestKey(&ed, &d, "esc")
	if ed.ShowRouteModal {
		t.Errorf("expected modal closed after esc")
	}

	// Dismiss via q
	sendTestKey(&ed, &d, "P")
	sendTestKey(&ed, &d, "q")
	if ed.ShowRouteModal {
		t.Errorf("expected modal closed after q")
	}

	// Dismiss via P toggle
	sendTestKey(&ed, &d, "P")
	sendTestKey(&ed, &d, "P")
	if ed.ShowRouteModal {
		t.Errorf("expected modal closed after P toggle")
	}

	// 3. Arrow / Space key navigation with active route
	ed.SelectRoute("thorough", &d)
	if ed.SlideIdx != 0 {
		t.Errorf("expected SlideIdx 0, got %d", ed.SlideIdx)
	}
	// space advances along route
	sendTestKey(&ed, &d, "space")
	if ed.SlideIdx != 1 || ed.RouteStep != 1 {
		t.Errorf("expected SlideIdx 1, RouteStep 1 after space, got slide=%d step=%d", ed.SlideIdx, ed.RouteStep)
	}
	// left moves back along route
	sendTestKey(&ed, &d, "left")
	if ed.SlideIdx != 0 || ed.RouteStep != 0 {
		t.Errorf("expected SlideIdx 0, RouteStep 0 after left, got slide=%d step=%d", ed.SlideIdx, ed.RouteStep)
	}

	// While in route on middle slide (index 1), taking decision branch [1] works
	sendTestKey(&ed, &d, "right") // move to slide 1
	if ed.SlideIdx != 1 {
		t.Fatalf("expected SlideIdx 1, got %d", ed.SlideIdx)
	}
	sendTestKey(&ed, &d, "1") // take branch 1 to outro
	if ed.SlideIdx != 2 {
		t.Errorf("expected branch 1 to take us to slide 2, got %d", ed.SlideIdx)
	}
	// Backspace pops back to slide 1
	sendTestKey(&ed, &d, "backspace")
	if ed.SlideIdx != 1 {
		t.Errorf("expected backspace to pop to slide 1, got %d", ed.SlideIdx)
	}
}

func TestEditorHistoryModal(t *testing.T) {
	src := `---
title: History Editor Test
---

::id s1
# Slide 1
::branch [1] To Slide 2 -> s2

---

::id s2
# Slide 2
::branch [1] To Slide 3 -> s3

---

::id s3
# Slide 3
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")

	// 1. Direct JumpToHistory calls
	ed.History = []int{0, 1}
	ed.SlideIdx = 2

	// Out of bounds step index
	if ed.JumpToHistory(-1, &d) || ed.JumpToHistory(5, &d) {
		t.Errorf("expected out of bounds JumpToHistory to return false")
	}

	// Rewind to step 0 (slide 0)
	if !ed.JumpToHistory(0, &d) {
		t.Fatalf("expected JumpToHistory(0) to succeed")
	}
	if ed.SlideIdx != 0 || len(ed.History) != 0 {
		t.Fatalf("expected rewound to slide 0 with empty history, got slide=%d hist=%v", ed.SlideIdx, ed.History)
	}

	// 2. Traversal & History Modal key interactions
	// Build history: slide 0 -> branch 1 -> slide 1 -> branch 1 -> slide 2
	sendTestKey(&ed, &d, "1")
	if ed.SlideIdx != 1 || len(ed.History) != 1 {
		t.Fatalf("expected on slide 1 with 1 history entry, got slide=%d hist=%v", ed.SlideIdx, ed.History)
	}
	sendTestKey(&ed, &d, "1")
	if ed.SlideIdx != 2 || len(ed.History) != 2 {
		t.Fatalf("expected on slide 2 with 2 history entries, got slide=%d hist=%v", ed.SlideIdx, ed.History)
	}

	// Press 'H' to open History Modal
	sendTestKey(&ed, &d, "H")
	if !ed.ShowHistoryModal {
		t.Fatalf("expected ShowHistoryModal true after 'H'")
	}
	if ed.HistoryCursor != 2 {
		t.Errorf("expected HistoryCursor 2, got %d", ed.HistoryCursor)
	}

	// Navigate up/down with k/j
	sendTestKey(&ed, &d, "k")
	if ed.HistoryCursor != 1 {
		t.Errorf("expected HistoryCursor 1 after 'k', got %d", ed.HistoryCursor)
	}
	sendTestKey(&ed, &d, "k")
	if ed.HistoryCursor != 0 {
		t.Errorf("expected HistoryCursor 0 after second 'k', got %d", ed.HistoryCursor)
	}
	sendTestKey(&ed, &d, "k")
	if ed.HistoryCursor != 0 {
		t.Errorf("expected HistoryCursor clamped at 0, got %d", ed.HistoryCursor)
	}
	sendTestKey(&ed, &d, "j")
	if ed.HistoryCursor != 1 {
		t.Errorf("expected HistoryCursor 1 after 'j', got %d", ed.HistoryCursor)
	}

	// Quick jump via '1' in modal to step 0 (slide 0)
	sendTestKey(&ed, &d, "1")
	if ed.ShowHistoryModal || ed.SlideIdx != 0 || len(ed.History) != 0 {
		t.Fatalf("expected modal closed and rewound to slide 0 after '1', got slide=%d hist=%v", ed.SlideIdx, ed.History)
	}

	// Rebuild history and test Enter on cursor
	sendTestKey(&ed, &d, "1")  // to slide 1
	sendTestKey(&ed, &d, "1")  // to slide 2
	sendTestKey(&ed, &d, "H")  // open modal
	sendTestKey(&ed, &d, "up") // cursor on step 1 (slide 1)
	sendTestKey(&ed, &d, "enter")
	if ed.ShowHistoryModal || ed.SlideIdx != 1 || len(ed.History) != 1 {
		t.Fatalf("expected rewound to slide 1 after Enter, got slide=%d hist=%v", ed.SlideIdx, ed.History)
	}

	// Test backspace inside modal
	sendTestKey(&ed, &d, "H")
	sendTestKey(&ed, &d, "backspace")
	if ed.SlideIdx != 0 || len(ed.History) != 0 {
		t.Fatalf("expected backspace in modal to pop history, got slide=%d hist=%v", ed.SlideIdx, ed.History)
	}

	// Test clear history via 'c'
	ed.History = []int{0, 1}
	sendTestKey(&ed, &d, "c")
	if len(ed.History) != 0 || ed.ShowHistoryModal {
		t.Fatalf("expected history cleared after 'c', got %v", ed.History)
	}

	// Test dismiss via esc / q / H
	sendTestKey(&ed, &d, "H")
	sendTestKey(&ed, &d, "esc")
	if ed.ShowHistoryModal {
		t.Errorf("expected modal closed after esc")
	}

	sendTestKey(&ed, &d, "H")
	sendTestKey(&ed, &d, "q")
	if ed.ShowHistoryModal {
		t.Errorf("expected modal closed after q")
	}

	sendTestKey(&ed, &d, "H")
	sendTestKey(&ed, &d, "H")
	if ed.ShowHistoryModal {
		t.Errorf("expected modal closed after H toggle")
	}
}

func TestEditorBranchHUDModal(t *testing.T) {
	src := `---
title: Branch HUD Test
---

::id hub
# Hub Slide
::branch [1] Path A -> target-a
::branch [2] Path B -> target-b

---

::id target-a
# Target A
Path A details

---

::id target-b
# Target B
Path B details

---

::id terminal
# Terminal Slide
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")

	// 1. Boundary & Helper tests
	ed.ToggleBranchHUD(nil) // nil deck safe
	if ed.ShowBranchHUD {
		t.Fatalf("expected ShowBranchHUD false for nil deck")
	}

	// JumpToForkOption with invalid index
	if ed.JumpToForkOption(BranchForkOption{TargetIndex: -1}, &d) || ed.JumpToForkOption(BranchForkOption{TargetIndex: 99}, &d) {
		t.Fatalf("expected JumpToForkOption to fail on invalid TargetIndex")
	}

	// 2. Open Branch HUD on Hub slide with 'J'
	sendTestKey(&ed, &d, "J")
	if !ed.ShowBranchHUD {
		t.Fatalf("expected ShowBranchHUD true after 'J'")
	}
	if ed.BranchHUDCursor != 0 {
		t.Errorf("expected initial BranchHUDCursor 0, got %d", ed.BranchHUDCursor)
	}

	// 3. Navigation inside HUD
	sendTestKey(&ed, &d, "down")
	if ed.BranchHUDCursor != 1 {
		t.Errorf("expected cursor 1 after 'down', got %d", ed.BranchHUDCursor)
	}
	sendTestKey(&ed, &d, "j")
	if ed.BranchHUDCursor != 1 {
		t.Errorf("expected cursor clamped at 1, got %d", ed.BranchHUDCursor)
	}
	sendTestKey(&ed, &d, "up")
	if ed.BranchHUDCursor != 0 {
		t.Errorf("expected cursor 0 after 'up', got %d", ed.BranchHUDCursor)
	}
	sendTestKey(&ed, &d, "k")
	if ed.BranchHUDCursor != 0 {
		t.Errorf("expected cursor clamped at 0, got %d", ed.BranchHUDCursor)
	}
	sendTestKey(&ed, &d, "G")
	if ed.BranchHUDCursor != 1 {
		t.Errorf("expected cursor 1 after 'G', got %d", ed.BranchHUDCursor)
	}
	sendTestKey(&ed, &d, "g")
	if ed.BranchHUDCursor != 0 {
		t.Errorf("expected cursor 0 after 'g', got %d", ed.BranchHUDCursor)
	}

	// 4. Dismissal via esc, q, J
	sendTestKey(&ed, &d, "esc")
	if ed.ShowBranchHUD {
		t.Fatalf("expected HUD closed after 'esc'")
	}
	sendTestKey(&ed, &d, "J")
	sendTestKey(&ed, &d, "q")
	if ed.ShowBranchHUD {
		t.Fatalf("expected HUD closed after 'q'")
	}
	sendTestKey(&ed, &d, "J")
	sendTestKey(&ed, &d, "J")
	if ed.ShowBranchHUD {
		t.Fatalf("expected HUD closed after toggle 'J'")
	}

	// 5. Jump via Enter on selected option (Target A at index 1)
	sendTestKey(&ed, &d, "J")
	sendTestKey(&ed, &d, "enter")
	if ed.ShowBranchHUD || ed.SlideIdx != 1 {
		t.Fatalf("expected jumped to slide 1 (Target A), got slide=%d", ed.SlideIdx)
	}
	if len(ed.History) != 1 || ed.History[0] != 0 {
		t.Fatalf("expected history to record slide 0, got %v", ed.History)
	}

	// 6. Jump via numeric key (e.g. '2' for Path B)
	ed.SlideIdx = 0
	ed.History = nil
	sendTestKey(&ed, &d, "J")
	sendTestKey(&ed, &d, "2")
	if ed.ShowBranchHUD || ed.SlideIdx != 2 {
		t.Fatalf("expected jumped to slide 2 (Target B) via '2', got slide=%d", ed.SlideIdx)
	}

	// 7. On terminal slide (slide 3)
	ed.SlideIdx = 3
	sendTestKey(&ed, &d, "J")
	if ed.ShowBranchHUD {
		t.Fatalf("expected HUD not to open on terminal slide")
	}
	if !strings.Contains(ed.Message, "terminal slide") {
		t.Errorf("expected terminal slide message, got %q", ed.Message)
	}
}

func TestEditorWaypointModal(t *testing.T) {
	src := `---
title: Waypoint Editor Test
---

::id s1
# Slide 1
::branch [1] To Step 2 -> s2

---

::id s2
# Slide 2
::next s3

---

::id s3
# Slide 3
Target Goal
::next s1

---

::id isolated
# Isolated Slide
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")

	// 1. Boundary & Helper tests
	ed.ToggleWaypointModal(nil)
	if ed.ShowWaypointModal {
		t.Fatalf("expected ShowWaypointModal false for nil deck")
	}

	singleDeck := Deck{Slides: []Slide{{}}}
	ed.ToggleWaypointModal(&singleDeck)
	if ed.ShowWaypointModal || !strings.Contains(ed.Message, "multiple slides") {
		t.Fatalf("expected multiple slides message for single deck")
	}

	// ApplyWaypointPath & StepWaypointPath with unreachable candidate
	unreachCand := WaypointCandidate{SlideIndex: 3, Reachable: false}
	if ed.ApplyWaypointPath(unreachCand, &d) || ed.StepWaypointPath(unreachCand, &d) {
		t.Fatalf("expected failure for unreachable candidate")
	}

	// 2. Open modal with 'W'
	sendTestKey(&ed, &d, "W")
	if !ed.ShowWaypointModal {
		t.Fatalf("expected ShowWaypointModal true after 'W'")
	}
	if ed.WaypointCursor != 0 {
		t.Errorf("expected initial WaypointCursor 0, got %d", ed.WaypointCursor)
	}

	// 3. Navigation with down/up/j/k/G/g
	sendTestKey(&ed, &d, "down")
	if ed.WaypointCursor != 1 {
		t.Errorf("expected cursor 1 after 'down', got %d", ed.WaypointCursor)
	}
	sendTestKey(&ed, &d, "j")
	if ed.WaypointCursor != 2 {
		t.Errorf("expected cursor 2 after 'j', got %d", ed.WaypointCursor)
	}
	sendTestKey(&ed, &d, "j")
	if ed.WaypointCursor != 2 {
		t.Errorf("expected cursor clamped at 2, got %d", ed.WaypointCursor)
	}
	sendTestKey(&ed, &d, "up")
	if ed.WaypointCursor != 1 {
		t.Errorf("expected cursor 1 after 'up', got %d", ed.WaypointCursor)
	}
	sendTestKey(&ed, &d, "g")
	if ed.WaypointCursor != 0 {
		t.Errorf("expected cursor 0 after 'g', got %d", ed.WaypointCursor)
	}
	sendTestKey(&ed, &d, "G")
	if ed.WaypointCursor != 2 {
		t.Errorf("expected cursor 2 after 'G', got %d", ed.WaypointCursor)
	}

	// 4. Dismiss via esc / W
	sendTestKey(&ed, &d, "esc")
	if ed.ShowWaypointModal {
		t.Fatalf("expected modal closed after 'esc'")
	}
	sendTestKey(&ed, &d, "W")
	sendTestKey(&ed, &d, "W")
	if ed.ShowWaypointModal {
		t.Fatalf("expected modal closed after toggle 'W'")
	}

	// 5. Query filtering with typing and backspace
	sendTestKey(&ed, &d, "W")
	sendTestKey(&ed, &d, "3")
	if ed.WaypointQuery != "3" {
		t.Errorf("expected query '3', got %q", ed.WaypointQuery)
	}
	sendTestKey(&ed, &d, "backspace")
	if ed.WaypointQuery != "" {
		t.Errorf("expected empty query after backspace, got %q", ed.WaypointQuery)
	}
	// Backspace on empty query closes modal
	sendTestKey(&ed, &d, "backspace")
	if ed.ShowWaypointModal {
		t.Fatalf("expected modal closed after backspace on empty query")
	}

	// 6. Enter on unreachable candidate produces message
	sendTestKey(&ed, &d, "W")
	sendTestKey(&ed, &d, "G") // cursor on isolated slide
	sendTestKey(&ed, &d, "enter")
	if !strings.Contains(ed.Message, "not reachable downstream") {
		t.Errorf("expected not reachable message, got %q", ed.Message)
	}
	if !ed.ShowWaypointModal {
		t.Errorf("expected modal to stay open on invalid target")
	}

	// 7. Enter on reachable candidate applies route
	sendTestKey(&ed, &d, "g") // cursor on slide 1 (s2)
	sendTestKey(&ed, &d, "enter")
	if ed.ShowWaypointModal {
		t.Fatalf("expected modal closed after valid Enter")
	}
	if !strings.Contains(ed.ActiveRoute, "waypoint-") {
		t.Errorf("expected active waypoint route, got %q", ed.ActiveRoute)
	}

	// 8. Test step 1 hop via 'w'
	ed.SlideIdx = 0
	ed.History = nil
	sendTestKey(&ed, &d, "W")
	sendTestKey(&ed, &d, "g") // target slide 1
	sendTestKey(&ed, &d, "w")
	if ed.ShowWaypointModal || ed.SlideIdx != 1 {
		t.Fatalf("expected stepped to slide 1 via 'w', got slide=%d", ed.SlideIdx)
	}
	if len(ed.History) != 1 || ed.History[0] != 0 {
		t.Fatalf("expected history to record slide 0, got %v", ed.History)
	}
}

func TestEditorRadarAndForkReturn(t *testing.T) {
	// 1. Boundary cases
	ed := NewEditor("test.deck.md")
	if ed.ReturnToUpstreamFork(nil) {
		t.Fatalf("expected false for nil deck")
	}
	emptyDeck := Deck{}
	if ed.ReturnToUpstreamFork(&emptyDeck) {
		t.Fatalf("expected false for empty deck")
	}
	if ed.FindLastForkIndex(nil) != -1 || ed.FindLastForkIndex(&emptyDeck) != -1 {
		t.Fatalf("expected -1 for empty deck fork search")
	}

	ed.ToggleRadarModal(nil)
	if ed.ShowRadarModal {
		t.Fatalf("expected ShowRadarModal false for nil deck")
	}
	ed.ToggleRadarModal(&emptyDeck)
	if ed.ShowRadarModal {
		t.Fatalf("expected ShowRadarModal false for empty deck")
	}
	if ed.JumpToRadarBranch(BranchCoverageItem{TargetIdx: -1}, &emptyDeck) {
		t.Fatalf("expected false for invalid branch target")
	}

	// 2. Setup presentation with decision hub
	src := `---
title: Radar & Fork Deck
---

::id hub
# Hub Slide
::branch [1] Arch -> arch
::branch [2] Runner -> runner

---

::id arch
# Arch Slide
::next conclusion

---

::id runner
# Runner Slide
::next conclusion

---

::id conclusion
# Conclusion
`
	d := ParseDeck(src)
	ed = NewEditor("test.deck.md")

	// 3. Open Radar Modal with 'V'
	sendTestKey(&ed, &d, "V")
	if !ed.ShowRadarModal {
		t.Fatalf("expected ShowRadarModal true after 'V'")
	}
	if ed.RadarCursor != 0 {
		t.Fatalf("expected initial cursor 0, got %d", ed.RadarCursor)
	}

	// Navigation within modal
	sendTestKey(&ed, &d, "down")
	if ed.RadarCursor != 1 {
		t.Fatalf("expected cursor 1 after down, got %d", ed.RadarCursor)
	}
	sendTestKey(&ed, &d, "up")
	if ed.RadarCursor != 0 {
		t.Fatalf("expected cursor 0 after up, got %d", ed.RadarCursor)
	}
	sendTestKey(&ed, &d, "G")
	if ed.RadarCursor != 1 {
		t.Fatalf("expected cursor 1 after G, got %d", ed.RadarCursor)
	}
	sendTestKey(&ed, &d, "g")
	if ed.RadarCursor != 0 {
		t.Fatalf("expected cursor 0 after g, got %d", ed.RadarCursor)
	}

	// Dismiss via 'V'
	sendTestKey(&ed, &d, "V")
	if ed.ShowRadarModal {
		t.Fatalf("expected modal closed after toggle V")
	}

	// 4. Select branch 2 via numeric key '2' inside radar modal
	sendTestKey(&ed, &d, "V")
	sendTestKey(&ed, &d, "2")
	if ed.ShowRadarModal {
		t.Fatalf("expected modal closed after selecting branch '2'")
	}
	if ed.SlideIdx != 2 {
		t.Fatalf("expected jumped to runner (slide 2), got slide=%d", ed.SlideIdx)
	}
	if len(ed.History) != 1 || ed.History[0] != 0 {
		t.Fatalf("expected history to contain hub (0), got %v", ed.History)
	}

	// 5. Test ReturnToUpstreamFork via 'U'
	sendTestKey(&ed, &d, "U")
	if ed.SlideIdx != 0 {
		t.Fatalf("expected returned to hub (slide 0), got slide=%d", ed.SlideIdx)
	}
	if !strings.Contains(ed.Message, "returned to upstream fork") {
		t.Fatalf("expected return message, got %q", ed.Message)
	}

	// Pressing U again when no forks in history
	sendTestKey(&ed, &d, "U")
	if !strings.Contains(ed.Message, "no prior fork found") {
		t.Fatalf("expected no prior fork message, got %q", ed.Message)
	}

	// 6. Test 'u' inside Radar Modal
	// Jump to slide 1 (arch) with history [0]
	ed.SlideIdx = 1
	ed.History = []int{0}
	sendTestKey(&ed, &d, "V")
	sendTestKey(&ed, &d, "u")
	if ed.ShowRadarModal {
		t.Fatalf("expected modal closed after 'u'")
	}
	if ed.SlideIdx != 0 {
		t.Fatalf("expected jumped to fork slide 0 via 'u', got %d", ed.SlideIdx)
	}

	// 7. Test Enter on focused branch in Radar Modal
	sendTestKey(&ed, &d, "V")
	sendTestKey(&ed, &d, "g") // cursor on branch 1 (Arch)
	sendTestKey(&ed, &d, "enter")
	if ed.ShowRadarModal || ed.SlideIdx != 1 {
		t.Fatalf("expected jumped to slide 1 via enter, got slide=%d", ed.SlideIdx)
	}
	if !strings.Contains(ed.Message, "jumped to branch [1]") {
		t.Fatalf("expected branch jump message, got %q", ed.Message)
	}
}

func TestEditorLoopIterationAndAutoExit(t *testing.T) {
	// 1. Boundary cases
	ed := NewEditor("test.deck.md")
	if pass, maxP, hasLoop := ed.CurrentLoopPass(0, nil); hasLoop || pass != 0 || maxP != 0 {
		t.Fatalf("expected hasLoop false for nil deck")
	}
	emptyDeck := Deck{}
	if pass, maxP, hasLoop := ed.CurrentLoopPass(0, &emptyDeck); hasLoop || pass != 0 || maxP != 0 {
		t.Fatalf("expected hasLoop false for empty deck")
	}
	if took, _ := ed.AdvanceLoop(0, nil); took {
		t.Fatalf("expected tookLoop false for nil deck")
	}
	if took, _ := ed.AdvanceLoop(0, &emptyDeck); took {
		t.Fatalf("expected tookLoop false for empty deck")
	}
	ed.ResetLoopCounter(99) // Should not panic on nonexistent slide
	ed.ResetAllLoops()

	// 2. Setup presentation deck with loop
	src := `---
title: TDD Iteration Loop
---

::id intro
# Intro

---

::id tdd-red
# Red: Failing Test

---

::id tdd-green
# Green: Make Pass

---

::id tdd-refactor
# Refactor: Clean Code
::loop [r] Red-Green-Refactor -> tdd-red max=3 next=summary

---

::id summary
# Presentation Summary
`
	d := ParseDeck(src)
	ed = NewEditor("test.deck.md")

	// 3. Navigate to refactor slide (slide 3)
	ed.SlideIdx = 3
	pass, maxPasses, hasLoop := ed.CurrentLoopPass(3, &d)
	if !hasLoop || pass != 0 || maxPasses != 3 {
		t.Fatalf("expected pass 0/3, got %d/%d (hasLoop=%v)", pass, maxPasses, hasLoop)
	}

	// 4. First pass: Space key loops back to tdd-red (slide 1)
	sendTestKey(&ed, &d, " ")
	if ed.SlideIdx != 1 {
		t.Fatalf("expected loop pass 1 to route to tdd-red (slide 1), got slide %d", ed.SlideIdx)
	}
	if !strings.Contains(ed.Message, "pass 1/3") {
		t.Errorf("expected pass 1/3 in message, got %q", ed.Message)
	}

	// 5. Progress to refactor again, second pass via 'enter'
	ed.SlideIdx = 3
	sendTestKey(&ed, &d, "enter")
	if ed.SlideIdx != 1 {
		t.Fatalf("expected loop pass 2 to route to tdd-red (slide 1), got slide %d", ed.SlideIdx)
	}
	if !strings.Contains(ed.Message, "pass 2/3") {
		t.Errorf("expected pass 2/3 in message, got %q", ed.Message)
	}

	// 6. Progress to refactor again, third pass via 'right'
	ed.SlideIdx = 3
	sendTestKey(&ed, &d, "right")
	if ed.SlideIdx != 1 {
		t.Fatalf("expected loop pass 3 to route to tdd-red (slide 1), got slide %d", ed.SlideIdx)
	}
	if !strings.Contains(ed.Message, "pass 3/3") {
		t.Errorf("expected pass 3/3 in message, got %q", ed.Message)
	}

	// 7. Loop exhaustion: when pass count reaches 3, next advance auto-exits to 'summary' (slide 4)
	ed.SlideIdx = 3
	pass, maxPasses, hasLoop = ed.CurrentLoopPass(3, &d)
	if pass != 3 || maxPasses != 3 {
		t.Fatalf("expected pass 3/3, got %d/%d", pass, maxPasses)
	}
	sendTestKey(&ed, &d, " ")
	if ed.SlideIdx != 4 {
		t.Fatalf("expected loop to auto-exit to summary (slide 4), got slide %d", ed.SlideIdx)
	}
	if !strings.Contains(ed.Message, "completed") {
		t.Errorf("expected completed in message, got %q", ed.Message)
	}

	// 8. Test loop hotkey 'r'
	ed.ResetLoopCounter(3)
	ed.SlideIdx = 3
	sendTestKey(&ed, &d, "r")
	if ed.SlideIdx != 1 {
		t.Fatalf("expected 'r' hotkey to trigger loop pass to slide 1, got slide %d", ed.SlideIdx)
	}

	// 9. Test restart on 'g' resets all loops
	ed.SlideIdx = 3
	sendTestKey(&ed, &d, " ") // increments counter
	sendTestKey(&ed, &d, "g")
	if ed.SlideIdx != 0 {
		t.Fatalf("expected 'g' to navigate to slide 0, got %d", ed.SlideIdx)
	}
	if p, _, _ := ed.CurrentLoopPass(3, &d); p != 0 {
		t.Fatalf("expected loop counter reset to 0 after 'g', got %d", p)
	}

	// 10. Test loop without explicit exit target (falls back to next slide)
	srcFallback := `---
title: Fallback Loop
---
::id start
# Start

---
::id loop-slide
# Looper
::loop start max=1

---
::id end
# End
`
	dFallback := ParseDeck(srcFallback)
	edFallback := NewEditor("test.deck.md")
	edFallback.SlideIdx = 1

	// Pass 1 -> start
	sendTestKey(&edFallback, &dFallback, " ")
	if edFallback.SlideIdx != 0 {
		t.Fatalf("expected loop to slide 0, got %d", edFallback.SlideIdx)
	}

	// Return to slide 1, pass exhausted -> linear next (slide 2)
	edFallback.SlideIdx = 1
	sendTestKey(&edFallback, &dFallback, " ")
	if edFallback.SlideIdx != 2 {
		t.Fatalf("expected fallback loop exit to slide 2, got %d", edFallback.SlideIdx)
	}
}
