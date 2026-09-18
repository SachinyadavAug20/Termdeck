package internal

import (
	"testing"
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
