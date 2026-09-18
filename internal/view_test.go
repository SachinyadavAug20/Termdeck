package internal

import (
	"strings"
	"testing"
)

func TestRenderHeadings(t *testing.T) {
	h1 := renderHeading("Title", 1)
	if strings.Contains(h1, "\n") {
		t.Errorf("expected h1 to contain no newlines, got %q", h1)
	}

	h2 := renderHeading("H2", 2)
	if strings.Contains(h2, "\n") {
		t.Errorf("expected h2 to contain no newlines, got %q", h2)
	}
}

func TestRenderLaserPointer(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Slide Title"},
					{Kind: BlockParagraph, Text: "Paragraph text"},
					{Kind: BlockImage, Src: "demo.png"},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")
	ed.BlockIdx = 0 // Pointing at heading

	out := View(d, ed, 80, 24)
	if !strings.Contains(out, "▶") {
		t.Errorf("expected laser pointer '▶' in view output")
	}

	// Make sure the pointer is right in front of the heading, not on a blank line above
	lines := strings.Split(out, "\n")
	foundPointer := false
	for _, l := range lines {
		if strings.Contains(l, "▶") {
			foundPointer = true
			if !strings.Contains(l, "Slide Title") {
				t.Errorf("expected laser pointer to be on same line as 'Slide Title', got %q", l)
			}
			break
		}
	}
	if !foundPointer {
		t.Errorf("did not find laser pointer line in view output")
	}
}

func TestRenderImageCard(t *testing.T) {
	card := renderImageCard("demo.png", "..", 60)
	if !strings.Contains(card, "demo.png") {
		t.Errorf("expected image card to contain filename 'demo.png', got %q", card)
	}
	if !strings.Contains(card, "🖼") {
		t.Errorf("expected image card to contain icon '🖼', got %q", card)
	}
}

func TestViewAlignment(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Left Aligned"},
				},
				Align: AlignLeft,
			},
		},
	}
	ed := NewEditor("test.deck.md")

	out := View(d, ed, 80, 24)
	if !strings.Contains(out, "Left Aligned") {
		t.Errorf("expected view output to contain 'Left Aligned'")
	}
	if !strings.Contains(out, "(left)") {
		t.Errorf("expected status bar to show '(left)', got %q", out)
	}
}
