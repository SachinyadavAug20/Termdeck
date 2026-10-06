package internal

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	directiveRegex = regexp.MustCompile(`^::\s*([a-zA-Z0-9_-]+)(.*)$`)
	reTrailingWS   = regexp.MustCompile(`[ \t]+$`)
)

// FormatDeck canonicalizes presentation markdown formatting, including frontmatter,
// slide separators, directive syntax, code fences, and whitespace.
func FormatDeck(src string) string {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.ReplaceAll(src, "\r", "\n")

	var frontmatter string
	body := src

	// Extract YAML frontmatter if present
	if strings.HasPrefix(strings.TrimLeft(src, " \t\n"), "---") {
		trimmed := strings.TrimLeft(src, " \t\n")
		rest := strings.TrimPrefix(trimmed, "---")
		if idx := strings.Index(rest, "\n---"); idx != -1 {
			fmContent := rest[:idx]
			afterFM := rest[idx+4:] // skip \n---
			frontmatter = formatFrontmatter(fmContent)
			body = afterFM
		}
	}

	formattedBody := formatDeckBody(body)

	var sb strings.Builder
	if frontmatter != "" {
		sb.WriteString("---\n")
		sb.WriteString(frontmatter)
		sb.WriteString("\n---\n\n")
	}

	sb.WriteString(formattedBody)
	res := sb.String()
	if !strings.HasSuffix(res, "\n") {
		res += "\n"
	}
	return res
}

func formatFrontmatter(fm string) string {
	lines := strings.Split(fm, "\n")
	var cleaned []string
	for _, l := range lines {
		trimmed := reTrailingWS.ReplaceAllString(l, "")
		cleaned = append(cleaned, trimmed)
	}

	// Trim leading/trailing empty lines
	for len(cleaned) > 0 && strings.TrimSpace(cleaned[0]) == "" {
		cleaned = cleaned[1:]
	}
	for len(cleaned) > 0 && strings.TrimSpace(cleaned[len(cleaned)-1]) == "" {
		cleaned = cleaned[:len(cleaned)-1]
	}

	return strings.Join(cleaned, "\n")
}

func formatDeckBody(body string) string {
	lines := strings.Split(body, "\n")
	var slides [][]string
	var currentSlide []string
	inCodeBlock := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if strings.HasPrefix(trimmedLine, "```") {
			inCodeBlock = !inCodeBlock
			currentSlide = append(currentSlide, reTrailingWS.ReplaceAllString(line, ""))
			continue
		}

		if !inCodeBlock && (trimmedLine == "---" || trimmedLine == "***") {
			slides = append(slides, currentSlide)
			currentSlide = nil
			continue
		}

		currentSlide = append(currentSlide, reTrailingWS.ReplaceAllString(line, ""))
	}
	if len(currentSlide) > 0 || len(slides) == 0 {
		slides = append(slides, currentSlide)
	}

	var formattedSlides []string
	for _, s := range slides {
		formattedSlide := formatSingleSlide(s)
		if strings.TrimSpace(formattedSlide) != "" {
			formattedSlides = append(formattedSlides, formattedSlide)
		}
	}

	return strings.Join(formattedSlides, "\n\n---\n\n")
}

func formatSingleSlide(lines []string) string {
	// Trim leading empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	// Trim trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	if len(lines) == 0 {
		return ""
	}

	var formatted []string
	inCode := false
	prevEmpty := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			formatted = append(formatted, line)
			prevEmpty = false
			continue
		}

		if inCode {
			formatted = append(formatted, line)
			prevEmpty = false
			continue
		}

		if trimmed == "" {
			if !prevEmpty {
				formatted = append(formatted, "")
				prevEmpty = true
			}
			continue
		}
		prevEmpty = false

		// Canonicalize directive syntax
		if strings.HasPrefix(trimmed, "::") {
			formatted = append(formatted, formatDirective(trimmed))
			continue
		}

		// Normalize markdown headings
		if strings.HasPrefix(trimmed, "#") {
			formatted = append(formatted, formatHeading(line))
			continue
		}

		formatted = append(formatted, line)
	}

	return strings.Join(formatted, "\n")
}

func formatDirective(line string) string {
	m := directiveRegex.FindStringSubmatch(line)
	if len(m) < 3 {
		return line
	}
	dirName := strings.ToLower(m[1])
	rest := strings.TrimSpace(m[2])

	switch dirName {
	case "tags", "tag":
		rest = strings.TrimPrefix(rest, ":")
		rest = strings.TrimSpace(rest)
		return fmt.Sprintf("::tags: %s", rest)
	case "track":
		rest = strings.TrimPrefix(rest, ":")
		rest = strings.TrimSpace(rest)
		return fmt.Sprintf("::track: %s", rest)
	case "route":
		rest = strings.TrimPrefix(rest, ":")
		rest = strings.TrimSpace(rest)
		return fmt.Sprintf("::route: %s", rest)
	case "loop":
		rest = strings.TrimPrefix(rest, ":")
		rest = strings.TrimSpace(rest)
		return fmt.Sprintf("::loop: %s", rest)
	case "bg":
		rest = strings.TrimPrefix(rest, ":")
		rest = strings.TrimSpace(rest)
		return fmt.Sprintf("::bg: %s", rest)
	case "align":
		rest = strings.TrimPrefix(rest, ":")
		rest = strings.TrimSpace(rest)
		return fmt.Sprintf("::align: %s", rest)
	case "note", "notes":
		rest = strings.TrimPrefix(rest, ":")
		rest = strings.TrimSpace(rest)
		return fmt.Sprintf("::note: %s", rest)
	case "branch":
		rest = strings.TrimPrefix(rest, ":")
		rest = strings.TrimSpace(rest)
		return fmt.Sprintf("::branch %s", rest)
	default:
		return line
	}
}

func formatHeading(line string) string {
	// Standardize space between '#' and title
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	hashes := line[:i]
	title := strings.TrimSpace(line[i:])
	return hashes + " " + title
}

// FormatDeckFile reads the specified file, formats it, and writes it back (unless checkOnly is true).
// It returns true if the file was already formatted, along with the formatted content.
func FormatDeckFile(path string, checkOnly bool) (isFormatted bool, formattedContent string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, "", err
	}
	original := string(data)
	formatted := FormatDeck(original)

	isFormatted = (original == formatted)
	if !isFormatted && !checkOnly {
		if err := os.WriteFile(path, []byte(formatted), 0644); err != nil {
			return false, formatted, err
		}
	}
	return isFormatted, formatted, nil
}
