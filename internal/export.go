package internal

import (
	"encoding/base64"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	reExportInlineCode = regexp.MustCompile("`([^`]+)`")
	reExportBold       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	reExportItalic     = regexp.MustCompile(`\*([^*]+)\*`)
)

func formatInlineHTML(s string) string {
	// 1. Extract code spans first to avoid asterisk style collision
	var codeSpans []string
	placeholder := "\x00CODE"
	replaced := reExportInlineCode.ReplaceAllStringFunc(s, func(m string) string {
		match := reExportInlineCode.FindStringSubmatch(m)
		idx := len(codeSpans)
		codeSpans = append(codeSpans, match[1])
		return fmt.Sprintf("%s%d\x00", placeholder, idx)
	})

	// 2. Escape HTML on general text
	escaped := html.EscapeString(replaced)

	// 3. Bold & Italic
	escaped = reExportBold.ReplaceAllString(escaped, "<strong>$1</strong>")
	escaped = reExportItalic.ReplaceAllString(escaped, "<em>$1</em>")

	// 4. Restore code spans with HTML escaping
	for i, code := range codeSpans {
		token := fmt.Sprintf("%s%d\x00", placeholder, i)
		escaped = strings.ReplaceAll(escaped, token, "<code>"+html.EscapeString(code)+"</code>")
	}

	return escaped
}

// ExportHTML serializes the given Deck into a self-contained, responsive, offline-ready HTML presentation.
func ExportHTML(d Deck, title string) string {
	if title == "" {
		if t, ok := d.Meta["title"]; ok && strings.TrimSpace(t) != "" {
			title = t
		} else if len(d.Slides) > 0 {
			title = d.Slides[0].Title()
		} else {
			title = "Termdeck Presentation"
		}
	}

	theme := ResolveTheme(d.Theme)

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	sb.WriteString("<meta charset=\"UTF-8\">\n")
	sb.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString("<title>" + html.EscapeString(title) + "</title>\n")
	sb.WriteString("<style>\n")
	sb.WriteString(generateDeckCSS(theme))
	sb.WriteString("</style>\n</head>\n<body>\n")

	sb.WriteString("<div class=\"deck-container\">\n")

	for sIdx, slide := range d.Slides {
		activeClass := ""
		if sIdx == 0 {
			activeClass = " active"
		}
		alignClass := " align-" + string(slide.Align)
		if slide.Align == "" {
			alignClass = " align-" + string(d.Align)
			if d.Align == "" {
				alignClass = " align-left"
			}
		}

		idAttr := html.EscapeString(slide.ID)
		slugAttr := html.EscapeString(slide.Slug())
		nextAttr := html.EscapeString(slide.NextID)
		prevAttr := html.EscapeString(slide.PrevID)
		sb.WriteString(fmt.Sprintf("  <section class=\"slide%s%s\" id=\"slide-%d\" data-slide=\"%d\" data-id=\"%s\" data-slug=\"%s\" data-next=\"%s\" data-prev=\"%s\">\n",
			activeClass, alignClass, sIdx+1, sIdx+1, idAttr, slugAttr, nextAttr, prevAttr))

		visIndices := slide.VisibleBlockIndices()
		for _, bIdx := range visIndices {
			blk := slide.Blocks[bIdx]
			sb.WriteString(renderBlockHTML(blk, d.BaseDir))
		}

		sb.WriteString("  </section>\n")
	}

	sb.WriteString("</div>\n")

	// Progress line
	sb.WriteString("<div class=\"progress-line\"><div class=\"progress-fill\" id=\"progress-fill\"></div></div>\n")

	// Status bar
	sb.WriteString("<div class=\"status-bar\">\n")
	sb.WriteString(fmt.Sprintf("  <div class=\"status-left\"><strong>Termdeck</strong> &middot; <span id=\"slide-counter\">1 / %d</span></div>\n", len(d.Slides)))
	sb.WriteString("  <div class=\"status-right\"><span>&larr; / &rarr; navigate &middot; 1-9: branch &middot; Backspace: back &middot; f: fullscreen &middot; X: run</span></div>\n")
	sb.WriteString("</div>\n")

	// Embedded Navigation Script
	sb.WriteString("<script>\n")
	sb.WriteString(generateDeckJS(len(d.Slides)))
	sb.WriteString("</script>\n</body>\n</html>\n")

	return sb.String()
}

func renderBlockHTML(blk Block, baseDir string) string {
	switch blk.Kind {
	case BlockHeading:
		lvl := blk.Level
		if lvl < 1 {
			lvl = 1
		}
		if lvl > 6 {
			lvl = 6
		}
		return fmt.Sprintf("    <h%d>%s</h%d>\n", lvl, formatInlineHTML(blk.Text), lvl)

	case BlockParagraph:
		return fmt.Sprintf("    <p>%s</p>\n", formatInlineHTML(blk.Text))

	case BlockCode:
		langBadge := ""
		langText := "code"
		if blk.Lang != "" {
			langBadge = fmt.Sprintf(" data-lang=\"%s\"", html.EscapeString(blk.Lang))
			langText = blk.Lang
		}
		var codeEscaped []string
		for _, line := range blk.Lines {
			codeEscaped = append(codeEscaped, html.EscapeString(line))
		}
		return fmt.Sprintf("    <div class=\"code-block\"%s><div class=\"code-header\"><span class=\"code-lang\">%s</span><button class=\"code-run-btn\" onclick=\"runCodeSnippet(this)\">▶ Run</button></div><pre><code>%s</code></pre><div class=\"code-output\" style=\"display:none;\"></div></div>\n", langBadge, html.EscapeString(langText), strings.Join(codeEscaped, "\n"))

	case BlockTable:
		var sb strings.Builder
		sb.WriteString("    <table class=\"deck-table\">\n")
		isHeader := true
		for _, line := range blk.Lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") {
				line = line[1 : len(line)-1]
			}
			parts := strings.Split(line, "|")
			if isHeader {
				sb.WriteString("      <thead><tr>\n")
				for _, p := range parts {
					sb.WriteString(fmt.Sprintf("        <th>%s</th>\n", formatInlineHTML(strings.TrimSpace(p))))
				}
				sb.WriteString("      </tr></thead><tbody>\n")
				isHeader = false
				continue
			}
			// Skip separator line e.g. |---|---|
			if len(parts) > 0 && strings.Contains(parts[0], "---") {
				continue
			}
			sb.WriteString("      <tr>\n")
			for _, p := range parts {
				sb.WriteString(fmt.Sprintf("        <td>%s</td>\n", formatInlineHTML(strings.TrimSpace(p))))
			}
			sb.WriteString("      </tr>\n")
		}
		sb.WriteString("    </tbody></table>\n")
		return sb.String()

	case BlockCallout:
		cType := blk.Callout
		if cType == "" {
			cType = "note"
		}
		icon := "ℹ"
		title := strings.ToUpper(cType)
		switch cType {
		case "tip":
			icon = "💡"
		case "warning":
			icon = "⚠"
		case "important":
			icon = "🚨"
		case "caution":
			icon = "🛑"
		case "quote":
			icon = "❝"
		}
		var innerLines []string
		if len(blk.Lines) > 0 {
			for _, l := range blk.Lines {
				innerLines = append(innerLines, formatInlineHTML(l))
			}
		} else {
			innerLines = append(innerLines, formatInlineHTML(blk.Text))
		}
		return fmt.Sprintf("    <div class=\"callout callout-%s\"><div class=\"callout-header\">%s %s</div><p>%s</p></div>\n",
			html.EscapeString(cType), icon, title, strings.Join(innerLines, "<br>\n"))

	case BlockList:
		trimmed := strings.TrimSpace(blk.Text)
		isDone := strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]")
		isTask := isDone || strings.HasPrefix(trimmed, "- [ ]")
		if isTask {
			checked := ""
			doneClass := ""
			label := strings.TrimSpace(trimmed[5:])
			if isDone {
				checked = " checked"
				doneClass = " task-done"
			}
			return fmt.Sprintf("    <div class=\"task-item%s\"><input type=\"checkbox\" disabled%s> <span>%s</span></div>\n",
				doneClass, checked, formatInlineHTML(label))
		}
		// Regular bullet list item
		bulletText := trimmed
		if strings.HasPrefix(bulletText, "- ") || strings.HasPrefix(bulletText, "* ") {
			bulletText = bulletText[2:]
		}
		return fmt.Sprintf("    <div class=\"list-item\"><span class=\"bullet\">&bull;</span> <span>%s</span></div>\n", formatInlineHTML(bulletText))

	case BlockDivider:
		return "    <hr class=\"deck-divider\">\n"

	case BlockBranch:
		keyStr := blk.BranchKey
		if keyStr == "" {
			keyStr = "→"
		}
		target := html.EscapeString(blk.BranchTarget)
		label := formatInlineHTML(blk.Text)
		return fmt.Sprintf("    <div class=\"branch-fork-card\" data-key=\"%s\" data-target=\"%s\" onclick=\"jumpToBranch('%s')\"><span class=\"branch-key\">[%s]</span> <span class=\"branch-label\">%s</span> <span class=\"branch-arrow\">&xrarr;</span> <span class=\"branch-target\">#%s</span></div>\n",
			html.EscapeString(keyStr), target, target, html.EscapeString(keyStr), label, target)

	case BlockImage:
		// Attempt to read and embed base64 image if exists, or use src path
		src := blk.Src
		if baseDir != "" {
			fullPath, found := ResolveImagePath(src, baseDir)
			if found {
				if bytes, err := os.ReadFile(fullPath); err == nil {
					mime := "image/png"
					ext := strings.ToLower(filepath.Ext(fullPath))
					switch ext {
					case ".jpg", ".jpeg":
						mime = "image/jpeg"
					case ".gif":
						mime = "image/gif"
					case ".svg":
						mime = "image/svg+xml"
					}
					src = fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(bytes))
				}
			}
		}
		alt := blk.Text
		if alt == "" {
			alt = blk.Src
		}
		return fmt.Sprintf("    <div class=\"image-card\"><img src=\"%s\" alt=\"%s\" /><div class=\"image-caption\">%s</div></div>\n",
			html.EscapeString(src), html.EscapeString(alt), html.EscapeString(alt))

	default:
		if strings.TrimSpace(blk.Text) != "" {
			return fmt.Sprintf("    <p>%s</p>\n", formatInlineHTML(blk.Text))
		}
		return ""
	}
}

func generateDeckCSS(theme Theme) string {
	accent := theme.Accent
	if accent == "" {
		accent = "#bb9af7"
	}
	return `
  :root {
    --bg: #12131a;
    --fg: #e2e8f0;
    --accent: ` + accent + `;
    --border: #242938;
    --card-bg: #1a1e2b;
    --code-bg: #0d1117;
    --success: #10b981;
    --warning: #f59e0b;
    --danger: #ef4444;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    background: var(--bg);
    color: var(--fg);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Fira Code", monospace;
    height: 100vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    user-select: none;
  }
  .deck-container {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    position: relative;
    overflow: hidden;
  }
  .slide {
    display: none;
    flex-direction: column;
    width: 100%;
    max-width: 960px;
    max-height: 84vh;
    overflow-y: auto;
    gap: 1.2rem;
    animation: slideFadeIn 0.15s ease-out;
    padding: 1.5rem 2rem;
    background: var(--card-bg);
    border: 1px solid var(--border);
    border-radius: 12px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.5);
  }
  .slide.active { display: flex; }
  .slide.align-center { align-items: center; text-align: center; }
  .slide.align-left { align-items: flex-start; text-align: left; }
  .slide.align-right { align-items: flex-end; text-align: right; }
  @keyframes slideFadeIn {
    from { opacity: 0; transform: translateY(6px); }
    to { opacity: 1; transform: translateY(0); }
  }
  h1 { font-size: 2.4rem; color: var(--accent); border-bottom: 2px solid var(--accent); padding-bottom: 0.4rem; width: 100%; }
  h2 { font-size: 1.8rem; color: #ffffff; }
  h3 { font-size: 1.4rem; color: #cbd5e1; }
  p { font-size: 1.15rem; line-height: 1.6; color: var(--fg); }
  strong { color: #ffffff; font-weight: 700; }
  em { color: #93c5fd; font-style: italic; }
  code {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    background: var(--code-bg);
    color: var(--accent);
    padding: 0.2rem 0.4rem;
    border-radius: 4px;
    font-size: 0.95em;
    border: 1px solid var(--border);
  }
  .code-block {
    position: relative;
    width: 100%;
    background: var(--code-bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    overflow: hidden;
  }
  .code-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.35rem 0.85rem;
    background: rgba(255, 255, 255, 0.04);
    border-bottom: 1px solid var(--border);
    font-size: 0.8rem;
  }
  .code-lang {
    color: var(--accent);
    font-weight: 700;
    text-transform: uppercase;
    font-family: monospace;
    letter-spacing: 0.05em;
  }
  .code-run-btn {
    background: var(--accent);
    color: #ffffff;
    border: none;
    border-radius: 4px;
    padding: 0.2rem 0.6rem;
    font-size: 0.75rem;
    cursor: pointer;
    font-weight: 700;
    transition: opacity 0.15s ease;
  }
  .code-run-btn:hover { opacity: 0.85; }
  .code-output {
    margin: 0 1rem 1rem 1rem;
    padding: 0.6rem 0.85rem;
    background: rgba(0, 0, 0, 0.45);
    border-left: 3px solid var(--accent);
    border-radius: 4px;
    font-family: ui-monospace, monospace;
    font-size: 0.85rem;
    color: #a7f3d0;
  }
  pre {
    padding: 1rem 1.25rem;
    overflow-x: auto;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 0.95rem;
    line-height: 1.5;
  }
  pre code {
    background: transparent;
    padding: 0;
    border: none;
    color: #e2e8f0;
  }
  .deck-table {
    width: 100%;
    border-collapse: collapse;
    margin: 0.5rem 0;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid var(--border);
  }
  .deck-table th, .deck-table td {
    padding: 0.6rem 1rem;
    border: 1px solid var(--border);
    text-align: left;
    font-size: 1rem;
  }
  .deck-table th {
    background: #1e2230;
    color: var(--accent);
    font-weight: 600;
  }
  .deck-table td {
    background: var(--card-bg);
  }
  .callout {
    width: 100%;
    border: 1px solid var(--border);
    border-left: 4px solid var(--accent);
    background: #181d2a;
    border-radius: 8px;
    padding: 1rem 1.25rem;
  }
  .callout-tip { border-left-color: var(--success); }
  .callout-warning { border-left-color: var(--warning); }
  .callout-important, .callout-caution { border-left-color: var(--danger); }
  .callout-header {
    font-size: 0.9rem;
    font-weight: 700;
    letter-spacing: 0.05em;
    color: var(--accent);
    margin-bottom: 0.4rem;
  }
  .task-item {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    font-size: 1.15rem;
  }
  .task-item input {
    accent-color: var(--success);
    width: 1.1rem;
    height: 1.1rem;
  }
  .task-done span {
    text-decoration: line-through;
    opacity: 0.5;
  }
  .list-item {
    display: flex;
    align-items: flex-start;
    gap: 0.6rem;
    font-size: 1.15rem;
  }
  .bullet {
    color: var(--accent);
  }
  .deck-divider {
    width: 100%;
    border: none;
    border-top: 1px solid var(--border);
  }
  .branch-fork-card {
    display: inline-flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.6rem 1.2rem;
    margin: 0.4rem 0;
    background: var(--code-bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.15s ease;
    font-size: 1.05rem;
    max-width: 680px;
    text-decoration: none;
  }
  .branch-fork-card:hover {
    border-color: var(--accent);
    transform: translateX(4px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.25);
  }
  .branch-key {
    background: var(--accent);
    color: #ffffff;
    font-weight: 700;
    padding: 0.15rem 0.45rem;
    border-radius: 4px;
    font-family: monospace;
    font-size: 0.9rem;
  }
  .branch-label {
    font-weight: 600;
    color: #ffffff;
  }
  .branch-arrow {
    color: #64748b;
  }
  .branch-target {
    color: #94a3b8;
    font-family: monospace;
    font-size: 0.85rem;
  }
  .image-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
  }
  .image-card img {
    max-width: 100%;
    max-height: 400px;
    border-radius: 8px;
    border: 1px solid var(--border);
  }
  .image-caption {
    font-size: 0.85rem;
    color: #64748b;
  }
  .progress-line {
    height: 3px;
    background: var(--border);
    width: 100%;
  }
  .progress-fill {
    height: 100%;
    background: var(--accent);
    transition: width 0.2s ease-out;
  }
  .status-bar {
    height: 36px;
    background: #0f1117;
    border-top: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 1.5rem;
    font-size: 0.85rem;
    color: #64748b;
  }
  .status-left strong { color: var(--accent); }
`
}

func generateDeckJS(totalSlides int) string {
	return fmt.Sprintf(`
  let currentSlide = 1;
  const totalSlides = %d;
  const historyStack = [];

  window.jumpToBranch = function(target) {
    if (!target) return;
    target = target.toLowerCase();
    const slides = document.querySelectorAll('.slide');
    for (let i = 0; i < slides.length; i++) {
      const s = slides[i];
      const sId = (s.getAttribute('data-id') || '').toLowerCase();
      const sSlug = (s.getAttribute('data-slug') || '').toLowerCase();
      if (sId === target || sSlug === target || (i + 1).toString() === target) {
        historyStack.push(currentSlide);
        updateSlide(i + 1);
        return;
      }
    }
  };

  function updateSlide(n) {
    if (n < 1) n = 1;
    if (n > totalSlides) n = totalSlides;
    currentSlide = n;

    document.querySelectorAll('.slide').forEach((el, idx) => {
      if (idx + 1 === currentSlide) {
        el.classList.add('active');
      } else {
        el.classList.remove('active');
      }
    });

    const progressPercent = totalSlides > 1 ? ((currentSlide - 1) / (totalSlides - 1)) * 100 : 100;
    const progressFill = document.getElementById('progress-fill');
    if (progressFill) progressFill.style.width = progressPercent + '%%';

    const counter = document.getElementById('slide-counter');
    if (counter) counter.innerText = currentSlide + ' / ' + totalSlides;
  }

  document.addEventListener('keydown', (e) => {
    if (e.key >= '1' && e.key <= '9') {
      const activeSlide = document.querySelector('.slide.active');
      if (activeSlide) {
        const btn = activeSlide.querySelector('.branch-fork-card[data-key="' + e.key + '"]');
        if (btn) {
          const tgt = btn.getAttribute('data-target');
          if (tgt) { jumpToBranch(tgt); return; }
        }
      }
    }

    switch (e.key) {
      case 'ArrowRight':
      case 'l':
      case ' ':
      case 'Enter':
      case 'PageDown': {
        const activeSlide = document.querySelector('.slide.active');
        const nextTarget = activeSlide ? activeSlide.getAttribute('data-next') : null;
        if (nextTarget) {
          jumpToBranch(nextTarget);
        } else {
          updateSlide(currentSlide + 1);
        }
        break;
      }
      case 'Backspace':
      case 'H':
        if (historyStack.length > 0) {
          updateSlide(historyStack.pop());
          break;
        }
        updateSlide(currentSlide - 1);
        break;
      case 'ArrowLeft':
      case 'h':
      case 'PageUp': {
        const activeSlide = document.querySelector('.slide.active');
        const prevTarget = activeSlide ? activeSlide.getAttribute('data-prev') : null;
        if (prevTarget) {
          jumpToBranch(prevTarget);
        } else {
          updateSlide(currentSlide - 1);
        }
        break;
      }
      case 'Home':
      case 'g':
        updateSlide(1);
        break;
      case 'End':
      case 'G':
        updateSlide(totalSlides);
        break;
      case 'f':
      case 'F':
        if (!document.fullscreenElement) {
          document.documentElement.requestFullscreen().catch(() => {});
        } else {
          document.exitFullscreen().catch(() => {});
        }
        break;
      case 'X': {
        const activeSlide = document.querySelector('.slide.active');
        if (activeSlide) {
          const runBtn = activeSlide.querySelector('.code-run-btn');
          if (runBtn) runCodeSnippet(runBtn);
        }
        break;
      }
    }
  });

  window.runCodeSnippet = function(btn) {
    const parent = btn.closest('.code-block');
    if (!parent) return;
    const out = parent.querySelector('.code-output');
    if (!out) return;
    if (out.style.display === 'none') {
      const lang = parent.getAttribute('data-lang') || 'code';
      out.textContent = '⚡ [' + lang + '] live run complete (exit 0 · simulated output)';
      out.style.display = 'block';
      btn.textContent = '✖ Close';
    } else {
      out.style.display = 'none';
      btn.textContent = '▶ Run';
    }
  };

  // Touch swipe support
  let touchStartX = 0;
  document.addEventListener('touchstart', (e) => {
    touchStartX = e.changedTouches[0].screenX;
  }, false);
  document.addEventListener('touchend', (e) => {
    const diff = e.changedTouches[0].screenX - touchStartX;
    if (diff < -50) updateSlide(currentSlide + 1);
    if (diff > 50) updateSlide(currentSlide - 1);
  }, false);

  updateSlide(1);
`, totalSlides)
}

// ExportHTMLFile writes the standalone HTML presentation to the specified output file path.
func ExportHTMLFile(d Deck, filePath string) error {
	content := ExportHTML(d, "")
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(filePath, []byte(content), 0644)
}

// ExportHTML on Editor computes default export path and invokes ExportHTMLFile.
func (e *Editor) ExportHTML(d *Deck) (string, error) {
	outPath := "deck.html"
	if e.FilePath != "" {
		base := e.FilePath
		if strings.HasSuffix(base, ".deck.md") {
			base = strings.TrimSuffix(base, ".deck.md")
		} else {
			base = strings.TrimSuffix(base, filepath.Ext(base))
		}
		outPath = base + ".html"
	}
	err := ExportHTMLFile(*d, outPath)
	if err != nil {
		e.Message = "export error: " + err.Error()
		return "", err
	}
	e.Message = "exported to " + filepath.Base(outPath)
	return outPath, nil
}
