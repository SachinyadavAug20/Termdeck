package internal

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// --- Styles ---

var (
	boldStyle      = lipgloss.NewStyle().Bold(true)
	italicStyle    = lipgloss.NewStyle().Italic(true)
	codeBlockStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	imageStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	dimStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dimBoldStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Bold(true)
	cursorStyle    = lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("0"))
	editStyle      = lipgloss.NewStyle().Background(lipgloss.Color("22")).Foreground(lipgloss.Color("252")).Padding(0, 1)
	messageStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	notesBoxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Foreground(lipgloss.Color("252")).Padding(0, 1)

	h2Style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	h3Style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#D8D8D8"))
	h4Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#B0B0B0"))
	h5Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	h6Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#606060"))

	reCodeSpan = regexp.MustCompile("`" + `([^` + "`" + `]+)` + "`")
	reBold     = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reItalic   = regexp.MustCompile(`\*([^*]+)\*`)
)

// --- Inline styling ---

func inlineStyle(text string) string {
	matches := reCodeSpan.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		result := reBold.ReplaceAllStringFunc(text, func(m string) string {
			return boldStyle.Render(m[2 : len(m)-2])
		})
		return reItalic.ReplaceAllStringFunc(result, func(m string) string {
			return italicStyle.Render(m[1 : len(m)-1])
		})
	}

	var b strings.Builder
	lastIdx := 0
	styledSpans := make([]string, len(matches))

	for i, m := range matches {
		b.WriteString(text[lastIdx:m[0]])
		inner := text[m[0]+1 : m[1]-1]
		styledSpans[i] = currentTheme.CodeSpanStyle.Render(inner)
		fmt.Fprintf(&b, "\x00CODE%d\x00", i)
		lastIdx = m[1]
	}
	b.WriteString(text[lastIdx:])
	result := b.String()

	result = reBold.ReplaceAllStringFunc(result, func(m string) string {
		return boldStyle.Render(m[2 : len(m)-2])
	})

	result = reItalic.ReplaceAllStringFunc(result, func(m string) string {
		return italicStyle.Render(m[1 : len(m)-1])
	})

	for i, styled := range styledSpans {
		ph := fmt.Sprintf("\x00CODE%d\x00", i)
		result = strings.Replace(result, ph, styled, 1)
	}

	return result
}

// --- Heading rendering ---

func renderHeading(text string, level int) string {
	switch level {
	case 1:
		return currentTheme.H1Style.Render(text)
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
	// Go
	"func": true, "return": true, "if": true, "else": true, "for": true,
	"range": true, "var": true, "const": true, "type": true, "struct": true,
	"package": true, "import": true, "defer": true, "go": true, "chan": true,
	"select": true, "case": true, "switch": true, "default": true, "break": true,
	"continue": true, "true": true, "false": true, "nil": true, "interface": true,
	"map": true, "make": true, "new": true, "panic": true, "recover": true, "iota": true,
	"any": true,

	// Python
	"def": true, "class": true, "from": true, "as": true,
	"print": true, "self": true, "None": true, "True": true, "False": true,
	"with": true, "try": true, "except": true, "finally": true, "raise": true,
	"yield": true, "lambda": true, "pass": true, "del": true, "global": true,
	"nonlocal": true, "assert": true, "is": true, "in": true, "not": true,
	"and": true, "or": true, "elif": true, "while": true,

	// JavaScript / TypeScript
	"let": true, "function": true, "async": true, "await": true, "export": true,
	"null": true, "undefined": true, "typeof": true, "instanceof": true, "throw": true,
	"catch": true, "enum": true, "implements": true, "extends": true,

	// Rust
	"fn": true, "mut": true, "impl": true, "trait": true, "match": true,
	"pub": true, "use": true, "mod": true, "loop": true, "where": true,
	"crate": true, "ref": true, "move": true, "dyn": true,

	// Shell / Bash
	"echo": true, "then": true, "fi": true, "do": true, "done": true,
	"esac": true, "exit": true, "local": true,

	// SQL
	"SELECT": true, "FROM": true, "WHERE": true, "INSERT": true, "INTO": true,
	"UPDATE": true, "DELETE": true, "JOIN": true, "LEFT": true, "RIGHT": true,
	"INNER": true, "OUTER": true, "GROUP": true, "BY": true, "ORDER": true,
	"HAVING": true, "LIMIT": true, "OFFSET": true, "CREATE": true, "TABLE": true,
	"DROP": true, "ALTER": true,
}

func highlightLine(line string) string {
	var result strings.Builder
	i := 0
	for i < len(line) {
		ch := line[i]

		if ch == '#' || (ch == '/' && i+1 < len(line) && line[i+1] == '/') {
			result.WriteString(currentTheme.SyntaxComment.Render(line[i:]))
			return result.String()
		}
		if ch == '-' && i+1 < len(line) && line[i+1] == '-' {
			result.WriteString(currentTheme.SyntaxComment.Render(line[i:]))
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
			result.WriteString(currentTheme.SyntaxString.Render(line[i:j]))
			i = j
			continue
		}

		if ch >= '0' && ch <= '9' {
			j := i
			for j < len(line) && ((line[j] >= '0' && line[j] <= '9') || line[j] == '.') {
				j++
			}
			result.WriteString(currentTheme.SyntaxNumber.Render(line[i:j]))
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
				result.WriteString(currentTheme.SyntaxKeyword.Render(word))
			} else if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
				result.WriteString(currentTheme.SyntaxType.Render(word))
			} else {
				result.WriteString(currentTheme.SyntaxPlain.Render(word))
			}
			i = j
			continue
		}

		result.WriteString(currentTheme.SyntaxPlain.Render(string(ch)))
		i++
	}

	return result.String()
}

func highlightCode(lines []string, lang string, showLineNumbers bool) string {
	normLang := strings.ToLower(strings.TrimSpace(lang))
	var result strings.Builder

	digits := 1
	if len(lines) > 0 {
		digits = len(fmt.Sprintf("%d", len(lines)))
	}

	getGutter := func(idx int) string {
		if !showLineNumbers {
			return ""
		}
		numStr := fmt.Sprintf("%*d │ ", digits, idx+1)
		return currentTheme.SyntaxComment.Render(numStr)
	}

	if normLang == "diff" || normLang == "patch" {
		for i, line := range lines {
			if i > 0 {
				result.WriteString("\n")
			}
			gutter := getGutter(i)
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				result.WriteString(gutter + currentTheme.SyntaxString.Render(line))
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				result.WriteString(gutter + currentTheme.LaserPointerStyle.Render(line))
			} else if strings.HasPrefix(line, "@@") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
				result.WriteString(gutter + currentTheme.SyntaxComment.Render(line))
			} else {
				result.WriteString(gutter + currentTheme.SyntaxPlain.Render(line))
			}
		}
		return result.String()
	}

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		gutter := getGutter(i)
		result.WriteString(gutter + highlightLine(line))
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

func renderBlock(blk Block, w int, maxBlockH int, baseDir string, isCursor bool, isEditing bool, editDraft string, cursorCol int, showLineNumbers bool) string {
	var b strings.Builder

	cursorMark := "  "
	if isCursor && !isEditing {
		cursorMark = currentTheme.LaserPointerStyle.Render("▶ ")
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
		if showLineNumbers && len(blk.Lines) > 0 {
			digits := len(fmt.Sprintf("%d", len(blk.Lines)))
			maxCodeLineLen += digits + 3
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

		codeContent := highlightCode(blk.Lines, blk.Lang, showLineNumbers)
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
			b.WriteString(editStyle.Width(w - 4).Render(text))
		} else {
			rendered := renderListItem(text)
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
		if strings.HasPrefix(blk.Directive, "::notes") {
			return ""
		}
		if isCursor {
			b.WriteString(cursorMark + dimStyle.Render(blk.Directive))
		} else {
			b.WriteString("  " + dimStyle.Render(blk.Directive))
		}

	case BlockTable:
		text := blk.Text
		if isEditing {
			text = editDraft
		}
		if isEditing {
			b.WriteString(editStyle.Width(w - 4).Render(text))
		} else {
			rendered := renderTable(blk, w)
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

	case BlockCallout:
		text := blk.Text
		if isEditing {
			text = editDraft
			b.WriteString(editStyle.Width(w - 4).Render(text))
		} else {
			rendered := renderCallout(blk, w)
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

	case BlockDivider:
		if isEditing {
			b.WriteString(editStyle.Width(w - 4).Render("***"))
		} else {
			divW := 32
			if divW > w-8 {
				divW = w - 8
			}
			divider := currentTheme.TableBorderStyle.Render(strings.Repeat("─", divW))
			if isCursor {
				b.WriteString(cursorMark + divider)
			} else {
				b.WriteString("  " + divider)
			}
		}
	}

	return b.String()
}

func renderListItem(text string) string {
	trimmed := strings.TrimSpace(text)

	prefix := ""
	if strings.HasPrefix(trimmed, "- ") {
		prefix = "- "
	} else if strings.HasPrefix(trimmed, "* ") {
		prefix = "* "
	}

	if prefix != "" {
		content := trimmed[len(prefix):]
		if strings.HasPrefix(content, "[x] ") || strings.HasPrefix(content, "[X] ") {
			itemText := content[4:]
			checkMark := currentTheme.SyntaxString.Render("✔ ")
			return checkMark + dimStyle.Render(inlineStyle(itemText))
		} else if strings.HasPrefix(content, "[ ] ") {
			itemText := content[4:]
			circleMark := currentTheme.ProgressLineDimStyle.Render("○ ")
			return circleMark + inlineStyle(itemText)
		}
		bullet := currentTheme.TableHeaderStyle.Render("• ")
		return bullet + inlineStyle(content)
	}

	if idx := strings.Index(trimmed, ". "); idx > 0 && idx < 5 {
		isNum := true
		for _, ch := range trimmed[:idx] {
			if ch < '0' || ch > '9' {
				isNum = false
				break
			}
		}
		if isNum {
			numPart := trimmed[:idx+2]
			content := trimmed[idx+2:]
			return currentTheme.TableHeaderStyle.Render(numPart) + inlineStyle(content)
		}
	}

	return inlineStyle(text)
}

func renderCallout(blk Block, w int) string {
	boxW := w - 8
	if boxW > 68 {
		boxW = 68
	}
	if boxW < 30 {
		boxW = 30
	}

	icon := "💡"
	title := "TIP"
	color := currentTheme.Success

	switch blk.Callout {
	case "note":
		icon = "ℹ"
		title = "NOTE"
		color = currentTheme.Accent
	case "warning":
		icon = "⚠"
		title = "WARNING"
		color = currentTheme.Warning
	case "important":
		icon = "🚨"
		title = "IMPORTANT"
		color = currentTheme.Secondary
	case "caution":
		icon = "🛑"
		title = "CAUTION"
		color = currentTheme.Laser
	case "quote":
		icon = "❝"
		title = "QUOTE"
		color = currentTheme.Comment
	default:
		icon = "💡"
		title = "TIP"
		color = currentTheme.Success
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color)).Render(icon + "  " + title)

	lines := blk.Lines
	if len(lines) == 0 && blk.Text != "" {
		lines = strings.Split(blk.Text, "\n")
	}

	var styledBody strings.Builder
	for i, l := range lines {
		if i > 0 {
			styledBody.WriteString("\n")
		}
		if blk.Callout == "quote" {
			styledBody.WriteString(lipgloss.NewStyle().Italic(true).Render(inlineStyle(l)))
		} else {
			styledBody.WriteString(inlineStyle(l))
		}
	}

	content := header + "\n" + styledBody.String()

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(color)).
		Padding(0, 1).
		Width(boxW).
		Render(content)

	return card
}

func renderTable(blk Block, w int) string {
	rawLines := blk.Lines
	if len(rawLines) == 0 && blk.Text != "" {
		rawLines = strings.Split(blk.Text, "\n")
	}
	if len(rawLines) == 0 {
		return ""
	}

	var rows [][]string
	for _, l := range rawLines {
		trimmed := strings.TrimSpace(l)
		if !strings.HasPrefix(trimmed, "|") {
			continue
		}
		clean := strings.Trim(trimmed, "|")
		isSep := true
		for _, ch := range clean {
			if ch != '-' && ch != ':' && ch != '|' && ch != ' ' {
				isSep = false
				break
			}
		}
		if isSep {
			continue
		}

		parts := strings.Split(trimmed, "|")
		var row []string
		for i := 1; i < len(parts)-1; i++ {
			row = append(row, strings.TrimSpace(parts[i]))
		}
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	if len(rows) == 0 {
		return ""
	}

	cols := 0
	for _, r := range rows {
		if len(r) > cols {
			cols = len(r)
		}
	}

	colWidths := make([]int, cols)
	for _, r := range rows {
		for i, cell := range r {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	for i := range colWidths {
		if colWidths[i] < 4 {
			colWidths[i] = 4
		}
	}

	var sb strings.Builder
	header := rows[0]
	sb.WriteString("│")
	for i := 0; i < cols; i++ {
		val := ""
		if i < len(header) {
			val = header[i]
		}
		padded := fmt.Sprintf(" %-*s ", colWidths[i], val)
		sb.WriteString(currentTheme.TableHeaderStyle.Render(padded))
		sb.WriteString(currentTheme.TableBorderStyle.Render("│"))
	}
	sb.WriteString("\n")

	sb.WriteString("├")
	for i := 0; i < cols; i++ {
		sb.WriteString(strings.Repeat("─", colWidths[i]+2))
		if i < cols-1 {
			sb.WriteString("┼")
		} else {
			sb.WriteString("┤")
		}
	}
	sb.WriteString("\n")

	for rIdx := 1; rIdx < len(rows); rIdx++ {
		r := rows[rIdx]
		sb.WriteString("│")
		for i := 0; i < cols; i++ {
			val := ""
			if i < len(r) {
				val = r[i]
			}
			padded := fmt.Sprintf(" %-*s ", colWidths[i], val)
			sb.WriteString(currentTheme.TableCellStyle.Render(padded))
			sb.WriteString(currentTheme.TableBorderStyle.Render("│"))
		}
		if rIdx < len(rows)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// --- Slide rendering ---

func renderSlide(slide Slide, w, h int, baseDir string, e Editor) string {
	var lines []string
	maxBlockH := h - 4
	if maxBlockH < 8 {
		maxBlockH = 8
	}
	visibleIndices := slide.VisibleBlockIndices()
	for _, idx := range visibleIndices {
		blk := slide.Blocks[idx]
		isCursor := idx == e.BlockIdx
		isEditing := isCursor && e.Mode == ModeEdit
		editDraft := ""
		if isEditing {
			editDraft = e.Draft
		}
		rendered := renderBlock(blk, w, maxBlockH, baseDir, isCursor, isEditing, editDraft, e.CursorCol, e.ShowLineNumbers)
		if rendered != "" {
			lines = append(lines, rendered)
		}
	}
	return strings.Join(lines, "\n\n")
}

func renderNotesOverlay(notes string, width, maxHeight int) string {
	if maxHeight < 3 {
		maxHeight = 3
	}
	content := notes
	if strings.TrimSpace(content) == "" {
		content = dimStyle.Render("(no speaker notes for this slide)")
	}
	title := currentTheme.NotesTitleStyle.Render("📝 Speaker Notes") + dimStyle.Render(" (press 'n' to hide · 't' theme)")
	boxW := width - 4
	if boxW < 20 {
		boxW = width
	}
	inner := title + "\n" + content
	return notesBoxStyle.BorderForeground(lipgloss.Color(currentTheme.Accent)).Width(boxW).MaxHeight(maxHeight).Render(inner)
}

func renderHelpModal(w, h int) string {
	boxW := 66
	if boxW > w-4 {
		boxW = w - 4
	}
	if boxW < 36 {
		boxW = 36
	}

	var sb strings.Builder
	sb.WriteString(currentTheme.HelpTitleStyle.Render("Termdeck Keyboard Controls"))
	sb.WriteString("\n" + dimStyle.Render("Press '?' or 'Esc' to close") + "\n\n")

	sb.WriteString(currentTheme.HelpHeaderStyle.Render("  NAVIGATION") + "\n")
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("→, l, Space, Enter"), currentTheme.HelpDescStyle.Render("Next slide")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("←, h, Backspace"), currentTheme.HelpDescStyle.Render("Previous slide")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("/"), currentTheme.HelpDescStyle.Render("Jump to slide (number or search)")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("o / O"), currentTheme.HelpDescStyle.Render("Slide overview & grid sorter")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("↓, j"), currentTheme.HelpDescStyle.Render("Move laser pointer down")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("↑, k"), currentTheme.HelpDescStyle.Render("Move laser pointer up")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("g / G"), currentTheme.HelpDescStyle.Render("First / Last slide")))

	sb.WriteString("\n" + currentTheme.HelpHeaderStyle.Render("  PRESENTATION") + "\n")
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("t, T, ctrl+t, f2"), currentTheme.HelpDescStyle.Render(fmt.Sprintf("Cycle color theme (%s)", currentTheme.Name))))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("z"), currentTheme.HelpDescStyle.Render("Toggle distraction-free zen mode")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("b / B"), currentTheme.HelpDescStyle.Render("Blank presentation screen")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("y / Y"), currentTheme.HelpDescStyle.Render("Copy block to clipboard (OSC 52)")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("c / C"), currentTheme.HelpDescStyle.Render("Toggle presentation timer / Reset timer")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("r / R"), currentTheme.HelpDescStyle.Render("Reload deck file from disk")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("L"), currentTheme.HelpDescStyle.Render("Toggle code block line numbers")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("x"), currentTheme.HelpDescStyle.Render("Toggle task item ([ ] ⇄ [x]) & auto-save")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("n"), currentTheme.HelpDescStyle.Render("Toggle speaker notes overlay")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("Tab / ctrl+a"), currentTheme.HelpDescStyle.Render("Cycle alignment (left/center/right)")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("p"), currentTheme.HelpDescStyle.Render("Open image in system viewer")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("E"), currentTheme.HelpDescStyle.Render("Export presentation to HTML")))

	sb.WriteString("\n" + currentTheme.HelpHeaderStyle.Render("  LIVE EDITOR") + "\n")
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("i"), currentTheme.HelpDescStyle.Render("Edit focused block")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("ctrl+t / f2"), currentTheme.HelpDescStyle.Render("Cycle color theme while editing")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("Enter"), currentTheme.HelpDescStyle.Render("Confirm edit & auto-save")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("Esc"), currentTheme.HelpDescStyle.Render("Cancel edit / Close help")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("ctrl+n / ctrl+d"), currentTheme.HelpDescStyle.Render("Add / Delete block")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("ctrl+k / ctrl+j"), currentTheme.HelpDescStyle.Render("Move block up / down")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("u / ctrl+r"), currentTheme.HelpDescStyle.Render("Undo / Redo")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("ctrl+s"), currentTheme.HelpDescStyle.Render("Save file manually")))

	sb.WriteString("\n" + fmt.Sprintf("  %-22s %s\n", currentTheme.HelpKeyStyle.Render("q / ctrl+c"), currentTheme.HelpDescStyle.Render("Quit (auto-saves changes)")))

	box := currentTheme.HelpBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

func renderProgressLine(curSlide, totalSlides, width int) string {
	if width <= 0 {
		return ""
	}
	if totalSlides <= 0 {
		return currentTheme.ProgressLineDimStyle.Render(strings.Repeat("─", width))
	}
	if curSlide < 1 {
		curSlide = 1
	}
	if curSlide > totalSlides {
		curSlide = totalSlides
	}

	filled := int(float64(curSlide) / float64(totalSlides) * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	unfilled := width - filled

	var b strings.Builder
	if filled > 0 {
		b.WriteString(currentTheme.ProgressLineFilledStyle.Render(strings.Repeat("─", filled)))
	}
	if unfilled > 0 {
		b.WriteString(currentTheme.ProgressLineDimStyle.Render(strings.Repeat("─", unfilled)))
	}
	return b.String()
}

func renderJumpModal(d Deck, e Editor, w, h int) string {
	boxW := 54
	if boxW > w-4 {
		boxW = w - 4
	}
	if boxW < 32 {
		boxW = 32
	}

	var sb strings.Builder
	title := currentTheme.HelpTitleStyle.Render("🔍 Jump to Slide")
	sb.WriteString(title + "\n")
	sb.WriteString(dimStyle.Render("Type slide number (1-"+fmt.Sprintf("%d", len(d.Slides))+") or title search:") + "\n\n")

	cursorDraft := e.Draft + currentTheme.LaserPointerStyle.Render("█")
	inputLine := currentTheme.TableHeaderStyle.Render("  > ") + cursorDraft
	sb.WriteString(inputLine + "\n\n")

	// Show matching or nearby slides preview
	query := strings.ToLower(strings.TrimSpace(e.Draft))
	var matches []int
	for idx, slide := range d.Slides {
		if query == "" {
			if len(matches) < 6 {
				matches = append(matches, idx)
			}
		} else {
			numStr := fmt.Sprintf("%d", idx+1)
			if strings.HasPrefix(numStr, query) || strings.Contains(strings.ToLower(slide.Title()), query) {
				if len(matches) < 6 {
					matches = append(matches, idx)
				}
			}
		}
	}

	for _, idx := range matches {
		slide := d.Slides[idx]
		t := slide.Title()
		if len(t) > 34 {
			t = t[:34] + "..."
		}
		mark := "  "
		if idx == e.SlideIdx {
			mark = currentTheme.LaserPointerStyle.Render("▶ ")
		}
		num := fmt.Sprintf("%2d. ", idx+1)
		line := mark + dimBoldStyle.Render(num) + currentTheme.TableCellStyle.Render(t)
		sb.WriteString(line + "\n")
	}

	sb.WriteString("\n" + dimStyle.Render("Press Enter to jump · Esc to cancel"))

	card := currentTheme.HelpBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, card)
}

func renderOverviewModal(d Deck, e Editor, w, h int) string {
	totalSlides := len(d.Slides)
	if totalSlides == 0 {
		return ""
	}

	modalW := w - 4
	if modalW > 84 {
		modalW = 84
	}
	if modalW < 36 {
		modalW = 36
	}

	cols := 3
	if modalW < 54 {
		cols = 1
	} else if modalW < 74 {
		cols = 2
	}

	// Width of each card
	cardW := (modalW - 6 - (cols-1)*2) / cols
	if cardW < 18 {
		cardW = 18
	}

	cursor := e.OverviewCursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= totalSlides {
		cursor = totalSlides - 1
	}

	cursorRow := cursor / cols
	totalRows := (totalSlides + cols - 1) / cols

	// Determine how many rows we can display
	maxVisibleRows := (h - 9) / 4
	if maxVisibleRows < 1 {
		maxVisibleRows = 1
	}
	if maxVisibleRows > 4 {
		maxVisibleRows = 4
	}

	startRow := 0
	if cursorRow >= maxVisibleRows {
		startRow = cursorRow - maxVisibleRows + 1
	}
	endRow := startRow + maxVisibleRows
	if endRow > totalRows {
		endRow = totalRows
	}

	var sb strings.Builder
	title := currentTheme.HelpTitleStyle.Render("🗂  Slide Overview & Grid Sorter")
	sb.WriteString(title + "\n")
	sb.WriteString(dimStyle.Render(fmt.Sprintf("%d slides · slide %d/%d selected", totalSlides, cursor+1, totalSlides)) + "\n\n")

	if startRow > 0 {
		sb.WriteString(dimStyle.Render("  ▲  more slides above") + "\n")
	}

	for r := startRow; r < endRow; r++ {
		var rowCards []string
		for c := 0; c < cols; c++ {
			idx := r*cols + c
			if idx >= totalSlides {
				emptyStyle := lipgloss.NewStyle().Width(cardW + 2)
				rowCards = append(rowCards, emptyStyle.Render(""))
				continue
			}

			slide := d.Slides[idx]
			isCursor := idx == cursor
			isCurrent := idx == e.SlideIdx

			// Header line: #N and badges
			numStr := fmt.Sprintf("#%d", idx+1)
			if isCursor {
				numStr = "▶ " + numStr
			} else {
				numStr = "  " + numStr
			}
			if isCurrent {
				numStr += " " + currentTheme.ProgressLineFilledStyle.Render("●")
			}

			// Title line: truncated
			titleText := slide.Title()
			maxTitleLen := cardW - 4
			if maxTitleLen < 8 {
				maxTitleLen = 8
			}
			if len(titleText) > maxTitleLen {
				titleText = titleText[:maxTitleLen-3] + "..."
			}

			// Summary line: blks, code, etc.
			sumText := slide.Summary()
			maxSumLen := cardW - 4
			if maxSumLen < 8 {
				maxSumLen = 8
			}
			if len(sumText) > maxSumLen {
				sumText = sumText[:maxSumLen-3] + "..."
			}

			var cardContent string
			if isCursor {
				cardContent = currentTheme.HelpKeyStyle.Render(numStr) + "\n" +
					dimBoldStyle.Render(titleText) + "\n" +
					dimStyle.Render(sumText)
			} else if isCurrent {
				cardContent = currentTheme.HelpDescStyle.Render(numStr) + "\n" +
					currentTheme.TableCellStyle.Render(titleText) + "\n" +
					dimStyle.Render(sumText)
			} else {
				cardContent = dimBoldStyle.Render(numStr) + "\n" +
					currentTheme.TableCellStyle.Render(titleText) + "\n" +
					dimStyle.Render(sumText)
			}

			// Border
			cardBorder := lipgloss.RoundedBorder()
			borderColor := lipgloss.Color(currentTheme.DimTrack)
			if isCursor {
				borderColor = lipgloss.Color(currentTheme.Laser)
			} else if isCurrent {
				borderColor = lipgloss.Color(currentTheme.Success)
			}

			cStyle := lipgloss.NewStyle().
				Border(cardBorder).
				BorderForeground(borderColor).
				Width(cardW).
				Height(3).
				Padding(0, 1)

			rowCards = append(rowCards, cStyle.Render(cardContent))
		}
		sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, rowCards...) + "\n")
	}

	if endRow < totalRows {
		sb.WriteString(dimStyle.Render("  ▼  more slides below") + "\n")
	}

	sb.WriteString("\n" + dimStyle.Render("←/→/↑/↓ or hjkl: navigate · Enter/Space: jump to slide · Esc/o: close"))

	card := currentTheme.HelpBoxStyle.Width(modalW).Render(sb.String())
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, card)
}

func renderBlankScreen(w, h int) string {
	msg := dimStyle.Render("●  presentation paused  ·  press any key to resume")
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, msg)
}

// --- Full view ---

func View(d Deck, e Editor, width, height int) string {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	themeName := e.Theme
	if themeName == "" {
		themeName = d.Theme
	}
	SetCurrentTheme(themeName)

	if e.ScreenBlank {
		return renderBlankScreen(width, height)
	}

	if e.ShowHelp {
		return renderHelpModal(width, height)
	}

	if e.ShowOverview {
		return renderOverviewModal(d, e, width, height)
	}

	if e.Mode == ModePrompt {
		return renderJumpModal(d, e, width, height)
	}

	notesOverlay := ""
	notesHeight := 0
	if e.ShowNotes {
		notes := ""
		if e.SlideIdx < len(d.Slides) {
			notes = d.Slides[e.SlideIdx].Notes()
		}
		maxNotesH := height / 3
		if maxNotesH < 4 {
			maxNotesH = 4
		}
		if maxNotesH > 8 {
			maxNotesH = 8
		}
		notesOverlay = renderNotesOverlay(notes, width, maxNotesH)
		notesHeight = lipgloss.Height(notesOverlay)
	}

	bodyHeight := height - 3
	if e.ZenMode {
		bodyHeight = height - 2
	}
	if notesHeight > 0 {
		bodyHeight -= notesHeight
	}
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
	if !e.ZenMode {
		if e.Mode == ModeEdit {
			status = editStatus(e, width)
		} else {
			status = navStatus(d, e, width)
		}
	}

	curSlide := e.SlideIdx + 1
	totalSlides := len(d.Slides)
	progressLine := renderProgressLine(curSlide, totalSlides, width)

	var sb strings.Builder
	sb.WriteString(body)
	if notesOverlay != "" {
		sb.WriteString("\n")
		sb.WriteString(notesOverlay)
	}
	if !e.ZenMode {
		sb.WriteString("\n")
		sb.WriteString(status)
	}
	sb.WriteString("\n")
	sb.WriteString(progressLine)

	return sb.String()
}

func navStatus(d Deck, e Editor, w int) string {
	align := d.Align
	if align == "" {
		align = AlignCenter
	}
	if e.SlideIdx < len(d.Slides) && d.Slides[e.SlideIdx].Align != "" {
		align = d.Slides[e.SlideIdx].Align
	}

	visibleCount := 0
	hasNotes := false
	if e.SlideIdx < len(d.Slides) {
		visibleCount = len(d.Slides[e.SlideIdx].VisibleBlockIndices())
		hasNotes = d.Slides[e.SlideIdx].Notes() != ""
	}

	totalSlides := len(d.Slides)
	curSlide := e.SlideIdx + 1

	left := fmt.Sprintf("slide %d/%d (%s)  ·  %s  ·  blocks %d", curSlide, totalSlides, align, currentTheme.Name, visibleCount)

	if hasNotes {
		if e.ShowNotes {
			left += "  ·  [n: notes open]"
		} else {
			left += "  ·  [n: notes]"
		}
	}
	if e.ShowLineNumbers {
		left += "  ·  [L: lines]"
	}
	if e.WatchMode {
		left += "  ·  [watch]"
	}
	if e.ShowTimer && !e.TimerStart.IsZero() {
		elapsed := int(time.Since(e.TimerStart).Seconds())
		if elapsed < 0 {
			elapsed = 0
		}
		mins := elapsed / 60
		secs := elapsed % 60
		timerStr := fmt.Sprintf("⏱ %02d:%02d", mins, secs)
		if mins >= 60 {
			hours := mins / 60
			mins = mins % 60
			timerStr = fmt.Sprintf("⏱ %d:%02d:%02d", hours, mins, secs)
		}
		left += "  ·  [" + timerStr + "]"
	}
	if e.Dirty {
		left += "  ·  [modified]"
	}
	if e.Message != "" {
		left += "  ·  " + e.Message
	}
	right := "? help · / jump · o grid · y yank · E export · b blank · c timer · r reload · L lines · z zen · x task · tab align · t theme · n notes · i edit · ^n add · ^d del · ^s save · u undo · q quit"
	status := left + "  ·  " + right
	return dimStyle.Width(w).Render(status)
}

func editStatus(e Editor, w int) string {
	left := fmt.Sprintf("editing  ·  %s  ·  col %d/%d", currentTheme.Name, e.CursorCol, len(e.Draft))
	right := "^t theme · esc cancel · enter confirm"
	status := left + "  ·  " + right
	if e.Message != "" {
		status = e.Message + "  ·  " + status
	}
	return messageStyle.Width(w).Render(status)
}
