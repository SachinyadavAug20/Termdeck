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

#### 18 September 2026
- [x] Text body alignment (`left`, `center`, `right`) with `Tab`/`Ctrl+A`
- [x] High-contrast laser pointer indicator (`▶ `)
- [x] Stepped opacity heading hierarchy (pink underline for H1)
- [x] Image presentation cards with system viewer integration (`p`)
- [x] Complete automated test suite (91.7% statement coverage)
- [x] Makefile developer automation (`make test`, `make coverage`, `make bench`, `make lint`)

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
├── main_test.go         # Model lifecycle & CLI bootstrap unit tests
├── internal/
│   ├── model.go         # Block types, Deck/Slide parsing, serialization
│   ├── model_test.go    # Parser & serializer tests, edge cases, benchmarks
│   ├── view.go          # Rendering, styles, syntax highlighting
│   ├── view_test.go     # Highlighting lexer, styling, view benchmarks
│   ├── editor.go        # Edit mode, block operations, undo/redo
│   ├── editor_test.go   # Navigation, keyboard dispatch, edit mode tests
│   ├── image.go         # Terminal image renderer (ANSI half-blocks)
│   └── image_test.go    # Path resolution, format probing, card tests
├── demo.deck.md         # Sample deck
├── Makefile             # Developer automation (test, coverage, lint, bench, build)
├── go.mod
├── go.sum
├── .gitignore
├── README.md
└── TTP_Documentation/
    ├── CHANGELOG.md
    ├── specs/
    │   └── format.md    # Format v0.1 specification
    ├── development/
    │   ├── testing.md   # Complete QA & testing guide
    │   └── images/      # Screenshots, diagrams
    └── api/             # API docs (future)
```

## Keys — Viewer

| Key | Action |
|-----|--------|
| `→` `l` `Space` `Enter` `↓` `j` `PageDown` | Next slide |
| `←` `h` `↑` `k` `PageUp` `Backspace` | Previous slide |
| `Tab` `Ctrl+A` | Cycle alignment (`left` → `center` → `right`) & auto-save |
| `p` | Open focused image in system viewer |
| `G` | Last slide |
| `g` | First slide |
| `q` `Ctrl+C` | Quit (auto-saves any unsaved changes) |
| `Esc` | Clear message status |

## Keys — Editor

Press `i` to enter edit mode on the selected block. Press `Esc` to exit edit mode.

| Key | Action |
|-----|--------|
| `↑` `k` | Move cursor up |
| `↓` `j` | Move cursor down |
| `i` | Enter edit mode |
| `Esc` `Ctrl+C` | Exit edit mode / cancel |
| `Enter` | Confirm edit & auto-save to file |
| `Ctrl+N` | Add new block |
| `Ctrl+D` | Delete block |
| `Ctrl+K` | Move block up |
| `Ctrl+J` | Move block down |
| `Ctrl+S` | Save file manually |
| `u` | Undo |
| `Ctrl+R` | Redo |

## Features

- Full-screen (alternate screen buffer)
- Auto-resizes on terminal resize
- Configurable alignment: `left`, `center`, `right` (via `Tab`, `::align`, or frontmatter)
- **Automatic saving**: Saves immediately on alignment toggle (`Tab`/`Ctrl+A`), on edit confirm (`Enter`), and on exit (`q`/`Ctrl+C`)
- Manual save anytime with `Ctrl+S`
- Inline styling: **bold**, *italic*, `code`
- Heading levels (h1–h6) with clean visual hierarchy
- `::code lang=X` blocks with syntax highlighting
- `::image` presentation cards with system viewer integration
- `::notes` (hidden in presentation)
- Block-based editor with live editing
- Undo/redo
- Save to `.deck.md`

## Testing & Development

Termdeck features an automated test suite achieving **91.7% statement coverage** with 32 unit tests and 2 performance benchmarks.

```bash
# Run all unit tests with coverage summary
make test

# Generate an interactive HTML coverage report
make coverage

# Display per-function statement coverage
make coverage-summary

# Run performance benchmarks
make bench

# Verify formatting and static analysis
make lint
```

For full details, see the [Developer Testing Guide](TTP_Documentation/development/testing.md).

## Format

See `TTP_Documentation/specs/format.md` for the full format specification.

