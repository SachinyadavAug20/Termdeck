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

#### 24 September 2026
- [x] **Live Terminal Code Runner (`X` / `ctrl+x`)**: execute focused code blocks (bash, sh, zsh, python, go, node, ruby) live during presentations with output card, green/red exit badges, and stdout/stderr capture
- [x] **Element Zoom & Focus Mode (`f` / `F`)**: maximize code snippets, architecture diagrams, and tables to full terminal viewport with dynamic line numbers, vertical scrolling (`j`/`k`), and integrated live execution
- [x] **CI/CD Automated Code Snippet Verification (`--test-code`)**: test and validate all executable code blocks across the presentation with execution time metrics and pass/fail summary
- [x] **Direct Slide Code Execution CLI (`--run-slide <N>`)**: run code blocks on slide N directly from shell scripts or automation hooks
- [x] **Non-Linear DAG Traversal Breadcrumb Tracking**: real-time journey path breadcrumbs (`[01:intro] ──► [15:hub] ──► [19:runner]`) in status bar and topology explorer
- [x] **Non-Linear Directed Graph (DAG) Presentation Engine**: transform linear slides into dynamic, interactive branching graphs
- [x] **Interactive Decision Branches**: author forks with `::branch [key] label -> target` or markdown arrows `-> [Label](target)`
- [x] **Instant Branch Jump Shortcuts**: press `1`..`9` anytime during presentation to follow a branch, or hit `Enter` on a focused branch card
- [x] **Graph Traversal History Stack**: backtrack along your visited path without losing context using `Backspace` or `H`
- [x] **Interactive Graph Map & DAG Explorer (`M`)**: visual terminal topology modal with breadcrumb path tracking, edge arrows, and Enter-to-jump
- [x] **Convergence & Non-Linear Edge Control**: `::next <slug>` and `::prev <slug>` directives to converge disparate branches back to a common conclusion
- [x] **Slide Identifiers & Markdown Header Slugs**: define custom slide targets via `::id <slug>` or standard markdown `# Title {#slug}`
- [x] **ASCII Topology Map CLI (`--graph`)**: render clean, styled ASCII DAG diagrams of slide connections directly in the terminal
- [x] **Mermaid Diagram Export CLI (`--mermaid`)**: emit standard `graph LR` diagram syntax for GitHub markdown documentation
- [x] **Interactive Offline HTML Branching**: exported standalone HTML decks feature clickable branch cards, keyboard shortcuts (`1-9`, `Backspace`), and DAG history tracking
- [x] **105 Automated Unit Tests** maintaining $\ge 90.0\%$ statement coverage in `internal/` with zero external runtime dependencies

#### 22 September 2026
- [x] Distraction-free Zen Mode (`z`) for clean presentations and video demos
- [x] Native Markdown code diff syntax highlighting (```` ```diff ````) with green additions and red deletions
- [x] Extended modern developer syntax highlighting (Go, Rust, TypeScript, Python, SQL)
- [x] Native Markdown Callout & Admonition boxes (`> [!TIP]`, `> [!NOTE]`, `> [!WARNING]`, `> [!IMPORTANT]`, `> [!CAUTION]`, `> quote`) with themed rounded borders
- [x] Quick Slide Jump modal (`/`) with instant numeric jumping and live slide title search
- [x] Interactive Task Checklists (`- [x]` / `- [ ]`) with live `x` key toggling, green checkmarks (`✔`), and auto-save
- [x] Native Horizontal Dividers (`***` / `___` / `::hr`) with subtle themed hairline styling
- [x] Code Block Line Numbers (`L`) toggleable on the fly with dimmed gutter styling
- [x] Presentation Stopwatch & Talk Pacing Timer (`c` / `C`) with live per-second ticking and status bar display
- [x] Live File Watch (`-w` / `--watch`) and manual reload (`r` / `R`) for dual-monitor workflows
- [x] Slide Overview & 2D Grid Sorter (`o` / `O`) for visual deck restructuring and fast multi-slide navigation
- [x] Code Yank to Clipboard (`y` / `Y`) with OSC 52 ANSI escape codes and system clipboard fallback
- [x] Presentation Screen Blackout (`b` / `B`) to refocus audience attention with instant any-key resume
- [x] Standalone Offline HTML Deck Export (`--export-html` flag and `E` key) with zero dependencies and embedded base64 assets
- [x] Talk Statistics & Sprint Velocity Metrics (`S` key and `--stats` CLI flag) with speaking time estimation (130 WPM) and checklist velocity bar
- [x] Auto-Advance & Rehearsal Pacing Mode (`A` key and `-a, --autoplay <sec>` flag) with live countdown badge (`[▶ auto: 5s (3s)]`), loop restart, and manual nav pause protection
- [x] 88 automated unit tests with comprehensive coverage (91.1% in internal/) and zero regressions

#### 20 September 2026
- [x] Dynamic Theme Engine (9 curated palettes + custom hex) with live cycling (`t`/`T`/`F2`)
- [x] Non-intrusive hairline slide progress indicator along bottom edge
- [x] Native Markdown tables with borders, alignment, and styled headers
- [x] Native Markdown fenced code blocks (```` ```lang ````) and syntax highlighting
- [x] Speaker notes privacy overlay (`n`) isolated from audience view
- [x] In-app keyboard shortcuts & controls help modal (`?`/`F1`)
- [x] Educational "Git Under The Hood" example presentation for CS students
- [x] 45 automated unit tests with 92.2% statement coverage

![Termdeck v0.3: Theme Engine, bottom progress line, and Git Under The Hood demo](TTP_Documentation/development/images/demo.gif)

#### 18 September 2026
- [x] Text body alignment (`left`, `center`, `right`) with `Tab`/`Ctrl+A`
- [x] High-contrast laser pointer indicator (`▶ `)
- [x] Stepped opacity heading hierarchy (pink underline for H1)
- [x] Image presentation cards with system viewer integration (`p`)
- [x] Auto-save on alignment toggle, edit confirm, and quit
- [x] Complete automated test suite (91.7% statement coverage)
- [x] Makefile developer automation (`make test`, `make coverage`, `make bench`, `make lint`)

![Termdeck v0.2: Left alignment, laser pointer, and image card](TTP_Documentation/development/images/2.png)

#### 17 September 2026
- [x] Basic start with core logic in main.go
- [x] Full-screen viewer with keyboard navigation
- [x] Inline styling, heading levels, code blocks
- [x] Block-based editor with undo/redo
- [x] Save to `.deck.md`

![Termdeck Initial Screenshot](TTP_Documentation/development/images/1.png)

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
│   ├── export.go        # Standalone HTML export, CSS/JS bundling, base64 images
│   ├── export_test.go   # HTML export tests, formatting, file output verification
│   ├── graph.go         # Directed Graph (DAG) topology, Mermaid export, ASCII map
│   ├── graph_test.go    # Graph builder, cycle detection, orphan analysis tests
│   ├── runner.go        # Subprocess live code executor (bash, python, go, etc.)
│   ├── runner_test.go   # Subprocess timeouts, dedent, CI code test assertions
│   ├── image.go         # Terminal image renderer (ANSI half-blocks)
│   ├── image_test.go    # Path resolution, format probing, card tests
│   ├── stats.go         # Talk statistics, density metrics, duration estimation
│   └── stats_test.go    # Metrics calculation, progress bar, CLI format tests
├── demo.deck.md         # Sample deck
├── git_under_the_hood.deck.md # Deep-dive example presentation
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
    │   ├── architecture.md # Technical architecture & deep-dive
    │   ├── testing.md      # Complete QA & testing guide
    │   ├── images/         # Screenshots & demos (1.png, 2.png, demo.gif)
    │   └── videos/         # Screen recordings
    └── api/                # API docs (future)
```

## Usage

```bash
deck [options] <file.deck.md>

# Options:
#   -s, --start-at <N>   Start at slide N
#   -t, --theme <name>   Set presentation theme
#       --list-themes    List all available themes
#   -w, --watch          Watch file for external changes and auto-reload
#   -a, --autoplay <sec> Auto-advance slides every N seconds (default: 5)
#       --graph          Print presentation topology map (ASCII DAG) to terminal
#       --mermaid        Print presentation topology as Mermaid diagram syntax
#       --test-code      Execute all runnable code blocks and assert zero errors (CI/CD)
#       --run-slide <N>  Execute code block on slide N directly in terminal
#       --stats          Print presentation statistics and metrics to terminal
#       --export-html    Export presentation to standalone HTML file
#   -h, --help           Show help
#   -v, --version        Show version
```

## Keys — Viewer

| Key | Action |
|-----|--------|
| `→` `l` `Space` `Enter` `PageDown` | Next slide / advance directed edge |
| `←` `h` `PageUp` | Previous slide |
| `1` – `9` | Jump directly along numbered branch / fork option |
| `Backspace` `H` | Backtrack along visited graph traversal history |
| `M` | Open interactive presentation graph map & DAG explorer modal |
| `X` `Ctrl+X` | Run focused code block live in background & show output card |
| `f` / `F` | Toggle Element Zoom & Focus Mode (full-viewport view with `j`/`k` scroll) |
| `↓` `j` | Move block cursor / laser pointer down (or scroll in Focus Mode) |
| `↑` `k` | Move block cursor / laser pointer up (or scroll in Focus Mode) |
| `/` | Quick Jump to slide (enter slide number or title search) |
| `o` / `O` | Slide Overview & 2D Grid Sorter (navigate cards, Enter to jump) |
| `?` `F1` | Toggle in-app keyboard shortcuts help modal |
| `t` `T` `F2` | Cycle color theme (`tokyo-night`, `dracula`, `nord`, etc.) |
| `z` | Toggle distraction-free zen mode (hides status bar) |
| `c` / `C` | Toggle presentation stopwatch (`c`) / Reset timer to 00:00 (`C`) |
| `A` | Toggle auto-advance slides & rehearsal pacing |
| `r` / `R` | Reload deck file from disk (manual refresh) |
| `y` / `Y` | Yank focused code block, text, or runner output to clipboard |
| `b` / `B` | Blank/blackout presentation screen (any key resumes) |
| `E` | Export deck to standalone offline HTML presentation |
| `S` | Talk statistics & sprint deck metrics modal |
| `L` | Toggle code block line numbers |
| `x` | Run focused code block live (or toggle task checklist `[ ]` ⇄ `[x]`) |
| `n` | Toggle speaker notes overlay (hidden from audience by default) |
| `Tab` `Ctrl+A` | Cycle alignment (`left` → `center` → `right`) & auto-save |
| `p` | Open focused image in system viewer |
| `G` | Last slide |
| `g` | First slide |
| `q` `Ctrl+C` | Quit (auto-saves any unsaved changes) |
| `Esc` | Dismiss runner card / exit focus mode / close modals / clear status |

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
- **Automatic saving**: Saves immediately on alignment toggle (`Tab`/`Ctrl+A`), on theme change (`t`/`T`/`F2`), on edit confirm (`Enter`), and on exit (`q`/`Ctrl+C`)
- Manual save anytime with `Ctrl+S`
- **Dynamic Theme Engine**: 9 curated color palettes (`tokyo-night`, `dracula`, `catppuccin`, `nord`, `gruvbox`, `monokai`, `solarized`, `cyberpunk`, `termdeck`) plus custom hex colors (`#3b82f6`)
- Live theme switching with `t` / `T` / `F2`
- Frontmatter theme specification (`theme: dracula`) and CLI option (`--theme <name>`, `--list-themes`)
- Inline styling: **bold**, *italic*, `code`
- Heading levels (h1–h6) with clean visual hierarchy
- **Standard Markdown Fenced Code Blocks** (```` ```lang ````) and `::code lang=X` blocks with syntax highlighting and `diff`/`patch` support
- **Code Block Line Numbers**: Press `L` anytime to toggle subtle line numbers in code and diff blocks with auto-expanding gutters
- **Presentation Stopwatch & Talk Timer**: Press `c` to toggle an active elapsed talk timer (`[⏱ MM:SS]`) with per-second live updates, and `C` to reset
- **Distraction-Free Zen Mode**: Toggle off all status bars with `z` for pure presentation focus
- **Standard Markdown Images** (`![alt](path)`) and `::image` presentation cards with system viewer integration (`p`)
- **Markdown Tables** (`| col1 | col2 |`) with formatted borders and headers
- **Callout & Admonition Cards**: Native `> [!TIP]`, `> [!NOTE]`, `> [!WARNING]`, `> [!IMPORTANT]`, `> [!CAUTION]`, and `> quote` with custom themed borders and icons
- **Interactive Task Checklists**: Native `- [x]` / `- [ ]` lists with styled green checkmarks (`✔`), dim completed state, and instant `x` key toggle
- **Horizontal Dividers**: Clean section separators (`***`, `___`, `::hr`) rendered as subtle themed hairlines
- **Quick Slide Jump Modal**: Press `/` to jump instantly by slide number or live fuzzy title search
- `::notes` speaker notes (hidden from audience canvas; toggleable presenter overlay via `n`)
- **In-App Help Modal** (`?` / `F1`) detailing all viewer, presenter, and editor controls
- **Code Block & Element Yank (`y` / `Y`)**: Copy focused code snippets, commands, tables, or text directly to system clipboard via ANSI OSC 52 (works over SSH and tmux) and native OS clipboard utilities (`pbcopy`, `wl-copy`, `xclip`, `clip`)
- **Presentation Screen Blanking (`b` / `B`)**: Temporarily blank/blackout the screen to direct audience focus to the speaker during key verbal explanations; any key instantly resumes the slide
- **Slide Overview & 2D Grid Sorter**: Press `o` or `O` anytime to open a visual grid map of all slides with titles, block element counts, cursor focus, active slide indicator, and 2D arrow/hjkl navigation
- **Live Terminal Code Runner**: Run focused shell commands, Go programs, Python scripts, Node.js, and Ruby code directly during presentations by pressing `X`, `ctrl+x`, or `x`. Terminal output is displayed in a framed card with colored exit code badges, execution time, and stdout/stderr capture. Close with `Esc` or yank output with `y`.
- **Element Zoom & Focus Mode**: Press `f` or `F` on any block to zoom into full-viewport focus. Ideal for large architecture diagrams, complex SQL queries, and multi-line code blocks. Supports vertical scrolling with `j`/`k`, dynamic line numbers, and live code execution inside focus view.
- **CI/CD Automated Code Testing (`--test-code`)**: Validate all executable code snippets across the deck in headless CI mode (`deck --test-code demo.deck.md`), returning exit code 0 if all snippets execute cleanly and non-zero on failure.
- **Direct Slide Execution CLI (`--run-slide <N>`)**: Run code blocks from slide N directly in the shell without entering the TUI.
- **Non-Linear DAG Traversal Breadcrumb Tracking**: Real-time journey breadcrumbs (`[01:intro] ──► [15:hub] ──► [19:runner]`) in the status bar and topology explorer modal, ensuring audiences and speakers never lose orientation during branching talks.
- **Non-Linear Directed Graph (DAG) Engine**: Break free from rigid linear slides! Author interactive decision forks (`::branch [key] label -> target` or `-> [label](target)`), direct numerical jumping (`1`–`9`), back-stack traversal (`Backspace` / `H`), convergence (`::next <slug>`), and interactive topology explorer modal (`M`)
- **ASCII DAG & Mermaid Diagrams**: Inspect deck topology directly in your terminal with `deck --graph <deck.md>` or export Mermaid syntax for GitHub with `deck --mermaid <deck.md>`
- **Standalone Offline HTML Export**: Press `E` or pass `--export-html` to generate a self-contained single-file HTML presentation with embedded CSS, base64 images, and interactive JavaScript navigation
- **Presentation Statistics & Deck Metrics**: Press `S` or pass `--stats` for talk duration estimates (130 WPM), code density metrics, block counts, and sprint task checklist velocity
- **Rehearsal & Autoplay Mode**: Press `A` or pass `-a, --autoplay [sec]` for automated rehearsal pacing with countdown timer, loop restart, and manual override protection
- **Live File Watch & Hot-Reload**: Start with `-w` or `--watch` to auto-reload on file edits from external editors/IDEs, or press `r` / `R` anytime to reload manually (safeguards protect active in-app edit sessions)
- **CLI Options**: `--graph`, `--mermaid`, `--test-code`, `--run-slide <N>`, `--export-html`, `--stats`, `--autoplay`, `--watch` (`-w`), `--theme <name>`, `--list-themes`, `--start-at N`, `--version`, `--help`
- Block-based editor with live editing
- Undo/redo
- Save to `.deck.md`

## Testing & Development

Termdeck features an automated test suite achieving **90.0% statement coverage** in `internal/` with 105 unit tests and 2 performance benchmarks.

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

For full details, see:
- [Technical Architecture Deep-Dive](TTP_Documentation/development/architecture.md)
- [Developer Testing & QA Guide](TTP_Documentation/development/testing.md)

## Format

See `TTP_Documentation/specs/format.md` for the full format specification.

