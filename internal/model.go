package internal

import (
	"fmt"
	"regexp"
	"strings"
)

var reMarkdownImage = regexp.MustCompile(`^!\[(.*?)\]\((.*?)\)$`)

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
)

type Block struct {
	Kind      BlockKind
	Level     int
	Text      string
	Lang      string
	Lines     []string
	Src       string
	Directive string
	Raw       string
}

// --- Slide & Deck ---

type AlignKind string

const (
	AlignLeft   AlignKind = "left"
	AlignCenter AlignKind = "center"
	AlignRight  AlignKind = "right"
)

type Slide struct {
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

type Deck struct {
	Meta    map[string]string
	Slides  []Slide
	BaseDir string
	Align   AlignKind
	Theme   string
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

func parseSlide(lines []string) Slide {
	var blocks []Block
	var slideAlign AlignKind
	inCode := false
	codeLang := ""
	var codeLines []string

	inNotes := false
	var noteLines []string

	inTable := false
	var tableLines []string

	flushCode := func() {
		blocks = append(blocks, Block{
			Kind:  BlockCode,
			Lang:  codeLang,
			Lines: codeLines,
		})
		codeLines = nil
		codeLang = ""
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

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

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
			inCode = true
			codeLang = strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
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

			if strings.HasPrefix(trimmed, "::code") {
				inCode = true
				_, val := ParseDirective(trimmed)
				codeLang = val
				continue
			}

			if strings.HasPrefix(trimmed, "::notes") {
				inNotes = true
				continue
			}

			key, val := ParseDirective(trimmed)
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
			continue
		}

		// 2. Markdown Table Detection: line starts with '|', ends with '|', has at least 2 '|'
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

	return Slide{Blocks: blocks, Align: slideAlign}
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
		fmt.Fprintf(&b, "::code lang=%s\n", blk.Lang)
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
	default:
		return blk.Text
	}
}
