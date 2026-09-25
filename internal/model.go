package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	reMarkdownImage = regexp.MustCompile(`^!\[(.*?)\]\((.*?)\)$`)
	reMarkdownLink  = regexp.MustCompile(`^\[(.*?)\]\((.*?)\)$`)
)

// --- Block types ---

type BlockKind int

const (
	BlockHeading BlockKind = iota
	BlockParagraph
	BlockCode
	BlockImage
	BlockDirective
	BlockList
	BlockTable
	BlockCallout
	BlockDivider
	BlockBranch
	BlockColumns
)

type Block struct {
	Kind         BlockKind
	Level        int
	Text         string
	Lang         string
	Lines        []string
	Src          string
	Directive    string
	Raw          string
	Callout      string
	BranchKey    string
	BranchTarget string
	NoEval       bool
	Columns      [][]Block
}

// --- Slide & Deck ---

type AlignKind string

const (
	AlignLeft   AlignKind = "left"
	AlignCenter AlignKind = "center"
	AlignRight  AlignKind = "right"
)

type Branch struct {
	Key    string
	Label  string
	Target string
}

type Slide struct {
	ID     string
	NextID string
	PrevID string
	Tags   []string
	Blocks []Block
	Align  AlignKind
}

func (s *Slide) VisibleBlockIndices() []int {
	var indices []int
	for i, b := range s.Blocks {
		if b.Kind == BlockDirective && strings.HasPrefix(b.Directive, "::notes") {
			continue
		}
		indices = append(indices, i)
	}
	return indices
}

func (s *Slide) Notes() string {
	for _, b := range s.Blocks {
		if b.Kind == BlockDirective && strings.HasPrefix(b.Directive, "::notes") {
			if len(b.Lines) > 0 {
				return strings.Join(b.Lines, "\n")
			}
			return b.Text
		}
	}
	return ""
}

func (s *Slide) Title() string {
	for _, b := range s.Blocks {
		if b.Kind == BlockHeading && strings.TrimSpace(b.Text) != "" {
			return strings.TrimSpace(b.Text)
		}
	}
	for _, b := range s.Blocks {
		if b.Kind == BlockParagraph && strings.TrimSpace(b.Text) != "" {
			txt := strings.TrimSpace(b.Text)
			if len(txt) > 30 {
				return txt[:30] + "..."
			}
			return txt
		}
	}
	return "Slide"
}

func (s *Slide) Branches() []Branch {
	var list []Branch
	for _, b := range s.Blocks {
		if b.Kind == BlockBranch {
			list = append(list, Branch{
				Key:    b.BranchKey,
				Label:  b.Text,
				Target: b.BranchTarget,
			})
		} else if b.Kind == BlockColumns {
			for _, col := range b.Columns {
				for _, inner := range col {
					if inner.Kind == BlockBranch {
						list = append(list, Branch{
							Key:    inner.BranchKey,
							Label:  inner.Text,
							Target: inner.BranchTarget,
						})
					}
				}
			}
		}
	}
	return list
}

func (s *Slide) Slug() string {
	if s.ID != "" {
		return s.ID
	}
	t := strings.ToLower(strings.TrimSpace(s.Title()))
	t = strings.ReplaceAll(t, " ", "-")
	var clean []rune
	for _, r := range t {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			clean = append(clean, r)
		}
	}
	slug := string(clean)
	if slug == "" {
		slug = "slide"
	}
	return slug
}

func (s *Slide) HasTag(tag string) bool {
	if tag == "" || strings.EqualFold(tag, "all") {
		return true
	}
	tagLower := strings.ToLower(strings.TrimSpace(tag))
	for _, t := range s.Tags {
		if strings.ToLower(strings.TrimSpace(t)) == tagLower {
			return true
		}
	}
	return false
}

func (s *Slide) Summary() string {
	var kinds []string
	hasCode := false
	hasImage := false
	hasTable := false
	hasCallout := false
	hasTask := false
	hasBranch := false
	hasColumns := false

	for _, b := range s.Blocks {
		switch b.Kind {
		case BlockCode:
			hasCode = true
		case BlockImage:
			hasImage = true
		case BlockTable:
			hasTable = true
		case BlockCallout:
			hasCallout = true
		case BlockBranch:
			hasBranch = true
		case BlockColumns:
			hasColumns = true
		case BlockList:
			trimmed := strings.TrimSpace(b.Text)
			if strings.HasPrefix(trimmed, "- [ ]") || strings.HasPrefix(trimmed, "- [x]") {
				hasTask = true
			}
		}
	}

	visCount := len(s.VisibleBlockIndices())
	summary := fmt.Sprintf("%d blk", visCount)
	if visCount != 1 {
		summary += "s"
	}

	if hasCode {
		kinds = append(kinds, "code")
	}
	if hasTable {
		kinds = append(kinds, "table")
	}
	if hasCallout {
		kinds = append(kinds, "card")
	}
	if hasImage {
		kinds = append(kinds, "img")
	}
	if hasTask {
		kinds = append(kinds, "task")
	}
	if hasBranch {
		kinds = append(kinds, "fork")
	}
	if hasColumns {
		kinds = append(kinds, "cols")
	}
	if len(kinds) > 0 {
		summary += " · " + strings.Join(kinds, ",")
	}
	return summary
}

func (s *Slide) FindBranchByKey(key string) *Branch {
	keyLower := strings.ToLower(strings.TrimSpace(key))
	for _, b := range s.Branches() {
		if strings.ToLower(b.Key) == keyLower {
			return &b
		}
	}
	return nil
}

type Deck struct {
	Meta    map[string]string
	Slides  []Slide
	BaseDir string
	Align   AlignKind
	Theme   string
}

func (d *Deck) AllTags() []string {
	seen := make(map[string]bool)
	var tags []string
	for _, s := range d.Slides {
		for _, t := range s.Tags {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				lower := strings.ToLower(trimmed)
				if !seen[lower] {
					seen[lower] = true
					tags = append(tags, trimmed)
				}
			}
		}
	}
	return tags
}

func (d *Deck) SlideIndicesForTag(tag string) []int {
	var indices []int
	for i, s := range d.Slides {
		if s.HasTag(tag) {
			indices = append(indices, i)
		}
	}
	return indices
}

func (d *Deck) FindSlideByID(target string) int {
	target = strings.TrimSpace(target)
	if target == "" {
		return -1
	}

	targetLower := strings.ToLower(target)

	// 1. Exact match on Slide.ID (case-insensitive)
	for i := range d.Slides {
		if d.Slides[i].ID != "" && strings.ToLower(d.Slides[i].ID) == targetLower {
			return i
		}
	}

	// 2. Exact match on Slide.Slug() (case-insensitive)
	for i := range d.Slides {
		if strings.ToLower(d.Slides[i].Slug()) == targetLower {
			return i
		}
	}

	// 3. Numeric 1-based index (e.g. "3" -> index 2)
	var num int
	if n, err := fmt.Sscanf(target, "%d", &num); err == nil && n == 1 {
		if num >= 1 && num <= len(d.Slides) {
			return num - 1
		}
	}

	// 4. Case-insensitive substring match on Slide.Title()
	for i := range d.Slides {
		if strings.Contains(strings.ToLower(d.Slides[i].Title()), targetLower) {
			return i
		}
	}

	return -1
}

func (d *Deck) HasBranches() bool {
	for i := range d.Slides {
		if len(d.Slides[i].Branches()) > 0 || d.Slides[i].NextID != "" {
			return true
		}
	}
	return false
}

// --- Parsing ---

func ParseDeck(src string) Deck {
	if strings.TrimSpace(src) == "" {
		return Deck{Meta: map[string]string{}, Align: AlignCenter, Theme: "termdeck"}
	}
	lines := strings.Split(src, "\n")
	meta := map[string]string{}

	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			trimmed := strings.TrimSpace(lines[i])
			if trimmed == "---" {
				lines = lines[i+1:]
				break
			}
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])
				val := strings.TrimSpace(trimmed[idx+1:])
				meta[key] = val
			}
		}
	}

	var slideGroups [][]string
	var cur []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" && len(cur) > 0 {
			slideGroups = append(slideGroups, cur)
			cur = nil
			continue
		}
		cur = append(cur, line)
	}
	if len(cur) > 0 {
		slideGroups = append(slideGroups, cur)
	}

	var slides []Slide
	for _, group := range slideGroups {
		slides = append(slides, parseSlide(group))
	}

	deckAlign := AlignCenter
	if a, ok := meta["align"]; ok {
		switch strings.ToLower(strings.TrimSpace(a)) {
		case "left":
			deckAlign = AlignLeft
		case "right":
			deckAlign = AlignRight
		case "center":
			deckAlign = AlignCenter
		}
	}

	deckTheme := "termdeck"
	if th, ok := meta["theme"]; ok {
		deckTheme = strings.ToLower(strings.TrimSpace(th))
	}

	return Deck{Meta: meta, Slides: slides, Align: deckAlign, Theme: deckTheme}
}

func parseCodeLangAndFlags(s string) (string, bool) {
	s = strings.TrimSpace(s)
	noEval := false
	lower := strings.ToLower(s)
	if strings.Contains(lower, "no-eval") || strings.Contains(lower, "eval=false") || strings.Contains(lower, "run=false") || strings.Contains(lower, "no_run") || strings.Contains(lower, "ignore") || strings.Contains(lower, "noexec") {
		noEval = true
	}

	lang := s
	if strings.HasPrefix(lower, "lang=") {
		parts := strings.Fields(s)
		lang = strings.TrimPrefix(parts[0], "lang=")
		lang = strings.TrimPrefix(lang, "LANG=")
	} else {
		parts := strings.Fields(s)
		if len(parts) > 0 {
			lang = parts[0]
		}
	}
	lang = strings.Trim(lang, `"',;`)
	return lang, noEval
}

func parseSlide(lines []string) Slide {
	var blocks []Block
	var slideAlign AlignKind
	var slideID string
	var nextID string
	var prevID string
	var tags []string
	branchCount := 0

	inCode := false
	codeLang := ""
	codeNoEval := false
	var codeLines []string

	inNotes := false
	var noteLines []string

	inTable := false
	var tableLines []string

	inCallout := false
	var calloutLines []string

	flushCode := func() {
		blocks = append(blocks, Block{
			Kind:   BlockCode,
			Lang:   codeLang,
			Lines:  codeLines,
			Text:   strings.Join(codeLines, "\n"),
			NoEval: codeNoEval,
		})
		codeLines = nil
		codeLang = ""
		codeNoEval = false
	}

	flushNotes := func() {
		blocks = append(blocks, Block{
			Kind:      BlockDirective,
			Directive: "::notes",
			Lines:     noteLines,
			Text:      strings.Join(noteLines, "\n"),
		})
		noteLines = nil
	}

	flushTable := func() {
		if len(tableLines) > 0 {
			blocks = append(blocks, Block{
				Kind:  BlockTable,
				Lines: tableLines,
				Text:  strings.Join(tableLines, "\n"),
			})
			tableLines = nil
		}
	}

	flushCallout := func() {
		if len(calloutLines) == 0 {
			return
		}
		first := strings.TrimSpace(calloutLines[0])
		first = strings.TrimPrefix(first, ">")
		first = strings.TrimSpace(first)

		calloutType := "quote"
		var bodyLines []string

		upperFirst := strings.ToUpper(first)
		switch {
		case strings.HasPrefix(upperFirst, "[!TIP]"):
			calloutType = "tip"
			rest := strings.TrimSpace(first[6:])
			if rest != "" {
				bodyLines = append(bodyLines, rest)
			}
		case strings.HasPrefix(upperFirst, "[!NOTE]"):
			calloutType = "note"
			rest := strings.TrimSpace(first[7:])
			if rest != "" {
				bodyLines = append(bodyLines, rest)
			}
		case strings.HasPrefix(upperFirst, "[!WARNING]"):
			calloutType = "warning"
			rest := strings.TrimSpace(first[10:])
			if rest != "" {
				bodyLines = append(bodyLines, rest)
			}
		case strings.HasPrefix(upperFirst, "[!IMPORTANT]"):
			calloutType = "important"
			rest := strings.TrimSpace(first[12:])
			if rest != "" {
				bodyLines = append(bodyLines, rest)
			}
		case strings.HasPrefix(upperFirst, "[!CAUTION]"):
			calloutType = "caution"
			rest := strings.TrimSpace(first[10:])
			if rest != "" {
				bodyLines = append(bodyLines, rest)
			}
		default:
			bodyLines = append(bodyLines, first)
		}

		for _, l := range calloutLines[1:] {
			cleaned := strings.TrimSpace(l)
			cleaned = strings.TrimPrefix(cleaned, ">")
			cleaned = strings.TrimPrefix(cleaned, " ")
			bodyLines = append(bodyLines, cleaned)
		}

		blocks = append(blocks, Block{
			Kind:    BlockCallout,
			Callout: calloutType,
			Lines:   bodyLines,
			Text:    strings.Join(bodyLines, "\n"),
		})
		calloutLines = nil
	}

	inColumns := false
	var colBlocks [][]Block
	var currentColLines []string
	hasColStarted := false

	flushColumns := func() {
		if hasColStarted && len(currentColLines) > 0 {
			colBlocks = append(colBlocks, parseColumnBlocks(currentColLines))
			currentColLines = nil
		}
		if len(colBlocks) > 0 {
			blocks = append(blocks, Block{
				Kind:    BlockColumns,
				Columns: colBlocks,
			})
			colBlocks = nil
		}
		inColumns = false
		hasColStarted = false
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Column parsing mode
		if inColumns {
			if trimmed == ":::" || trimmed == ":::end" || trimmed == "::end" || trimmed == "::columns end" {
				flushColumns()
				continue
			}
			if trimmed == ":::col" || trimmed == "::col" || strings.HasPrefix(trimmed, ":::col ") || strings.HasPrefix(trimmed, "::col ") {
				if hasColStarted && len(currentColLines) > 0 {
					colBlocks = append(colBlocks, parseColumnBlocks(currentColLines))
					currentColLines = nil
				}
				hasColStarted = true
				continue
			}
			currentColLines = append(currentColLines, line)
			continue
		}

		if strings.HasPrefix(trimmed, ":::columns") || strings.HasPrefix(trimmed, "::columns") || strings.HasPrefix(trimmed, ":::split") || strings.HasPrefix(trimmed, "::split") {
			if inCode {
				flushCode()
				inCode = false
			}
			if inNotes {
				flushNotes()
				inNotes = false
			}
			if inTable {
				flushTable()
				inTable = false
			}
			if inCallout {
				flushCallout()
				inCallout = false
			}
			inColumns = true
			hasColStarted = true
			colBlocks = nil
			currentColLines = nil
			continue
		}

		// 1. Standard markdown code fences (```)
		if strings.HasPrefix(trimmed, "```") {
			if inCode {
				flushCode()
				inCode = false
				continue
			}
			if inNotes {
				flushNotes()
				inNotes = false
			}
			if inTable {
				flushTable()
				inTable = false
			}
			if inCallout {
				flushCallout()
				inCallout = false
			}
			inCode = true
			codeLang, codeNoEval = parseCodeLangAndFlags(strings.TrimPrefix(trimmed, "```"))
			continue
		}

		if strings.HasPrefix(trimmed, "::") {
			if inCode {
				flushCode()
				inCode = false
				if strings.HasPrefix(trimmed, "::code") && trimmed == "::code" {
					continue
				}
			}
			if inNotes {
				flushNotes()
				inNotes = false
				if strings.HasPrefix(trimmed, "::notes") && trimmed == "::notes" {
					continue
				}
			}
			if inTable {
				flushTable()
				inTable = false
			}
			if inCallout {
				flushCallout()
				inCallout = false
			}

			if strings.HasPrefix(trimmed, "::code") {
				inCode = true
				_, val := ParseDirective(trimmed)
				codeLang, codeNoEval = parseCodeLangAndFlags(val)
				continue
			}

			if strings.HasPrefix(trimmed, "::notes") {
				inNotes = true
				continue
			}

			key, val := ParseDirective(trimmed)
			if key == "id" || key == "slug" {
				slideID = strings.TrimSpace(val)
				continue
			} else if key == "next" {
				nextID = strings.TrimSpace(val)
				continue
			} else if key == "prev" {
				prevID = strings.TrimSpace(val)
				continue
			} else if key == "tags" {
				for _, t := range strings.Split(val, ",") {
					if s := strings.TrimSpace(t); s != "" {
						tags = append(tags, s)
					}
				}
				continue
			} else if key == "branch" || key == "fork" {
				if k, label, target, ok := parseBranchLine(line); ok {
					if k == "" {
						branchCount++
						k = strconv.Itoa(branchCount)
					}
					blocks = append(blocks, Block{
						Kind:         BlockBranch,
						Text:         label,
						BranchKey:    k,
						BranchTarget: target,
						Raw:          line,
					})
					continue
				}
			}

			if key == "align" {
				val = strings.ToLower(strings.TrimSpace(val))
				if val == "left" || val == "center" || val == "right" {
					slideAlign = AlignKind(val)
				}
				continue
			} else if key == "left" || key == "center" || key == "right" {
				slideAlign = AlignKind(key)
				continue
			}

			if key == "image" {
				blocks = append(blocks, Block{Kind: BlockImage, Src: val})
			} else if key == "hr" {
				blocks = append(blocks, Block{Kind: BlockDivider, Raw: line})
			} else {
				blocks = append(blocks, Block{Kind: BlockDirective, Directive: trimmed, Raw: line})
			}
			continue
		}

		if inCode {
			codeLines = append(codeLines, line)
			continue
		}

		if inNotes {
			noteLines = append(noteLines, line)
			continue
		}

		if trimmed == "" {
			if inTable {
				flushTable()
				inTable = false
			}
			if inCallout {
				flushCallout()
				inCallout = false
			}
			continue
		}

		// 2. Callout / Blockquote: line starts with '>'
		if strings.HasPrefix(trimmed, ">") {
			if inTable {
				flushTable()
				inTable = false
			}
			inCallout = true
			calloutLines = append(calloutLines, trimmed)
			continue
		} else if inCallout {
			flushCallout()
			inCallout = false
		}

		// 3. Horizontal divider: "***", "___", or "::hr"
		if trimmed == "***" || trimmed == "___" || trimmed == "::hr" {
			if inTable {
				flushTable()
				inTable = false
			}
			if inCallout {
				flushCallout()
				inCallout = false
			}
			blocks = append(blocks, Block{Kind: BlockDivider, Raw: line})
			continue
		}

		// 4. Markdown Table Detection: line starts with '|', ends with '|', has at least 2 '|'
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") && strings.Count(trimmed, "|") >= 2 {
			inTable = true
			tableLines = append(tableLines, trimmed)
			continue
		} else if inTable {
			flushTable()
			inTable = false
		}

		// 3. Headings (#)
		if strings.HasPrefix(trimmed, "#") {
			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
				} else {
					break
				}
			}
			text := strings.TrimSpace(trimmed[level:])
			if idx := strings.Index(text, "{#"); idx != -1 && strings.HasSuffix(text, "}") {
				slugPart := strings.TrimSuffix(text[idx+2:], "}")
				slugPart = strings.TrimSpace(slugPart)
				if slideID == "" && slugPart != "" {
					slideID = slugPart
				}
				text = strings.TrimSpace(text[:idx])
			}
			blocks = append(blocks, Block{Kind: BlockHeading, Level: level, Text: text, Raw: line})
			continue
		}

		// 4. Standard Markdown Image: ![alt](path)
		if m := reMarkdownImage.FindStringSubmatch(trimmed); len(m) == 3 {
			blocks = append(blocks, Block{Kind: BlockImage, Src: strings.TrimSpace(m[2]), Text: strings.TrimSpace(m[1]), Raw: line})
			continue
		}

		// 5. Lists
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			blocks = append(blocks, Block{Kind: BlockList, Text: trimmed, Raw: line})
			continue
		}
		if len(trimmed) > 2 && trimmed[0] >= '0' && trimmed[0] <= '9' && trimmed[1] == '.' {
			blocks = append(blocks, Block{Kind: BlockList, Text: trimmed, Raw: line})
			continue
		}

		// 6. Branch / Fork links: "-> [1] Deep Dive -> arch" or "-> [Deep Dive](arch)"
		if k, label, target, ok := parseBranchLine(line); ok {
			if k == "" {
				branchCount++
				k = strconv.Itoa(branchCount)
			}
			blocks = append(blocks, Block{
				Kind:         BlockBranch,
				Text:         label,
				BranchKey:    k,
				BranchTarget: target,
				Raw:          line,
			})
			continue
		}

		blocks = append(blocks, Block{Kind: BlockParagraph, Text: trimmed, Raw: line})
	}

	if inCode {
		flushCode()
	}
	if inNotes {
		flushNotes()
	}
	if inTable {
		flushTable()
	}
	if inCallout {
		flushCallout()
	}
	if inColumns {
		flushColumns()
	}

	return Slide{
		ID:     slideID,
		NextID: nextID,
		PrevID: prevID,
		Tags:   tags,
		Blocks: blocks,
		Align:  slideAlign,
	}
}

func parseColumnBlocks(lines []string) []Block {
	slide := parseSlide(lines)
	return slide.Blocks
}

func ParseDirective(line string) (key, value string) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "::") {
		return "", ""
	}
	content := trimmed[2:]
	if idx := strings.IndexAny(content, " ="); idx != -1 {
		key = content[:idx]
		rest := strings.TrimSpace(content[idx:])
		if eq := strings.Index(rest, "="); eq != -1 {
			value = strings.TrimSpace(rest[eq+1:])
		} else {
			value = rest
		}
	} else {
		key = content
	}
	return key, value
}

func parseBranchLine(line string) (key, label, target string, ok bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return "", "", "", false
	}

	if strings.HasPrefix(trimmed, "::branch") || strings.HasPrefix(trimmed, "::fork") {
		rest := ""
		if strings.HasPrefix(trimmed, "::branch") {
			rest = strings.TrimSpace(trimmed[len("::branch"):])
		} else {
			rest = strings.TrimSpace(trimmed[len("::fork"):])
		}
		return parseBranchContent(rest)
	}

	if strings.HasPrefix(trimmed, "->") || strings.HasPrefix(trimmed, "=>") {
		rest := strings.TrimSpace(trimmed[2:])
		return parseBranchContent(rest)
	}

	if strings.HasPrefix(trimmed, "↳") || strings.HasPrefix(trimmed, "↪") {
		r := []rune(trimmed)
		rest := strings.TrimSpace(string(r[1:]))
		return parseBranchContent(rest)
	}

	if strings.HasPrefix(trimmed, "[") {
		closeIdx := strings.Index(trimmed, "]")
		if closeIdx > 1 && len(trimmed) > closeIdx+1 {
			k := trimmed[1:closeIdx]
			after := strings.TrimSpace(trimmed[closeIdx+1:])
			// [1] -> label -> target
			if strings.HasPrefix(after, "->") || strings.HasPrefix(after, "=>") {
				rest := strings.TrimSpace(after[2:])
				_, l, tgt, valid := parseBranchContent(rest)
				if valid {
					return k, l, tgt, true
				}
			}
			// [1] label -> target or [1] [label](target)
			if strings.Contains(after, "->") || strings.Contains(after, "=>") || (strings.HasPrefix(after, "[") && strings.Contains(after, "](")) {
				_, l, tgt, valid := parseBranchContent(after)
				if valid {
					return k, l, tgt, true
				}
			}
		}
	}

	return "", "", "", false
}

func parseBranchContent(content string) (key, label, target string, ok bool) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", "", "", false
	}

	// Extract [key] prefix if present: e.g. "[1] Deep Dive -> arch"
	if strings.HasPrefix(content, "[") {
		closeBracket := strings.Index(content, "]")
		if closeBracket != -1 {
			afterBracket := strings.TrimSpace(content[closeBracket+1:])
			if !strings.HasPrefix(afterBracket, "(") {
				key = strings.TrimSpace(content[1:closeBracket])
				content = afterBracket
			}
		}
	}

	// Check for markdown link: [label](target)
	if m := reMarkdownLink.FindStringSubmatch(content); len(m) == 3 {
		label = strings.TrimSpace(m[1])
		target = strings.TrimSpace(m[2])
		return key, label, target, true
	}

	// Check for "->" or "=>" separator: label -> target
	sep := ""
	if strings.Contains(content, "->") {
		sep = "->"
	} else if strings.Contains(content, "=>") {
		sep = "=>"
	}

	if sep != "" {
		parts := strings.SplitN(content, sep, 2)
		label = strings.TrimSpace(parts[0])
		target = strings.TrimSpace(parts[1])
		target = strings.TrimPrefix(target, "(")
		target = strings.TrimSuffix(target, ")")
		target = strings.TrimSpace(target)
		if label == "" {
			label = target
		}
		return key, label, target, true
	}

	// If single token without spaces, treat as target
	if content != "" && !strings.Contains(content, " ") {
		return key, content, content, true
	}

	return "", "", "", false
}

// --- Serialization ---

func SerializeDeck(d Deck) string {
	var b strings.Builder

	meta := make(map[string]string)
	for k, v := range d.Meta {
		meta[k] = v
	}
	if d.Align != "" && d.Align != AlignCenter {
		meta["align"] = string(d.Align)
	}
	if d.Theme != "" && d.Theme != "termdeck" {
		meta["theme"] = d.Theme
	}

	if len(meta) > 0 {
		b.WriteString("---\n")
		for k, v := range meta {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
		b.WriteString("---\n\n")
	}

	for i, slide := range d.Slides {
		if i > 0 {
			b.WriteString("\n---\n\n")
		}
		if slide.ID != "" {
			fmt.Fprintf(&b, "::id %s\n", slide.ID)
		}
		if slide.NextID != "" {
			fmt.Fprintf(&b, "::next %s\n", slide.NextID)
		}
		if slide.PrevID != "" {
			fmt.Fprintf(&b, "::prev %s\n", slide.PrevID)
		}
		if len(slide.Tags) > 0 {
			fmt.Fprintf(&b, "::tags %s\n", strings.Join(slide.Tags, ", "))
		}
		if slide.Align != "" && slide.Align != d.Align {
			fmt.Fprintf(&b, "::align %s\n", slide.Align)
		}
		for j, block := range slide.Blocks {
			if j > 0 {
				b.WriteString("\n")
			}
			b.WriteString(SerializeBlock(block))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func SerializeBlock(blk Block) string {
	switch blk.Kind {
	case BlockHeading:
		prefix := strings.Repeat("#", blk.Level)
		return fmt.Sprintf("%s %s", prefix, blk.Text)
	case BlockParagraph:
		return blk.Text
	case BlockCode:
		var b strings.Builder
		if blk.NoEval {
			fmt.Fprintf(&b, "::code lang=%s eval=false\n", blk.Lang)
		} else {
			fmt.Fprintf(&b, "::code lang=%s\n", blk.Lang)
		}
		for _, line := range blk.Lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
		return strings.TrimRight(b.String(), "\n")
	case BlockImage:
		return fmt.Sprintf("::image %s", blk.Src)
	case BlockDirective:
		if strings.HasPrefix(blk.Directive, "::notes") {
			var b strings.Builder
			b.WriteString("::notes\n")
			if len(blk.Lines) > 0 {
				for _, l := range blk.Lines {
					b.WriteString(l)
					b.WriteString("\n")
				}
			} else if blk.Text != "" {
				b.WriteString(blk.Text)
				b.WriteString("\n")
			}
			return strings.TrimRight(b.String(), "\n")
		}
		return blk.Directive
	case BlockList:
		return blk.Text
	case BlockTable:
		if len(blk.Lines) > 0 {
			return strings.Join(blk.Lines, "\n")
		}
		return blk.Text
	case BlockCallout:
		var b strings.Builder
		if blk.Callout != "" && blk.Callout != "quote" {
			fmt.Fprintf(&b, "> [!%s]\n", strings.ToUpper(blk.Callout))
		}
		lines := blk.Lines
		if len(lines) == 0 && blk.Text != "" {
			lines = strings.Split(blk.Text, "\n")
		}
		for i, line := range lines {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString("> " + line)
		}
		return b.String()
	case BlockDivider:
		return "***"
	case BlockBranch:
		if blk.BranchKey != "" {
			return fmt.Sprintf("::branch [%s] %s -> %s", blk.BranchKey, blk.Text, blk.BranchTarget)
		}
		return fmt.Sprintf("::branch %s -> %s", blk.Text, blk.BranchTarget)
	case BlockColumns:
		var b strings.Builder
		b.WriteString(":::columns\n")
		for _, col := range blk.Columns {
			b.WriteString(":::col\n")
			for j, inner := range col {
				if j > 0 {
					b.WriteString("\n")
				}
				b.WriteString(SerializeBlock(inner))
				b.WriteString("\n")
			}
		}
		b.WriteString(":::")
		return b.String()
	default:
		return blk.Text
	}
}
