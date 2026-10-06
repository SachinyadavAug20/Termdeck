package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatDeckBasic(t *testing.T) {
	input := `# Slide 1   
Some text here   

---

# Slide 2
Another slide
`
	expected := `# Slide 1
Some text here

---

# Slide 2
Another slide
`
	formatted := FormatDeck(input)
	if formatted != expected {
		t.Fatalf("expected:\n%q\ngot:\n%q", expected, formatted)
	}
}

func TestFormatDeckWithFrontmatter(t *testing.T) {
	input := `---
title: My Presentation   
theme: tokyo-night   
---

# Title Slide
Hello World!
`
	expected := `---
title: My Presentation
theme: tokyo-night
---

# Title Slide
Hello World!
`
	formatted := FormatDeck(input)
	if formatted != expected {
		t.Fatalf("expected:\n%q\ngot:\n%q", expected, formatted)
	}
}

func TestFormatDeckDirectives(t *testing.T) {
	input := `# Directives Test
:: tags :  intro, guide  
:: branch   [1] Learn More -> learn
:: track:  devs
:: route:  quick-tour
:: loop:  5
:: bg:  #1a1b26
:: align:  center
:: note:  Remember to explain this well!
:: custom:  value
`
	formatted := FormatDeck(input)
	if !strings.Contains(formatted, "::tags: intro, guide") {
		t.Errorf("expected normalized ::tags, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "::branch [1] Learn More -> learn") {
		t.Errorf("expected normalized ::branch, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "::track: devs") {
		t.Errorf("expected normalized ::track, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "::route: quick-tour") {
		t.Errorf("expected normalized ::route, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "::loop: 5") {
		t.Errorf("expected normalized ::loop, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "::bg: #1a1b26") {
		t.Errorf("expected normalized ::bg, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "::align: center") {
		t.Errorf("expected normalized ::align, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, "::note: Remember to explain this well!") {
		t.Errorf("expected normalized ::note, got:\n%s", formatted)
	}
	if !strings.Contains(formatted, ":: custom:  value") {
		t.Errorf("expected custom directive preserved, got:\n%s", formatted)
	}
}

func TestFormatDeckCodeBlockPreservation(t *testing.T) {
	input := `# Code Slide

` + "```go" + `
package main

func main() {
    println("hello")
}
` + "```" + `
`
	formatted := FormatDeck(input)
	if !strings.Contains(formatted, "func main() {\n    println(\"hello\")\n}") {
		t.Fatalf("code block indentation was modified unexpectedly:\n%s", formatted)
	}
}

func TestFormatDeckConsecutiveBlankLines(t *testing.T) {
	input := `# Slide 1



Paragraph with extra blank lines above it.




---



# Slide 2
End.
`
	formatted := FormatDeck(input)
	if strings.Contains(formatted, "\n\n\n") {
		// Should not have triple newlines within a slide
		// Separator between slides is \n\n---\n\n which has at most 2 newlines consecutively
		for _, part := range strings.Split(formatted, "\n\n---\n\n") {
			if strings.Contains(part, "\n\n\n") {
				t.Errorf("found excessive consecutive blank lines inside slide:\n%s", part)
			}
		}
	}
}

func TestFormatDeckFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sample.deck.md")

	unformatted := "# Title   \n\n\nHello   \n"
	if err := os.WriteFile(filePath, []byte(unformatted), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 1. Check only: returns isFormatted=false without modifying file
	isFmt, _, err := FormatDeckFile(filePath, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isFmt {
		t.Errorf("expected isFormatted=false for unformatted file")
	}
	data, _ := os.ReadFile(filePath)
	if string(data) != unformatted {
		t.Errorf("file was modified during checkOnly")
	}

	// 2. Format in-place: writes changes to disk
	isFmt2, out2, err2 := FormatDeckFile(filePath, false)
	if err2 != nil {
		t.Fatalf("unexpected error: %v", err2)
	}
	if isFmt2 {
		t.Errorf("expected isFormatted=false before write")
	}
	data2, _ := os.ReadFile(filePath)
	if string(data2) != out2 {
		t.Errorf("file content on disk did not match formatted content")
	}

	// 3. Re-run format on already formatted file
	isFmt3, _, err3 := FormatDeckFile(filePath, true)
	if err3 != nil {
		t.Fatalf("unexpected error: %v", err3)
	}
	if !isFmt3 {
		t.Errorf("expected isFormatted=true on already formatted file")
	}

	// 4. Non-existent file error
	_, _, errMissing := FormatDeckFile(filepath.Join(tmpDir, "missing.md"), false)
	if errMissing == nil {
		t.Errorf("expected error for missing file")
	}
}
