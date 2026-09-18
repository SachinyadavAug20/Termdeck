package main

import (
	"fmt"
	"os"
	"path/filepath"

	"deck/internal"
	tea "github.com/charmbracelet/bubbletea"
)

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

	case tea.KeyMsg:
		cmd := m.editor.HandleKey(msg, &m.deck)
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

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: deck <file.deck.md>")
		os.Exit(1)
	}

	m, err := buildModel(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
