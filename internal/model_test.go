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

func TestParseDeckEdgeCases(t *testing.T) {
	// 1. Empty and whitespace-only strings
	dEmpty := ParseDeck("")
	if len(dEmpty.Slides) != 0 {
		t.Errorf("expected 0 slides for empty string, got %d", len(dEmpty.Slides))
	}
	dSpaces := ParseDeck("   \n\n\t  \n")
	if len(dSpaces.Slides) != 0 {
		t.Errorf("expected 0 slides for whitespace string, got %d", len(dSpaces.Slides))
	}

	// 2. Deck with NO frontmatter
	srcNoFront := "# Slide One\n\nSome text.\n---\n# Slide Two\n"
	dNoFront := ParseDeck(srcNoFront)
	if len(dNoFront.Slides) != 2 {
		t.Fatalf("expected 2 slides without frontmatter, got %d", len(dNoFront.Slides))
	}
	if dNoFront.Slides[0].Blocks[0].Text != "Slide One" {
		t.Errorf("expected Slide One, got %q", dNoFront.Slides[0].Blocks[0].Text)
	}

	// 3. Numbered lists and asterisk lists
	srcLists := `---
title: Lists
---
1. First item
2. Second item
* Asterisk item
`
	dLists := ParseDeck(srcLists)
	if len(dLists.Slides) != 1 {
		t.Fatalf("expected 1 slide for lists, got %d", len(dLists.Slides))
	}
	s := dLists.Slides[0]
	if len(s.Blocks) != 3 {
		t.Fatalf("expected 3 list blocks, got %d", len(s.Blocks))
	}
	if s.Blocks[0].Kind != BlockList || s.Blocks[0].Text != "1. First item" {
		t.Errorf("expected numbered list 1, got %+v", s.Blocks[0])
	}
	if s.Blocks[1].Kind != BlockList || s.Blocks[1].Text != "2. Second item" {
		t.Errorf("expected numbered list 2, got %+v", s.Blocks[1])
	}
	if s.Blocks[2].Kind != BlockList || s.Blocks[2].Text != "* Asterisk item" {
		t.Errorf("expected asterisk list, got %+v", s.Blocks[2])
	}

	// 4. Code block closed with ::code
	srcCode := `---
title: Code
---
::code lang=python
print("hello")
::code

Normal text
`
	dCode := ParseDeck(srcCode)
	if len(dCode.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(dCode.Slides))
	}
	codeSlide := dCode.Slides[0]
	var foundCode, foundText bool
	for _, b := range codeSlide.Blocks {
		if b.Kind == BlockCode && b.Lang == "python" {
			foundCode = true
			if len(b.Lines) != 1 || b.Lines[0] != `print("hello")` {
				t.Errorf("expected code lines print(hello), got %v", b.Lines)
			}
		}
		if b.Kind == BlockParagraph && b.Text == "Normal text" {
			foundText = true
		}
	}
	if !foundCode {
		t.Errorf("expected code block")
	}
	if !foundText {
		t.Errorf("expected paragraph after closed code block")
	}

	// 5. Directives: shorthand align and custom directives
	srcDir := `---
title: Directives
---
::left
# Left Slide
::custom directive content
`
	dDir := ParseDeck(srcDir)
	if len(dDir.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(dDir.Slides))
	}
	if dDir.Slides[0].Align != AlignLeft {
		t.Errorf("expected slide align left from ::left, got %q", dDir.Slides[0].Align)
	}
	foundCustom := false
	for _, b := range dDir.Slides[0].Blocks {
		if b.Kind == BlockDirective && strings.Contains(b.Directive, "::custom") {
			foundCustom = true
		}
	}
	if !foundCustom {
		t.Errorf("expected custom directive block")
	}
}

func TestSerializeDeckEdgeCases(t *testing.T) {
	d := Deck{
		Meta: map[string]string{
			"title": "Edge Case Deck",
		},
		Align: AlignRight,
		Slides: []Slide{
			{
				Align: AlignCenter,
				Blocks: []Block{
					{Kind: BlockHeading, Level: 3, Text: "H3 Title"},
					{Kind: BlockParagraph, Text: "Body text"},
					{Kind: BlockCode, Lang: "sh", Lines: []string{"echo hello", "exit 0"}},
					{Kind: BlockList, Text: "- bullet"},
					{Kind: BlockImage, Src: "img.png"},
					{Kind: BlockDirective, Directive: "::custom directive", Raw: "::custom directive"},
				},
			},
		},
	}

	serialized := SerializeDeck(d)
	if !strings.Contains(serialized, "align: right") {
		t.Errorf("expected serialized deck to contain 'align: right', got %s", serialized)
	}
	if !strings.Contains(serialized, "::align center") {
		t.Errorf("expected serialized slide to contain '::align center', got %s", serialized)
	}
	if !strings.Contains(serialized, "### H3 Title") {
		t.Errorf("expected serialized heading '### H3 Title', got %s", serialized)
	}
	if !strings.Contains(serialized, "::code lang=sh") {
		t.Errorf("expected serialized code directive, got %s", serialized)
	}
	if !strings.Contains(serialized, "::image img.png") {
		t.Errorf("expected serialized image directive, got %s", serialized)
	}
	if !strings.Contains(serialized, "::custom directive") {
		t.Errorf("expected serialized custom directive, got %s", serialized)
	}

	// Verify roundtrip preservation
	reparsed := ParseDeck(serialized)
	if reparsed.Align != AlignRight {
		t.Errorf("expected reparsed deck align right, got %s", reparsed.Align)
	}
	if reparsed.Slides[0].Align != AlignCenter {
		t.Errorf("expected reparsed slide align center, got %s", reparsed.Slides[0].Align)
	}
}

func TestSpeakerNotesModel(t *testing.T) {
	src := `# Title

Visible paragraph

- Bullet 1

::notes
This is line 1 of notes.
This is line 2 of notes.
`
	deck := ParseDeck(src)
	if len(deck.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(deck.Slides))
	}
	slide := deck.Slides[0]
	// Slide has 3 visible blocks + 1 notes directive block
	if len(slide.Blocks) != 4 {
		t.Fatalf("expected 4 blocks total, got %d", len(slide.Blocks))
	}
	vis := slide.VisibleBlockIndices()
	if len(vis) != 3 {
		t.Fatalf("expected 3 visible blocks, got %d", len(vis))
	}
	if vis[0] != 0 || vis[1] != 1 || vis[2] != 2 {
		t.Errorf("unexpected visible block indices: %v", vis)
	}

	notes := slide.Notes()
	if !strings.Contains(notes, "This is line 1 of notes.") || !strings.Contains(notes, "This is line 2 of notes.") {
		t.Errorf("expected notes to contain both lines, got %q", notes)
	}

	// Test serialization round trip of notes
	serialized := SerializeDeck(deck)
	if !strings.Contains(serialized, "::notes\nThis is line 1 of notes.\nThis is line 2 of notes.") {
		t.Errorf("expected serialized notes, got %s", serialized)
	}
}

func BenchmarkParseDeck(b *testing.B) {
	src := `---
format: 0.1
title: Benchmarking Deck
align: center
---

# Benchmark Slide 1
This is a paragraph with **bold** and *italic* and ` + "`code`" + `.

- Item 1
- Item 2
- Item 3

---

# Benchmark Slide 2
::code lang=go
package main
func main() { println("Hello") }

::image test.png
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseDeck(src)
	}
}
