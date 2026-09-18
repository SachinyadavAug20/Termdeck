# termdeck format v0.1

A plain-text slide format for terminal presentations.

## File extension

`.deck.md`

## Structure

A deck is a Markdown file with three sections:

```
---              ← opening delimiter
format: 0.1      ← YAML frontmatter (key: value)
title: My Deck
---              ← closing delimiter

# Slide 1        ← slide content

---

# Slide 2
...
```

### Frontmatter

Optional YAML block between `---` delimiters. Keys:

| Key | Required | Description |
|-----|----------|-------------|
| `format` | yes | Version string, e.g. `0.1` |
| `title` | no | Deck title |
| `author` | no | Author name |

### Slides

Separated by `---` on its own line. The first slide starts after the frontmatter closing `---`.

## Block types

### Headings

`#` through `######` — rendered with decreasing visual weight.

```
# Title
## Section
### Subsection
```

### Text

Plain text lines. Supports inline formatting:

| Syntax | Rendered as |
|--------|-------------|
| `**bold**` | **bold** |
| `*italic*` | *italic* |
| `` `code` `` | `inline code` |

### Lists

Lines starting with `-` (unordered) or `1.` (ordered). Displayed as-is in v0.1.

### Directives

Lines starting with `::` — special instructions for the viewer.

| Directive | Syntax | Description |
|-----------|--------|-------------|
| `::code` | `::code lang=X` | Code block (until next `::` or `---`) |
| `::image` | `::image src=X` | Image placeholder |
| `::notes` | `::notes` | Speaker notes (hidden in presentation) |

#### `::code`

```
::code lang=python
  def hello():
      print("world")
```

Rendered as a bordered block with the language label. Content between `::code` and the next `::` or `---` is the code body.

#### `::image`

```
::image screenshot.png
```

Rendered as a placeholder: `[ image: screenshot.png ]`. Native image rendering (Sixel, Kitty) is planned for a future version.

#### `::notes`

```
::notes
  These are speaker notes.
  Not visible to the audience.
```

Skipped by the viewer. Reserved for future presenter-mode use.

## Slide rendering

- Slides are vertically and horizontally centered in the terminal.
- Footer shows: `slide N/M · ←/→ navigate · q quit`
- Content fills the available height minus footer.

## Keyboard shortcuts (viewer)

| Key | Action |
|-----|--------|
| `→`, `l`, `Space`, `Enter`, `↓`, `j`, `PageDown` | Next slide |
| `←`, `h`, `↑`, `k`, `PageUp`, `Backspace` | Previous slide |
| `g`, `Home` | First slide |
| `G`, `End` | Last slide |
| `q`, `Ctrl+C`, `Esc` | Quit |

## Example

See `demo.deck.md`.
