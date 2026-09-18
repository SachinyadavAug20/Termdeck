package internal

import (
	"strings"
	"testing"
)

func TestParseDeck(t *testing.T) {
	src := `---
format: 0.1
title: Test Presentation
author: Alice
---

# Title Slide

This is a paragraph with **bold** and *italic*.

- Item 1
- Item 2

---

# Code Slide

::code lang=go
package main
func main() {}

::image test.png

::notes
Private speaker notes.
`

	deck := ParseDeck(src)

	if deck.Meta["format"] != "0.1" {
		t.Errorf("expected format 0.1, got %q", deck.Meta["format"])
	}
	if deck.Meta["title"] != "Test Presentation" {
		t.Errorf("expected title 'Test Presentation', got %q", deck.Meta["title"])
	}
	if len(deck.Slides) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(deck.Slides))
	}

	// Slide 1 checks
	s1 := deck.Slides[0]
	if len(s1.Blocks) < 4 {
		t.Fatalf("expected at least 4 blocks on slide 1, got %d", len(s1.Blocks))
	}
	if s1.Blocks[0].Kind != BlockHeading || s1.Blocks[0].Level != 1 || s1.Blocks[0].Text != "Title Slide" {
		t.Errorf("unexpected block 0: %+v", s1.Blocks[0])
	}
	if s1.Blocks[1].Kind != BlockParagraph {
		t.Errorf("expected paragraph, got %+v", s1.Blocks[1])
	}
	if s1.Blocks[2].Kind != BlockList || s1.Blocks[2].Text != "- Item 1" {
		t.Errorf("unexpected block 2: %+v", s1.Blocks[2])
	}

	// Slide 2 checks
	s2 := deck.Slides[1]
	var hasCode, hasImage, hasDirective bool
	for _, blk := range s2.Blocks {
		if blk.Kind == BlockCode {
			hasCode = true
			if blk.Lang != "go" {
				t.Errorf("expected code lang 'go', got %q", blk.Lang)
			}
		}
		if blk.Kind == BlockImage {
			hasImage = true
			if blk.Src != "test.png" {
				t.Errorf("expected image src 'test.png', got %q", blk.Src)
			}
		}
		if blk.Kind == BlockDirective && strings.HasPrefix(blk.Directive, "::notes") {
			hasDirective = true
		}
	}
	if !hasCode {
		t.Error("expected code block on slide 2")
	}
	if !hasImage {
		t.Error("expected image block on slide 2")
	}
	if !hasDirective {
		t.Error("expected directive block on slide 2")
	}
}

func TestSerializeRoundTrip(t *testing.T) {
	src := `---
format: 0.1
title: Roundtrip
---

# Heading 1

Some text.

---

::image my_pic.png
`
	deck := ParseDeck(src)
	serialized := SerializeDeck(deck)
	reparsed := ParseDeck(serialized)

	if len(reparsed.Slides) != len(deck.Slides) {
		t.Fatalf("slide count mismatch: %d vs %d", len(reparsed.Slides), len(deck.Slides))
	}
	if reparsed.Meta["title"] != deck.Meta["title"] {
		t.Errorf("title mismatch: %q vs %q", reparsed.Meta["title"], deck.Meta["title"])
	}
}

func TestParseDirective(t *testing.T) {
	k, v := ParseDirective("::code lang=python")
	if k != "code" || v != "python" {
		t.Errorf("expected code/python, got %q/%q", k, v)
	}

	k2, v2 := ParseDirective("::image /path/to/img.png")
	if k2 != "image" || v2 != "/path/to/img.png" {
		t.Errorf("expected image//path/to/img.png, got %q/%q", k2, v2)
	}
}

func TestParseAlignment(t *testing.T) {
	src := `---
format: 0.1
title: Align Deck
align: left
---

# Slide 1 (left by default)

---

::align right

# Slide 2 (right aligned)

---

::center

# Slide 3 (center aligned)
`
	deck := ParseDeck(src)
	if deck.Align != AlignLeft {
		t.Errorf("expected deck align left, got %q", deck.Align)
	}
	if len(deck.Slides) != 3 {
		t.Fatalf("expected 3 slides, got %d", len(deck.Slides))
	}
	if deck.Slides[0].Align != "" {
		t.Errorf("expected slide 0 to inherit deck align, got %q", deck.Slides[0].Align)
	}
	if deck.Slides[1].Align != AlignRight {
		t.Errorf("expected slide 1 align right, got %q", deck.Slides[1].Align)
	}
	if deck.Slides[2].Align != AlignCenter {
		t.Errorf("expected slide 2 align center, got %q", deck.Slides[2].Align)
	}

	// Serialization check
	serialized := SerializeDeck(deck)
	reparsed := ParseDeck(serialized)
	if reparsed.Align != AlignLeft {
		t.Errorf("expected reparsed deck align left, got %q", reparsed.Align)
	}
	if reparsed.Slides[1].Align != AlignRight {
		t.Errorf("expected reparsed slide 1 align right, got %q", reparsed.Slides[1].Align)
	}
}
