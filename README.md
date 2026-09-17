# Termdeck

Terminal presentation tool. Write slides in markdown, present full-screen in your terminal.

## Install

```bash
go build -o deck .
```

## Usage

```bash
deck demo.deck.md
```

## Format

Slides are separated by `---` (three dashes on a line).

```markdown
---
format: 0.1
title: My Presentation
author: Name
---

# Slide Title

- bullet points
- **bold** text
- `inline code`

---

# Another Slide

Content here...
```

## Keys

| Key | Action |
|-----|--------|
| `→` `j` `Space` `Enter` | Next slide |
| `←` `k` `Backspace` | Previous slide |
| `G` | Last slide |
| `g` | First slide |
| `q` `Esc` `Ctrl+C` | Quit |

## Features

- Full-screen (alternate screen buffer)
- Auto-resizes on terminal resize
- Markdown-like syntax
- Git-friendly plain text files
# Termdeck
