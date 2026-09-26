package internal

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
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

	// Other languages: python, bash, rust, typescript, sql
	py := highlightLine("def foo(self): print(None)")
	if !strings.Contains(py, "def") {
		t.Errorf("expected python keywords: %q", py)
	}
	sh := highlightLine("if [ -z $x ]; then echo done; fi")
	if !strings.Contains(sh, "then") {
		t.Errorf("expected bash keywords: %q", sh)
	}
	rs := highlightLine("pub fn run(mut item: State) -> Result")
	if !strings.Contains(rs, "fn") || !strings.Contains(rs, "mut") || !strings.Contains(rs, "pub") {
		t.Errorf("expected rust keywords: %q", rs)
	}
	ts := highlightLine("async function getData(): Promise { let x = await fetch(); }")
	if !strings.Contains(ts, "async") || !strings.Contains(ts, "await") || !strings.Contains(ts, "let") {
		t.Errorf("expected typescript keywords: %q", ts)
	}
	sql := highlightLine("SELECT name, age FROM users WHERE id = 10")
	if !strings.Contains(sql, "SELECT") || !strings.Contains(sql, "FROM") || !strings.Contains(sql, "WHERE") {
		t.Errorf("expected sql keywords: %q", sql)
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
	out := highlightCode(lines, "go", false)
	if !strings.Contains(out, "package") || !strings.Contains(out, "Hello") {
		t.Errorf("expected highlighted code output, got %q", out)
	}

	// Line numbering enabled
	outLines := highlightCode(lines, "go", true)
	if !strings.Contains(outLines, "1 │") || !strings.Contains(outLines, "5 │") {
		t.Errorf("expected line numbers in highlighted code, got %q", outLines)
	}

	// Diff syntax highlighting
	diffLines := []string{
		"--- a/main.go",
		"+++ b/main.go",
		"@@ -10,3 +10,3 @@",
		"- func fetch(id int)",
		"+ func fetch(ctx context.Context, id int)",
		"  context line",
	}
	diffOut := highlightCode(diffLines, "diff", false)
	if !strings.Contains(diffOut, "+ func fetch") || !strings.Contains(diffOut, "- func fetch") {
		t.Errorf("expected diff additions and deletions, got %q", diffOut)
	}
	if !strings.Contains(diffOut, "@@ -10,3 +10,3 @@") {
		t.Errorf("expected diff chunk header, got %q", diffOut)
	}

	// Diff with line numbers
	diffWithLines := highlightCode(diffLines, "diff", true)
	if !strings.Contains(diffWithLines, "1 │") || !strings.Contains(diffWithLines, "+ func fetch") {
		t.Errorf("expected line numbers with diff, got %q", diffWithLines)
	}

	// Patch language alias
	patchOut := highlightCode([]string{"+ added"}, "patch", false)
	if !strings.Contains(patchOut, "+ added") {
		t.Errorf("expected patch addition, got %q", patchOut)
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
	outH := stripANSI(renderBlock(hBlk, w, maxH, baseDir, true, false, "", 0, false))
	if !strings.Contains(outH, "Title") || !strings.Contains(outH, "▶") {
		t.Errorf("expected laser pointer and title in heading, got %q", outH)
	}
	outHEdit := stripANSI(renderBlock(hBlk, w, maxH, baseDir, true, true, "Editing Title", 5, false))
	if !strings.Contains(outHEdit, "Editing Title") {
		t.Errorf("expected edit draft in heading edit: %q", outHEdit)
	}

	// 2. Paragraph normal & editing
	pBlk := Block{Kind: BlockParagraph, Text: "Some paragraph"}
	outP := stripANSI(renderBlock(pBlk, w, maxH, baseDir, false, false, "", 0, false))
	if !strings.Contains(outP, "Some paragraph") {
		t.Errorf("expected paragraph text: %q", outP)
	}
	outPEdit := stripANSI(renderBlock(pBlk, w, maxH, baseDir, true, true, "Draft Para", 2, false))
	if !strings.Contains(outPEdit, "Draft Para") {
		t.Errorf("expected edit draft in paragraph edit: %q", outPEdit)
	}

	// 3. Code block normal & editing & with line numbers
	cBlk := Block{Kind: BlockCode, Lang: "go", Lines: []string{"func foo() {}", "var x = 10"}}
	outC := stripANSI(renderBlock(cBlk, w, maxH, baseDir, true, false, "", 0, false))
	if !strings.Contains(outC, "go") || !strings.Contains(outC, "func") {
		t.Errorf("expected code block output: %q", outC)
	}
	outCLines := stripANSI(renderBlock(cBlk, w, maxH, baseDir, false, false, "", 0, true))
	if !strings.Contains(outCLines, "1 │") || !strings.Contains(outCLines, "2 │") {
		t.Errorf("expected line numbers in code block render, got: %q", outCLines)
	}
	outCEdit := stripANSI(renderBlock(cBlk, w, maxH, baseDir, true, true, "draft code", 0, false))
	if !strings.Contains(outCEdit, "draft code") {
		t.Errorf("expected code edit output: %q", outCEdit)
	}

	// 4. Image block editing
	iBlk := Block{Kind: BlockImage, Src: "test.png"}
	outIEdit := stripANSI(renderBlock(iBlk, w, maxH, baseDir, true, true, "new_src.png", 0, false))
	if !strings.Contains(outIEdit, "new_src.png") {
		t.Errorf("expected image draft edit output: %q", outIEdit)
	}

	// 5. List block normal & editing
	lBlk := Block{Kind: BlockList, Text: "- List Item"}
	outL := stripANSI(renderBlock(lBlk, w, maxH, baseDir, true, false, "", 0, false))
	if !strings.Contains(outL, "List Item") {
		t.Errorf("expected list block output: %q", outL)
	}
	outLEdit := stripANSI(renderBlock(lBlk, w, maxH, baseDir, true, true, "- New List Item", 0, false))
	if !strings.Contains(outLEdit, "- New List Item") {
		t.Errorf("expected list draft edit output: %q", outLEdit)
	}

	// 6. Directive block
	dBlk := Block{Kind: BlockDirective, Directive: "::plugin param"}
	outD := stripANSI(renderBlock(dBlk, w, maxH, baseDir, true, false, "", 0, false))
	if !strings.Contains(outD, "::plugin") || !strings.Contains(outD, "▶") {
		t.Errorf("expected laser pointer and directive: %q", outD)
	}
	outDNoCursor := stripANSI(renderBlock(dBlk, w, maxH, baseDir, false, false, "", 0, false))
	if !strings.Contains(outDNoCursor, "::plugin") {
		t.Errorf("expected directive: %q", outDNoCursor)
	}

	// 7. Speaker notes directive is hidden from canvas
	notesBlk := Block{Kind: BlockDirective, Directive: "::notes"}
	outNotes := renderBlock(notesBlk, w, maxH, baseDir, false, false, "", 0, false)
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

func TestRenderCallouts(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockCallout, Callout: "tip", Lines: []string{"Use indexing for fast lookups."}},
					{Kind: BlockCallout, Callout: "note", Lines: []string{"Default limit is 100."}},
					{Kind: BlockCallout, Callout: "warning", Lines: []string{"Avoid N+1 queries."}},
					{Kind: BlockCallout, Callout: "important", Lines: []string{"Migration requires restart."}},
					{Kind: BlockCallout, Callout: "caution", Lines: []string{"Data drop is irreversible."}},
					{Kind: BlockCallout, Callout: "quote", Lines: []string{"Code is read more than written."}},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")

	out := stripANSI(View(d, ed, 80, 40))
	if !strings.Contains(out, "TIP") || !strings.Contains(out, "Use indexing for fast lookups.") {
		t.Errorf("expected Tip callout in output: %q", out)
	}
	if !strings.Contains(out, "NOTE") || !strings.Contains(out, "Default limit is 100.") {
		t.Errorf("expected Note callout in output: %q", out)
	}
	if !strings.Contains(out, "WARNING") || !strings.Contains(out, "Avoid N+1 queries.") {
		t.Errorf("expected Warning callout in output: %q", out)
	}
	if !strings.Contains(out, "IMPORTANT") || !strings.Contains(out, "Migration requires restart.") {
		t.Errorf("expected Important callout in output: %q", out)
	}
	if !strings.Contains(out, "CAUTION") || !strings.Contains(out, "Data drop is irreversible.") {
		t.Errorf("expected Caution callout in output: %q", out)
	}
	if !strings.Contains(out, "QUOTE") || !strings.Contains(out, "Code is read more than written.") {
		t.Errorf("expected Quote callout in output: %q", out)
	}

	// Editing callout block
	ed.Mode = ModeEdit
	ed.BlockIdx = 0
	ed.Draft = "Edited tip draft"
	editOut := stripANSI(View(d, ed, 80, 40))
	if !strings.Contains(editOut, "Edited tip draft") {
		t.Errorf("expected edit draft in callout edit view: %q", editOut)
	}
}

func TestRenderJumpModal(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Intro Slide"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Architecture"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Benchmarks"}}},
		},
	}
	ed := NewEditor("test.deck.md")
	ed.Mode = ModePrompt
	ed.Draft = "arch"

	out := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(out, "Jump to Slide") {
		t.Errorf("expected Jump to Slide modal title: %q", out)
	}
	if !strings.Contains(out, "Architecture") {
		t.Errorf("expected matching slide title 'Architecture': %q", out)
	}
	if !strings.Contains(out, "Press Enter to jump") {
		t.Errorf("expected jump instructions: %q", out)
	}
}

func TestRenderListItem(t *testing.T) {
	// 1. Checked task
	chk := renderListItem("- [x] Deploy to production")
	if !strings.Contains(chk, "✔") || !strings.Contains(chk, "Deploy to production") {
		t.Errorf("expected checked task rendering, got %q", chk)
	}

	// 2. Unchecked task
	unchk := renderListItem("- [ ] Migrate database")
	if !strings.Contains(unchk, "○") || !strings.Contains(unchk, "Migrate database") {
		t.Errorf("expected unchecked task rendering, got %q", unchk)
	}

	// 3. Regular bullet
	bullet := renderListItem("- Standard bullet point")
	if !strings.Contains(bullet, "•") || !strings.Contains(bullet, "Standard bullet point") {
		t.Errorf("expected bullet rendering, got %q", bullet)
	}

	// 4. Numbered list
	num := renderListItem("1. First step")
	if !strings.Contains(num, "1.") || !strings.Contains(num, "First step") {
		t.Errorf("expected numbered list rendering, got %q", num)
	}
}

func TestRenderDivider(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Title"},
					{Kind: BlockDivider},
					{Kind: BlockParagraph, Text: "Below line"},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")

	out := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(out, "Title") || !strings.Contains(out, "Below line") {
		t.Errorf("expected slide content in divider view: %q", out)
	}
	if !strings.Contains(out, "──") {
		t.Errorf("expected divider horizontal line in view: %q", out)
	}

	// Edit mode on divider
	ed.Mode = ModeEdit
	ed.BlockIdx = 1
	editOut := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(editOut, "***") {
		t.Errorf("expected '***' in edit mode on divider: %q", editOut)
	}
}

func TestCodeLineNumbersView(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Code Slide"},
					{Kind: BlockCode, Lang: "go", Lines: []string{"func sum(a, b int) int {", "    return a + b", "}"}},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")

	// Standard view without line numbers
	outNoLines := stripANSI(View(d, ed, 80, 24))
	if strings.Contains(outNoLines, "1 │") {
		t.Errorf("expected no line numbers by default, got: %q", outNoLines)
	}
	if strings.Contains(outNoLines, "[L: lines]") {
		t.Errorf("expected no [L: lines] badge by default, got: %q", outNoLines)
	}

	// View with line numbers enabled
	ed.ShowLineNumbers = true
	outWithLines := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(outWithLines, "1 │") || !strings.Contains(outWithLines, "2 │") || !strings.Contains(outWithLines, "3 │") {
		t.Errorf("expected line numbers 1-3 in view, got: %q", outWithLines)
	}
	if !strings.Contains(outWithLines, "[L: lines]") {
		t.Errorf("expected [L: lines] badge in status bar, got: %q", outWithLines)
	}
}

func TestTimerView(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Timer Presentation"}}},
		},
	}
	ed := NewEditor("test.deck.md")

	// Timer disabled
	out := stripANSI(View(d, ed, 80, 24))
	if strings.Contains(out, "⏱") {
		t.Errorf("expected no stopwatch glyph when timer is disabled")
	}

	// Timer enabled: 5 mins 23 secs ago
	ed.ShowTimer = true
	ed.TimerStart = time.Now().Add(-5*time.Minute - 23*time.Second)
	out = stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(out, "⏱ 05:23") {
		t.Errorf("expected [⏱ 05:23] in status bar, got: %q", out)
	}

	// Long presentation: 1 hour 12 mins 4 secs
	ed.TimerStart = time.Now().Add(-1*time.Hour - 12*time.Minute - 4*time.Second)
	out = stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(out, "⏱ 1:12:04") {
		t.Errorf("expected [⏱ 1:12:04] in status bar, got: %q", out)
	}
}

func TestWatchModeView(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Watch Deck"}}},
		},
	}
	ed := NewEditor("test.deck.md")

	// WatchMode off
	out := stripANSI(View(d, ed, 80, 24))
	if strings.Contains(out, "[watch]") {
		t.Errorf("expected no [watch] badge when watch mode is off")
	}

	// WatchMode on
	ed.WatchMode = true
	out = stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(out, "[watch]") {
		t.Errorf("expected [watch] badge in status bar when watch mode is on")
	}

	// Help modal should contain reload shortcut
	help := stripANSI(renderHelpModal(80, 24))
	if !strings.Contains(help, "r / R") || !strings.Contains(help, "Reload deck file from disk") {
		t.Errorf("expected help modal to document 'r / R' reload, got: %s", help)
	}
}

func TestOverviewModalView(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Architecture Overview"}}},
			{Blocks: []Block{
				{Kind: BlockHeading, Level: 1, Text: "Code Details"},
				{Kind: BlockCode, Lang: "go", Lines: []string{"func run() {}"}},
			}},
			{Blocks: []Block{
				{Kind: BlockHeading, Level: 1, Text: "Task List"},
				{Kind: BlockList, Text: "- [x] Done"},
			}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 4"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 5"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 6"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 7"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Slide 8"}}},
		},
	}
	ed := NewEditor("test.deck.md")
	ed.SlideIdx = 0
	ed.ShowOverview = true
	ed.OverviewCursor = 1 // Cursor on Slide 2

	// 1. Standard 80x24 view
	out := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(out, "Slide Overview & Grid Sorter") {
		t.Errorf("expected overview title in view output, got: %s", out)
	}
	if !strings.Contains(out, "Architecture") {
		t.Errorf("expected slide 1 title 'Architecture' in overview")
	}
	if !strings.Contains(out, "Code Details") {
		t.Errorf("expected slide 2 title in overview")
	}
	if !strings.Contains(out, "▶ #2") {
		t.Errorf("expected cursor pointer on slide #2, got: %s", out)
	}
	if !strings.Contains(out, "code") {
		t.Errorf("expected 'code' summary in slide #2 card")
	}

	// 2. Narrow terminal width (e.g. 50 cols -> 1 or 2 cols)
	narrowOut := stripANSI(View(d, ed, 50, 20))
	if !strings.Contains(narrowOut, "Slide Overview & Grid Sorter") {
		t.Errorf("expected overview title in narrow view")
	}

	// 3. Very small height triggers pagination
	shortOut := stripANSI(View(d, ed, 80, 14))
	if !strings.Contains(shortOut, "Slide Overview & Grid Sorter") {
		t.Errorf("expected overview title in short view")
	}

	// 4. Help modal lists 'o / O'
	help := stripANSI(renderHelpModal(80, 24))
	if !strings.Contains(help, "o / O") || !strings.Contains(help, "Slide overview & grid sorter") {
		t.Errorf("expected help modal to document 'o / O', got: %s", help)
	}

	// 5. navStatus lists 'o grid'
	ed.ShowOverview = false
	status := stripANSI(navStatus(d, ed, 120))
	if !strings.Contains(status, "o grid") {
		t.Errorf("expected navStatus to show 'o grid', got: %s", status)
	}
}

func TestBlankScreenView(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Normal Presentation Slide"}}},
		},
	}
	ed := NewEditor("test.deck.md")
	ed.ScreenBlank = true

	out := stripANSI(View(d, ed, 80, 24))
	if !strings.Contains(out, "presentation paused") || !strings.Contains(out, "press any key to resume") {
		t.Errorf("expected blank screen placeholder message, got: %s", out)
	}
	if strings.Contains(out, "Normal Presentation Slide") {
		t.Errorf("expected slide content hidden when screen is blanked")
	}

	// Verify help modal entries
	help := stripANSI(renderHelpModal(80, 24))
	if !strings.Contains(help, "y / Y") || !strings.Contains(help, "Copy block to clipboard (OSC 52)") {
		t.Errorf("expected help modal to document y/Y, got: %s", help)
	}
	if !strings.Contains(help, "b / B") || !strings.Contains(help, "Blank presentation screen") {
		t.Errorf("expected help modal to document b/B, got: %s", help)
	}
	if !strings.Contains(help, "E") || !strings.Contains(help, "Export presentation to HTML") {
		t.Errorf("expected help modal to document E export, got: %s", help)
	}

	// Verify navStatus entries
	ed.ScreenBlank = false
	nav := stripANSI(navStatus(d, ed, 160))
	if !strings.Contains(nav, "y yank") || !strings.Contains(nav, "E export") || !strings.Contains(nav, "b blank") {
		t.Errorf("expected navStatus to show y yank, E export, and b blank, got: %s", nav)
	}
}

func TestAutoplayView(t *testing.T) {
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Autoplay Demo"}}},
		},
	}
	ed := NewEditor("test.deck.md")
	ed.Autoplay = true
	ed.AutoplayInterval = 8
	ed.AutoplayCountdown = 4

	nav := stripANSI(navStatus(d, ed, 240))
	if !strings.Contains(nav, "[▶ auto: 8s (4s)]") {
		t.Errorf("expected navStatus to contain [▶ auto: 8s (4s)], got: %s", nav)
	}
	if !strings.Contains(nav, "A auto") {
		t.Errorf("expected navStatus to contain 'A auto', got: %s", nav)
	}

	help := stripANSI(renderHelpModal(80, 24))
	if !strings.Contains(help, "A") || !strings.Contains(help, "Toggle auto-advance") {
		t.Errorf("expected help modal to document 'A' auto-advance, got: %s", help)
	}
}

func TestRenderBranchBlock(t *testing.T) {
	SetCurrentTheme("termdeck")
	blk := Block{
		Kind:         BlockBranch,
		BranchKey:    "1",
		Text:         "Storage Architecture",
		BranchTarget: "storage",
	}

	rendered := stripANSI(renderBlock(blk, 80, 20, "", false, false, "", 0, false))
	if !strings.Contains(rendered, "[1]") {
		t.Errorf("expected '[1]' in rendered branch block, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "Storage Architecture") {
		t.Errorf("expected label in rendered branch block, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "──►") || !strings.Contains(rendered, "#storage") {
		t.Errorf("expected arrow and target in rendered branch block, got:\n%s", rendered)
	}

	// Test cursor mark
	renderedCursor := stripANSI(renderBlock(blk, 80, 20, "", true, false, "", 0, false))
	if !strings.Contains(renderedCursor, "▶") {
		t.Errorf("expected cursor pointer '▶' on focused branch card, got:\n%s", renderedCursor)
	}

	// Test edit mode
	renderedEdit := renderBlock(blk, 80, 20, "", true, true, "custom edit draft", 0, false)
	if !strings.Contains(renderedEdit, "custom edit draft") {
		t.Errorf("expected edit draft in edit mode, got:\n%s", renderedEdit)
	}
}

func TestRenderGraphModal(t *testing.T) {
	SetCurrentTheme("termdeck")
	src := `---
title: Graph Deck
---

# Architecture
::branch [1] Storage -> storage
::branch [2] Network -> network

---

::id storage
# Storage
::next conclusion
LSM trees.

---

::id conclusion
# Conclusion
Done.`

	d := ParseDeck(src)
	ed := NewEditor("")
	ed.ShowGraphMap = true

	modal := stripANSI(renderGraphModal(d, ed, 80, 24))
	if !strings.Contains(modal, "Presentation Topology Map") {
		t.Errorf("expected title in graph modal, got:\n%s", modal)
	}
	if !strings.Contains(modal, "[01]") || !strings.Contains(modal, "Architecture") {
		t.Errorf("expected slide 1 in graph modal, got:\n%s", modal)
	}
	if !strings.Contains(modal, "[1] ──►") {
		t.Errorf("expected branch edge in graph modal, got:\n%s", modal)
	}

	// Test with history breadcrumbs
	ed.History = []int{0, 1}
	ed.SlideIdx = 2
	modalHist := stripANSI(renderGraphModal(d, ed, 80, 24))
	if !strings.Contains(modalHist, "Path:") || !strings.Contains(modalHist, "[01] ──► [02] ──► [03]") {
		t.Errorf("expected breadcrumbs in graph modal with history, got:\n%s", modalHist)
	}

	// Empty deck
	emptyModal := renderGraphModal(Deck{}, ed, 80, 24)
	if emptyModal != "" {
		t.Errorf("expected empty string for empty deck, got:\n%s", emptyModal)
	}
}

func TestNavStatusForkAndHistory(t *testing.T) {
	SetCurrentTheme("termdeck")
	src := `---
title: Branch Status Deck
---

# Slide 1
::branch [1] Next Part -> part2
`
	d := ParseDeck(src)
	ed := NewEditor("")
	ed.History = []int{0}

	status := stripANSI(navStatus(d, ed, 240))
	if !strings.Contains(status, "[fork: 1 paths") || !strings.Contains(status, "Next Part") {
		t.Errorf("expected fork badge in nav status, got:\n%s", status)
	}
	if !strings.Contains(status, "[history: 1 (H)]") {
		t.Errorf("expected history badge in nav status, got:\n%s", status)
	}
	if !strings.Contains(status, "M map") {
		t.Errorf("expected 'M map' hint in nav status, got:\n%s", status)
	}
}

func TestHelpModalGraphShortcuts(t *testing.T) {
	SetCurrentTheme("termdeck")
	help := stripANSI(renderHelpModal(80, 24))
	if !strings.Contains(help, "1 - 9") {
		t.Errorf("expected '1 - 9' in help modal, got:\n%s", help)
	}
	if !strings.Contains(help, "Backspace") || !strings.Contains(help, "H") {
		t.Errorf("expected 'Backspace' and 'H' in help modal, got:\n%s", help)
	}
	if !strings.Contains(help, "M") || !strings.Contains(help, "graph map") {
		t.Errorf("expected 'M' graph map in help modal, got:\n%s", help)
	}
	if !strings.Contains(help, "f / F") {
		t.Errorf("expected 'f / F' in help modal, got:\n%s", help)
	}
	if !strings.Contains(help, "X / ctrl+x") {
		t.Errorf("expected 'X / ctrl+x' in help modal, got:\n%s", help)
	}
}

func TestRenderRunnerCardAndView(t *testing.T) {
	SetCurrentTheme("tokyo-night")

	// Nil result
	if s := renderRunnerCard(nil, 80); s != "" {
		t.Fatalf("expected empty string for nil result, got %q", s)
	}

	// Success result
	resSuccess := &ExecResult{
		Language: "bash",
		ExitCode: 0,
		Duration: 12 * time.Millisecond,
		Stdout:   "All systems operational",
	}
	cardSuccess := stripANSI(renderRunnerCard(resSuccess, 80))
	if !strings.Contains(cardSuccess, "EXIT 0") || !strings.Contains(cardSuccess, "All systems operational") {
		t.Fatalf("unexpected success card: %s", cardSuccess)
	}

	// Failure result with stderr
	resFail := &ExecResult{
		Language: "python",
		ExitCode: 1,
		Duration: 25 * time.Millisecond,
		Stderr:   "NameError: name 'x' is not defined",
		Error:    "exit status 1",
	}
	cardFail := stripANSI(renderRunnerCard(resFail, 80))
	if !strings.Contains(cardFail, "EXIT 1") || !strings.Contains(cardFail, "NameError") {
		t.Fatalf("unexpected failure card: %s", cardFail)
	}

	// Running banner
	banner := stripANSI(renderRunningBanner("go", 80))
	if !strings.Contains(banner, "Executing [go]") {
		t.Fatalf("unexpected banner: %s", banner)
	}

	// View with runner active
	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Live Run Demo"},
					{Kind: BlockCode, Lang: "sh", Text: "echo 'hello'"},
				},
			},
		},
	}
	ed := NewEditor("")
	ed.ShowRunner = true
	ed.RunnerResult = resSuccess

	viewStr := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(viewStr, "Live Terminal Runner") || !strings.Contains(viewStr, "All systems operational") {
		t.Fatalf("expected runner card in View, got:\n%s", viewStr)
	}

	// View with code running
	ed.ShowRunner = false
	ed.RunningCode = true
	ed.BlockIdx = 1
	viewRunning := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(viewRunning, "Executing [sh]") {
		t.Fatalf("expected running banner in View, got:\n%s", viewRunning)
	}
}

func TestRenderFocusModeView(t *testing.T) {
	SetCurrentTheme("dracula")

	ed := NewEditor("")

	// Empty deck
	if s := renderFocusMode(Deck{}, ed, 80, 24); s != "" {
		t.Fatalf("expected empty focus view for empty deck, got %q", s)
	}

	d := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Focus Demo"},
					{Kind: BlockCode, Lang: "go", Lines: []string{"func main() {", `    println("focus mode")`, "}"}},
					{Kind: BlockTable, Text: "| A | B |\n|---|---|\n| 1 | 2 |"},
					{Kind: BlockCallout, Callout: "note", Text: "Note card"},
					{Kind: BlockBranch, BranchKey: "1", Text: "Next", BranchTarget: "target"},
				},
			},
		},
	}

	ed.SlideIdx = 0
	ed.BlockIdx = 1 // Focused on code block
	ed.FocusMode = true

	focusView := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(focusView, "ZOOM FOCUS MODE") || !strings.Contains(focusView, "Code: go") {
		t.Fatalf("expected focus mode header in View, got:\n%s", focusView)
	}
	if !strings.Contains(focusView, "println(\"focus mode\")") {
		t.Fatalf("expected focused block body in View, got:\n%s", focusView)
	}

	// Test scrolling in focus mode
	ed.FocusScroll = 2
	scrollFocus := stripANSI(renderFocusMode(d, ed, 80, 20))
	if !strings.Contains(scrollFocus, "more content above") {
		t.Fatalf("expected scroll up indicator, got:\n%s", scrollFocus)
	}

	// Test focus mode on table
	ed.BlockIdx = 2
	tableFocus := stripANSI(renderFocusMode(d, ed, 80, 24))
	if !strings.Contains(tableFocus, "Table") {
		t.Fatalf("expected Table kind in focus view, got:\n%s", tableFocus)
	}

	// Test focus mode on callout
	ed.BlockIdx = 3
	calloutFocus := stripANSI(renderFocusMode(d, ed, 80, 24))
	if !strings.Contains(calloutFocus, "Card: note") {
		t.Fatalf("expected Card: note kind in focus view, got:\n%s", calloutFocus)
	}

	// Test focus mode on branch
	ed.BlockIdx = 4
	branchFocus := stripANSI(renderFocusMode(d, ed, 80, 24))
	if !strings.Contains(branchFocus, "Branch Fork") {
		t.Fatalf("expected Branch Fork kind in focus view, got:\n%s", branchFocus)
	}

	// Focus mode with runner card attached
	ed.BlockIdx = 1
	ed.ShowRunner = true
	ed.RunnerResult = &ExecResult{Language: "go", ExitCode: 0, Stdout: "focus runner result"}
	focusWithRunner := stripANSI(renderFocusMode(d, ed, 80, 30))
	if !strings.Contains(focusWithRunner, "focus runner result") {
		t.Fatalf("expected runner output inside focus mode, got:\n%s", focusWithRunner)
	}

	// Nav status with focus and running code
	ed.RunningCode = true
	nav := stripANSI(navStatus(d, ed, 240))
	if !strings.Contains(nav, "running code") || !strings.Contains(nav, "zoom: on") {
		t.Fatalf("expected navStatus badges for running code and zoom, got:\n%s", nav)
	}
}

func TestRenderColumnsAndTrackModal(t *testing.T) {
	src := `---
title: Columns and Track View Test
---

::id s1
::tags backend,arch
# Slide 1

:::columns
### Left
Left side content.
:::col
### Right
Right side content.
:::

---

::id s2
::tags frontend
# Slide 2
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")

	// 1. Render normal view with columns
	viewNormal := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(viewNormal, "Left") || !strings.Contains(viewNormal, "Right") {
		t.Fatalf("expected Left and Right column content in view, got:\n%s", viewNormal)
	}

	// 2. Render focus mode on columns
	ed.BlockIdx = 1 // columns block
	ed.FocusMode = true
	viewFocus := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(viewFocus, "Columns: 2 split") {
		t.Fatalf("expected 'Columns: 2 split' in focus mode header, got:\n%s", viewFocus)
	}
	ed.FocusMode = false

	// 3. Render Track Modal
	ed.ShowTrackModal = true
	viewTrack := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(viewTrack, "Audience Tracks") || !strings.Contains(viewTrack, "backend") || !strings.Contains(viewTrack, "frontend") {
		t.Fatalf("expected Audience Tracks modal with tags, got:\n%s", viewTrack)
	}
	ed.ShowTrackModal = false

	// 4. Render Graph Modal with active track
	ed.ActiveTrack = "backend"
	ed.ShowGraphMap = true
	viewGraph := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(viewGraph, "★ Track: backend") {
		t.Fatalf("expected Track header in graph modal, got:\n%s", viewGraph)
	}
	if !strings.Contains(viewGraph, "★") {
		t.Fatalf("expected star badge for track nodes in graph modal, got:\n%s", viewGraph)
	}
	ed.ShowGraphMap = false

	// 5. Render navStatus with active track
	status := stripANSI(navStatus(d, ed, 200))
	if !strings.Contains(status, "[★ track: backend]") {
		t.Fatalf("expected '[★ track: backend]' in nav status, got:\n%s", status)
	}
}

func TestRenderRouteModal(t *testing.T) {
	// 1. Empty routes
	dEmpty := Deck{Slides: []Slide{{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Solo"}}}}}
	ed := NewEditor("test.deck.md")
	emptyModal := stripANSI(renderRouteModal(dEmpty, ed, 100, 30))
	if !strings.Contains(emptyModal, "Preset Graph Routes") || !strings.Contains(emptyModal, "No preset routes defined.") {
		t.Fatalf("expected empty route modal message, got:\n%s", emptyModal)
	}

	// 2. Populated routes
	src := `---
title: Route View Test
routes:
  quick: intro -> outro
  deepdive: intro -> deep -> outro
---

::id intro
# Intro
Welcome to the talk.

---

::id deep
# Deep Dive
Technical details and architectures.

---

::id outro
# Conclusion
Summary and thank you.
`
	d := ParseDeck(src)
	ed = NewEditor("test.deck.md")
	ed.ShowRouteModal = true

	modalView := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(modalView, "Preset Graph Routes") {
		t.Fatalf("expected Preset Graph Routes title in View, got:\n%s", modalView)
	}
	if !strings.Contains(modalView, "quick") || !strings.Contains(modalView, "deepdive") {
		t.Fatalf("expected route names in modal view, got:\n%s", modalView)
	}
	if !strings.Contains(modalView, "slides") || !strings.Contains(modalView, "──►") {
		t.Fatalf("expected slide count and path arrows in modal, got:\n%s", modalView)
	}

	// Active route badge
	ed.ActiveRoute = "quick"
	ed.RouteStep = 0
	activeModal := stripANSI(renderRouteModal(d, ed, 100, 30))
	if !strings.Contains(activeModal, "● active") {
		t.Fatalf("expected ● active badge for active route, got:\n%s", activeModal)
	}

	// 3. navStatus with active route
	ed.ShowRouteModal = false
	status := stripANSI(navStatus(d, ed, 200))
	if !strings.Contains(status, "[⚡ route: quick (1/2)]") || !strings.Contains(status, "P route") {
		t.Fatalf("expected route status and P hint in navStatus, got:\n%s", status)
	}

	// 4. Graph modal with active route
	ed.ShowGraphMap = true
	graphView := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(graphView, "⚡ Route: quick") {
		t.Fatalf("expected active route in graph modal header, got:\n%s", graphView)
	}
	if !strings.Contains(graphView, "#1") {
		t.Fatalf("expected #1 route step badge in graph modal, got:\n%s", graphView)
	}
}

func TestRenderHistoryModal(t *testing.T) {
	// 1. Empty history
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Intro"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Architecture"}}},
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Conclusion"}}},
		},
	}
	ed := NewEditor("test.deck.md")
	emptyModal := stripANSI(renderHistoryModal(d, ed, 100, 30))
	if !strings.Contains(emptyModal, "Presentation Traversal History") || !strings.Contains(emptyModal, "start of your presentation traversal") {
		t.Fatalf("expected empty history modal message, got:\n%s", emptyModal)
	}

	// 2. Populated history
	ed.History = []int{0, 1}
	ed.SlideIdx = 2
	ed.ShowHistoryModal = true

	modalView := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(modalView, "Presentation Traversal History") {
		t.Fatalf("expected title in modalView, got:\n%s", modalView)
	}
	if !strings.Contains(modalView, "Intro") || !strings.Contains(modalView, "Architecture") || !strings.Contains(modalView, "Conclusion") {
		t.Fatalf("expected slide titles in history modal, got:\n%s", modalView)
	}
	if !strings.Contains(modalView, "● CURRENT") || !strings.Contains(modalView, "step back") {
		t.Fatalf("expected CURRENT and step back badges, got:\n%s", modalView)
	}

	// 3. navStatus with history
	ed.ShowHistoryModal = false
	status := stripANSI(navStatus(d, ed, 200))
	if !strings.Contains(status, "[history: 2 (H)]") || !strings.Contains(status, "H history") {
		t.Fatalf("expected history status and H hint in navStatus, got:\n%s", status)
	}
}

func TestRenderBranchHUDModal(t *testing.T) {
	src := `---
title: Branch View Test
routes:
  demo: hub -> target-a
---

::id hub
# Hub Slide
::branch [1] Option Alpha -> target-a
::branch [2] Option Beta -> target-b

---

::id target-a
::tags backend
# Target Alpha
Live microservices code:
` + "```go\nfunc StreamEvents() {}\n```" + `

---

::id target-b
# Target Beta
Simple summary text.

---

::id terminal
# Terminal Slide
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")

	// 1. Terminal slide with no outgoing branches
	ed.SlideIdx = 3
	emptyHUD := stripANSI(renderBranchHUDModal(d, ed, 100, 30))
	if !strings.Contains(emptyHUD, "Branch Decision Fork HUD") || !strings.Contains(emptyHUD, "No outgoing branches or links") {
		t.Fatalf("expected terminal slide message in HUD, got:\n%s", emptyHUD)
	}

	// 2. Hub slide HUD with options and previews
	ed.SlideIdx = 0
	ed.ShowBranchHUD = true
	ed.ActiveTrack = "backend"
	ed.ActiveRoute = "demo"

	hudView := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(hudView, "Branch Decision Fork HUD") {
		t.Fatalf("expected Branch Decision Fork HUD title in view, got:\n%s", hudView)
	}
	if !strings.Contains(hudView, "Option Alpha") || !strings.Contains(hudView, "Option Beta") {
		t.Fatalf("expected Option Alpha and Option Beta in HUD, got:\n%s", hudView)
	}
	if !strings.Contains(hudView, "Track Match") {
		t.Fatalf("expected Track Match badge in HUD, got:\n%s", hudView)
	}
	if !strings.Contains(hudView, "Route Step") {
		t.Fatalf("expected Route Step badge in HUD, got:\n%s", hudView)
	}
	if !strings.Contains(hudView, "Destination Preview: [02] Target Alpha") {
		t.Fatalf("expected destination preview box for selected target, got:\n%s", hudView)
	}
	if !strings.Contains(hudView, "StreamEvents") {
		t.Fatalf("expected code preview in destination preview card, got:\n%s", hudView)
	}

	// 3. Test cursor on Option Beta
	ed.BranchHUDCursor = 1
	hudBeta := stripANSI(renderBranchHUDModal(d, ed, 100, 30))
	if !strings.Contains(hudBeta, "Destination Preview: [03] Target Beta") || !strings.Contains(hudBeta, "Simple summary text") {
		t.Fatalf("expected Beta preview content, got:\n%s", hudBeta)
	}

	// 4. Test navStatus with fork (J) hint
	ed.ShowBranchHUD = false
	status := stripANSI(navStatus(d, ed, 240))
	if !strings.Contains(status, "[fork: 2 paths (J)") || !strings.Contains(status, "J fork") {
		t.Fatalf("expected fork (J) and J fork in navStatus, got:\n%s", status)
	}

	// 5. Test renderHelpModal documents J
	help := stripANSI(renderHelpModal(100, 40))
	if !strings.Contains(help, "Branch fork HUD") {
		t.Fatalf("expected J documented in help modal, got:\n%s", help)
	}
}

func TestRenderWaypointModal(t *testing.T) {
	src := `---
title: Waypoint Pathfinder View Test
---

::id s1
# Origin Slide
::branch [1] To Mid -> s2

---

::id s2
::tags arch,core
# Mid Slide
::next s3

---

::id s3
# Goal Slide
::next s1

---

::id s4
# Slide Four
::next s1

---

::id s5
# Slide Five
::next s1

---

::id s6
# Slide Six
::next s1

---

::id isolated
# Isolated Slide
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")

	// 1. Render normal View with ShowWaypointModal = true
	ed.SlideIdx = 0
	ed.ShowWaypointModal = true
	viewModal := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(viewModal, "Waypoint Pathfinder & Graph Routing") {
		t.Fatalf("expected modal title in view, got:\n%s", viewModal)
	}
	if !strings.Contains(viewModal, "Origin: [01] Origin Slide") {
		t.Fatalf("expected origin slide in modal, got:\n%s", viewModal)
	}
	if !strings.Contains(viewModal, "Search: (all destinations)") {
		t.Fatalf("expected all destinations search prompt, got:\n%s", viewModal)
	}
	if !strings.Contains(viewModal, "[02] Mid Slide") || !strings.Contains(viewModal, "──[1]──►") {
		t.Fatalf("expected Mid Slide with branch key edge trail, got:\n%s", viewModal)
	}
	if !strings.Contains(viewModal, "talk time") || !strings.Contains(viewModal, "[arch,core]") {
		t.Fatalf("expected metrics and tags, got:\n%s", viewModal)
	}

	// 2. Cursor scrolling and pagination (> maxVisible)
	ed.WaypointCursor = 5
	viewScroll := stripANSI(renderWaypointModal(d, ed, 100, 30))
	if !strings.Contains(viewScroll, "showing") || !strings.Contains(viewScroll, "candidates") {
		t.Fatalf("expected candidate pagination indicator, got:\n%s", viewScroll)
	}
	if !strings.Contains(viewScroll, "No downstream path from current slide") {
		t.Fatalf("expected isolated slide to be marked unreachable, got:\n%s", viewScroll)
	}

	// 3. Search query with matches and without matches
	ed.WaypointQuery = "Goal"
	ed.WaypointCursor = 0
	viewQueryMatch := stripANSI(renderWaypointModal(d, ed, 100, 30))
	if !strings.Contains(viewQueryMatch, "Search: Goal") || !strings.Contains(viewQueryMatch, "[03] Goal Slide") {
		t.Fatalf("expected Goal Slide query match, got:\n%s", viewQueryMatch)
	}

	ed.WaypointQuery = "nonexistent-query-string"
	viewNoMatch := stripANSI(renderWaypointModal(d, ed, 100, 30))
	if !strings.Contains(viewNoMatch, `No destinations matching "nonexistent-query-string"`) {
		t.Fatalf("expected no matches message, got:\n%s", viewNoMatch)
	}

	// 4. Single-slide deck empty candidate display
	singleDeck := Deck{Slides: []Slide{{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Solo"}}}}}
	soloEd := NewEditor("test.deck.md")
	viewSolo := stripANSI(renderWaypointModal(singleDeck, soloEd, 100, 30))
	if !strings.Contains(viewSolo, "No other slides in presentation") {
		t.Fatalf("expected no other slides message in solo deck, got:\n%s", viewSolo)
	}

	// 5. Small terminal dimensions constraint
	smallView := stripANSI(renderWaypointModal(d, ed, 30, 20))
	if !strings.Contains(smallView, "Waypoint Pathfinder") {
		t.Fatalf("expected waypoint rendered in small terminal, got:\n%s", smallView)
	}

	// 6. Navigation status bar contains W waypoint
	ed.ShowWaypointModal = false
	status := stripANSI(navStatus(d, ed, 240))
	if !strings.Contains(status, "W waypoint") {
		t.Fatalf("expected 'W waypoint' in navStatus, got:\n%s", status)
	}

	// 7. Help modal documents Waypoint Pathfinder
	help := stripANSI(renderHelpModal(100, 40))
	if !strings.Contains(help, "Waypoint pathfinder") {
		t.Fatalf("expected W documented in help modal, got:\n%s", help)
	}
}

func TestRenderRadarModal(t *testing.T) {
	src := `---
title: Radar View Test
---

::id hub
# Hub Slide
::branch [1] Branch A -> bA
::branch [2] Branch B -> bB
::branch [3] Branch C -> bC
::branch [4] Branch D -> bD
::branch [5] Branch E -> bE
::branch [6] Branch F -> bF
::branch [7] Branch G -> bG

---

::id bA
# Slide A
::next conclusion

---

::id bB
# Slide B
::next conclusion

---

::id bC
# Slide C
::next conclusion

---

::id bD
# Slide D
::next conclusion

---

::id bE
# Slide E
::next conclusion

---

::id bF
# Slide F
::next conclusion

---

::id bG
# Slide G
::next conclusion

---

::id conclusion
# Conclusion
`
	d := ParseDeck(src)
	ed := NewEditor("test.deck.md")

	// 1. Render normal View with ShowRadarModal = true
	ed.SlideIdx = 0
	ed.ShowRadarModal = true
	viewModal := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(viewModal, "Graph Exploration Radar & Branch Coverage") {
		t.Fatalf("expected radar title in view, got:\n%s", viewModal)
	}
	if !strings.Contains(viewModal, "Coverage:") || !strings.Contains(viewModal, "Speaking:") {
		t.Fatalf("expected coverage and speaking stats, got:\n%s", viewModal)
	}
	if !strings.Contains(viewModal, "Fork [01] #hub: Hub Slide") {
		t.Fatalf("expected fork header in modal, got:\n%s", viewModal)
	}
	if !strings.Contains(viewModal, "[1] Branch A") || !strings.Contains(viewModal, "[2] Branch B") {
		t.Fatalf("expected branch rows, got:\n%s", viewModal)
	}

	// 2. Cursor scrolling and pagination (> 5 branches)
	ed.RadarCursor = 6
	viewPaginated := stripANSI(renderRadarModal(d, ed, 100, 30))
	if !strings.Contains(viewPaginated, "showing") || !strings.Contains(viewPaginated, "branches") {
		t.Fatalf("expected pagination indicator in radar modal, got:\n%s", viewPaginated)
	}

	// 3. Linear deck with no branches
	linearDeck := Deck{Slides: []Slide{{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Linear"}}}}}
	soloEd := NewEditor("test.deck.md")
	viewLinear := stripANSI(renderRadarModal(linearDeck, soloEd, 100, 30))
	if !strings.Contains(viewLinear, "No branching decision points in presentation") {
		t.Fatalf("expected strictly linear notice, got:\n%s", viewLinear)
	}

	// 4. Small terminal width constraint
	smallView := stripANSI(renderRadarModal(d, ed, 30, 20))
	if !strings.Contains(smallView, "Graph Exploration Radar") {
		t.Fatalf("expected radar rendered in small terminal, got:\n%s", smallView)
	}

	// 5. Navigation status bar with radar badge and fork return
	ed.ShowRadarModal = false
	ed.SlideIdx = 1
	ed.History = []int{0}
	status := stripANSI(navStatus(d, ed, 240))
	if !strings.Contains(status, "[U: return to fork]") {
		t.Fatalf("expected '[U: return to fork]' in navStatus, got:\n%s", status)
	}
	if !strings.Contains(status, "radar:") || !strings.Contains(status, "(V)") {
		t.Fatalf("expected radar badge in navStatus, got:\n%s", status)
	}
	if !strings.Contains(status, "V radar") || !strings.Contains(status, "U fork") {
		t.Fatalf("expected 'V radar' and 'U fork' hints in navStatus, got:\n%s", status)
	}

	// 6. Help modal documents V and U
	help := stripANSI(renderHelpModal(100, 40))
	if !strings.Contains(help, "Graph exploration radar") || !strings.Contains(help, "Return to upstream branch fork") {
		t.Fatalf("expected V and U documented in help modal, got:\n%s", help)
	}
}

func TestRenderLoopCardAndView(t *testing.T) {
	// 1. Boundary cases
	if card := renderLoopCard(nil, 0, 3, 80); card != "" {
		t.Fatalf("expected empty card for nil loop config")
	}

	lcfg := &LoopConfig{
		Key:        "r",
		Label:      "TDD Loop",
		Target:     "tdd-red",
		MaxPasses:  3,
		ExitTarget: "tdd-summary",
	}

	// 2. Active loop card (pass 0/3)
	activeCard := stripANSI(renderLoopCard(lcfg, 0, 3, 80))
	if !strings.Contains(activeCard, "LOOP ITERATION: TDD Loop") || !strings.Contains(activeCard, "[pass 1/3]") {
		t.Fatalf("expected active loop header in card, got:\n%s", activeCard)
	}
	if !strings.Contains(activeCard, "Loop target: ──► #tdd-red") || !strings.Contains(activeCard, "Remaining iterations: 2") {
		t.Fatalf("expected loop target and remaining iterations in card, got:\n%s", activeCard)
	}
	if !strings.Contains(activeCard, "r: Loop Hotkey") {
		t.Fatalf("expected 'r: Loop Hotkey' in card hints, got:\n%s", activeCard)
	}

	// 3. Completed loop card (pass 3/3)
	completedCard := stripANSI(renderLoopCard(lcfg, 3, 3, 80))
	if !strings.Contains(completedCard, "LOOP COMPLETED: TDD Loop (3/3 passes)") {
		t.Fatalf("expected completed loop header, got:\n%s", completedCard)
	}
	if !strings.Contains(completedCard, "Exit target: ──► #tdd-summary") {
		t.Fatalf("expected exit target in completed card, got:\n%s", completedCard)
	}

	// 4. Slide rendering with loop card in View
	d := Deck{
		Slides: []Slide{
			{
				ID: "tdd-refactor",
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Refactoring Phase"},
					{Kind: BlockParagraph, Text: "Clean up code while tests are green."},
				},
				Loop: lcfg,
			},
			{
				ID: "tdd-red",
				Blocks: []Block{
					{Kind: BlockHeading, Level: 1, Text: "Failing Test Phase"},
				},
			},
		},
	}
	ed := NewEditor("test.deck.md")
	ed.SlideIdx = 0

	viewOutput := stripANSI(View(d, ed, 100, 30))
	if !strings.Contains(viewOutput, "LOOP ITERATION: TDD Loop") {
		t.Fatalf("expected loop card in View, got:\n%s", viewOutput)
	}

	// Hidden in focus mode
	ed.FocusMode = true
	focusView := stripANSI(View(d, ed, 100, 30))
	if strings.Contains(focusView, "LOOP ITERATION: TDD Loop") {
		t.Fatalf("expected loop card hidden in focus mode, got:\n%s", focusView)
	}
	ed.FocusMode = false

	// 5. navStatus badge
	statusActive := stripANSI(navStatus(d, ed, 200))
	if !strings.Contains(statusActive, "[⟳ loop: pass 1/3 (TDD Loop ──► tdd-red)]") {
		t.Fatalf("expected active loop badge in navStatus, got:\n%s", statusActive)
	}

	// Completed status badge
	ed.LoopCounters[0] = 3
	statusDone := stripANSI(navStatus(d, ed, 200))
	if !strings.Contains(statusDone, "[✔ loop: 3/3 done (TDD Loop)]") {
		t.Fatalf("expected done loop badge in navStatus, got:\n%s", statusDone)
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
