package internal

import (
	"strings"
	"testing"
)

func TestDoctorCleanDeck(t *testing.T) {
	src := `---
title: Clean Deck
theme: tokyo-night
---

# Slide 1
Welcome to the talk.

---

# Slide 2
All clean and good.
`
	deck := ParseDeck(src)
	report := RunDeckDoctor(deck, "clean.deck.md")

	if report.Errors != 0 {
		t.Fatalf("expected 0 errors on clean deck, got %d", report.Errors)
	}
	if report.Warnings != 0 {
		t.Fatalf("expected 0 warnings on clean deck, got %d", report.Warnings)
	}

	theme := ResolveTheme("tokyo-night")
	out := FormatDoctorCLI(report, theme)
	if !strings.Contains(out, "All systems verified") && !strings.Contains(out, "verified") {
		t.Fatalf("expected success message in doctor output, got:\n%s", out)
	}
}

func TestDoctorBrokenGraphDAG(t *testing.T) {
	src := `# Fork Slide
::branch [1] Missing -> non_existent_target
`
	deck := ParseDeck(src)
	report := RunDeckDoctor(deck, "broken_dag.deck.md")

	if report.Errors == 0 {
		t.Fatalf("expected error for broken link target, got 0")
	}

	theme := ResolveTheme("tokyo-night")
	out := FormatDoctorCLI(report, theme)
	if !strings.Contains(out, "[ERROR]") {
		t.Fatalf("expected [ERROR] badge in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Graph Topology") {
		t.Fatalf("expected 'Graph Topology' category in output, got:\n%s", out)
	}
}

func TestDoctorMissingImageAsset(t *testing.T) {
	src := `# Image Slide
![Architecture](does_not_exist_987654.png)
`
	deck := ParseDeck(src)
	report := RunDeckDoctor(deck, "missing_img.deck.md")

	if report.Errors == 0 {
		t.Fatalf("expected error for missing image asset, got 0")
	}

	foundImgError := false
	for _, res := range report.Results {
		if res.Category == "Asset Availability" && res.Severity == DoctorSeverityError {
			foundImgError = true
			break
		}
	}
	if !foundImgError {
		t.Fatalf("expected 'Asset Availability' error in report, got: %+v", report.Results)
	}
}

func TestDoctorEmptySlide(t *testing.T) {
	deck := Deck{
		Slides: []Slide{
			{Blocks: nil},
		},
	}
	report := RunDeckDoctor(deck, "empty.deck.md")

	foundEmpty := false
	for _, res := range report.Results {
		if res.Category == "Slide Content" && res.Severity == DoctorSeverityWarning {
			foundEmpty = true
			break
		}
	}
	if !foundEmpty {
		t.Fatalf("expected 'Slide Content' warning for empty slide")
	}
}

func TestDoctorViewportOverflow(t *testing.T) {
	var lines []string
	for i := 0; i < 40; i++ {
		lines = append(lines, "Bullet item line number")
	}
	deck := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockParagraph, Text: strings.Join(lines, "\n")},
				},
			},
		},
	}
	report := RunDeckDoctor(deck, "tall.deck.md")

	foundViewport := false
	for _, res := range report.Results {
		if res.Category == "Viewport Fit" {
			foundViewport = true
			break
		}
	}
	if !foundViewport {
		t.Fatalf("expected 'Viewport Fit' warning for tall slide")
	}
}

func TestDoctorToolchainCheck(t *testing.T) {
	langs := []string{
		"bash", "sh", "python", "py", "python3", "node", "js", "ts", "typescript",
		"ruby", "go", "rust", "c", "cpp", "lua", "perl", "php", "unknown_lang",
	}
	for _, l := range langs {
		_, _ = checkLanguageToolchain(l)
	}
}

func TestDoctorColumnsAndToolchains(t *testing.T) {
	deck := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{
						Kind: BlockColumns,
						Columns: [][]Block{
							{
								{Kind: BlockImage, Src: "missing_col_img.png"},
								{Kind: BlockCode, Lang: "python", Text: "print('hi')"},
								{Kind: BlockCode, Lang: "bash", Text: "echo 'col'"},
							},
							{
								{Kind: BlockList, Text: "- item 1\n- item 2"},
								{Kind: BlockHeading, Text: "Column Title"},
							},
						},
					},
				},
			},
		},
	}
	report := RunDeckDoctor(deck, "cols.deck.md")
	if report.Errors == 0 {
		t.Fatalf("expected error for missing image inside column")
	}

	theme := ResolveTheme("matrix")
	out := FormatDoctorCLI(report, theme)
	if !strings.Contains(out, "missing_col_img.png") {
		t.Fatalf("expected missing image in output: %s", out)
	}
}

func TestDoctorWarningsOnlyFormat(t *testing.T) {
	report := DoctorReport{
		DeckPath:    "warn.deck.md",
		TotalSlides: 1,
		TotalChecks: 1,
		Warnings:    1,
		Results: []DoctorCheckResult{
			{
				Category: "Viewport Fit",
				Severity: DoctorSeverityWarning,
				Message:  "Slide has too many lines",
				Advice:   "Break it up",
			},
		},
	}
	theme := ResolveTheme("dracula")
	out := FormatDoctorCLI(report, theme)
	if !strings.Contains(out, "[WARN]") {
		t.Fatalf("expected [WARN] in output: %s", out)
	}
	if !strings.Contains(out, "Found 1 warning(s)") {
		t.Fatalf("expected warning count in output: %s", out)
	}
}

type MockSpellCheck struct{}

func (m *MockSpellCheck) Name() string { return "Spell Checker" }

func (m *MockSpellCheck) Check(d Deck, filePath string) []DoctorCheckResult {
	return []DoctorCheckResult{
		{
			Category: "Spelling",
			Severity: DoctorSeverityInfo,
			SlideNum: 1,
			Message:  "Slide 1 has possible typo in word 'recieve'",
			Advice:   "Did you mean 'receive'?",
		},
	}
}

func TestCustomDoctorCheck_OCP(t *testing.T) {
	engine := NewDoctorEngine()
	engine.Register(&MockSpellCheck{})
	d := Deck{
		Slides: []Slide{
			{Blocks: []Block{{Kind: BlockHeading, Text: "Test"}}},
		},
	}
	report := engine.Run(d, "test.deck.md")
	foundSpell := false
	for _, res := range report.Results {
		if res.Category == "Spelling" {
			foundSpell = true
			break
		}
	}
	if !foundSpell {
		t.Fatalf("expected custom 'Spelling' check to be executed by engine")
	}
}


