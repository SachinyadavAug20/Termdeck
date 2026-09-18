package internal

import (
	"fmt"
	"strings"
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

type Deck struct {
	Meta    map[string]string
	Slides  []Slide
	BaseDir string
	Align   AlignKind
}

// --- Parsing ---

func ParseDeck(src string) Deck {
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

	return Deck{Meta: meta, Slides: slides, Align: deckAlign}
}

func parseSlide(lines []string) Slide {
	var blocks []Block
	var slideAlign AlignKind
	inCode := false
	codeLang := ""
	var codeLines []string

	flushCode := func() {
		blocks = append(blocks, Block{
			Kind:  BlockCode,
			Lang:  codeLang,
			Lines: codeLines,
		})
		codeLines = nil
		codeLang = ""
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "::") {
			if inCode {
				flushCode()
				inCode = false
				if strings.HasPrefix(trimmed, "::code") && trimmed == "::code" {
					continue
				}
			}

			if strings.HasPrefix(trimmed, "::code") {
				inCode = true
				_, val := ParseDirective(trimmed)
				codeLang = val
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

		if trimmed == "" {
			continue
		}

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
		return blk.Directive
	case BlockList:
		return blk.Text
	default:
		return blk.Text
	}
}
