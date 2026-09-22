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
| `theme` | no | Color theme: `tokyo-night`, `dracula`, `catppuccin`, `nord`, `gruvbox`, `monokai`, `solarized`, `cyberpunk`, `termdeck` (or `#hex`) |

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

### Lists & Task Checklists

Lines starting with `-` or `*` (unordered) or `1.` (ordered).
- **Checklists**: `- [ ]` renders as an unchecked item (`○ `), and `- [x]` renders with a bold green checkmark (`✔ `) with dimmed completed text. Press `x` in navigation mode to toggle task state (`[ ]` ⇄ `[x]`) and auto-save immediately to file.
- **Bullets**: Rendered with clean dot glyphs (`• `) with theme accenting.

### Directives

Lines starting with `::` — special instructions for the viewer.

| Directive | Syntax | Description |
|-----------|--------|-------------|
| `::align` | `::align left/center/right` | Slide alignment (also `::left`, `::center`, `::right`) |
| `::code` | `::code lang=X` | Code block (until next `::` or `---`) |
| `::image` | `::image src=X` | Image card (or `::image filename.png`) |
| `::hr` | `::hr` | Horizontal divider hairline within slide (also `***`, `___`) |
| `::notes` | `::notes` | Speaker notes (hidden in presentation) |

### Horizontal Dividers

Lines containing `***`, `___`, or `::hr` render a clean horizontal hairline (`────────`) in theme border styling to separate concepts within a single slide.

#### `::align`

```
::align left
```

Sets the text alignment for the slide. Supported values: `left`, `center`, `right`. Shorthand directives `::left`, `::center`, and `::right` are also supported.

#### `::code` / Standard Code Fences

```
::code lang=python
  def hello():
      print("world")
```

Standard Markdown code fences (```` ```lang ````) and `::code` blocks are natively supported with syntax highlighting for Go, Python, TypeScript, Rust, Shell, SQL, and `diff`/`patch` (with green additions and red deletions). Content between code fences is rendered as a bordered code block with language labeling and dynamic box sizing.

#### `::image` / Standard Markdown Images

```
::image screenshot.png
```
Or standard markdown syntax:
```
![Alt Text](screenshot.png)
```

Rendered as a clean presentation card with file information, dimensions, and format. Press `p` in navigation mode to open the high-resolution image directly in your system viewer (`xdg-open` / `open`).

**Path Resolution Order**:
1. Absolute path (if path begins with `/`)
2. Relative to the `.deck.md` file directory
3. Relative to `images/` or `assets/` subdirectories next to the `.deck.md` file
4. Relative to current working directory

If the file is not found, a styled placeholder `[ image not found: <filename> ]` is displayed.

#### Markdown Tables

```
| Feature | SVN | Git |
|---|---|---|
| Model | Delta | Snapshot |
| Branching | Slow | Instant |
```

Standard GitHub-flavored Markdown tables are parsed into `BlockTable` blocks and rendered with styled borders, padded columns, and bold highlighted headers.

#### Callouts & Admonitions

```markdown
> [!TIP]
> Keep database transactions short to minimize lock contention.

> [!NOTE]
> Backward compatible with v1 API.

> [!WARNING]
> Breaking change in v2.

> "Simplicity is prerequisite for reliability."
> — Edsger W. Dijkstra
```

Callouts are parsed into `BlockCallout` blocks and rendered with styled borders matching theme accent, warning, and success colors, prefixed with clear icons (`💡 TIP`, `ℹ NOTE`, `⚠ WARNING`, `🚨 IMPORTANT`, `🛑 CAUTION`, `❝ QUOTE`).

#### `::notes`

```
::notes
  These are speaker notes.
  Not visible to the audience canvas.
```

Speaker notes are completely omitted from the audience canvas by default. The block cursor and laser pointer (`▶ `) skip notes blocks entirely. Presenters can press `n` at any time to toggle a styled speaker notes overlay box at the bottom of the terminal. When notes exist on the current slide, the status bar displays an `[n: notes]` indicator.

## Slide rendering

- Slides are vertically centered in the terminal with configurable horizontal alignment (`left`, `center`, `right`).
- Footer shows: `slide N/M (align) · blocks B · [n: notes] · ? help · tab align · t theme · n notes · i edit · ^n add · ^d del · ^s save · u undo · q quit`
- Subtle hairline progress line (`─`) across the bottom edge of the terminal.
- Content fills the available height minus footer (and notes overlay when active).

## Keyboard shortcuts (viewer)

| Key | Action |
|-----|--------|
| `→`, `l`, `Space`, `Enter`, `PageDown` | Next slide |
| `←`, `h`, `PageUp`, `Backspace` | Previous slide |
| `↓`, `j` | Move block cursor / laser pointer down |
| `↑`, `k` | Move block cursor / laser pointer up |
| `/` | Quick Jump to slide (by number or title search) |
| `o`, `O` | Slide Overview & 2D Grid Sorter (navigate cards, Enter to jump) |
| `?`, `F1` | Toggle in-app keyboard shortcuts help modal |
| `t`, `T`, `F2` | Cycle color theme (`tokyo-night`, `dracula`, `nord`, etc.) & auto-save |
| `z` | Toggle distraction-free zen mode (hides status bar) |
| `c`, `C` | Toggle presentation stopwatch (`c`) / Reset timer to 00:00 (`C`) |
| `r`, `R` | Reload deck file from disk |
| `y`, `Y` | Copy focused code/block to clipboard (OSC 52 + system) |
| `b`, `B` | Blank/blackout presentation screen (any key resumes) |
| `E` | Export deck to standalone HTML presentation |
| `L` | Toggle code block line numbers |
| `x` | Toggle task checklist item (`[ ]` ⇄ `[x]`) & auto-save |
| `n` | Toggle speaker notes overlay box |
| `Tab`, `Ctrl+A` | Cycle alignment (`left` → `center` → `right`) & auto-save |
| `p` | Open focused image in system viewer |
| `g`, `Home` | First slide |
| `G`, `End` | Last slide |
| `q`, `Ctrl+C` | Quit (auto-saves any unsaved changes) |
| `Esc` | Close help modal / clear message status |

## Example

See `demo.deck.md`.
