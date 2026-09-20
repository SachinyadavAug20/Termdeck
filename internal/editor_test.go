package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	for _, k := range []string{"i", "a", "o", "I", "A", "O"} {
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
}
