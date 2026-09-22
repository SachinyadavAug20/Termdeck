package internal

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

var reANSI = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return reANSI.ReplaceAllString(s, "")
}

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
			if !strings.Contains(stripANSI(l), "Slide Title") {
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

	out := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(out, "Left Aligned") {
		t.Errorf("expected view output to contain 'Left Aligned'")
	}
	if !strings.Contains(out, "(left)") {
		t.Errorf("expected status bar to show '(left)', got %q", out)
	}
}

func TestRenderHeadingLevels(t *testing.T) {
	for lvl := 1; lvl <= 7; lvl++ {
		rendered := stripANSI(renderHeading("Heading Text", lvl))
		if !strings.Contains(rendered, "Heading Text") {
			t.Errorf("expected rendered heading for level %d to contain 'Heading Text', got %q", lvl, rendered)
		}
	}
}

func TestHighlightLine(t *testing.T) {
	// Comments
	c1 := highlightLine("# python comment")
	if !strings.Contains(c1, "# python comment") {
		t.Errorf("expected comment in output: %q", c1)
	}
	c2 := highlightLine("// go comment")
	if !strings.Contains(c2, "// go comment") {
		t.Errorf("expected go comment in output: %q", c2)
	}
	c3 := highlightLine("-- sql comment")
	if !strings.Contains(c3, "-- sql comment") {
		t.Errorf("expected sql comment in output: %q", c3)
	}

	// Strings
	s1 := highlightLine(`"hello \"world\""`)
	if !strings.Contains(s1, "hello") {
		t.Errorf("expected string in output: %q", s1)
	}
	s2 := highlightLine(`'single quote'`)
	if !strings.Contains(s2, "single quote") {
		t.Errorf("expected single quote string: %q", s2)
	}
	s3 := highlightLine("`backtick`")
	if !strings.Contains(s3, "backtick") {
		t.Errorf("expected backtick string: %q", s3)
	}

	// Numbers
	n1 := highlightLine("x = 42 + 3.14")
	if !strings.Contains(n1, "42") || !strings.Contains(n1, "3.14") {
		t.Errorf("expected numbers in output: %q", n1)
	}

	// Keywords and types
	kw := highlightLine("func MyFunc() { return nil }")
	if !strings.Contains(kw, "func") || !strings.Contains(kw, "MyFunc") {
		t.Errorf("expected keywords and types: %q", kw)
	}

	// Other languages: python, bash
	py := highlightLine("def foo(self): print(None)")
	if !strings.Contains(py, "def") {
		t.Errorf("expected python keywords: %q", py)
	}
	sh := highlightLine("if [ -z $x ]; then echo done; fi")
	if !strings.Contains(sh, "then") {
		t.Errorf("expected bash keywords: %q", sh)
	}
}

func TestHighlightCode(t *testing.T) {
	lines := []string{
		"package main",
		"",
		"func main() {",
		`    println("Hello")`,
		"}",
	}
	out := highlightCode(lines, "go")
	if !strings.Contains(out, "package") || !strings.Contains(out, "Hello") {
		t.Errorf("expected highlighted code output, got %q", out)
	}
}

func TestInlineStyle(t *testing.T) {
	// Plain text
	plain := inlineStyle("just normal text")
	if !strings.Contains(plain, "just normal text") {
		t.Errorf("expected plain text to be preserved: %q", plain)
	}

	// Bold
	bold := inlineStyle("This is **bold** text")
	if !strings.Contains(bold, "bold") {
		t.Errorf("expected bold text rendered: %q", bold)
	}

	// Italic
	italic := inlineStyle("This is *italic* text")
	if !strings.Contains(italic, "italic") {
		t.Errorf("expected italic text rendered: %q", italic)
	}

	// Code span
	code := inlineStyle("Use `go test` here")
	if !strings.Contains(code, "go test") {
		t.Errorf("expected code span rendered: %q", code)
	}

	// Combined styling
	combined := inlineStyle("Check `code` and **bold** and *italic* together.")
	if !strings.Contains(combined, "code") || !strings.Contains(combined, "bold") || !strings.Contains(combined, "italic") {
		t.Errorf("expected all elements rendered in combined: %q", combined)
	}

	// Multiple code spans on one line (prevents slice bounds out of range panics)
	multiCode := inlineStyle("`cat .git/HEAD` -> `ref: refs/heads/main`")
	if !strings.Contains(multiCode, "cat .git/HEAD") || !strings.Contains(multiCode, "ref: refs/heads/main") {
		t.Errorf("expected multiple code spans rendered: %q", multiCode)
	}

	threeCode := inlineStyle("Bad: `fixed bug`, `asdf`, `wip`")
	if !strings.Contains(threeCode, "fixed bug") || !strings.Contains(threeCode, "asdf") || !strings.Contains(threeCode, "wip") {
		t.Errorf("expected three code spans rendered: %q", threeCode)
	}

	codeWithAsterisk := inlineStyle("Artifacts: `node_modules/`, `*.exe`, `.env`")
	if !strings.Contains(codeWithAsterisk, "*.exe") {
		t.Errorf("expected code span with asterisk preserved: %q", codeWithAsterisk)
	}
}

func TestRenderBlockVariants(t *testing.T) {
	w := 80
	maxH := 20
	baseDir := ""

	// 1. Heading normal & editing
	hBlk := Block{Kind: BlockHeading, Level: 1, Text: "Title"}
	outH := stripANSI(renderBlock(hBlk, w, maxH, baseDir, true, false, "", 0))
	if !strings.Contains(outH, "Title") || !strings.Contains(outH, "▶") {
		t.Errorf("expected laser pointer and title in heading, got %q", outH)
	}
	outHEdit := stripANSI(renderBlock(hBlk, w, maxH, baseDir, true, true, "Editing Title", 5))
	if !strings.Contains(outHEdit, "Editing Title") {
		t.Errorf("expected edit draft in heading edit: %q", outHEdit)
	}

	// 2. Paragraph normal & editing
	pBlk := Block{Kind: BlockParagraph, Text: "Some paragraph"}
	outP := stripANSI(renderBlock(pBlk, w, maxH, baseDir, false, false, "", 0))
	if !strings.Contains(outP, "Some paragraph") {
		t.Errorf("expected paragraph text: %q", outP)
	}
	outPEdit := stripANSI(renderBlock(pBlk, w, maxH, baseDir, true, true, "Draft Para", 2))
	if !strings.Contains(outPEdit, "Draft Para") {
		t.Errorf("expected edit draft in paragraph edit: %q", outPEdit)
	}

	// 3. Code block normal & editing
	cBlk := Block{Kind: BlockCode, Lang: "go", Lines: []string{"func foo() {}", "var x = 10"}}
	outC := stripANSI(renderBlock(cBlk, w, maxH, baseDir, true, false, "", 0))
	if !strings.Contains(outC, "go") || !strings.Contains(outC, "func") {
		t.Errorf("expected code block output: %q", outC)
	}
	outCEdit := stripANSI(renderBlock(cBlk, w, maxH, baseDir, true, true, "draft code", 0))
	if !strings.Contains(outCEdit, "draft code") {
		t.Errorf("expected code edit output: %q", outCEdit)
	}

	// 4. Image block editing
	iBlk := Block{Kind: BlockImage, Src: "test.png"}
	outIEdit := stripANSI(renderBlock(iBlk, w, maxH, baseDir, true, true, "new_src.png", 0))
	if !strings.Contains(outIEdit, "new_src.png") {
		t.Errorf("expected image draft edit output: %q", outIEdit)
	}

	// 5. List block normal & editing
	lBlk := Block{Kind: BlockList, Text: "- List Item"}
	outL := stripANSI(renderBlock(lBlk, w, maxH, baseDir, true, false, "", 0))
	if !strings.Contains(outL, "List Item") {
		t.Errorf("expected list block output: %q", outL)
	}
	outLEdit := stripANSI(renderBlock(lBlk, w, maxH, baseDir, true, true, "- New List Item", 0))
	if !strings.Contains(outLEdit, "- New List Item") {
		t.Errorf("expected list draft edit output: %q", outLEdit)
	}

	// 6. Directive block
	dBlk := Block{Kind: BlockDirective, Directive: "::plugin param"}
	outD := stripANSI(renderBlock(dBlk, w, maxH, baseDir, true, false, "", 0))
	if !strings.Contains(outD, "::plugin") || !strings.Contains(outD, "▶") {
		t.Errorf("expected laser pointer and directive: %q", outD)
	}
	outDNoCursor := stripANSI(renderBlock(dBlk, w, maxH, baseDir, false, false, "", 0))
	if !strings.Contains(outDNoCursor, "::plugin") {
		t.Errorf("expected directive: %q", outDNoCursor)
	}

	// 7. Speaker notes directive is hidden from canvas
	notesBlk := Block{Kind: BlockDirective, Directive: "::notes"}
	outNotes := renderBlock(notesBlk, w, maxH, baseDir, false, false, "", 0)
	if outNotes != "" {
		t.Errorf("expected notes directive to be hidden from canvas, got: %q", outNotes)
	}
}

func TestStatusBars(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockParagraph, Text: "Content"}}},
		},
		Align: AlignLeft,
	}
	ed := NewEditor("test.deck.md")
	ed.Dirty = true
	ed.Message = "Test Alert"

	// navStatus
	ns := navStatus(d, ed, 100)
	if !strings.Contains(ns, "[modified]") {
		t.Errorf("expected [modified] in navStatus: %q", ns)
	}
	if !strings.Contains(ns, "Test Alert") {
		t.Errorf("expected message in navStatus: %q", ns)
	}
	if !strings.Contains(ns, "(left)") {
		t.Errorf("expected (left) in navStatus: %q", ns)
	}

	// editStatus
	ed.Mode = ModeEdit
	ed.CursorCol = 5
	ed.Draft = "Hello World"
	ed.Message = "Save Note"
	es := editStatus(ed, 100)
	if !strings.Contains(es, "editing") || !strings.Contains(es, "col 5/11") {
		t.Errorf("expected editing info in editStatus: %q", es)
	}
	if !strings.Contains(es, "Save Note") {
		t.Errorf("expected message in editStatus: %q", es)
	}
}

func TestViewDimensionsAndModes(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Right Slide"},
				},
				Align: AlignRight,
			},
		},
	}
	ed := NewEditor("deck.md")

	// Test fallback dimensions (0, 0)
	outZero := stripANSI(View(d, ed, 0, 0))
	if !strings.Contains(outZero, "Right Slide") {
		t.Errorf("expected View with 0 dimensions to render content: %q", outZero)
	}
	if !strings.Contains(outZero, "(right)") {
		t.Errorf("expected right align status: %q", outZero)
	}

	// Test Edit mode rendering in View
	ed.Mode = ModeEdit
	ed.Draft = "Editing In Full View"
	outEdit := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(outEdit, "editing") {
		t.Errorf("expected edit mode status bar in View: %q", outEdit)
	}

	// Test Center alignment default
	dCenter := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockParagraph, Text: "Center Slide"}}},
		},
	}
	edCenter := NewEditor("center.deck.md")
	outCenter := stripANSI(View(dCenter, edCenter, 80, 24))
	if !strings.Contains(outCenter, "(center)") {
		t.Errorf("expected center status in View: %q", outCenter)
	}
}

func TestSpeakerNotesView(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Public Heading"},
					{Kind: BlockParagraph, Text: "Public Content"},
					{Kind: BlockDirective, Directive: "::notes", Lines: []string{"Confidential Speaker Note Line"}},
				},
			},
			{
				Blocks: []Block{
					{Kind: BlockParagraph, Text: "Slide without notes"},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")

	// 1. Audience view (default ShowNotes: false)
	outDefault := stripANSI(View(d, ed, 80, 24))
	if strings.Contains(outDefault, "Confidential Speaker Note Line") {
		t.Errorf("critical error: speaker note text leaked onto audience canvas: %q", outDefault)
	}
	if strings.Contains(outDefault, "::notes") {
		t.Errorf("critical error: ::notes directive rendered onto audience canvas: %q", outDefault)
	}
	// Blocks count must only count visible blocks (2), not the hidden notes block (3)
	if !strings.Contains(outDefault, "blocks 2") {
		t.Errorf("expected status bar to report 'blocks 2', got: %q", outDefault)
	}
	// Status bar should indicate notes are available
	if !strings.Contains(outDefault, "[n: notes]") {
		t.Errorf("expected status bar to indicate '[n: notes]', got: %q", outDefault)
	}

	// 2. Presenter notes overlay open (ShowNotes: true)
	ed.ShowNotes = true
	outWithNotes := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(outWithNotes, "📝 Speaker Notes") {
		t.Errorf("expected notes title box in view: %q", outWithNotes)
	}
	if !strings.Contains(outWithNotes, "Confidential Speaker Note Line") {
		t.Errorf("expected notes text in speaker overlay: %q", outWithNotes)
	}
	if !strings.Contains(outWithNotes, "[n: notes open]") {
		t.Errorf("expected status bar to show '[n: notes open]', got: %q", outWithNotes)
	}

	// 3. Presenter navigates to slide without notes while ShowNotes is still true
	ed.SlideIdx = 1
	outSlide2 := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(outSlide2, "no speaker notes for this slide") {
		t.Errorf("expected empty notes message for slide without notes: %q", outSlide2)
	}
	if strings.Contains(outSlide2, "[n: notes]") {
		t.Errorf("expected no notes badge on slide without notes: %q", outSlide2)
	}

	// 4. Verification on the real sample deck file (git_under_the_hood.deck.md)
	realDeckData, err := os.ReadFile("../git_under_the_hood.deck.md")
	if err == nil {
		realDeck := ParseDeck(string(realDeckData))
		realEd := NewEditor("../git_under_the_hood.deck.md")
		realEd.SlideIdx = 1 // Slide 2: "Mental Model: Git is a Content-Addressable Filesystem"
		realOut := stripANSI(View(realDeck, realEd, 100, 30))
		if strings.Contains(realOut, "Hook the audience immediately.") {
			t.Errorf("critical error: speaker note leaked into canvas on slide 2 of git_under_the_hood.deck.md")
		}
		if strings.Contains(realOut, "::notes") {
			t.Errorf("critical error: ::notes directive visible on canvas on slide 2 of git_under_the_hood.deck.md")
		}
		if !strings.Contains(realOut, "[n: notes]") {
			t.Errorf("expected [n: notes] badge in status bar on slide 2")
		}

		// When 'n' is toggled, notes appear in overlay box
		realEd.ShowNotes = true
		realOutWithNotes := stripANSI(View(realDeck, realEd, 100, 30))
		if !strings.Contains(realOutWithNotes, "📝 Speaker Notes") {
			t.Errorf("expected notes overlay box title on slide 2")
		}
		if !strings.Contains(realOutWithNotes, "Hook the audience immediately.") {
			t.Errorf("expected notes text in overlay box on slide 2, got: %q", realOutWithNotes)
		}

		// Verify every slide in git_under_the_hood.deck.md renders with 0 panics
		for sIdx := range realDeck.Slides {
			realEd.SlideIdx = sIdx
			realEd.BlockIdx = 0
			realEd.ShowNotes = false
			out := View(realDeck, realEd, 100, 30)
			if len(out) == 0 {
				t.Errorf("slide %d produced empty output", sIdx)
			}
			realEd.ShowNotes = true
			outNotes := View(realDeck, realEd, 100, 30)
			if len(outNotes) == 0 {
				t.Errorf("slide %d with notes produced empty output", sIdx)
			}
		}
	}
}

func TestTableAndHelpModalView(t *testing.T) {
	// 1. Table rendering
	tblBlock := Block{
		Kind: BlockTable,
		Lines: []string{
			"| Command | Action |",
			"|---|---|",
			"| git status | view state |",
			"| git log | view history |",
		},
	}
	renderedTbl := stripANSI(renderTable(tblBlock, 80))
	if !strings.Contains(renderedTbl, "Command") || !strings.Contains(renderedTbl, "view history") {
		t.Errorf("expected table content in rendered table: %q", renderedTbl)
	}

	// 2. Help Modal view
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{tblBlock}},
			{Blocks: []Block{{Kind: BlockParagraph, Text: "Slide 2"}}},
		},
	}
	ed := NewEditor("test.deck.md")
	ed.ShowHelp = true
	helpView := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(helpView, "Termdeck Keyboard Controls") || !strings.Contains(helpView, "NAVIGATION") {
		t.Errorf("expected help modal in View: %q", helpView)
	}

	// 3. Status bar with clean slide info and bottom progress line
	ed.ShowHelp = false
	statusView := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(statusView, "slide 1/2") {
		t.Errorf("expected slide status in view: %q", statusView)
	}
	if !strings.Contains(statusView, "─") {
		t.Errorf("expected bottom progress line in status view: %q", statusView)
	}
}

func TestRenderProgressLine(t *testing.T) {
	// Zero width
	if res := renderProgressLine(1, 5, 0); res != "" {
		t.Errorf("expected empty progress line for 0 width, got %q", res)
	}

	// Zero total slides
	if res := stripANSI(renderProgressLine(1, 0, 40)); len([]rune(res)) != 40 {
		t.Errorf("expected 40 chars dim line for 0 total slides, got %d chars", len([]rune(res)))
	}

	// 1 of 4 slides at width 40 -> 10 chars filled, 30 dim
	res := renderProgressLine(1, 4, 40)
	clean := stripANSI(res)
	if len([]rune(clean)) != 40 {
		t.Errorf("expected exactly 40 chars, got %d", len([]rune(clean)))
	}
	// Verify raw ANSI has styling
	if !strings.Contains(res, "─") {
		t.Errorf("expected progress line to contain horizontal line characters: %q", res)
	}

	// Clamping
	resClamped := stripANSI(renderProgressLine(10, 4, 20))
	if len([]rune(resClamped)) != 20 {
		t.Errorf("expected clamped progress line to be 20 chars, got %d", len([]rune(resClamped)))
	}
}

func TestViewZenMode(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Zen Title"},
					{Kind: BlockParagraph, Text: "Clean presentation content."},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")

	// Normal view has status line
	normalOut := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(normalOut, "slide 1/1") || !strings.Contains(normalOut, "help") {
		t.Errorf("expected normal view to contain status bar, got %q", normalOut)
	}

	// Zen mode hides status bar but retains slide content and progress line
	ed.ZenMode = true
	zenOut := stripANSI(View(d, ed, 80, 24))
	if strings.Contains(zenOut, "slide 1/1") || strings.Contains(zenOut, "help") {
		t.Errorf("expected zen mode to omit status bar, got %q", zenOut)
	}
	if !strings.Contains(zenOut, "Zen Title") {
		t.Errorf("expected zen mode to contain slide heading, got %q", zenOut)
	}
	if !strings.Contains(zenOut, "Clean presentation content.") {
		t.Errorf("expected zen mode to contain slide body, got %q", zenOut)
	}
	if !strings.Contains(zenOut, "─") {
		t.Errorf("expected zen mode to retain progress line, got %q", zenOut)
	}
}

func BenchmarkRenderView(b *testing.B) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Benchmark Heading"},
					{Kind: BlockParagraph, Text: "Some paragraph with **bold** and `code` span."},
					{Kind: BlockCode, Lang: "go", Lines: []string{"func main() {", `    println("hello")`, "}"}},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = View(d, ed, 120, 30)
	}
}
