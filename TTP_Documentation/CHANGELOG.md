# Changelog

## v0.3 — 20 September 2026

### Added
- Standard Markdown compatibility: native support for triple-backtick code fences (```` ```lang ````) and standard image tags (`![alt](path)`).
- Markdown Tables (`| Col 1 | Col 2 |`): native parsing and formatted rendering with Lipgloss box borders and highlighted headers.
- In-App Help Modal (`?` / `F1`): centered overlay dialog detailing all viewer, presenter, and editor keybindings.
- Live Presentation Stopwatch (`t` / `Ctrl+T`): integrated speaking timer in the status line with real-time second ticks, pause, and reset.
- Subtle Bottom Progress Line: replaced bulky status text track with an elegant, non-intrusive full-width progress line at the bottom of the screen that advances smoothly with slide progression.
- Standard CLI Flags: `--help` / `-h`, `--version` / `-v`, and `--start-at <N>` / `-s <N>` in `main.go`.
- Test suite expanded to 39 automated unit tests achieving **92.3% statement coverage** in `deck/internal`.

## v0.2 — 18 September 2026

### Added
- Automated test suite reaching **91.7% statement coverage** across all packages with 32 unit tests and 2 performance benchmarks.
- Developer tooling [`Makefile`](file:///home/sachin/Projects/tpp/Makefile) with targets for `test`, `coverage`, `coverage-summary`, `bench`, `lint`, `build`, and `clean`.
- Developer testing guide (`TTP_Documentation/development/testing.md`).
- Slide text alignment options: `left`, `center`, `right` (via `Tab`, `Ctrl+A`, `::align`, or frontmatter default).
- Prominent laser pointer marker (`▶ ` in `#FF2A55`) aligned directly with the focused slide element.
- Refined heading visual hierarchy: H1 pink with underline, H2–H6 stepped white opacity fade.
- Formatted presentation image cards with dimensions, format tag, and `'p'` shortcut to open in system viewer.
- Fallback ANSI half-block image renderer in `internal/image.go`.
- Auto-save engine: Automatically persists changes to disk when toggling slide alignment (`Tab` / `Ctrl+A`), confirming live edits (`Enter`), and exiting the presentation (`q` / `Ctrl+C`). Graceful exit saving in `main.go`.
- Speaker notes privacy & overlay toggle: Multi-line notes under `::notes` directive are strictly grouped and hidden from the audience canvas, the laser pointer cursor skips notes blocks, status bar indicates note presence (`[n: notes]`), and presenters can toggle viewing notes on demand via `'n'` in a floating bottom panel.

## v0.1 — 17 September 2026

### Added
- Format spec v0.1 (`TTP_Documentation/specs/format.md`)
- Full-screen viewer with keyboard navigation
- Inline styling: **bold**, *italic*, `code`
- Heading levels (h1–h6)
- `::code lang=X` blocks with syntax highlighting
- `::image` placeholders
- `::notes` (hidden in presentation)
- Block-based editor with live editing
- Undo/redo stack
- Save to `.deck.md`
- Block reordering (Ctrl+K/J)
- Slide add/delete
