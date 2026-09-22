# Changelog

## v0.4 — 22 September 2026

### Added
- Distraction-Free Zen Mode (`z`): Toggle off all status bars, hints, and indicators for clean screen-sharing, video demos, and conference presenting, retaining only the hairline progress line at the bottom.
- Code Diff Highlighting (```` ```diff ````): Native syntax highlighting for git diffs and patches featuring styled green additions (`+`), red deletions (`-`), cyan hunk headers (`@@`), and dim metadata (`---`/`+++`).
- Extended Developer Syntax Highlighting: Native keyword, comment, string, and type highlighting expanded to modern systems and backend languages: Go, Rust, TypeScript, Python, SQL, and Shell.
- Native Callout & Admonition Cards: Support for GitHub-flavored markdown callouts (`> [!TIP]`, `> [!NOTE]`, `> [!WARNING]`, `> [!IMPORTANT]`, `> [!CAUTION]`, and `> quote`) rendered as rounded cards with themed border colors and contextual icons (`💡`, `ℹ`, `⚠`, `🚨`, `🛑`, `❝`).
- Quick Slide Jump Modal (`/`): Activated interactive modal for instant slide navigation with live numeric jumping (e.g. `/5` jumps to slide 5) and real-time fuzzy title search with visual match preview and laser cursor.
- Interactive Task Checklists: Native markdown task lists (`- [ ]`, `- [x]`) rendered with clean bullet markers (`○`, `✔`) and dim completed text. Press `x` in viewer mode to toggle task completion with immediate auto-save to disk.
- Native Horizontal Dividers (`***`, `___`, `::hr`): Subtle themed hairline section dividers within slides to structure complex technical ideas cleanly.
- Code Block Line Numbers (`L`): Toggle dimmed line numbers (` 1 │ `, ` 2 │ `) across code and diff blocks with dynamic gutter bounding box adjustment and `[L: lines]` status badge.
- Presentation Stopwatch & Talk Pacing Timer (`c` / `C`): Built-in elapsed talk timer displaying `[⏱ MM:SS]` (or `[⏱ H:MM:SS]`) in the status bar to assist speakers during timed tech talks, lightning talks, and sprint demos. Press `c` to toggle, `C` to reset to `00:00`.
- Live File Watch & Auto-Reload (`-w` / `--watch`): Background file monitor for live coding presentations and dual-monitor deck editing. When running with `-w`, Termdeck checks file modification timestamps every 500ms and reloads the deck instantly, strictly preserving user edit sessions if an in-app edit is active or dirty.
- Manual Deck Reload (`r` / `R`): Instantly refresh deck contents from disk at any time without leaving the presentation or losing the current slide index.
- Slide Overview & 2D Grid Sorter (`o` / `O`): Visual multi-column deck overview modal presenting all slides as structured cards with titles, block element counts (code, tables, cards, tasks, images), laser cursor focus, active slide badge, and 2D grid arrow/hjkl navigation with instant Enter-to-jump.
- Code Block & Element Yank (`y` / `Y`): Instant copy of focused code snippets, terminal commands, markdown tables, callout blocks, or paragraphs straight to the system clipboard via ANSI OSC 52 sequences (fully functional over SSH and tmux sessions) and native OS clipboard utilities (`pbcopy`, `wl-copy`, `xclip`, `clip`).
- Presentation Screen Blanking (`b` / `B`): Toggle a minimalist blackout presentation screen (`presentation paused · press any key to resume`) to redirect audience attention to the speaker during key verbal explanations; any key instantly resumes presentation view.
- Test suite expanded to **76 automated unit tests** achieving **90.6% statement coverage** in `deck/internal` with zero regressions and clean `go vet`/`gofmt`.

## v0.3 — 20 September 2026

### Added
- Standard Markdown compatibility: native support for triple-backtick code fences (```` ```lang ````) and standard image tags (`![alt](path)`).
- Markdown Tables (`| Col 1 | Col 2 |`): native parsing and formatted rendering with Lipgloss box borders and highlighted headers.
- Dynamic Theme Engine: 9 curated developer color themes (Tokyo Night, Dracula, Catppuccin Mocha, Nord, Gruvbox, Monokai, Solarized, Cyberpunk, Termdeck Pink) with live cycling (`t` / `T` / `f2`), frontmatter `theme: <name>`, custom hex colors, and `--theme` / `--list-themes` CLI options.
- Subtle Bottom Progress Line: replaced bulky status text track with an elegant, non-intrusive full-width progress line at the bottom of the screen that advances smoothly with slide progression.
- Standard CLI Flags: `--help` / `-h`, `--version` / `-v`, `--start-at <N>` / `-s <N>`, `--theme <name>` / `-t <name>`, and `--list-themes` in `main.go`.
- Test suite expanded to 45 automated unit tests achieving **92.2% statement coverage** in `deck/internal`.

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
