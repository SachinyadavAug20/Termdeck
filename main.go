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

const version = "0.3.0"

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type model struct {
	deck    internal.Deck
	editor  internal.Editor
	width   int
	height  int
	resized bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.resized {
			m.resized = true
			return m, tea.ClearScreen
		}

	case tickMsg:
		if m.editor.TimerRunning {
			return m, tickCmd()
		}
		return m, nil

	case tea.KeyMsg:
		cmd := m.editor.HandleKey(msg, &m.deck)
		if m.editor.TimerRunning && msg.String() == "t" {
			return m, tea.Batch(cmd, tickCmd())
		}
		if cmd != nil {
			return m, cmd
		}
	}

	return m, nil
}

func (m model) View() string {
	return internal.View(m.deck, m.editor, m.width, m.height)
}

func buildModel(filePath string) (model, error) {
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
	return model{
		deck:   deck,
		editor: editor,
	}, nil
}

func printHelp() {
	fmt.Println(`Termdeck - Terminal presentation tool

Usage:
  deck [options] <file.deck.md>

Options:
  -s, --start-at <N>   Start presentation at slide N (1-based)
  -v, --version        Show version information
  -h, --help           Show this help message

Controls:
  Navigation:   → / l / Space / Enter (next), ← / h / Backspace (prev)
  Pointer:      ↓ / j (down), ↑ / k (up)
  Jumps:        g (first slide), G (last slide)
  Notes:        n (toggle speaker notes overlay)
  Timer:        t (start/pause elapsed timer), ctrl+t (reset)
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

	args := os.Args[1:]
	var fileArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help":
			showHelp = true
		case arg == "-v" || arg == "--version":
			showVer = true
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

	if len(fileArgs) < 1 {
		fmt.Fprintln(os.Stderr, "error: missing deck file")
		fmt.Fprintln(os.Stderr, "usage: deck [options] <file.deck.md>")
		fmt.Fprintln(os.Stderr, "try 'deck --help' for more information")
		os.Exit(1)
	}

	m, err := buildModel(fileArgs[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
