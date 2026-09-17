package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type slide []string

type model struct {
	slides  []slide
	idx     int
	width   int
	height  int
	resized bool
}

var (
	boldStyle      = lipgloss.NewStyle().Bold(true)
	italicStyle    = lipgloss.NewStyle().Italic(true)
	codeSpanStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Background(lipgloss.Color("236")).Padding(0, 1)
	codeBlockStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	imageStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dimBoldStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Bold(true)

	// headings
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	h2Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	h3Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	h4Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("248"))
	h5Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("246"))
	h6Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))

	reCodeSpan = regexp.MustCompile("`" + `([^` + "`" + `]+)` + "`")
	reBold     = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reItalic   = regexp.MustCompile(`\*([^*]+)\*`)
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

func parseDirective(line string) (key, value string) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "::") {
		return "", ""
	}
	content := trimmed[2:]
	if idx := strings.IndexAny(content, " ="); idx != -1 {
		key = content[:idx]
		value = strings.TrimSpace(content[idx+1:])
		// handle key=value
		if eq := strings.Index(value, "="); eq != -1 {
			value = value[eq+1:]
		}
	} else {
		key = content
	}
	return key, value
}

// inlineStyle applies **bold**, *italic*, and `code` formatting
func inlineStyle(text string) string {
	// process in order: code spans first (atomic), then bold, then italic
	type span struct {
		start, end int
		styled     string
	}

	// code spans
	var spans []span
	for _, m := range reCodeSpan.FindAllStringIndex(text, -1) {
		inner := text[m[0]+1 : m[1]-1]
		spans = append(spans, span{m[0], m[1], codeSpanStyle.Render(inner)})
	}
	// replace code spans with placeholders
	result := text
	for i, s := range spans {
		ph := fmt.Sprintf("\x00CODE%d\x00", i)
		result = result[:s.start] + ph + result[s.end:]
	}

	// bold (**text**)
	result = reBold.ReplaceAllStringFunc(result, func(m string) string {
		inner := m[2 : len(m)-2]
		return boldStyle.Render(inner)
	})

	// italic (*text*) — but not inside bold markers
	result = reItalic.ReplaceAllStringFunc(result, func(m string) string {
		inner := m[1 : len(m)-1]
		return italicStyle.Render(inner)
	})

	// restore code spans
	for i, s := range spans {
		ph := fmt.Sprintf("\x00CODE%d\x00", i)
		result = strings.Replace(result, ph, s.styled, 1)
	}

	return result
}

func renderHeading(line string) string {
	trimmed := strings.TrimSpace(line)
	count := 0
	for _, ch := range trimmed {
		if ch == '#' {
			count++
		} else {
			break
		}
	}
	text := strings.TrimSpace(trimmed[count:])
	switch count {
	case 1:
		return titleStyle.Render(text)
	case 2:
		return h2Style.Render(text)
	case 3:
		return h3Style.Render(text)
	case 4:
		return h4Style.Render(text)
	case 5:
		return h5Style.Render(text)
	default:
		return h6Style.Render(text)
	}
}

func renderSlide(s slide, w int) string {
	var b strings.Builder
	inCodeBlock := false
	codeLang := ""
	var codeLines []string

	flushCode := func() {
		var content strings.Builder
		for i, line := range codeLines {
			if i > 0 {
				content.WriteString("\n")
			}
			content.WriteString(line)
		}
		header := ""
		if codeLang != "" {
			header = dimBoldStyle.Render(" "+codeLang) + "\n"
		}
		rendered := codeBlockStyle.Width(w - 4).Render(content.String())
		b.WriteString(header + rendered + "\n")
		codeLines = nil
		codeLang = ""
	}

	for _, line := range s {
		trimmed := strings.TrimSpace(line)

		// ::code block handling
		if strings.HasPrefix(trimmed, "::code") {
			if inCodeBlock {
				flushCode()
				inCodeBlock = false
				continue
			}
			inCodeBlock = true
			_, val := parseDirective(trimmed)
			codeLang = val
			continue
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}

		// other directives — skip
		if strings.HasPrefix(trimmed, "::") {
			key, val := parseDirective(trimmed)
			if key == "image" && val != "" {
				b.WriteString(imageStyle.Render(fmt.Sprintf("[ image: %s ]", val)) + "\n")
			}
			continue
		}

		// empty line
		if trimmed == "" {
			b.WriteString("\n")
			continue
		}

		// headings
		if strings.HasPrefix(trimmed, "#") {
			b.WriteString(renderHeading(line) + "\n")
			continue
		}

		// regular text with inline styling
		b.WriteString(inlineStyle(line) + "\n")
	}

	if inCodeBlock {
		flushCode()
	}

	return strings.TrimRight(b.String(), "\n")
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
		if !m.resized {
			m.resized = true
			return m, tea.ClearScreen
		}
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
	bodyHeight := h - 2
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	rendered := renderSlide(m.slides[m.idx], w)
	body := lipgloss.NewStyle().
		Width(w).
		Height(bodyHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Render(rendered)
	footerLine := dimStyle.Width(w).Render(fmt.Sprintf("slide %d/%d  ·  ←/→ navigate · q quit", m.idx+1, len(m.slides)))
	return body + "\n" + footerLine
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
