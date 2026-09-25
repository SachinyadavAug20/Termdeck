package main

import (
	"fmt"
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

func printHelp() {
	fmt.Println(`Termdeck - Terminal presentation tool

Usage:
  deck [options] <file.deck.md>

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
      --stats          Print presentation statistics and deck metrics to terminal
      --export-html    Export presentation to standalone HTML file
      --lint           Validate DAG topology for broken links, unreachable slides, and dead ends
      --test-code      Execute and verify all code snippets in presentation
      --run-slide <N>  Execute code block on slide N and print output
  -v, --version        Show version information
  -h, --help           Show this help message

Controls:
  Navigation:   → / l / Space / Enter (next / advance edge), ← / h (prev)
  Branching:    1-9 (follow branch option), Backspace (pop step)
  Audience:     K (audience tracks & subgraph filter), [ / ] (hop along track)
  Routes:       P (preset graph routes & guided paths)
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
	var testCode bool
	var runSlideNum int
	var cliTrack string
	var cliRoute string
	var lintDAG bool

	args := os.Args[1:]
	var fileArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			showHelp = true
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
		case arg == "--test-code":
			testCode = true
		case arg == "--run-slide":
			if i+1 < len(args) {
				i++
				fmt.Sscanf(args[i], "%d", &runSlideNum)
			}
		case strings.HasPrefix(arg, "--run-slide="):
			fmt.Sscanf(strings.TrimPrefix(arg, "--run-slide="), "%d", &runSlideNum)
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
		printHelp()
		return
	}
	if showVer {
		fmt.Printf("Termdeck v%s\n", version)
		return
	}
	if listThemes {
		fmt.Println("Available Termdeck Color Themes:")
		for _, th := range internal.AvailableThemes() {
			fmt.Printf("  %-14s %-18s (accent: %s)\n", th.ID, th.Name, th.Accent)
		}
		fmt.Println("\nTip: Pass '--theme <name>' or set 'theme: <name>' in deck frontmatter.")
		return
	}

	if exportHTML {
		if len(fileArgs) < 1 {
			fmt.Fprintln(os.Stderr, "error: missing deck file for --export-html")
			os.Exit(1)
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", deckFile, err)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "export error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported presentation to %s\n", outPath)
		return
	}

	if showStats {
		if len(fileArgs) < 1 {
			fmt.Fprintln(os.Stderr, "error: missing deck file for --stats")
			os.Exit(1)
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", deckFile, err)
			os.Exit(1)
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		if cliTheme != "" {
			d.Theme = cliTheme
		}
		theme := internal.ResolveTheme(d.Theme)
		stats := internal.CalculateStats(&d, 0)
		fmt.Print(internal.FormatStatsCLI(stats, theme))
		return
	}

	if showGraph {
		if len(fileArgs) < 1 {
			fmt.Fprintln(os.Stderr, "error: missing deck file for --graph")
			os.Exit(1)
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", deckFile, err)
			os.Exit(1)
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		if cliTheme != "" {
			d.Theme = cliTheme
		}
		theme := internal.ResolveTheme(d.Theme)
		if cliRoute != "" {
			fmt.Print(internal.FormatGraphCLIWithRoute(d, theme, cliRoute))
		} else if cliTrack != "" {
			fmt.Print(internal.FormatGraphCLIWithTrack(d, theme, cliTrack))
		} else {
			fmt.Print(internal.FormatGraphCLI(d, theme))
		}
		return
	}

	if showMermaid {
		if len(fileArgs) < 1 {
			fmt.Fprintln(os.Stderr, "error: missing deck file for --mermaid")
			os.Exit(1)
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", deckFile, err)
			os.Exit(1)
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		g := internal.BuildGraph(d)
		if cliRoute != "" {
			fmt.Print(g.ToMermaidWithRoute(cliRoute, d))
		} else if cliTrack != "" {
			fmt.Print(g.ToMermaidWithTrack(cliTrack))
		} else {
			fmt.Print(g.ToMermaid())
		}
		return
	}

	if lintDAG {
		if len(fileArgs) < 1 {
			fmt.Fprintln(os.Stderr, "error: missing deck file for --lint")
			os.Exit(1)
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", deckFile, err)
			os.Exit(1)
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		issues := internal.LintGraph(d)
		theme := internal.ResolveTheme(d.Theme)
		if cliTheme != "" {
			theme = internal.ResolveTheme(cliTheme)
		}
		out, errCount := internal.FormatLintCLI(issues, theme, deckFile)
		fmt.Print(out)
		if errCount > 0 {
			os.Exit(1)
		}
		return
	}

	if testCode {
		if len(fileArgs) < 1 {
			fmt.Fprintln(os.Stderr, "error: missing deck file for --test-code")
			os.Exit(1)
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", deckFile, err)
			os.Exit(1)
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		start := time.Now()
		passed, failed, results := internal.TestAllDeckCode(d, 5*time.Second)
		fmt.Print(internal.FormatTestCodeCLI(deckFile, passed, failed, results, time.Since(start)))
		if failed > 0 {
			os.Exit(1)
		}
		return
	}

	if runSlideNum > 0 {
		if len(fileArgs) < 1 {
			fmt.Fprintln(os.Stderr, "error: missing deck file for --run-slide")
			os.Exit(1)
		}
		deckFile := fileArgs[0]
		src, err := os.ReadFile(deckFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading %s: %v\n", deckFile, err)
			os.Exit(1)
		}
		d := internal.ParseDeck(string(src))
		d.BaseDir = filepath.Dir(deckFile)
		if runSlideNum > len(d.Slides) {
			fmt.Fprintf(os.Stderr, "error: slide %d exceeds total slide count (%d)\n", runSlideNum, len(d.Slides))
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "error: slide %d contains no code blocks\n", runSlideNum)
			os.Exit(1)
		}
		res := internal.ExecuteBlock(*targetBlock, 10*time.Second)
		if res.Stdout != "" {
			fmt.Print(res.Stdout)
			if !strings.HasSuffix(res.Stdout, "\n") {
				fmt.Println()
			}
		}
		if res.Stderr != "" {
			fmt.Fprint(os.Stderr, res.Stderr)
			if !strings.HasSuffix(res.Stderr, "\n") {
				fmt.Fprintln(os.Stderr)
			}
		}
		if res.Error != "" && !strings.Contains(res.Stderr, res.Error) {
			fmt.Fprintf(os.Stderr, "error: %s\n", res.Error)
		}
		if res.ExitCode != 0 {
			os.Exit(res.ExitCode)
		}
		return
	}

	if len(fileArgs) < 1 {
		fmt.Fprintln(os.Stderr, "error: missing deck file")
		fmt.Fprintln(os.Stderr, "usage: deck [options] <file.deck.md>")
		fmt.Fprintln(os.Stderr, "try 'deck --help' for more information")
		os.Exit(1)
	}

	m, err := buildModel(fileArgs[0], watchMode)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if fm, ok := finalModel.(model); ok && fm.editor.Dirty && fm.editor.FilePath != "" {
		fm.editor.Save(fm.deck)
	}
}
