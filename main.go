package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"deck/internal"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.7.0"

type model struct {
	deck        internal.Deck
	editor      internal.Editor
	width       int
	height      int
	resized     bool
	lastModTime time.Time
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	if m.editor.WatchMode {
		cmds = append(cmds, internal.WatchCmd())
	}
	if m.editor.ShowTimer || m.editor.Autoplay {
		cmds = append(cmds, internal.TickCmd())
	}
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.resized {
			m.resized = true
			return m, tea.ClearScreen
		}

	case tea.KeyMsg:
		cmd := m.editor.HandleKey(msg, &m.deck)
		if cmd != nil {
			return m, cmd
		}

	case internal.ExecFinishedMsg:
		m.editor.RunningCode = false
		m.editor.ShowRunner = true
		m.editor.RunnerResult = &msg.Result
		m.editor.Message = fmt.Sprintf("run completed (exit %d · %v)", msg.Result.ExitCode, msg.Result.Duration.Round(time.Millisecond))
		return m, nil

	case internal.TickMsg:
		var cmd tea.Cmd
		if m.editor.ShowTimer || m.editor.Autoplay {
			if m.editor.Autoplay {
				m.editor.TickAutoplay(&m.deck)
			}
			cmd = internal.TickCmd()
		}
		return m, cmd

	case internal.WatchMsg:
		if m.editor.WatchMode && m.editor.FilePath != "" {
			if fi, err := os.Stat(m.editor.FilePath); err == nil {
				modTime := fi.ModTime()
				if !m.lastModTime.IsZero() && modTime.After(m.lastModTime) {
					m.lastModTime = modTime
					if m.editor.Mode != internal.ModeEdit && !m.editor.Dirty {
						_ = m.editor.Reload(&m.deck)
					}
				} else if m.lastModTime.IsZero() {
					m.lastModTime = modTime
				}
			}
			return m, internal.WatchCmd()
		}
	}

	return m, nil
}

func (m model) View() string {
	return internal.View(m.deck, m.editor, m.width, m.height)
}

func buildModel(filePath string, watchMode ...bool) (model, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return model{}, err
	}

	deck := internal.ParseDeck(string(src))
	deck.BaseDir = filepath.Dir(filePath)
	if len(deck.Slides) == 0 {
		return model{}, fmt.Errorf("no slides found")
	}

	editor := internal.NewEditor(filePath)
	editor.Theme = deck.Theme
	isWatch := len(watchMode) > 0 && watchMode[0]
	editor.WatchMode = isWatch
	var modTime time.Time
	if fi, err := os.Stat(filePath); err == nil {
		modTime = fi.ModTime()
	}
	return model{
		deck:        deck,
		editor:      editor,
		lastModTime: modTime,
	}, nil
}

func scaffoldStarterDeck(targetPath string) error {
	if targetPath == "" {
		targetPath = "presentation.deck.md"
	}
	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("file %q already exists; refusing to overwrite", targetPath)
	}

	content := `---
title: Welcome to Termdeck
theme: tokyo-night
routes:
  quick: intro -> interactive -> conclusion
  full: intro -> interactive -> branching -> code -> conclusion
---

# Welcome to Termdeck {#intro}
::tags: welcome, overview

### Terminal-native Presentation Engine & DAG Explorer
Elevate your tech talks directly within your terminal.

- Fast, distraction-free markdown presentations
- Non-linear branching decision graphs & DAG topology
- Dynamic shortest-path waypoint pathfinding
- Built-in sandboxed live code runner

> Press Space or Enter to advance · Press : or Ctrl+P for Command Palette

---

# Presentation Power Tools {#interactive}
::tags: guide, shortcuts

Everything you need to navigate and present effortlessly:

- Command Palette (: / Ctrl+P): Search and execute any presentation command
- Help Modal (?): Full interactive cheat sheet & shortcut reference
- DAG Map (M): Bird's-eye topology view of all interconnected slides
- Overview Grid (o): 2D thumbnail card sorter of the entire deck
- Audience Tracks (K): Filter presentation paths for different audiences
- Live Code Execution (X): Run shell, python, and go snippets on the fly

---

# Non-Linear Branching {#branching}
::tags: branching, graph

Termdeck presentations aren't just sequential slides — they form a directed acyclic graph (DAG).
Audiences can choose their own adventure at decision forks!

::branch [1] Deep Dive into Code -> code
::branch [2] Jump straight to Conclusion -> conclusion

---

# Live Code Execution {#code}
::tags: code, interactive

Run live demonstrations right inside the terminal presentation without switching windows.

` + "```" + `bash
echo "Hello from Termdeck!"
uname -s -m
date
` + "```" + `

> Press X or Ctrl+X to execute this block directly in the terminal!

---

# Summary & Next Steps {#conclusion}
::tags: summary

You are now ready to build stunning terminal presentations.

- Edit this file directly in markdown or press i in Termdeck
- Export standalone HTML slides using E or deck --export-html
- Check your presentation graph topology with deck --lint

Happy Presenting!
`
	return os.WriteFile(targetPath, []byte(content), 0644)
}

func printHelp(w ...io.Writer) {
	out := io.Writer(os.Stdout)
	if len(w) > 0 && w[0] != nil {
		out = w[0]
	}
	fmt.Fprintln(out, `Termdeck - Terminal presentation tool

Usage:
  deck [options] <file.deck.md>
  deck init [filename.deck.md]
  deck fmt [--check] <file.deck.md>
  deck doctor <file.deck.md>

Commands:
  init [name]          Scaffold a new starter presentation template (default: presentation.deck.md)
  fmt [--check] <file> Canonicalize markdown formatting, directives, and slide whitespace
  doctor <file>        Run comprehensive presentation integrity, asset, and runtime diagnostics

Options:
  -s, --start-at <N>   Start presentation at slide N (1-based)
  -k, --track <name>   Filter slides and DAG navigation to audience track
      --route <name>   Follow pre-planned graph presentation route
  -t, --theme <name>   Set presentation color theme
      --list-themes    List all available color themes
  -w, --watch          Watch deck file for external changes and auto-reload
  -a, --autoplay [sec] Auto-advance slides every N seconds (default: 5)
      --graph          Print presentation topology map (ASCII DAG) to terminal
      --mermaid        Print presentation topology as Mermaid diagram syntax
      --dot, --graphviz Export presentation topology to Graphviz DOT digraph
      --stats          Print presentation statistics and deck metrics to terminal
      --radar, --coverage Print presentation graph exploration radar and branch coverage
      --export-html    Export presentation to standalone HTML file
      --config <path>  Load configuration from config file (INI-style key: value)
      --lint           Validate DAG topology for broken links, unreachable slides, and dead ends
      --test-code      Execute and verify all code snippets in presentation
      --run-slide <N>  Execute code block on slide N and print output
  -v, --version        Show version information
  -h, --help           Show this help message

Controls:
  Navigation:   → / l / Space / Enter (next / advance edge), ← / h (prev)
  Branching:    1-9 (follow branch option), J (fork HUD & preview), Backspace (pop step), U (return to fork)
  Audience:     K (audience tracks & subgraph filter), [ / ] (hop along track)
  Routes:       P (preset graph routes & guided paths), W (waypoint pathfinder), V (radar & coverage)
  Palette:      : / ctrl+p (fuzzy command palette & action launcher)
  History:      H (traversal history & visual reflog modal)
  Graph Map:    M (presentation graph map & DAG explorer)
  Pointer:      ↓ / j (down), ↑ / k (up)
  Jumps:        / (jump to slide by number/search), g (first), G (last)
  Overview:     o / O (slide overview & 2D grid sorter)
  Focus Mode:   f / F (zoom focused block to fill terminal viewport)
  Live Runner:  X / ctrl+x (run code block), x (run code / toggle task)
  Theme:        t / T / f2 (cycle color themes: tokyo-night, dracula, nord, ...)
  Zen Mode:     z (toggle distraction-free zen mode)
  Line numbers: L (toggle code block line numbers)
  Timer:        c (toggle presentation timer), C (reset timer)
  Auto-play:    A (toggle auto-advance / rehearsal pacing)
  Reload:       r / R (reload deck from disk)
  Yank:         y / Y (yank focused code/block/runner to clipboard)
  Blank Screen: b / B (blank presentation screen, any key resumes)
  Export HTML:  E (export deck to standalone HTML presentation)
  Stats:        S (presentation statistics & deck metrics)
  Notes:        n (toggle speaker notes overlay)
  Alignment:    Tab / ctrl+a (cycle left/center/right alignment)
  Media:        p (open focused image card in desktop viewer)
  Help:         ? / f1 (in-app help modal)
  Editor:       i (edit block), Enter (save), Esc (cancel)
  Quit:         q / ctrl+c (auto-saves any changes)`)
}

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

func runCLI(args []string, stdout, stderr io.Writer) int {
	var startAt int
	var showHelp bool
	var showVer bool
	var cliTheme string
	var listThemes bool
	var watchMode bool
	var exportHTML bool
	var exportOutPath string
	var showStats bool
	var autoplayMode bool
	var autoplaySec int
	var showGraph bool
	var showMermaid bool
	var showDOT bool
	var testCode bool
	var runSlideNum int
	var cliTrack string
	var cliRoute string
	var lintDAG bool
	var showRadar bool
	var cfgPath string

	// Pre-scan for --config to load settings before flag overrides
	for i := 0; i < len(args); i++ {
		if args[i] == "--config" && i+1 < len(args) {
			cfgPath = args[i+1]
			break
		} else if strings.HasPrefix(args[i], "--config=") {
			cfgPath = strings.TrimPrefix(args[i], "--config=")
			break
		}
	}

	var cfg internal.Config
	var err error
	if cfgPath != "" {
		cfg, err = internal.LoadConfigFromPath(cfgPath)
		if err != nil {
			fmt.Fprintf(stderr, "warning: could not load config from %s: %v\n", cfgPath, err)
		}
	} else {
		cfg, err = internal.LoadConfig()
		if err != nil {
			fmt.Fprintf(stderr, "warning: could not load config: %v\n", err)
		}
	}

	// Apply config as defaults (CLI flags override config)
	if cfg.Theme != "" {
		cliTheme = cfg.Theme
	}
	if cfg.StartAt > 0 {
		startAt = cfg.StartAt
	}
	if cfg.Track != "" {
		cliTrack = cfg.Track
	}
	if cfg.Route != "" {
		cliRoute = cfg.Route
	}
	if cfg.Watch {
		watchMode = true
	}
	if cfg.Autoplay {
		autoplayMode = true
		if cfg.AutoplaySec > 0 {
			autoplaySec = cfg.AutoplaySec
		}
	}
	if cfg.ShowGraph {
		showGraph = true
	}
	if cfg.ShowMermaid {
		showMermaid = true
	}
	if cfg.ShowDOT {
		showDOT = true
	}
	if cfg.TestCode {
		testCode = true
	}
	if cfg.RunSlide > 0 {
		runSlideNum = cfg.RunSlide
	}
	if cfg.ShowStats {
		showStats = true
	}
	if cfg.ShowRadar {
		showRadar = true
	}
	if cfg.ExportHTML {
		exportHTML = true
	}
	if cfg.ExportOutPath != "" {
		exportOutPath = cfg.ExportOutPath
	}

	var fileArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			showHelp = true
		case arg == "--config":
			if i+1 < len(args) {
				i++ // already handled in pre-scan
			}
		case strings.HasPrefix(arg, "--config="):
			// already handled in pre-scan
		case arg == "-v" || arg == "--version":
			showVer = true
		case arg == "--list-themes":
			listThemes = true
		case arg == "-w" || arg == "--watch":
			watchMode = true
		case arg == "-k" || arg == "--track":
			if i+1 < len(args) {
				i++
				cliTrack = args[i]
			}
		case strings.HasPrefix(arg, "--track="):
			cliTrack = strings.TrimPrefix(arg, "--track=")
		case arg == "--route":
			if i+1 < len(args) {
				i++
				cliRoute = args[i]
			}
		case strings.HasPrefix(arg, "--route="):
			cliRoute = strings.TrimPrefix(arg, "--route=")
		case arg == "--lint" || arg == "--lint-graph":
			lintDAG = true
		case arg == "-a" || arg == "--autoplay":
			autoplayMode = true
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && !strings.HasSuffix(args[i+1], ".deck.md") && !strings.HasSuffix(args[i+1], ".md") {
				var sec int
				if n, _ := fmt.Sscanf(args[i+1], "%d", &sec); n == 1 && sec > 0 {
					i++
					autoplaySec = sec
				}
			}
		case strings.HasPrefix(arg, "--autoplay="):
			autoplayMode = true
			fmt.Sscanf(strings.TrimPrefix(arg, "--autoplay="), "%d", &autoplaySec)
		case arg == "--graph":
			showGraph = true
		case arg == "--mermaid":
			showMermaid = true
		case arg == "--dot" || arg == "--graphviz":
			showDOT = true
		case arg == "--test-code":
			testCode = true
		case arg == "--run-slide":
			if i+1 < len(args) {
				i++
				fmt.Sscanf(args[i], "%d", &runSlideNum)
			}
		case strings.HasPrefix(arg, "--run-slide="):
			fmt.Sscanf(strings.TrimPrefix(arg, "--run-slide="), "%d", &runSlideNum)
		case arg == "--radar" || arg == "--coverage":
			showRadar = true
		case arg == "--stats":
			showStats = true
		case arg == "--export-html":
			exportHTML = true
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && !strings.HasSuffix(args[i+1], ".deck.md") && !strings.HasSuffix(args[i+1], ".md") {
				i++
				exportOutPath = args[i]
			}
		case strings.HasPrefix(arg, "--export-html="):
			exportHTML = true
			exportOutPath = strings.TrimPrefix(arg, "--export-html=")
		case arg == "-t" || arg == "--theme":
			if i+1 < len(args) {
				i++
				cliTheme = args[i]
			}
		case strings.HasPrefix(arg, "--theme="):
			cliTheme = strings.TrimPrefix(arg, "--theme=")
		case arg == "-s" || arg == "--start-at":
			if i+1 < len(args) {
				i++
				fmt.Sscanf(args[i], "%d", &startAt)
			}
		case strings.HasPrefix(arg, "--start-at="):
			fmt.Sscanf(strings.TrimPrefix(arg, "--start-at="), "%d", &startAt)
		default:
			fileArgs = append(fileArgs, arg)
		}
	}

	if showHelp {
		printHelp(stdout)
		return 0
	}
	if showVer {
		fmt.Fprintf(stdout, "Termdeck v%s\n", version)
		return 0
	}
	if listThemes {
		fmt.Fprintln(stdout, "Available Termdeck Color Themes:")
		for _, th := range internal.AvailableThemes() {
			fmt.Fprintf(stdout, "  %-14s %-18s (accent: %s)\n", th.ID, th.Name, th.Accent)
		}
		fmt.Fprintln(stdout, "\nTip: Pass '--theme <name>' or set 'theme: <name>' in deck frontmatter.")
		return 0
	}

	if len(fileArgs) > 0 && (fileArgs[0] == "init" || fileArgs[0] == "new") {
		target := "presentation.deck.md"
		if len(fileArgs) > 1 {
			target = fileArgs[1]
		}
		if err := scaffoldStarterDeck(target); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Created starter presentation at %s\n", target)
		fmt.Fprintf(stdout, "Run 'deck %s' to launch your presentation.\n", target)
		return 0
	}

	if len(fileArgs) > 0 && (fileArgs[0] == "fmt" || fileArgs[0] == "format") {
		isCheck := false
		var targetFiles []string
		for _, a := range fileArgs[1:] {
			if a == "--check" {
				isCheck = true
			} else if !strings.HasPrefix(a, "-") {
				targetFiles = append(targetFiles, a)
			}
		}
		if len(targetFiles) == 0 {
			fmt.Fprintln(stderr, "error: missing deck file for fmt")
			fmt.Fprintln(stderr, "usage: deck fmt [--check] <file.deck.md>")
			return 1
		}
		hasUnformatted := false
		for _, f := range targetFiles {
			isFmt, _, err := internal.FormatDeckFile(f, isCheck)
			if err != nil {
				fmt.Fprintf(stderr, "error formatting %s: %v\n", f, err)
				return 1
			}
			if isCheck {
				if !isFmt {
					fmt.Fprintf(stdout, "✕ %s needs formatting\n", f)
					hasUnformatted = true
				} else {
					fmt.Fprintf(stdout, "✓ %s is formatted\n", f)
				}
			} else {
				if !isFmt {
					fmt.Fprintf(stdout, "Formatted %s\n", f)
				} else {
					fmt.Fprintf(stdout, "%s already formatted\n", f)
				}
			}
		}
		if hasUnformatted {
			return 1
		}
		return 0
	}

	if len(fileArgs) > 0 && (fileArgs[0] == "doctor" || fileArgs[0] == "check") {
		if len(fileArgs) < 2 {
			fmt.Fprintln(stderr, "error: missing deck file for doctor")
			fmt.Fprintln(stderr, "usage: deck doctor <file.deck.md>")
			return 1
		}
		deckFile := fileArgs[1]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		if cliTheme != "" {
			d.Theme = cliTheme
		}
		report := internal.RunDeckDoctor(d, deckFile)
		theme := internal.ResolveTheme(d.Theme)
		fmt.Fprint(stdout, internal.FormatDoctorCLI(report, theme))
		if report.Errors > 0 {
			return 1
		}
		return 0
	}

	if exportHTML {
		if len(fileArgs) < 1 {
			fmt.Fprintln(stderr, "error: missing deck file for --export-html")
			return 1
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		if cliTheme != "" {
			d.Theme = cliTheme
		}
		outPath := exportOutPath
		if outPath == "" {
			ext := filepath.Ext(deckFile)
			outPath = strings.TrimSuffix(deckFile, ext) + ".html"
		}
		if err := internal.ExportHTMLFile(d, outPath); err != nil {
			fmt.Fprintf(stderr, "export error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Exported presentation to %s\n", outPath)
		return 0
	}

	if showRadar {
		if len(fileArgs) < 1 {
			fmt.Fprintln(stderr, "error: missing deck file for --radar")
			return 1
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		visited := map[int]bool{0: true}
		fmt.Fprint(stdout, internal.FormatRadarCLI(d, visited, 0))
		return 0
	}

	if showStats {
		if len(fileArgs) < 1 {
			fmt.Fprintln(stderr, "error: missing deck file for --stats")
			return 1
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		if cliTheme != "" {
			d.Theme = cliTheme
		}
		theme := internal.ResolveTheme(d.Theme)
		stats := internal.CalculateStats(&d, 0)
		fmt.Fprint(stdout, internal.FormatStatsCLI(stats, theme))
		return 0
	}

	if showGraph {
		if len(fileArgs) < 1 {
			fmt.Fprintln(stderr, "error: missing deck file for --graph")
			return 1
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		if cliTheme != "" {
			d.Theme = cliTheme
		}
		theme := internal.ResolveTheme(d.Theme)
		if cliRoute != "" {
			fmt.Fprint(stdout, internal.FormatGraphCLIWithRoute(d, theme, cliRoute))
		} else if cliTrack != "" {
			fmt.Fprint(stdout, internal.FormatGraphCLIWithTrack(d, theme, cliTrack))
		} else {
			fmt.Fprint(stdout, internal.FormatGraphCLI(d, theme))
		}
		return 0
	}

	if showMermaid {
		if len(fileArgs) < 1 {
			fmt.Fprintln(stderr, "error: missing deck file for --mermaid")
			return 1
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		g := internal.BuildGraph(d)
		if cliRoute != "" {
			fmt.Fprint(stdout, g.ToMermaidWithRoute(cliRoute, d))
		} else if cliTrack != "" {
			fmt.Fprint(stdout, g.ToMermaidWithTrack(cliTrack))
		} else {
			fmt.Fprint(stdout, g.ToMermaid())
		}
		return 0
	}

	if showDOT {
		if len(fileArgs) < 1 {
			fmt.Fprintln(stderr, "error: missing deck file for --dot")
			return 1
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		g := internal.BuildGraph(d)
		if cliTrack != "" {
			fmt.Fprint(stdout, g.ToGraphvizDOTWithTrack(cliTrack))
		} else {
			fmt.Fprint(stdout, g.ToGraphvizDOT())
		}
		return 0
	}

	if lintDAG {
		if len(fileArgs) < 1 {
			fmt.Fprintln(stderr, "error: missing deck file for --lint")
			return 1
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		issues := internal.LintGraph(d)
		theme := internal.ResolveTheme(d.Theme)
		if cliTheme != "" {
			theme = internal.ResolveTheme(cliTheme)
		}
		out, errCount := internal.FormatLintCLI(issues, theme, deckFile)
		fmt.Fprint(stdout, out)
		if errCount > 0 {
			return 1
		}
		return 0
	}

	if testCode {
		if len(fileArgs) < 1 {
			fmt.Fprintln(stderr, "error: missing deck file for --test-code")
			return 1
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		start := time.Now()
		passed, failed, results := internal.TestAllDeckCode(d, 5*time.Second)
		fmt.Fprint(stdout, internal.FormatTestCodeCLI(deckFile, passed, failed, results, time.Since(start)))
		if failed > 0 {
			return 1
		}
		return 0
	}

	if runSlideNum > 0 {
		if len(fileArgs) < 1 {
			fmt.Fprintln(stderr, "error: missing deck file for --run-slide")
			return 1
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(stderr, "error reading %s: %v\n", deckFile, err)
			return 1
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		if runSlideNum > len(d.Slides) {
			fmt.Fprintf(stderr, "error: slide %d exceeds total slide count (%d)\n", runSlideNum, len(d.Slides))
			return 1
		}
		slide := d.Slides[runSlideNum-1]
		var targetBlock *internal.Block
		for _, b := range slide.Blocks {
			if b.Kind == internal.BlockCode {
				targetBlock = &b
				break
			}
		}
		if targetBlock == nil {
			fmt.Fprintf(stderr, "error: slide %d contains no code blocks\n", runSlideNum)
			return 1
		}
		res := internal.ExecuteBlock(*targetBlock, 10*time.Second, nil)
		if res.Stdout != "" {
			fmt.Fprint(stdout, res.Stdout)
			if !strings.HasSuffix(res.Stdout, "\n") {
				fmt.Fprintln(stdout)
			}
		}
		if res.Stderr != "" {
			fmt.Fprint(stderr, res.Stderr)
			if !strings.HasSuffix(res.Stderr, "\n") {
				fmt.Fprintln(stderr)
			}
		}
		if res.Error != "" && !strings.Contains(res.Stderr, res.Error) {
			fmt.Fprintf(stderr, "error: %s\n", res.Error)
		}
		if res.ExitCode != 0 {
			return res.ExitCode
		}
		return 0
	}

	if len(fileArgs) < 1 {
		fmt.Fprintln(stderr, "error: missing deck file")
		fmt.Fprintln(stderr, "usage: deck [options] <file.deck.md>")
		fmt.Fprintln(stderr, "       deck init [filename.deck.md]   # create starter presentation")
		fmt.Fprintln(stderr, "       deck fmt [--check] <file.deck.md>")
		fmt.Fprintln(stderr, "       deck doctor <file.deck.md>")
		fmt.Fprintln(stderr, "try 'deck --help' for more information")
		return 1
	}

	m, err := buildModel(fileArgs[0], watchMode)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if cliTheme != "" {
		m.deck.Theme = cliTheme
		m.editor.Theme = cliTheme
	}

	if cliTrack != "" {
		m.editor.SelectTrack(cliTrack, &m.deck)
	}

	if cliRoute != "" {
		m.editor.SelectRoute(cliRoute, &m.deck)
	}

	if startAt > 0 {
		idx := startAt - 1
		if idx >= len(m.deck.Slides) {
			idx = len(m.deck.Slides) - 1
		}
		m.editor.SlideIdx = idx
		m.editor.ClampBlockIdx(&m.deck)
	}

	if autoplayMode {
		m.editor.Autoplay = true
		if autoplaySec <= 0 {
			autoplaySec = 5
		}
		m.editor.AutoplayInterval = autoplaySec
		m.editor.AutoplayCountdown = autoplaySec
	}

	if os.Getenv("TERMDECK_NO_RUN") == "1" {
		return 0
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if fm, ok := finalModel.(model); ok && fm.editor.Dirty && fm.editor.FilePath != "" {
		fm.editor.Save(fm.deck)
	}
	return 0
}
