package internal

import (
	"fmt"
	"path/filepath"
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

	h1Style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Underline(true)
	h2Style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	h3Style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D8D8D8"))
	h4Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#B0B0B0"))
	h5Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	h6Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#606060"))

	laserPointerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF2A55"))

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

func renderHeading(text string, level int) string {
	switch level {
	case 1:
		return h1Style.Render(text)
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

// --- Image card rendering ---

func renderImageCard(src string, baseDir string, maxW int) string {
	info := GetImageInfo(src, baseDir)

	cardW := maxW
	if cardW > 46 {
		cardW = 46
	}
	if cardW < 28 {
		cardW = 28
	}

	cardBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(cardW)

	if !info.Exists {
		msg := fmt.Sprintf("⚠  image not found: %s", src)
		return cardBorder.BorderForeground(lipgloss.Color("196")).Render(
			lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(msg),
		)
	}

	name := filepath.Base(src)
	titleLine := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Render("🖼  " + name)

	detail := ""
	if info.Width > 0 && info.Height > 0 {
		detail = fmt.Sprintf("%d × %d px", info.Width, info.Height)
		if info.Format != "" {
			detail += "  ·  " + info.Format
		}
	} else if info.Format != "" {
		detail = info.Format
	}

	detailLine := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true).Render(detail)
	hintLine := lipgloss.NewStyle().Foreground(lipgloss.Color("239")).Render("press 'p' to open")

	content := titleLine
	if detail != "" {
		content += "\n   " + detailLine
	}
	content += "\n   " + hintLine

	return cardBorder.Render(content)
}

// --- Block rendering ---

func renderBlock(blk Block, w int, maxBlockH int, baseDir string, isCursor bool, isEditing bool, editDraft string, cursorCol int) string {
	var b strings.Builder

	cursorMark := "  "
	if isCursor && !isEditing {
		cursorMark = laserPointerStyle.Render("▶ ")
	}

	switch blk.Kind {
	case BlockHeading:
		text := blk.Text
		if isEditing {
			text = editDraft
		}
		rendered := renderHeading(text, blk.Level)
		if isEditing {
			b.WriteString(editStyle.Width(w - 4).Render(text))
		} else {
			lines := strings.Split(rendered, "\n")
			for idx, l := range lines {
				if idx > 0 {
					b.WriteString("\n")
				}
				if idx == 0 && isCursor {
					b.WriteString(cursorMark + l)
				} else {
					b.WriteString("  " + l)
				}
			}
		}

	case BlockParagraph:
		text := blk.Text
		if isEditing {
			text = editDraft
		}
		rendered := inlineStyle(text)
		if isEditing {
			b.WriteString(editStyle.Width(w - 4).Render(text))
		} else {
			lines := strings.Split(rendered, "\n")
			for idx, l := range lines {
				if idx > 0 {
					b.WriteString("\n")
				}
				if idx == 0 && isCursor {
					b.WriteString(cursorMark + l)
				} else {
					b.WriteString("  " + l)
				}
			}
		}

	case BlockCode:
		content := strings.Join(blk.Lines, "\n")
		if isEditing {
			content = editDraft
			b.WriteString(editStyle.Width(w - 4).Render(content))
			break
		}

		maxCodeLineLen := 0
		for _, l := range blk.Lines {
			if len(l) > maxCodeLineLen {
				maxCodeLineLen = len(l)
			}
		}
		if blk.Lang != "" && len(blk.Lang)+6 > maxCodeLineLen {
			maxCodeLineLen = len(blk.Lang) + 6
		}
		boxW := maxCodeLineLen + 6
		if boxW > w-8 {
			boxW = w - 8
		}
		if boxW < 32 {
			boxW = 32
		}

		header := ""
		if blk.Lang != "" {
			langLabel := dimBoldStyle.Render(" " + blk.Lang + " ")
			gap := boxW - lipgloss.Width(langLabel)
			if gap < 0 {
				gap = 0
			}
			header = strings.Repeat(" ", gap) + langLabel + "\n"
		}

		codeContent := highlightCode(blk.Lines, blk.Lang)
		rendered := codeBlockStyle.Width(boxW).Render(codeContent)

		fullCode := header + rendered
		lines := strings.Split(fullCode, "\n")
		targetLine := 0
		if header != "" {
			targetLine = 1
		}
		for idx, l := range lines {
			if idx > 0 {
				b.WriteString("\n")
			}
			if idx == targetLine && isCursor {
				b.WriteString(cursorMark + l)
			} else {
				b.WriteString("  " + l)
			}
		}

	case BlockImage:
		if isEditing {
			b.WriteString(editStyle.Width(w - 4).Render(editDraft))
			break
		}

		card := renderImageCard(blk.Src, baseDir, w-6)
		lines := strings.Split(card, "\n")
		for idx, l := range lines {
			if idx > 0 {
				b.WriteString("\n")
			}
			if idx == 0 && isCursor {
				b.WriteString(cursorMark + l)
			} else {
				b.WriteString("  " + l)
			}
		}

	case BlockList:
		text := blk.Text
		if isEditing {
			text = editDraft
		}
		rendered := inlineStyle(text)
		if isEditing {
			b.WriteString(editStyle.Width(w - 4).Render(text))
		} else {
			lines := strings.Split(rendered, "\n")
			for idx, l := range lines {
				if idx > 0 {
					b.WriteString("\n")
				}
				if idx == 0 && isCursor {
					b.WriteString(cursorMark + l)
				} else {
					b.WriteString("  " + l)
				}
			}
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

func renderSlide(slide Slide, w, h int, baseDir string, e Editor) string {
	var lines []string
	maxBlockH := h - 4
	if maxBlockH < 8 {
		maxBlockH = 8
	}
	for i, blk := range slide.Blocks {
		isCursor := i == e.BlockIdx
		isEditing := isCursor && e.Mode == ModeEdit
		editDraft := ""
		if isEditing {
			editDraft = e.Draft
		}
		rendered := renderBlock(blk, w, maxBlockH, baseDir, isCursor, isEditing, editDraft, e.CursorCol)
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

	baseDir := d.BaseDir
	if baseDir == "" && e.FilePath != "" {
		baseDir = filepath.Dir(e.FilePath)
	}

	align := d.Align
	if align == "" {
		align = AlignCenter
	}
	if e.SlideIdx < len(d.Slides) && d.Slides[e.SlideIdx].Align != "" {
		align = d.Slides[e.SlideIdx].Align
	}

	content := ""
	if e.SlideIdx < len(d.Slides) {
		content = renderSlide(d.Slides[e.SlideIdx], width, bodyHeight, baseDir, e)
	}

	var hAlign lipgloss.Position
	padLeft := 0
	padRight := 0

	switch align {
	case AlignLeft:
		hAlign = lipgloss.Left
		padLeft = 8
	case AlignRight:
		hAlign = lipgloss.Right
		padRight = 8
	default:
		hAlign = lipgloss.Center
	}

	body := lipgloss.NewStyle().
		Width(width).
		Height(bodyHeight).
		Align(hAlign, lipgloss.Center).
		PaddingLeft(padLeft).
		PaddingRight(padRight).
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
	align := d.Align
	if align == "" {
		align = AlignCenter
	}
	if e.SlideIdx < len(d.Slides) && d.Slides[e.SlideIdx].Align != "" {
		align = d.Slides[e.SlideIdx].Align
	}

	left := fmt.Sprintf("slide %d/%d (%s)  ·  blocks %d", e.SlideIdx+1, len(d.Slides), align, len(d.Slides[e.SlideIdx].Blocks))
	if e.Dirty {
		left += "  ·  [modified]"
	}
	if e.Message != "" {
		left += "  ·  " + e.Message
	}
	right := "tab align · i edit · ^n add · ^d del · ^s save · u undo · q quit"
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
