package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatInlineHTML(t *testing.T) {
	// 1. Plain text with HTML special characters
	plain := formatInlineHTML("x < y && y > z & \"quotes\"")
	if !strings.Contains(plain, "&lt;") || !strings.Contains(plain, "&gt;") || !strings.Contains(plain, "&amp;") {
		t.Errorf("expected HTML escaping, got %q", plain)
	}

	// 2. Bold and italic
	styled := formatInlineHTML("**bold text** and *italic text*")
	if !strings.Contains(styled, "<strong>bold text</strong>") || !strings.Contains(styled, "<em>italic text</em>") {
		t.Errorf("expected bold and italic tags, got %q", styled)
	}

	// 3. Inline code span
	codeSpan := formatInlineHTML("use `git commit -m \"msg\"` command")
	if !strings.Contains(codeSpan, "<code>git commit -m &#34;msg&#34;</code>") {
		t.Errorf("expected code tags with escaped inner content, got %q", codeSpan)
	}

	// 4. Code span containing asterisks (collision-free check)
	asteriskCode := formatInlineHTML("pointer `*val` with **bold**")
	if !strings.Contains(asteriskCode, "<code>*val</code>") || !strings.Contains(asteriskCode, "<strong>bold</strong>") {
		t.Errorf("expected code span with asterisk preserved, got %q", asteriskCode)
	}
}

func TestExportHTML(t *testing.T) {
	src := `---
title: Demo Presentation
theme: dracula
align: left
---
# Welcome Slide
This is an introductory paragraph with **bold** and *italic* and ` + "`code`" + `.

---

# Code & Tables
` + "```go\nfunc main() {\n    println(\"hello world\")\n}\n```" + `

| Feature | Support |
|---|---|
| SSH | Yes |
| Tmux | Yes |

---

# Callouts & Tasks
> [!TIP]
> Keep database transactions small.

> [!WARNING]
> Breaking change in v2.

- [ ] Write integration tests
- [x] Implement export feature
- Regular bullet point

***

::image demo.png

::notes
These speaker notes must NOT be rendered on audience slides.
`
	d := ParseDeck(src)
	d.BaseDir = t.TempDir()

	htmlOut := ExportHTML(d, "")

	// 1. Structure checks
	if !strings.Contains(htmlOut, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE html")
	}
	if !strings.Contains(htmlOut, "<title>Demo Presentation</title>") {
		t.Errorf("expected title 'Demo Presentation' in HTML header, got %s", htmlOut)
	}
	if !strings.Contains(htmlOut, "class=\"deck-container\"") {
		t.Errorf("expected deck-container")
	}

	// 2. Slide count
	if strings.Count(htmlOut, "<section class=\"slide") != 3 {
		t.Errorf("expected 3 slides in HTML output, found %d", strings.Count(htmlOut, "<section class=\"slide"))
	}

	// 3. First slide active
	if !strings.Contains(htmlOut, "<section class=\"slide active") {
		t.Errorf("expected first slide to have 'active' class")
	}

	// 4. Code block
	if !strings.Contains(htmlOut, "class=\"code-block\" data-lang=\"go\"") {
		t.Errorf("expected code-block with data-lang='go'")
	}
	if !strings.Contains(htmlOut, "func main()") {
		t.Errorf("expected code line in HTML")
	}

	// 5. Table block
	if !strings.Contains(htmlOut, "<table class=\"deck-table\">") {
		t.Errorf("expected deck-table")
	}
	if !strings.Contains(htmlOut, "<th>Feature</th>") || !strings.Contains(htmlOut, "<td>SSH</td>") {
		t.Errorf("expected table cells in HTML")
	}

	// 6. Callout cards
	if !strings.Contains(htmlOut, "class=\"callout callout-tip\"") {
		t.Errorf("expected callout-tip in HTML")
	}
	if !strings.Contains(htmlOut, "Keep database transactions small.") {
		t.Errorf("expected callout text in HTML")
	}
	if !strings.Contains(htmlOut, "class=\"callout callout-warning\"") {
		t.Errorf("expected callout-warning in HTML")
	}

	// 7. Tasks
	if !strings.Contains(htmlOut, "class=\"task-item\"") {
		t.Errorf("expected task-item in HTML")
	}
	if !strings.Contains(htmlOut, "class=\"task-item task-done\"") {
		t.Errorf("expected task-done in HTML")
	}

	// 8. Divider
	if !strings.Contains(htmlOut, "<hr class=\"deck-divider\">") {
		t.Errorf("expected deck-divider in HTML")
	}

	// 9. Notes privacy check
	if strings.Contains(htmlOut, "These speaker notes must NOT be rendered") {
		t.Errorf("speaker notes leaked into audience HTML canvas!")
	}

	// 10. Navigation script
	if !strings.Contains(htmlOut, "const totalSlides = 3;") {
		t.Errorf("expected JS totalSlides = 3")
	}
	if !strings.Contains(htmlOut, "updateSlide(currentSlide + 1);") {
		t.Errorf("expected JS keyboard handler")
	}
}

func TestExportHTMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "sub", "deck.html")

	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Exported Title"}}},
		},
	}

	err := ExportHTMLFile(d, outPath)
	if err != nil {
		t.Fatalf("unexpected error exporting HTML file: %v", err)
	}

	bytes, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read exported HTML file: %v", err)
	}
	if !strings.Contains(string(bytes), "Exported Title") {
		t.Errorf("expected 'Exported Title' in exported file")
	}
}

func TestEditorExportHTML(t *testing.T) {
	tmpDir := t.TempDir()
	deckPath := filepath.Join(tmpDir, "my_talk.deck.md")

	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Level: 1, Text: "Talk Slide"}}},
		},
	}

	ed := NewEditor(deckPath)
	outPath, err := ed.ExportHTML(&d)
	if err != nil {
		t.Fatalf("unexpected error from ed.ExportHTML: %v", err)
	}
	expectedOut := filepath.Join(tmpDir, "my_talk.html")
	if outPath != expectedOut {
		t.Errorf("expected output path %q, got %q", expectedOut, outPath)
	}
	if ed.Message != "exported to my_talk.html" {
		t.Errorf("expected status message 'exported to my_talk.html', got %q", ed.Message)
	}

	// Fallback when FilePath is empty
	edEmpty := NewEditor("")
	outEmpty, err := edEmpty.ExportHTML(&d)
	if err != nil {
		t.Fatalf("unexpected error from edEmpty.ExportHTML: %v", err)
	}
	if outEmpty != "deck.html" {
		t.Errorf("expected default 'deck.html', got %q", outEmpty)
	}
	_ = os.Remove("deck.html") // Clean up generated file
}
