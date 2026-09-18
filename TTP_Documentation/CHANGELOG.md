# Changelog

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
