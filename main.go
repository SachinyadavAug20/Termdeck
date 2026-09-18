package main

import (
	"fmt"
	"os"

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

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: deck <file.deck.md>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	src, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	deck := internal.ParseDeck(string(src))
	if len(deck.Slides) == 0 {
		fmt.Fprintln(os.Stderr, "no slides found")
		os.Exit(1)
	}

	editor := internal.NewEditor(filePath)

	p := tea.NewProgram(model{
		deck:   deck,
		editor: editor,
	}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
