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

func TestBranchParsingAndModel(t *testing.T) {
	src := `---
title: Branching Deck
---

# Architecture Overview {#arch-overview}
::next summary
::prev intro
::tags arch,backend

Choose your deep dive:
::branch [1] Backend Storage Engine -> backend-storage
::fork [2] Frontend Reactive UI -> frontend-ui
-> [Concurrency Patterns](concurrency)
=> [Memory Optimizations](memory)
-> [5] Profiling Tips -> profiling
[6] Distributed Tracing -> tracing

---

::id backend-storage
# Backend Storage Engine
Details on LSM trees.

---

::id summary
# Conclusion
Wrap up.`

	deck := ParseDeck(src)
	if len(deck.Slides) != 3 {
		t.Fatalf("expected 3 slides, got %d", len(deck.Slides))
	}

	s1 := deck.Slides[0]
	if s1.ID != "arch-overview" {
		t.Errorf("expected ID 'arch-overview', got %q", s1.ID)
	}
	if s1.NextID != "summary" {
		t.Errorf("expected NextID 'summary', got %q", s1.NextID)
	}
	if s1.PrevID != "intro" {
		t.Errorf("expected PrevID 'intro', got %q", s1.PrevID)
	}
	if len(s1.Tags) != 2 || s1.Tags[0] != "arch" || s1.Tags[1] != "backend" {
		t.Errorf("unexpected tags: %+v", s1.Tags)
	}

	branches := s1.Branches()
	if len(branches) != 6 {
		t.Fatalf("expected 6 branches, got %d: %+v", len(branches), branches)
	}

	if branches[0].Key != "1" || branches[0].Target != "backend-storage" || branches[0].Label != "Backend Storage Engine" {
		t.Errorf("unexpected branch 0: %+v", branches[0])
	}
	if branches[1].Key != "2" || branches[1].Target != "frontend-ui" || branches[1].Label != "Frontend Reactive UI" {
		t.Errorf("unexpected branch 1: %+v", branches[1])
	}
	if branches[2].Target != "concurrency" || branches[2].Label != "Concurrency Patterns" {
		t.Errorf("unexpected branch 2: %+v", branches[2])
	}
	if branches[3].Target != "memory" || branches[3].Label != "Memory Optimizations" {
		t.Errorf("unexpected branch 3: %+v", branches[3])
	}
	if branches[4].Key != "5" || branches[4].Target != "profiling" {
		t.Errorf("unexpected branch 4: %+v", branches[4])
	}
	if branches[5].Key != "6" || branches[5].Target != "tracing" {
		t.Errorf("unexpected branch 5: %+v", branches[5])
	}

	// Test FindBranchByKey
	b1 := s1.FindBranchByKey("1")
	if b1 == nil || b1.Target != "backend-storage" {
		t.Errorf("FindBranchByKey('1') failed: %+v", b1)
	}
	bMissing := s1.FindBranchByKey("99")
	if bMissing != nil {
		t.Errorf("expected nil for missing key, got %+v", bMissing)
	}

	// Test Summary contains 'fork'
	if !strings.Contains(s1.Summary(), "fork") {
		t.Errorf("expected summary to contain 'fork', got %q", s1.Summary())
	}

	// Test Slug
	if s1.Slug() != "arch-overview" {
		t.Errorf("expected slug 'arch-overview', got %q", s1.Slug())
	}
	s2 := deck.Slides[1]
	if s2.Slug() != "backend-storage" {
		t.Errorf("expected slug 'backend-storage', got %q", s2.Slug())
	}
	s3 := deck.Slides[2]
	if s3.Slug() != "summary" {
		t.Errorf("expected slug 'summary', got %q", s3.Slug())
	}
}

func TestFindSlideByID(t *testing.T) {
	src := `---
title: Graph Deck
---

# Intro Slide
Welcome.

---

::id deep-dive
# Deep Dive
Technical details.

---

# Performance Benchmarks {#benchmarks}
Speed matters.
`
	deck := ParseDeck(src)
	if !deck.HasBranches() {
		// Even though it has no branches, let's test FindSlideByID
	}

	// 1. Exact ID
	if idx := deck.FindSlideByID("deep-dive"); idx != 1 {
		t.Errorf("expected index 1 for 'deep-dive', got %d", idx)
	}
	if idx := deck.FindSlideByID("benchmarks"); idx != 2 {
		t.Errorf("expected index 2 for 'benchmarks', got %d", idx)
	}

	// 2. Numeric 1-based index string
	if idx := deck.FindSlideByID("1"); idx != 0 {
		t.Errorf("expected index 0 for '1', got %d", idx)
	}
	if idx := deck.FindSlideByID("3"); idx != 2 {
		t.Errorf("expected index 2 for '3', got %d", idx)
	}

	// 3. Title substring
	if idx := deck.FindSlideByID("bench"); idx != 2 {
		t.Errorf("expected index 2 for 'bench', got %d", idx)
	}
	if idx := deck.FindSlideByID("Intro"); idx != 0 {
		t.Errorf("expected index 0 for 'Intro', got %d", idx)
	}

	// 4. Missing or invalid
	if idx := deck.FindSlideByID("non-existent"); idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}
	if idx := deck.FindSlideByID(""); idx != -1 {
		t.Errorf("expected -1, got %d", idx)
	}
	if idx := deck.FindSlideByID("999"); idx != -1 {
		t.Errorf("expected -1 for out-of-bounds number, got %d", idx)
	}
}

func TestBranchSerialization(t *testing.T) {
	src := `---
title: Branch Deck
---

::id intro
::next conclusion
::prev start
::tags overview,arch
# Welcome to Graph

::branch [1] Deep Dive -> dive
::branch [2] Live Demo -> demo
`
	deck := ParseDeck(src)
	serialized := SerializeDeck(deck)
	reparsed := ParseDeck(serialized)

	if len(reparsed.Slides) != 1 {
		t.Fatalf("expected 1 slide, got %d", len(reparsed.Slides))
	}
	s := reparsed.Slides[0]
	if s.ID != "intro" {
		t.Errorf("expected ID 'intro', got %q", s.ID)
	}
	if s.NextID != "conclusion" {
		t.Errorf("expected NextID 'conclusion', got %q", s.NextID)
	}
	if s.PrevID != "start" {
		t.Errorf("expected PrevID 'start', got %q", s.PrevID)
	}
	if len(s.Tags) != 2 || s.Tags[0] != "overview" || s.Tags[1] != "arch" {
		t.Errorf("unexpected tags: %+v", s.Tags)
	}
	branches := s.Branches()
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
	if branches[0].Key != "1" || branches[0].Target != "dive" {
		t.Errorf("unexpected branch 0: %+v", branches[0])
	}
	if branches[1].Key != "2" || branches[1].Target != "demo" {
		t.Errorf("unexpected branch 1: %+v", branches[1])
	}
}

func TestColumnsParsingAndSerialization(t *testing.T) {
	src := `---
title: Split Columns Demo
---

# Slide With Columns

:::columns
### Left Column
Left paragraph text.

- Left item 1
- Left item 2
:::col
### Right Column
Right paragraph text.

::code lang=go
fmt.Println("right")
::code
:::

---

# Split Alternative Syntax

::split
Column A
::col
Column B
::split
`
	deck := ParseDeck(src)
	if len(deck.Slides) != 2 {
		t.Fatalf("expected 2 slides, got %d", len(deck.Slides))
	}

	s1 := deck.Slides[0]
	var colBlock *Block
	for _, b := range s1.Blocks {
		if b.Kind == BlockColumns {
			colBlock = &b
			break
		}
	}
	if colBlock == nil {
		t.Fatalf("expected BlockColumns on slide 1")
	}
	if len(colBlock.Columns) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(colBlock.Columns))
	}
	if len(colBlock.Columns[0]) < 2 {
		t.Errorf("expected left column to have at least 2 blocks, got %d", len(colBlock.Columns[0]))
	}
	if len(colBlock.Columns[1]) < 2 {
		t.Errorf("expected right column to have at least 2 blocks, got %d", len(colBlock.Columns[1]))
	}

	summary := s1.Summary()
	if !strings.Contains(summary, "cols") {
		t.Errorf("expected summary to include 'cols', got %q", summary)
	}

	// Test alternative ::split syntax
	s2 := deck.Slides[1]
	var colBlock2 *Block
	for _, b := range s2.Blocks {
		if b.Kind == BlockColumns {
			colBlock2 = &b
			break
		}
	}
	if colBlock2 == nil {
		t.Fatalf("expected BlockColumns on slide 2 from ::split")
	}
	if len(colBlock2.Columns) != 2 {
		t.Fatalf("expected 2 columns on slide 2, got %d", len(colBlock2.Columns))
	}

	// Test serialization round-trip
	serialized := SerializeDeck(deck)
	reparsed := ParseDeck(serialized)
	if len(reparsed.Slides) != 2 {
		t.Fatalf("expected 2 reparsed slides, got %d", len(reparsed.Slides))
	}
	var reparsedCol *Block
	for _, b := range reparsed.Slides[0].Blocks {
		if b.Kind == BlockColumns {
			reparsedCol = &b
			break
		}
	}
	if reparsedCol == nil || len(reparsedCol.Columns) != 2 {
		t.Fatalf("expected reparsed BlockColumns with 2 columns")
	}
}

func TestSlideTagsAndAudienceTracks(t *testing.T) {
	src := `---
title: Tag Test
---

::id s1
::tags backend,arch
# Slide 1

---

::id s2
::tags frontend,ui
# Slide 2

---

::id s3
::tags backend,performance
# Slide 3
`
	deck := ParseDeck(src)
	if len(deck.Slides) != 3 {
		t.Fatalf("expected 3 slides, got %d", len(deck.Slides))
	}

	s1 := &deck.Slides[0]
	if !s1.HasTag("backend") {
		t.Errorf("expected s1 to have tag backend")
	}
	if !s1.HasTag("BACKEND") {
		t.Errorf("expected s1 to have tag BACKEND (case insensitive)")
	}
	if !s1.HasTag("arch") {
		t.Errorf("expected s1 to have tag arch")
	}
	if s1.HasTag("frontend") {
		t.Errorf("s1 should not have frontend tag")
	}
	if !s1.HasTag("") {
		t.Errorf("empty tag should match any slide")
	}
	if !s1.HasTag("all") {
		t.Errorf("'all' tag should match any slide")
	}

	allTags := deck.AllTags()
	expectedTags := []string{"backend", "arch", "frontend", "ui", "performance"}
	if len(allTags) != len(expectedTags) {
		t.Fatalf("expected %d unique tags, got %d: %+v", len(expectedTags), len(allTags), allTags)
	}

	backendIndices := deck.SlideIndicesForTag("backend")
	if len(backendIndices) != 2 || backendIndices[0] != 0 || backendIndices[1] != 2 {
		t.Errorf("expected backend indices [0, 2], got %+v", backendIndices)
	}

	frontendIndices := deck.SlideIndicesForTag("frontend")
	if len(frontendIndices) != 1 || frontendIndices[0] != 1 {
		t.Errorf("expected frontend indices [1], got %+v", frontendIndices)
	}

	nonexistentIndices := deck.SlideIndicesForTag("devops")
	if len(nonexistentIndices) != 0 {
		t.Errorf("expected 0 indices for nonexistent tag, got %+v", nonexistentIndices)
	}
}

func TestSlideTitleAndBranches(t *testing.T) {
	// 1. Heading slide
	sHeading := Slide{Blocks: []Block{{Kind: BlockHeading, Text: "My Heading"}}}
	if sHeading.Title() != "My Heading" {
		t.Errorf("expected 'My Heading', got %q", sHeading.Title())
	}

	// 2. Short paragraph slide
	sShortPara := Slide{Blocks: []Block{{Kind: BlockParagraph, Text: "Short paragraph"}}}
	if sShortPara.Title() != "Short paragraph" {
		t.Errorf("expected 'Short paragraph', got %q", sShortPara.Title())
	}

	// 3. Long paragraph slide (> 30 chars)
	sLongPara := Slide{Blocks: []Block{{Kind: BlockParagraph, Text: "This is a very long paragraph that definitely exceeds thirty characters."}}}
	if !strings.HasSuffix(sLongPara.Title(), "...") {
		t.Errorf("expected truncated title with ellipsis, got %q", sLongPara.Title())
	}

	// 4. Empty slide with no heading or paragraph
	sEmpty := Slide{Blocks: []Block{{Kind: BlockCode, Text: "code"}}}
	if sEmpty.Title() != "Slide" {
		t.Errorf("expected fallback 'Slide', got %q", sEmpty.Title())
	}

	// 5. HasBranches check
	deckNoBranches := Deck{Slides: []Slide{sHeading, sShortPara}}
	if deckNoBranches.HasBranches() {
		t.Errorf("expected HasBranches false for linear deck")
	}

	deckWithNext := Deck{Slides: []Slide{{NextID: "outro"}}}
	if !deckWithNext.HasBranches() {
		t.Errorf("expected HasBranches true for deck with NextID")
	}
}

func TestParseCodeLangAndFlags(t *testing.T) {
	cases := []struct {
		input        string
		expectedL    string
		expectedEval bool
	}{
		{"go", "go", false},
		{"python no-eval", "python", true},
		{"lang=bash eval=false", "bash", true},
		{"sh run=false", "sh", true},
		{"rust ignore", "rust", true},
		{"ts noexec", "ts", true},
		{`"javascript"`, "javascript", false},
	}

	for _, tc := range cases {
		lang, noEval := parseCodeLangAndFlags(tc.input)
		if lang != tc.expectedL || noEval != tc.expectedEval {
			t.Errorf("parseCodeLangAndFlags(%q) = (%q, %v); expected (%q, %v)", tc.input, lang, noEval, tc.expectedL, tc.expectedEval)
		}
	}
}

func TestRouteModelAndParsing(t *testing.T) {
	// 1. parseRouteLine tests
	name, slugs, ok := parseRouteLine("lightning: intro -> middle -> outro")
	if !ok || name != "lightning" || len(slugs) != 3 || slugs[0] != "intro" || slugs[1] != "middle" || slugs[2] != "outro" {
		t.Fatalf("unexpected parseRouteLine arrow: %s %v %v", name, slugs, ok)
	}

	name2, slugs2, ok2 := parseRouteLine("fast: intro => demo => end")
	if !ok2 || name2 != "fast" || len(slugs2) != 3 || slugs2[1] != "demo" {
		t.Fatalf("unexpected parseRouteLine double arrow: %s %v %v", name2, slugs2, ok2)
	}

	name3, slugs3, ok3 := parseRouteLine("comma: s1, s2, s3")
	if !ok3 || name3 != "comma" || len(slugs3) != 3 || slugs3[2] != "s3" {
		t.Fatalf("unexpected parseRouteLine comma: %s %v %v", name3, slugs3, ok3)
	}

	// invalid lines
	if _, _, ok := parseRouteLine("no colon line"); ok {
		t.Errorf("expected failure on line with no colon")
	}
	if _, _, ok := parseRouteLine(": no name"); ok {
		t.Errorf("expected failure on empty name")
	}
	if _, _, ok := parseRouteLine("name: "); ok {
		t.Errorf("expected failure on empty target list")
	}

	// 2. ParseDeck frontmatter & inline directives
	src := `---
title: Route Test
routes:
  talk: intro -> arch -> end
  quick: intro, end
route.extra: arch -> end
---

::id intro
# Intro Slide

---

::id arch
# Architecture Slide

::route inline-path: intro -> arch -> end

---

::id end
# End Slide
`
	deck := ParseDeck(src)
	routes := deck.AllRouteNames()
	expectedRoutes := []string{"extra", "inline-path", "quick", "talk"}
	if len(routes) != len(expectedRoutes) {
		t.Fatalf("expected %d routes, got %v", len(expectedRoutes), routes)
	}
	for i, r := range expectedRoutes {
		if routes[i] != r {
			t.Errorf("expected route %d to be %q, got %q", i, r, routes[i])
		}
	}

	// 3. RouteSlideIndices
	indices := deck.RouteSlideIndices("talk")
	if len(indices) != 3 || indices[0] != 0 || indices[1] != 1 || indices[2] != 2 {
		t.Errorf("expected talk route indices [0, 1, 2], got %v", indices)
	}

	// Unknown route returns empty
	emptyIdx := deck.RouteSlideIndices("nonexistent")
	if len(emptyIdx) != 0 {
		t.Errorf("expected empty indices for nonexistent route, got %v", emptyIdx)
	}

	// Route with unknown slide slug
	deck.Routes["broken"] = []string{"intro", "missing-slug", "end"}
	brokenIdx := deck.RouteSlideIndices("broken")
	if len(brokenIdx) != 2 || brokenIdx[0] != 0 || brokenIdx[1] != 2 {
		t.Errorf("expected broken route to skip missing slug and return [0, 2], got %v", brokenIdx)
	}

	// 4. Serialization
	serialized := SerializeDeck(deck)
	if !strings.Contains(serialized, "::route") {
		t.Errorf("expected serialized deck to contain ::route, got:\n%s", serialized)
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
