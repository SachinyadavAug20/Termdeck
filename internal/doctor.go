package internal

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DoctorSeverity defines the severity level of a check finding.
type DoctorSeverity string

const (
	DoctorSeverityError   DoctorSeverity = "ERROR"
	DoctorSeverityWarning DoctorSeverity = "WARNING"
	DoctorSeverityInfo    DoctorSeverity = "INFO"
	DoctorSeverityPass    DoctorSeverity = "PASS"
)

// DoctorCheckResult represents a single diagnostic finding.
type DoctorCheckResult struct {
	Category string
	Severity DoctorSeverity
	Message  string
	SlideNum int // 1-based; 0 for deck-wide checks
	Advice   string
}

// DoctorReport aggregates all diagnostic checks performed across a presentation.
type DoctorReport struct {
	DeckPath    string
	TotalSlides int
	TotalChecks int
	Errors      int
	Warnings    int
	Infos       int
	Passes      int
	Results     []DoctorCheckResult
}

// DoctorCheck defines the interface for a pluggable presentation diagnostic check (SRP & ISP).
type DoctorCheck interface {
	Name() string
	Check(d Deck, filePath string) []DoctorCheckResult
}

// DoctorEngine orchestrates diagnostic checks across a presentation (DIP & OCP).
type DoctorEngine struct {
	checks []DoctorCheck
}

// NewDoctorEngine initializes an engine with standard default presentation checks.
func NewDoctorEngine() *DoctorEngine {
	e := &DoctorEngine{}
	e.Register(&DAGTopologyCheck{})
	e.Register(&AssetAvailabilityCheck{})
	e.Register(&ToolchainAvailabilityCheck{})
	e.Register(&ViewportFitCheck{})
	e.Register(&SlideContentCheck{})
	return e
}

// Register adds a new diagnostic check rule without modifying existing check logic (OCP).
func (e *DoctorEngine) Register(check DoctorCheck) {
	e.checks = append(e.checks, check)
}

// Run executes all registered diagnostic checks on the specified deck.
func (e *DoctorEngine) Run(d Deck, filePath string) DoctorReport {
	report := DoctorReport{
		DeckPath:    filePath,
		TotalSlides: len(d.Slides),
	}

	for _, check := range e.checks {
		results := check.Check(d, filePath)
		for _, res := range results {
			report.addResult(res)
		}
	}

	if len(report.Results) == 0 {
		report.addResult(DoctorCheckResult{
			Category: "Health Check",
			Severity: DoctorSeverityPass,
			Message:  "All presentation integrity, asset, and runtime checks passed flawlessly.",
			SlideNum: 0,
			Advice:   "Your deck is fully verified and ready for live presentation.",
		})
	}

	return report
}

// DefaultDoctorEngine is the global doctor diagnostic engine instance.
var DefaultDoctorEngine = NewDoctorEngine()

// RegisterDoctorCheck adds a new diagnostic check to the global engine (OCP).
func RegisterDoctorCheck(check DoctorCheck) {
	DefaultDoctorEngine.Register(check)
}

// RunDeckDoctor performs an exhaustive audit of presentation topology, assets,
// toolchain prerequisites, and slide layout metrics using the default engine.
func RunDeckDoctor(d Deck, filePath string) DoctorReport {
	return DefaultDoctorEngine.Run(d, filePath)
}

// DAGTopologyCheck validates graph DAG topology, broken branch targets, and cycles.
type DAGTopologyCheck struct{}

func (c *DAGTopologyCheck) Name() string { return "DAG Topology" }

func (c *DAGTopologyCheck) Check(d Deck, filePath string) []DoctorCheckResult {
	var results []DoctorCheckResult
	lintIssues := LintGraph(d)
	for _, issue := range lintIssues {
		sev := DoctorSeverityWarning
		if issue.Severity == SeverityError {
			sev = DoctorSeverityError
		}
		msg := issue.Message
		if issue.Title != "" {
			msg = fmt.Sprintf("[%s] %s", issue.Title, issue.Message)
		}
		slideNum := issue.SlideIdx + 1
		results = append(results, DoctorCheckResult{
			Category: "Graph Topology",
			Severity: sev,
			Message:  msg,
			SlideNum: slideNum,
			Advice:   "Verify slide anchors {#id} or branch destinations in presentation.",
		})
	}
	return results
}

// AssetAvailabilityCheck verifies all local image assets referenced in slides exist on disk.
type AssetAvailabilityCheck struct{}

func (c *AssetAvailabilityCheck) Name() string { return "Asset Availability" }

func (c *AssetAvailabilityCheck) Check(d Deck, filePath string) []DoctorCheckResult {
	var results []DoctorCheckResult
	for i, slide := range d.Slides {
		slideNum := i + 1
		checkSlideAssets(slide, slideNum, d.BaseDir, func(res DoctorCheckResult) {
			results = append(results, res)
		})
	}
	return results
}

// ToolchainAvailabilityCheck verifies compiler and interpreter binaries in $PATH.
type ToolchainAvailabilityCheck struct{}

func (c *ToolchainAvailabilityCheck) Name() string { return "Runtime Toolchains" }

func (c *ToolchainAvailabilityCheck) Check(d Deck, filePath string) []DoctorCheckResult {
	var results []DoctorCheckResult
	checkedLangs := make(map[string]bool)
	for i, slide := range d.Slides {
		slideNum := i + 1
		checkSlideToolchains(slide, slideNum, checkedLangs, func(res DoctorCheckResult) {
			results = append(results, res)
		})
	}
	return results
}

// ViewportFitCheck verifies that slides fit within standard terminal display boundaries.
type ViewportFitCheck struct{}

func (c *ViewportFitCheck) Name() string { return "Viewport Fit" }

func (c *ViewportFitCheck) Check(d Deck, filePath string) []DoctorCheckResult {
	var results []DoctorCheckResult
	for i, slide := range d.Slides {
		slideNum := i + 1
		checkSlideViewport(slide, slideNum, func(res DoctorCheckResult) {
			results = append(results, res)
		})
	}
	return results
}

// SlideContentCheck verifies that slides are not completely empty.
type SlideContentCheck struct{}

func (c *SlideContentCheck) Name() string { return "Slide Content" }

func (c *SlideContentCheck) Check(d Deck, filePath string) []DoctorCheckResult {
	var results []DoctorCheckResult
	for i, slide := range d.Slides {
		if len(slide.Blocks) == 0 {
			results = append(results, DoctorCheckResult{
				Category: "Slide Content",
				Severity: DoctorSeverityWarning,
				SlideNum: i + 1,
				Message:  fmt.Sprintf("Slide %d is empty and contains no content blocks", i+1),
				Advice:   "Add presentation content or remove the empty slide delimiter.",
			})
		}
	}
	return results
}

func (r *DoctorReport) addResult(res DoctorCheckResult) {
	r.Results = append(r.Results, res)
	r.TotalChecks++
	switch res.Severity {
	case DoctorSeverityError:
		r.Errors++
	case DoctorSeverityWarning:
		r.Warnings++
	case DoctorSeverityInfo:
		r.Infos++
	case DoctorSeverityPass:
		r.Passes++
	}
}

func checkSlideAssets(slide Slide, slideNum int, baseDir string, emit func(DoctorCheckResult)) {
	var checkBlock func(b Block)
	checkBlock = func(b Block) {
		if b.Kind == BlockImage {
			src := b.Src
			if !strings.HasPrefix(src, "http://") && !strings.HasPrefix(src, "https://") {
				if _, found := ResolveImagePath(src, baseDir); !found {
					emit(DoctorCheckResult{
						Category: "Asset Availability",
						Severity: DoctorSeverityError,
						SlideNum: slideNum,
						Message:  fmt.Sprintf("Slide %d references image %q which could not be found on disk", slideNum, src),
						Advice:   fmt.Sprintf("Place image at %q relative to deck or in ./images/ or ./assets/ directory.", src),
					})
				}
			}
		}
		if b.Kind == BlockColumns {
			for _, col := range b.Columns {
				for _, subBlock := range col {
					checkBlock(subBlock)
				}
			}
		}
	}

	for _, b := range slide.Blocks {
		checkBlock(b)
	}
}

func checkSlideToolchains(slide Slide, slideNum int, checkedLangs map[string]bool, emit func(DoctorCheckResult)) {
	var checkBlock func(b Block)
	checkBlock = func(b Block) {
		if b.Kind == BlockCode && !b.NoEval && b.Lang != "" && IsExecutableLanguage(b.Lang) {
			lang := strings.ToLower(strings.TrimSpace(b.Lang))
			if !checkedLangs[lang] {
				checkedLangs[lang] = true
				binName, exists := checkLanguageToolchain(lang)
				if !exists {
					emit(DoctorCheckResult{
						Category: "Runtime Toolchains",
						Severity: DoctorSeverityWarning,
						SlideNum: slideNum,
						Message:  fmt.Sprintf("Slide %d code snippet uses %q, but runtime binary %q is not installed in $PATH", slideNum, lang, binName),
						Advice:   fmt.Sprintf("Install %q on the presenter machine or mark snippet 'noeval' if non-executable.", binName),
					})
				}
			}
		}
		if b.Kind == BlockColumns {
			for _, col := range b.Columns {
				for _, subBlock := range col {
					checkBlock(subBlock)
				}
			}
		}
	}

	for _, b := range slide.Blocks {
		checkBlock(b)
	}
}

func checkLanguageToolchain(lang string) (string, bool) {
	switch lang {
	case "python", "py", "python3":
		if _, err := exec.LookPath("python3"); err == nil {
			return "python3", true
		}
		if _, err := exec.LookPath("python"); err == nil {
			return "python", true
		}
		return "python3", false

	case "node", "js", "javascript":
		if _, err := exec.LookPath("node"); err == nil {
			return "node", true
		}
		if _, err := exec.LookPath("bun"); err == nil {
			return "bun", true
		}
		if _, err := exec.LookPath("deno"); err == nil {
			return "deno", true
		}
		return "node", false

	case "ts", "typescript":
		if _, err := exec.LookPath("deno"); err == nil {
			return "deno", true
		}
		if _, err := exec.LookPath("bun"); err == nil {
			return "bun", true
		}
		if _, err := exec.LookPath("tsx"); err == nil {
			return "tsx", true
		}
		if _, err := exec.LookPath("ts-node"); err == nil {
			return "ts-node", true
		}
		return "deno / bun / tsx", false

	case "ruby", "rb":
		if _, err := exec.LookPath("ruby"); err == nil {
			return "ruby", true
		}
		return "ruby", false

	case "go", "golang":
		if _, err := exec.LookPath("go"); err == nil {
			return "go", true
		}
		return "go", false

	case "rust", "rs":
		if _, err := exec.LookPath("rustc"); err == nil {
			return "rustc", true
		}
		return "rustc", false

	case "c":
		if _, err := exec.LookPath("gcc"); err == nil {
			return "gcc", true
		}
		if _, err := exec.LookPath("clang"); err == nil {
			return "clang", true
		}
		return "gcc", false

	case "cpp", "c++", "cc":
		if _, err := exec.LookPath("g++"); err == nil {
			return "g++", true
		}
		if _, err := exec.LookPath("clang++"); err == nil {
			return "clang++", true
		}
		return "g++", false

	case "lua", "luajit":
		if _, err := exec.LookPath("lua"); err == nil {
			return "lua", true
		}
		if _, err := exec.LookPath("luajit"); err == nil {
			return "luajit", true
		}
		return "lua", false

	case "perl", "pl":
		if _, err := exec.LookPath("perl"); err == nil {
			return "perl", true
		}
		return "perl", false

	case "php":
		if _, err := exec.LookPath("php"); err == nil {
			return "php", true
		}
		return "php", false

	default: // shell
		if _, err := exec.LookPath("bash"); err == nil {
			return "bash", true
		}
		if _, err := exec.LookPath("sh"); err == nil {
			return "sh", true
		}
		return "sh", false
	}
}

func checkSlideViewport(slide Slide, slideNum int, emit func(DoctorCheckResult)) {
	if len(slide.Blocks) == 0 {
		return
	}

	lineCount := 0
	maxLineLen := 0

	for _, b := range slide.Blocks {
		switch b.Kind {
		case BlockHeading:
			lineCount += 2
			if len([]rune(b.Text)) > maxLineLen {
				maxLineLen = len([]rune(b.Text))
			}
		case BlockParagraph, BlockCallout:
			lines := strings.Split(b.Text, "\n")
			lineCount += len(lines) + 1
			for _, l := range lines {
				if len([]rune(l)) > maxLineLen {
					maxLineLen = len([]rune(l))
				}
			}
		case BlockCode:
			lines := b.Lines
			if len(lines) == 0 {
				lines = strings.Split(b.Text, "\n")
			}
			lineCount += len(lines) + 2
			for _, l := range lines {
				if len([]rune(l)) > maxLineLen {
					maxLineLen = len([]rune(l))
				}
			}
		case BlockList:
			lines := strings.Split(b.Text, "\n")
			lineCount += len(lines)
			for _, l := range lines {
				if len([]rune(l)) > maxLineLen {
					maxLineLen = len([]rune(l))
				}
			}
		default:
			lineCount++
		}
	}

	if lineCount > 32 {
		emit(DoctorCheckResult{
			Category: "Viewport Fit",
			Severity: DoctorSeverityWarning,
			SlideNum: slideNum,
			Message:  fmt.Sprintf("Slide %d has ~%d vertical lines (recommended <= 30 to avoid terminal clipping)", slideNum, lineCount),
			Advice:   "Split into multiple slides or use columns (::: cols) to utilize horizontal terminal width.",
		})
	}

	if maxLineLen > 100 {
		emit(DoctorCheckResult{
			Category: "Viewport Fit",
			Severity: DoctorSeverityInfo,
			SlideNum: slideNum,
			Message:  fmt.Sprintf("Slide %d has lines exceeding 100 characters (longest: %d chars)", slideNum, maxLineLen),
			Advice:   "Wrap long sentences or bullets so text reads comfortably on compact screens.",
		})
	}
}

// FormatDoctorCLI renders the diagnostic report using beautiful terminal formatting.
func FormatDoctorCLI(report DoctorReport, theme Theme) string {
	accent := lipgloss.Color(theme.Accent)
	errColor := lipgloss.Color("#f7768e")
	warnColor := lipgloss.Color("#e0af68")
	infoColor := lipgloss.Color("#7dcfff")
	passColor := lipgloss.Color("#9ece6a")
	muted := lipgloss.Color(theme.Muted)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(accent)
	boxStyle := lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(0, 1)

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(titleStyle.Render(fmt.Sprintf(" Termdeck Doctor: %s", report.DeckPath)))
	sb.WriteString("\n")

	summaryText := fmt.Sprintf("Slides: %d  ·  Checks: %d  ·  Errors: %d  ·  Warnings: %d",
		report.TotalSlides, report.TotalChecks, report.Errors, report.Warnings)
	sb.WriteString(lipgloss.NewStyle().Foreground(muted).Render(" " + summaryText))
	sb.WriteString("\n\n")

	if len(report.Results) == 0 || (report.Errors == 0 && report.Warnings == 0 && report.Infos == 0) {
		successMsg := lipgloss.NewStyle().Foreground(passColor).Bold(true).Render("  ✓ All systems verified! Presentation DAG, media assets, and runtimes are clean.")
		sb.WriteString(boxStyle.Render(successMsg))
		sb.WriteString("\n\n")
		return sb.String()
	}

	for _, res := range report.Results {
		var badge string
		switch res.Severity {
		case DoctorSeverityError:
			badge = lipgloss.NewStyle().Bold(true).Foreground(errColor).Render(" [ERROR] ")
		case DoctorSeverityWarning:
			badge = lipgloss.NewStyle().Bold(true).Foreground(warnColor).Render(" [WARN]  ")
		case DoctorSeverityInfo:
			badge = lipgloss.NewStyle().Foreground(infoColor).Render(" [INFO]  ")
		case DoctorSeverityPass:
			badge = lipgloss.NewStyle().Bold(true).Foreground(passColor).Render(" [PASS]  ")
		}

		categoryTag := lipgloss.NewStyle().Foreground(accent).Render(fmt.Sprintf("[%-18s]", res.Category))
		sb.WriteString(fmt.Sprintf("%s %s %s\n", badge, categoryTag, res.Message))
		if res.Advice != "" {
			adviceText := lipgloss.NewStyle().Foreground(muted).Render(fmt.Sprintf("           ↳ Tip: %s", res.Advice))
			sb.WriteString(adviceText + "\n")
		}
	}

	sb.WriteString("\n")
	if report.Errors > 0 {
		verdict := lipgloss.NewStyle().Foreground(errColor).Bold(true).Render(fmt.Sprintf("  ✕ Found %d critical error(s) that should be resolved before presenting.", report.Errors))
		sb.WriteString(boxStyle.Render(verdict))
	} else if report.Warnings > 0 {
		verdict := lipgloss.NewStyle().Foreground(warnColor).Bold(true).Render(fmt.Sprintf("  ⚠ Found %d warning(s). Presentation will run, but consider the suggestions above.", report.Warnings))
		sb.WriteString(boxStyle.Render(verdict))
	} else {
		verdict := lipgloss.NewStyle().Foreground(passColor).Bold(true).Render("  ✓ Deck verified successfully!")
		sb.WriteString(boxStyle.Render(verdict))
	}
	sb.WriteString("\n\n")

	return sb.String()
}
