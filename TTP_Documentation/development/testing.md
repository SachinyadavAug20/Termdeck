# Termdeck Testing & Quality Assurance Guide

This document outlines the testing architecture, developer tooling, coverage metrics, and guidelines for contributing tests to **Termdeck (`deck`)**.

---

## 1. Testing Philosophy

- **Zero External Test Frameworks**: All tests are written using the Go standard library `testing` package without heavy assertion libraries or mock frameworks.
- **Fast Execution**: The complete test suite runs in under **0.5 seconds** locally.
- **High Coverage Threshold**: Core business logic and rendering paths maintain **>90% statement coverage**.
- **ANSI Resilience**: Terminal styles produced by Lipgloss are tested using robust ANSI stripping helpers (`stripANSI`) to ensure consistent test results across varying terminal environments and CI pipelines.

---

## 2. Test Inventory & Architecture

```
tpp/
├── main_test.go             # Root CLI & Bubble Tea engine tests
├── internal/
│   ├── editor_test.go       # Navigation, edit mode, undo/redo, block mutations
│   ├── view_test.go         # Syntax highlighting, inline styles, Lipgloss layout
│   ├── model_test.go        # Parser edge cases, directives, round-trip serialization
│   └── image_test.go        # Image path resolution, format probing, ASCII cards
└── Makefile                 # Developer automation
```

### A. Root Package (`deck`) — `main_test.go`
Tests the Bubble Tea model lifecycle and CLI bootstrapping logic:
- `TestModelInit`: Asserts initial Bubble Tea command is `nil`.
- `TestModelUpdateWindowSize`: Verifies terminal resize updates dimensions and emits `tea.ClearScreen` on initial resize.
- `TestModelUpdateKeyMsg`: Verifies key events dispatch correctly to the editor state machine.
- `TestModelUpdateQuitKey`: Asserts pressing `"q"` returns `tea.Quit`.
- `TestModelView`: Asserts full-screen rendering and status bar contents.
- `TestBuildModel`: Tests loading valid decks, non-existent files, and empty files (asserting "no slides found").

### B. Editor State Machine — `internal/editor_test.go`
Tests the navigation and editing state machine:
- `TestEditorNavigation`: Block pointer bounds and movement (`MoveUp`, `MoveDown`).
- `TestEditorBlockEditing`: Draft buffer lifecycle and commit in `EnterEdit`/`ExitEdit`.
- `TestEditorAddDeleteBlock`: Block insertion and deletion with undo snapshots.
- `TestEditorUndoRedo`: Multi-step undo/redo stack verification.
- `TestEditorToggleAlign`: Cycling through `left` $\to$ `center` $\to$ `right` $\to$ `left`.
- `TestEditorHandleKeyNav`: Keyboard event handling for all navigation keys:
  - Slide jumping (`g`, `G`)
  - Directional navigation (`right`, `left`, `down`, `up`, `j`, `k`, `h`, `l`)
  - Block mutations (`ctrl+n`, `ctrl+d`, `ctrl+k`, `ctrl+j`)
  - Slide mutations (`ctrl+N`, `ctrl+D`)
  - Image opening (`p`)
  - Edit trigger keys (`i`, `a`, `o`, `I`, `A`, `O`)
- `TestEditorHandleKeyEdit`: In-place typing, cursor navigation (`home`, `end`, `left`, `right`), deletions (`backspace`, `delete`), canceling (`esc`), and saving (`enter`).
- `TestEditorSaveAndError`: Disk serialization and unwritable directory error handling.
- `TestEditorEdgeCases`: Boundaries (preventing deleting last block or last slide, empty undo/redo stacks, out-of-range indices).

### C. View & Syntax Highlighter — `internal/view_test.go`
Tests visual layout and terminal text styling:
- `TestHighlightLine`: Syntax highlighter lexer covering:
  - Comments (`//`, `#`, `--`)
  - String literals (`"..."`, `'...'`, `` `...` ``)
  - Numeric literals (`42`, `3.14`)
  - Keywords (`func`, `return`, `def`, `class`, `echo`, etc.)
  - PascalCase types
- `TestHighlightCode`: Multiline code syntax highlighting.
- `TestInlineStyle`: Markdown inline styling (`**bold**`, `*italic*`, `` `code` `` spans).
- `TestRenderHeadingLevels`: Hierarchy rendering for H1 (pink underline) and H2–H6 (stepped white opacity).
- `TestRenderLaserPointer`: Verifies the laser pointer marker (`▶ ` in `#FF2A55`) renders on the active line.
- `TestRenderImageCard`: Formatted image card with dimension metadata and open hints.
- `TestRenderBlockVariants`: Renders headings, paragraphs, code blocks with language tags, image cards, list items, and directives in both display and edit modes.
- `TestStatusBars`: Validates `navStatus` (slide position, alignment badge, dirty indicator) and `editStatus` (cursor column position, messages).
- `TestViewDimensionsAndModes`: Validates fallback on zero dimensions, edit mode rendering, and alignment padding.
- `BenchmarkRenderView`: Measures frames-per-second rendering efficiency.

### D. Model & Parser — `internal/model_test.go`
Tests markdown AST parsing and serialization:
- `TestParseDeck`: Standard presentation parsing with YAML frontmatter.
- `TestSerializeRoundTrip`: AST to Markdown string serialization consistency.
- `TestParseDirective`: Directives extraction (`::code lang=python`, `::image path.png`).
- `TestParseAlignment`: Slide-level and deck-level alignment specifications.
- `TestParseDeckEdgeCases`: Empty decks, missing frontmatter, numbered lists, bullet points, closed code blocks, and custom directives.
- `TestSerializeDeckEdgeCases`: Serialization with alignments and multiple block types.
- `BenchmarkParseDeck`: Measures markdown parser throughput.

### E. Terminal Images — `internal/image_test.go`
Tests image resolution and ANSI rendering:
- `TestResolveImagePath`: Path lookup relative to `.deck.md` and current working directory.
- `TestRenderImageMissing`: Missing file fallback placeholder.
- `TestRenderImageOutput`: ANSI half-block (`▀`) truecolor character emission.
- `TestRenderRealImages`: Rendering test fixtures.
- `TestGetImageInfo`: Decoding image dimensions and format types.

---

## 3. Developer Tooling (`Makefile`)

A `Makefile` is provided at the repository root to standardize development:

```bash
# Run all unit tests with statement coverage summary
make test

# Generate an interactive HTML coverage report at coverage.html
make coverage

# Display function-level coverage breakdown table in terminal
make coverage-summary

# Run performance benchmarks with memory allocations
make bench

# Check Go formatting and run go vet
make lint

# Compile the release binary (deck)
make build

# Clean build binaries and coverage profiles
make clean

# Display available targets and descriptions
make help
```

---

## 4. Coverage Metrics

Statement coverage across packages:

| Package | Statement Coverage | Status |
|---|---|---|
| `deck` (root) | 65.6% | Fully covers all model & helper logic (excluding CLI `os.Exit` loop) |
| `deck/internal` | 92.5% | Passes >90% target across all subsystems |
| **Total Project** | **91.7%** | **PASSED** |

---

## 5. Performance Benchmarks

Performance targets ensure real-time terminal responsiveness:

```bash
$ make bench
BenchmarkParseDeck-8       415594       3045 ns/op        4176 B/op      24 allocs/op
BenchmarkRenderView-8        4814     262283 ns/op       68454 B/op     630 allocs/op
```

- **Markdown Parsing**: ~3.0 microseconds per slide deck.
- **Terminal Rendering**: ~0.26 milliseconds per full terminal frame (equivalent to >3000 FPS rendering capability).

---

## 6. Guidelines for Writing New Tests

1. **Use `t.TempDir()`**: When testing file I/O (`Save`, `buildModel`, image path resolution), always create test files inside `t.TempDir()` so tests remain hermetic.
2. **Strip ANSI Escape Codes**: When testing strings rendered with Lipgloss, use the `stripANSI` helper:
   ```go
   var reANSI = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
   func stripANSI(s string) string {
       return reANSI.ReplaceAllString(s, "")
   }
   ```
3. **Simulate Bubble Tea Keys**: Use `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")}` or standard key types (`tea.KeyEnter`, `tea.KeyEsc`, etc.) to trigger state transitions via `ed.HandleKey`.
4. **Format Before Committing**: Always run `make lint` prior to committing to ensure strict formatting compliance with `gofmt`.
