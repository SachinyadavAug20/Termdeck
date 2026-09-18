package internal

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---

var (
	boldStyle      = lipgloss.NewStyle().Bold(true)
	italicStyle    = lipgloss.NewStyle().Italic(true)
	codeSpanStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Background(lipgloss.Color("236")).Padding(0, 1)
	codeBlockStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	imageStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dimBoldStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Bold(true)
	cursorStyle    = lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("0"))
	editStyle      = lipgloss.NewStyle().Background(lipgloss.Color("22")).Foreground(lipgloss.Color("252")).Padding(0, 1)
	messageStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	h2Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	h3Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	h4Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("248"))
	h5Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("246"))
	h6Style    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("244"))

	syntaxKeyword = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	syntaxString  = lipgloss.NewStyle().Foreground(lipgloss.Color("114"))
	syntaxComment = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	syntaxNumber  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	syntaxType    = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))
	syntaxPlain   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	reCodeSpan = regexp.MustCompile("`" + `([^` + "`" + `]+)` + "`")
	reBold     = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reItalic   = regexp.MustCompile(`\*([^*]+)\*`)
)

// --- Inline styling ---

func inlineStyle(text string) string {
	type span struct {
		start, end int
		styled     string
	}

	var spans []span
	for _, m := range reCodeSpan.FindAllStringIndex(text, -1) {
		inner := text[m[0]+1 : m[1]-1]
		spans = append(spans, span{m[0], m[1], codeSpanStyle.Render(inner)})
	}
	result := text
	for i, s := range spans {
		ph := fmt.Sprintf("\x00CODE%d\x00", i)
		result = result[:s.start] + ph + result[s.end:]
	}

	result = reBold.ReplaceAllStringFunc(result, func(m string) string {
		return boldStyle.Render(m[2 : len(m)-2])
	})

	result = reItalic.ReplaceAllStringFunc(result, func(m string) string {
		return italicStyle.Render(m[1 : len(m)-1])
	})

	for i, s := range spans {
		ph := fmt.Sprintf("\x00CODE%d\x00", i)
		result = strings.Replace(result, ph, s.styled, 1)
	}

	return result
}

// --- Heading rendering ---

var (
	h1BorderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			Underline(true)
)

func renderHeading(text string, level int, w int) string {
	switch level {
	case 1:
		// h1: underline with spacing
		return "\n" + h1BorderStyle.Render(text) + "\n"
	case 2:
		// h2: bold with extra spacing
		return "\n" + h2Style.Render(text) + "\n"
	case 3:
		// h3: bold with single spacing
		return "\n" + h3Style.Render(text)
	case 4:
		// h4: dimmer
		return h4Style.Render(text)
	case 5:
		// h5: dimmer
		return h5Style.Render(text)
	default:
		// h6: dimmest
		return h6Style.Render(text)
	}
}

// --- Syntax highlighting ---

var simpleKeywords = map[string]bool{
	"func": true, "return": true, "if": true, "else": true, "for": true,
	"range": true, "var": true, "const": true, "type": true, "struct": true,
	"package": true, "import": true, "defer": true, "go": true, "chan": true,
	"select": true, "case": true, "switch": true, "default": true, "break": true,
	"continue": true, "true": true, "false": true, "nil": true,
	"def": true, "class": true, "from": true, "as": true,
	"print": true, "self": true, "None": true,
	"with": true, "try": true, "except": true, "finally": true, "raise": true,
	"yield": true, "lambda": true, "pass": true, "del": true, "global": true,
	"echo": true, "then": true, "fi": true,
	"elif": true, "while": true, "do": true, "done": true,
	"esac": true, "function": true, "exit": true, "local": true, "export": true,
	"null": true, "undefined": true,
}

func highlightLine(line string) string {
	var result strings.Builder
	i := 0
	for i < len(line) {
		ch := line[i]

		if ch == '#' || (ch == '/' && i+1 < len(line) && line[i+1] == '/') {
			result.WriteString(syntaxComment.Render(line[i:]))
			return result.String()
		}
		if ch == '-' && i+1 < len(line) && line[i+1] == '-' {
			result.WriteString(syntaxComment.Render(line[i:]))
			return result.String()
		}

		if ch == '"' || ch == '\'' || ch == '`' {
			quote := ch
			j := i + 1
			for j < len(line) && line[j] != quote {
				if line[j] == '\\' {
					j++
				}
				j++
			}
			if j < len(line) {
				j++
			}
			result.WriteString(syntaxString.Render(line[i:j]))
			i = j
			continue
		}

		if ch >= '0' && ch <= '9' {
			j := i
			for j < len(line) && ((line[j] >= '0' && line[j] <= '9') || line[j] == '.') {
				j++
			}
			result.WriteString(syntaxNumber.Render(line[i:j]))
			i = j
			continue
		}

		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
			j := i
			for j < len(line) && ((line[j] >= 'a' && line[j] <= 'z') || (line[j] >= 'A' && line[j] <= 'Z') || (line[j] >= '0' && line[j] <= '9') || line[j] == '_') {
				j++
			}
			word := line[i:j]
			if simpleKeywords[word] {
				result.WriteString(syntaxKeyword.Render(word))
			} else if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
				result.WriteString(syntaxType.Render(word))
			} else {
				result.WriteString(syntaxPlain.Render(word))
			}
			i = j
			continue
		}

		result.WriteString(syntaxPlain.Render(string(ch)))
		i++
	}

	return result.String()
}

func highlightCode(lines []string, lang string) string {
	var result strings.Builder
	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(highlightLine(line))
	}
	return result.String()
}

// --- Block rendering ---

func renderBlock(blk Block, w int, isCursor bool, isEditing bool, editDraft string, cursorCol int) string {
	var b strings.Builder

	cursorMark := "  "
	if isCursor && !isEditing {
		cursorMark = dimBoldStyle.Render("▸ ")
	}

	switch blk.Kind {
	case BlockHeading:
		text := blk.Text
		if isEditing {
			text = editDraft
		}
		rendered := renderHeading(text, blk.Level, w)
		if isCursor && !isEditing {
			b.WriteString(cursorMark + rendered)
		} else if isEditing {
			b.WriteString(editStyle.Width(w - 4).Render(text))
		} else {
			b.WriteString("  " + rendered)
		}

	case BlockParagraph:
		text := blk.Text
		if isEditing {
			text = editDraft
		}
		rendered := inlineStyle(text)
		if isCursor && !isEditing {
			b.WriteString(cursorMark + rendered)
		} else if isEditing {
			b.WriteString(editStyle.Width(w - 4).Render(text))
		} else {
			b.WriteString("  " + rendered)
		}

	case BlockCode:
		content := strings.Join(blk.Lines, "\n")
		if isEditing {
			content = editDraft
		}

		header := ""
		if blk.Lang != "" {
			langLabel := dimBoldStyle.Render(" " + blk.Lang + " ")
			gap := w - 4 - lipgloss.Width(langLabel)
			if gap < 0 {
				gap = 0
			}
			header = strings.Repeat(" ", gap) + langLabel + "\n"
		}

		var codeContent string
		if isEditing {
			codeContent = editStyle.Width(w - 6).Render(content)
		} else {
			codeContent = highlightCode(blk.Lines, blk.Lang)
		}

		rendered := codeBlockStyle.Width(w - 4).Render(codeContent)
		if isCursor && !isEditing {
			b.WriteString(cursorMark + header + rendered)
		} else {
			b.WriteString("  " + header + rendered)
		}

	case BlockImage:
		placeholder := fmt.Sprintf("[ image: %s ]", blk.Src)
		if isCursor {
			b.WriteString(cursorMark + imageStyle.Render(placeholder))
		} else {
			b.WriteString("  " + imageStyle.Render(placeholder))
		}

	case BlockList:
		text := blk.Text
		if isEditing {
			text = editDraft
		}
		rendered := inlineStyle(text)
		if isCursor && !isEditing {
			b.WriteString(cursorMark + rendered)
		} else if isEditing {
			b.WriteString(editStyle.Width(w - 4).Render(text))
		} else {
			b.WriteString("  " + rendered)
		}

	case BlockDirective:
		if isCursor {
			b.WriteString(cursorMark + dimStyle.Render(blk.Directive))
		} else {
			b.WriteString("  " + dimStyle.Render(blk.Directive))
		}
	}

	return b.String()
}

// --- Slide rendering ---

func renderSlide(slide Slide, w int, e Editor) string {
	var lines []string
	for i, blk := range slide.Blocks {
		isCursor := i == e.BlockIdx
		isEditing := isCursor && e.Mode == ModeEdit
		editDraft := ""
		if isEditing {
			editDraft = e.Draft
		}
		rendered := renderBlock(blk, w, isCursor, isEditing, editDraft, e.CursorCol)
		lines = append(lines, rendered)
	}
	return strings.Join(lines, "\n\n")
}

// --- Full view ---

func View(d Deck, e Editor, width, height int) string {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	bodyHeight := height - 2
	if bodyHeight < 1 {
		bodyHeight = 1
	}

	content := ""
	if e.SlideIdx < len(d.Slides) {
		content = renderSlide(d.Slides[e.SlideIdx], width, e)
	}

	body := lipgloss.NewStyle().
		Width(width).
		Height(bodyHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)

	status := ""
	if e.Mode == ModeEdit {
		status = editStatus(e, width)
	} else {
		status = navStatus(d, e, width)
	}

	return body + "\n" + status
}

func navStatus(d Deck, e Editor, w int) string {
	left := fmt.Sprintf("slide %d/%d  ·  blocks %d", e.SlideIdx+1, len(d.Slides), len(d.Slides[e.SlideIdx].Blocks))
	if e.Dirty {
		left += "  ·  [modified]"
	}
	right := "i edit · ^n add · ^d del · ^s save · u undo · q quit"
	status := left + "  ·  " + right
	return dimStyle.Width(w).Render(status)
}

func editStatus(e Editor, w int) string {
	left := fmt.Sprintf("editing  ·  col %d/%d", e.CursorCol, len(e.Draft))
	right := "esc cancel · enter confirm"
	status := left + "  ·  " + right
	if e.Message != "" {
		status = e.Message + "  ·  " + status
	}
	return messageStyle.Width(w).Render(status)
}
