package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type slide []string

type model struct {
	slides []slide
	idx    int
	width  int
	height int
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func parseDeck(src string) []slide {
	lines := strings.Split(src, "\n")
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				lines = lines[i+1:]
				break
			}
		}
	}
	var slides []slide
	var cur slide
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" && len(cur) > 0 {
			slides = append(slides, cur)
			cur = nil
			continue
		}
		cur = append(cur, line)
	}
	if len(cur) > 0 {
		slides = append(slides, cur)
	}
	return slides
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "right", "l", " ", "enter", "down", "j", "pgdown":
			if m.idx < len(m.slides)-1 {
				m.idx++
			}
		case "left", "h", "up", "k", "pgup", "backspace":
			if m.idx > 0 {
				m.idx--
			}
		case "g", "home":
			m.idx = 0
		case "G", "end":
			m.idx = len(m.slides) - 1
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m model) View() string {
	w, h := m.width, m.height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}
	var b strings.Builder
	for _, line := range m.slides[m.idx] {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "::") {
			continue
		}
		if strings.HasPrefix(line, "# ") {
			b.WriteString(titleStyle.Render(strings.TrimPrefix(line, "# ")) + "\n\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	body := lipgloss.NewStyle().
		Width(w).
		Height(h - 2).
		Align(lipgloss.Center, lipgloss.Center).
		Render(strings.TrimRight(b.String(), "\n"))
	footer := dimStyle.Render(fmt.Sprintf("slide %d/%d  ·  ←/→ navigate · q quit", m.idx+1, len(m.slides)))
	return body + "\n" + footer
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: deck <file.deck.md>")
		os.Exit(1)
	}
	src, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	slides := parseDeck(string(src))
	if len(slides) == 0 {
		fmt.Fprintln(os.Stderr, "no slides found")
		os.Exit(1)
	}
	p := tea.NewProgram(model{slides: slides}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}