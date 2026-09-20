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
	boldStyle       = lipgloss.NewStyle().Bold(true)
	italicStyle     = lipgloss.NewStyle().Italic(true)
	codeSpanStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Background(lipgloss.Color("236")).Padding(0, 1)
	codeBlockStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1)
	imageStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	dimStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	dimBoldStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Bold(true)
	cursorStyle     = lipgloss.NewStyle().Background(lipgloss.Color("240")).Foreground(lipgloss.Color("0"))
	editStyle       = lipgloss.NewStyle().Background(lipgloss.Color("22")).Foreground(lipgloss.Color("252")).Padding(0, 1)
	messageStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	notesBoxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Foreground(lipgloss.Color("252")).Padding(0, 1)
	notesTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))

	tableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Padding(0, 1)
	tableCellStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Padding(0, 1)
	tableBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	helpBoxStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("212")).Foreground(lipgloss.Color("252")).Padding(1, 2)
	helpTitleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	helpHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	helpKeyStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("114"))
	helpDescStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	progressStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	timerRunningStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("114"))
	timerPausedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))

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
		styledSpans[i] = codeSpanStyle.Render(inner)
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
	}

	return b.String()
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
		sb.WriteString(tableHeaderStyle.Render(padded))
		sb.WriteString(tableBorderStyle.Render("│"))
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
			sb.WriteString(tableCellStyle.Render(padded))
			sb.WriteString(tableBorderStyle.Render("│"))
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
		rendered := renderBlock(blk, w, maxBlockH, baseDir, isCursor, isEditing, editDraft, e.CursorCol)
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
	title := notesTitleStyle.Render("📝 Speaker Notes") + dimStyle.Render(" (press 'n' to hide)")
	boxW := width - 4
	if boxW < 20 {
		boxW = width
	}
	inner := title + "\n" + content
	return notesBoxStyle.Width(boxW).MaxHeight(maxHeight).Render(inner)
}

func renderHelpModal(w, h int) string {
	boxW := 62
	if boxW > w-4 {
		boxW = w - 4
	}
	if boxW < 36 {
		boxW = 36
	}

	var sb strings.Builder
	sb.WriteString(helpTitleStyle.Render("Termdeck Keyboard Controls"))
	sb.WriteString("\n" + dimStyle.Render("Press '?' or 'Esc' to close") + "\n\n")

	sb.WriteString(helpHeaderStyle.Render("  NAVIGATION") + "\n")
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("→, l, Space, Enter"), helpDescStyle.Render("Next slide")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("←, h, Backspace"), helpDescStyle.Render("Previous slide")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("↓, j"), helpDescStyle.Render("Move laser pointer down")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("↑, k"), helpDescStyle.Render("Move laser pointer up")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("g / G"), helpDescStyle.Render("First / Last slide")))

	sb.WriteString("\n" + helpHeaderStyle.Render("  PRESENTATION") + "\n")
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("n"), helpDescStyle.Render("Toggle speaker notes overlay")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("t"), helpDescStyle.Render("Start / Pause elapsed timer")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("ctrl+t"), helpDescStyle.Render("Reset timer to 00:00")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("Tab / ctrl+a"), helpDescStyle.Render("Cycle alignment (left/center/right)")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("p"), helpDescStyle.Render("Open image in system viewer")))

	sb.WriteString("\n" + helpHeaderStyle.Render("  LIVE EDITOR") + "\n")
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("i"), helpDescStyle.Render("Edit focused block")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("Enter"), helpDescStyle.Render("Confirm edit & auto-save")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("Esc"), helpDescStyle.Render("Cancel edit / Close help")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("ctrl+n / ctrl+d"), helpDescStyle.Render("Add / Delete block")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("ctrl+k / ctrl+j"), helpDescStyle.Render("Move block up / down")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("u / ctrl+r"), helpDescStyle.Render("Undo / Redo")))
	sb.WriteString(fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("ctrl+s"), helpDescStyle.Render("Save file manually")))

	sb.WriteString("\n" + fmt.Sprintf("  %-22s %s\n", helpKeyStyle.Render("q / ctrl+c"), helpDescStyle.Render("Quit (auto-saves changes)")))

	box := helpBoxStyle.Width(boxW).Render(sb.String())
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}

// --- Full view ---

func View(d Deck, e Editor, width, height int) string {
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	if e.ShowHelp {
		return renderHelpModal(width, height)
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

	bodyHeight := height - 2
	if notesHeight > 0 {
		bodyHeight = height - 2 - notesHeight
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
	if e.Mode == ModeEdit {
		status = editStatus(e, width)
	} else {
		status = navStatus(d, e, width)
	}

	if notesOverlay != "" {
		return body + "\n" + notesOverlay + "\n" + status
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

	visibleCount := 0
	hasNotes := false
	if e.SlideIdx < len(d.Slides) {
		visibleCount = len(d.Slides[e.SlideIdx].VisibleBlockIndices())
		hasNotes = d.Slides[e.SlideIdx].Notes() != ""
	}

	totalSlides := len(d.Slides)
	curSlide := e.SlideIdx + 1
	pct := 0
	if totalSlides > 0 {
		pct = int(float64(curSlide) / float64(totalSlides) * 100)
	}

	// Visual progress bar track
	barLen := 8
	filled := 0
	if totalSlides > 0 {
		filled = (curSlide * barLen) / totalSlides
	}
	if filled > barLen {
		filled = barLen
	}
	progressTrack := fmt.Sprintf("[%s%s] %d%%", strings.Repeat("█", filled), strings.Repeat("░", barLen-filled), pct)

	timerStr := ""
	if e.TimerRunning || e.ElapsedTime() > 0 {
		timerIcon := "⏱"
		style := timerRunningStyle
		if !e.TimerRunning {
			timerIcon = "⏸"
			style = timerPausedStyle
		}
		timerStr = "  ·  " + style.Render(fmt.Sprintf("[%s %s]", timerIcon, e.FormatTimer()))
	}

	left := fmt.Sprintf("slide %d/%d %s (%s)  ·  blocks %d", curSlide, totalSlides, progressStyle.Render(progressTrack), align, visibleCount)
	left += timerStr

	if hasNotes {
		if e.ShowNotes {
			left += "  ·  [n: notes open]"
		} else {
			left += "  ·  [n: notes]"
		}
	}
	if e.Dirty {
		left += "  ·  [modified]"
	}
	if e.Message != "" {
		left += "  ·  " + e.Message
	}
	right := "? help · tab align · n notes · t timer · i edit · ^n add · ^d del · ^s save · u undo · q quit"
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
