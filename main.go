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

const version = "0.4.0"

type model struct {
	deck        internal.Deck
	editor      internal.Editor
	width       int
	height      int
	resized     bool
	lastModTime time.Time
}

func (m model) Init() tea.Cmd {
	if m.editor.WatchMode {
		return internal.WatchCmd()
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

	case internal.TickMsg:
		if m.editor.ShowTimer {
			return m, internal.TickCmd()
		}

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
  -t, --theme <name>   Set presentation color theme
      --list-themes    List all available color themes
  -w, --watch          Watch deck file for external changes and auto-reload
  -v, --version        Show version information
  -h, --help           Show this help message

Controls:
  Navigation:   → / l / Space / Enter (next), ← / h / Backspace (prev)
  Pointer:      ↓ / j (down), ↑ / k (up)
  Jumps:        / (jump to slide by number/search), g (first), G (last)
  Overview:     o / O (slide overview & 2D grid sorter)
  Theme:        t / T / f2 (cycle color themes: tokyo-night, dracula, nord, ...)
  Zen Mode:     z (toggle distraction-free zen mode)
  Line numbers: L (toggle code block line numbers)
  Timer:        c (toggle presentation timer), C (reset timer)
  Reload:       r / R (reload deck from disk)
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

	if startAt > 0 {
		idx := startAt - 1
		if idx >= len(m.deck.Slides) {
			idx = len(m.deck.Slides) - 1
		}
		m.editor.SlideIdx = idx
		m.editor.ClampBlockIdx(&m.deck)
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
