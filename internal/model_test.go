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

func TestStandardMarkdownFeatures(t *testing.T) {
	src := `# Title

` + "```python\ndef hello():\n    print('hi')\n```" + `

![System Architecture](assets/arch.png)

| Command | Action |
|---|---|
| git status | check tree |
| git commit | save commit |
`
	deck := ParseDeck(src)
	if len(deck.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(deck.Slides))
	}
	s := deck.Slides[0]

	if len(s.Blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(s.Blocks))
	}

	// 1. Standard fenced code
	bCode := s.Blocks[1]
	if bCode.Kind != BlockCode || bCode.Lang != "python" || len(bCode.Lines) != 2 {
		t.Errorf("unexpected code block: %+v", bCode)
	}

	// 2. Standard markdown image
	bImg := s.Blocks[2]
	if bImg.Kind != BlockImage || bImg.Src != "assets/arch.png" || bImg.Text != "System Architecture" {
		t.Errorf("unexpected image block: %+v", bImg)
	}

	// 3. Markdown table
	bTable := s.Blocks[3]
	if bTable.Kind != BlockTable || len(bTable.Lines) != 4 {
		t.Errorf("unexpected table block: %+v", bTable)
	}

	// 4. Test serialization of table
	ser := SerializeBlock(bTable)
	if !strings.Contains(ser, "| Command | Action |") || !strings.Contains(ser, "| git commit | save commit |") {
		t.Errorf("unexpected serialized table: %s", ser)
	}
}

func TestCalloutBlocks(t *testing.T) {
	src := `---
title: Callout Deck
---
# Slide 1

> [!TIP]
> Use connection pooling to reduce database latency.
> Keep connections warm.

> [!NOTE]
> Backward compatible with v1 API.

> [!WARNING]
> Do not execute during peak production hours.

> [!IMPORTANT]
> Requires Redis 7.0 or higher.

> [!CAUTION]
> Irreversible data migration.

> "Simplicity is prerequisite for reliability."
> — Edsger W. Dijkstra
`
	d := ParseDeck(src)
	if len(d.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(d.Slides))
	}
	s := d.Slides[0]
	// Expected blocks: Heading (1) + 6 Callouts = 7 blocks
	if len(s.Blocks) != 7 {
		t.Fatalf("expected 7 blocks, got %d", len(s.Blocks))
	}

	// 1. Tip
	bTip := s.Blocks[1]
	if bTip.Kind != BlockCallout || bTip.Callout != "tip" || len(bTip.Lines) != 2 {
		t.Errorf("expected Tip callout, got %+v", bTip)
	}

	// 2. Note
	bNote := s.Blocks[2]
	if bNote.Kind != BlockCallout || bNote.Callout != "note" || bNote.Lines[0] != "Backward compatible with v1 API." {
		t.Errorf("expected Note callout, got %+v", bNote)
	}

	// 3. Warning
	bWarn := s.Blocks[3]
	if bWarn.Kind != BlockCallout || bWarn.Callout != "warning" {
		t.Errorf("expected Warning callout, got %+v", bWarn)
	}

	// 4. Important
	bImp := s.Blocks[4]
	if bImp.Kind != BlockCallout || bImp.Callout != "important" {
		t.Errorf("expected Important callout, got %+v", bImp)
	}

	// 5. Caution
	bCst := s.Blocks[5]
	if bCst.Kind != BlockCallout || bCst.Callout != "caution" {
		t.Errorf("expected Caution callout, got %+v", bCst)
	}

	// 6. Quote
	bQuote := s.Blocks[6]
	if bQuote.Kind != BlockCallout || bQuote.Callout != "quote" || len(bQuote.Lines) != 2 {
		t.Errorf("expected Quote callout, got %+v", bQuote)
	}

	// Test serialization round-trip
	ser := SerializeDeck(d)
	d2 := ParseDeck(ser)
	if len(d2.Slides[0].Blocks) != 7 {
		t.Fatalf("expected 7 blocks after serialization round-trip, got %d", len(d2.Slides[0].Blocks))
	}
	if d2.Slides[0].Blocks[1].Callout != "tip" || d2.Slides[0].Blocks[6].Callout != "quote" {
		t.Errorf("mismatch in serialized callouts")
	}
}

func TestBlockDivider(t *testing.T) {
	src := `---
title: Divider Deck
---
# Section 1
First concept

***

Second concept

___

Third concept

::hr

Fourth concept
`
	d := ParseDeck(src)
	if len(d.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(d.Slides))
	}
	s := d.Slides[0]
	// Heading (1) + 4 Paragraphs + 3 Dividers = 8 blocks
	if len(s.Blocks) != 8 {
		t.Fatalf("expected 8 blocks, got %d", len(s.Blocks))
	}
	if s.Blocks[2].Kind != BlockDivider {
		t.Errorf("expected BlockDivider from '***', got %+v", s.Blocks[2])
	}
	if s.Blocks[4].Kind != BlockDivider {
		t.Errorf("expected BlockDivider from '___', got %+v", s.Blocks[4])
	}
	if s.Blocks[6].Kind != BlockDivider {
		t.Errorf("expected BlockDivider from '::hr', got %+v", s.Blocks[6])
	}

	ser := SerializeBlock(s.Blocks[2])
	if ser != "***" {
		t.Errorf("expected serialized divider '***', got %q", ser)
	}
}

func TestSlideSummary(t *testing.T) {
	s1 := Slide{
		Blocks: []Block{
			{Kind: BlockHeading, Level: 1, Text: "Introduction"},
			{Kind: BlockParagraph, Text: "Welcome to the talk."},
		},
	}
	if sum := s1.Summary(); sum != "2 blks" {
		t.Errorf("expected '2 blks', got %q", sum)
	}

	s2 := Slide{
		Blocks: []Block{
			{Kind: BlockHeading, Level: 1, Text: "Code Sample"},
			{Kind: BlockCode, Lang: "go", Lines: []string{"func main() {}"}},
			{Kind: BlockTable, Lines: []string{"| a | b |", "|---|---|"}},
		},
	}
	if sum := s2.Summary(); sum != "3 blks · code,table" {
		t.Errorf("expected '3 blks · code,table', got %q", sum)
	}

	s3 := Slide{
		Blocks: []Block{
			{Kind: BlockHeading, Level: 1, Text: "Sprint Tasks"},
			{Kind: BlockList, Text: "- [ ] Write tests"},
			{Kind: BlockCallout, Callout: "tip", Text: "Keep it DRY"},
			{Kind: BlockImage, Src: "demo.png"},
			{Kind: BlockDirective, Directive: "::notes reminder"},
		},
	}
	// Note block is not visible, so 4 visible blocks
	if sum := s3.Summary(); sum != "4 blks · card,img,task" {
		t.Errorf("expected '4 blks · card,img,task', got %q", sum)
	}

	sSingle := Slide{
		Blocks: []Block{
			{Kind: BlockHeading, Level: 1, Text: "Single Block"},
		},
	}
	if sum := sSingle.Summary(); sum != "1 blk" {
		t.Errorf("expected '1 blk', got %q", sum)
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
