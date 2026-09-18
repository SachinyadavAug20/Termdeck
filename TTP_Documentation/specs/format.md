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
| `align` | no | Default alignment: `left`, `center`, `right` |

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
| `::align` | `::align left/center/right` | Slide alignment (also `::left`, `::center`, `::right`) |
| `::code` | `::code lang=X` | Code block (until next `::` or `---`) |
| `::image` | `::image src=X` | Image card (or `::image filename.png`) |
| `::notes` | `::notes` | Speaker notes (hidden in presentation) |

#### `::align`

```
::align left
```

Sets the text alignment for the slide. Supported values: `left`, `center`, `right`. Shorthand directives `::left`, `::center`, and `::right` are also supported.

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

Rendered as a clean presentation card with file information, dimensions, and format. Press `p` in navigation mode to open the high-resolution image directly in your system viewer (`xdg-open` / `open`).

**Path Resolution Order**:
1. Absolute path (if path begins with `/`)
2. Relative to the `.deck.md` file directory
3. Relative to `images/` or `assets/` subdirectories next to the `.deck.md` file
4. Relative to current working directory

If the file is not found, a styled placeholder `[ image not found: <filename> ]` is displayed.

#### `::notes`

```
::notes
  These are speaker notes.
  Not visible to the audience.
```

Skipped by the viewer. Reserved for future presenter-mode use.

## Slide rendering

- Slides are vertically centered in the terminal with configurable horizontal alignment (`left`, `center`, `right`).
- Footer shows: `slide N/M (align) · tab align · i edit · ^s save · q quit`
- Content fills the available height minus footer.

## Keyboard shortcuts (viewer)

| Key | Action |
|-----|--------|
| `→`, `l`, `Space`, `Enter`, `↓`, `j`, `PageDown` | Next slide |
| `←`, `h`, `↑`, `k`, `PageUp`, `Backspace` | Previous slide |
| `Tab`, `Ctrl+A` | Cycle alignment (`left` → `center` → `right`) |
| `p` | Open focused image in system viewer |
| `g`, `Home` | First slide |
| `G`, `End` | Last slide |
| `q`, `Ctrl+C`, `Esc` | Quit |

## Example

See `demo.deck.md`.
