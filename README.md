# Termdeck

Terminal presentation tool. Write slides in markdown, present and edit full-screen in your terminal.

## Install

```bash
go build -o deck .
```

## Usage

```bash
deck demo.deck.md
```

## Project Status

#### 17 September 2026
- [x] Basic start with core logic in main.go
- [x] Full-screen viewer with keyboard navigation
- [x] Inline styling, heading levels, code blocks
- [x] Block-based editor with undo/redo
- [x] Save to `.deck.md`

![Termdeck Screenshot](TTP_Documentation/development/images/1.png)

## Project Structure

```
tpp/
├── main.go              # Entry point, tea.Model
├── internal/
│   ├── model.go         # Block types, Deck/Slide parsing, serialization
│   ├── view.go          # Rendering, styles, syntax highlighting
│   ├── editor.go        # Edit mode, block operations, undo/redo
│   └── image.go         # Terminal image renderer (ANSI half-blocks)
├── demo.deck.md         # Sample deck
├── go.mod
├── go.sum
├── .gitignore
├── README.md
└── TTP_Documentation/
    ├── CHANGELOG.md
    ├── specs/
    │   └── format.md    # Format v0.1 specification
    ├── development/
    │   └── images/      # Screenshots, diagrams
    └── api/             # API docs (future)
```

## Keys — Viewer

| Key | Action |
|-----|--------|
| `→` `l` `Space` `Enter` `↓` `j` `PageDown` | Next slide |
| `←` `h` `↑` `k` `PageUp` `Backspace` | Previous slide |
| `Tab` `Ctrl+A` | Cycle alignment (`left` → `center` → `right`) |
| `p` | Open focused image in system viewer |
| `G` | Last slide |
| `g` | First slide |
| `q` `Esc` `Ctrl+C` | Quit |

## Keys — Editor

Press `i` to enter edit mode on the selected block. Press `Esc` to exit edit mode.

| Key | Action |
|-----|--------|
| `↑` `k` | Move cursor up |
| `↓` `j` | Move cursor down |
| `i` | Enter edit mode |
| `Esc` | Exit edit mode / cancel |
| `Enter` | Confirm edit |
| `Ctrl+N` | Add new block |
| `Ctrl+D` | Delete block |
| `Ctrl+K` | Move block up |
| `Ctrl+J` | Move block down |
| `Ctrl+S` | Save file |
| `u` | Undo |
| `Ctrl+R` | Redo |

## Features

- Full-screen (alternate screen buffer)
- Auto-resizes on terminal resize
- Configurable alignment: `left`, `center`, `right` (via `Tab`, `::align`, or frontmatter)
- Inline styling: **bold**, *italic*, `code`
- Heading levels (h1–h6) with clean visual hierarchy
- `::code lang=X` blocks with syntax highlighting
- `::image` presentation cards with system viewer integration
- `::notes` (hidden in presentation)
- Block-based editor with live editing
- Undo/redo
- Save to `.deck.md`

## Format

See `TTP_Documentation/specs/format.md` for the full format specification.
