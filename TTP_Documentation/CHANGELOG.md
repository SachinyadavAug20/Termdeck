# Changelog

## v0.6 — 24 September 2026

### Added
- **Live Terminal Code Runner (`internal/runner.go`)**: Direct in-presentation code block execution for developer demos, live technical talks, and interactive workshops.
  - Supports `bash`, `sh`, `zsh`, `python` / `python3`, `go`, `node` / `nodejs` / `js`, and `ruby`.
  - Non-blocking execution via Bubble Tea background commands (`tea.Cmd`) with `context.WithTimeout` (5s default) to guarantee the presenter interface never hangs.
  - Automatic indentation stripping (`dedent()`) so indented script blocks inside markdown parse and execute properly without Python indentation errors.
  - Language whitelisting (`IsExecutableLanguage`) prevents accidental execution of structural displays (diffs, text diagrams, YAML, JSON, SQL).
  - Terminal output overlay card (`renderRunnerCard`) rendering exit status badges (`✔ exit 0` / `✖ exit 1`), elapsed execution time (`142ms`), stdout/stderr separation, and truncation protection (16KB / 300 lines limit).
  - Keystrokes: `X` or `ctrl+x` to run focused code block; `x` runs code if focused on a code block or toggles tasks on list items; `Esc` dismisses output; `y` / `Y` yanks execution output to system clipboard.
- **Element Zoom & Focus Mode (`f` / `F`)**: Full terminal viewport maximization for deep-dive inspection of wide ASCII diagrams, intricate code algorithms, or dense tables.
  - Features dynamic line numbering with gutter padding.
  - Vertical scrolling (`j`/`k` or arrow keys) for code blocks and diagrams taller than terminal viewport.
  - Seamless integrated live execution (`X` or `x`) directly inside Focus Mode with sticky output drawer.
  - Press `Esc`, `f`, or `F` to return smoothly to normal slide view.
- **Automated CI/CD Code Snippet Testing (`--test-code`)**: Headless CLI tool that executes every runnable code block across all slides in a deck and returns exit code 0 on full success or 1 if any snippet fails.
  - Displays formatted pass/fail summary (`FormatTestCodeCLI`) with execution timing for every slide and block.
  - Supports `eval=false`, `no-eval`, `no_run`, and `noexec` flags on code blocks (e.g. ```` ```bash no-eval ```` or `::code lang=bash eval=false`) to mark illustrative or destructive commands that should not be auto-run.
- **Direct Slide Execution CLI (`--run-slide <N>`)**: Allows running the code block on slide N directly from shell scripts, terminal pipelines, or automation jobs without opening the TUI.
- **Non-Linear DAG Traversal Breadcrumb Tracking**: Real-time journey path tracking rendered both in the status bar (`[01:intro] ──► [15:hub] ──► [19:runner] (step 3)`) and in the graph topology explorer modal (`M`), providing instant spatial orientation during complex branching presentations.
- **Shortest Path Graph Traversal (`DeckGraph.ShortestPath`)**: Breadth-first search (BFS) path finding between any two slides in the deck graph topology.
- **Interactive Offline HTML Live Runner Simulation**: Exported standalone HTML decks now render code blocks with `.code-header`, language badges, a Run button (`▶ Run`), and an interactive output drawer simulating live terminal output with keyboard shortcuts (`X`, `f`).
- Expanded automated test suite to **105 tests** maintaining **90.0% statement coverage** in `deck/internal` with zero regressions and zero lint warnings.

## v0.5 — 24 September 2026

### Added
- **Non-Linear Directed Graph (DAG) Presentation Engine**: Replaces traditional rigid linear slide constraints with a dynamic directed acyclic graph architecture, allowing presenters to fork into deep-dive topic tracks based on audience feedback and converge back to common conclusion slides.
- **Interactive Decision Branches**: Author presentation forks with directives (`::branch [1] Backend Architecture -> arch`, `::fork [2] Frontend UI -> ui`) or native markdown arrow links (`-> [Concurrency Patterns](concurrency)`, `=> [Memory Optimization](memory)`). Unkeyed branches are automatically assigned sequential numbers (`[1]`, `[2]`, ...).
- **Direct Numerical Branch Jumping**: Press `1` through `9` during presentation to instantly follow the corresponding branch on the slide without interrupting talk flow. Presenters can also navigate down to any branch card with the laser pointer and hit `Enter` to follow the branch.
- **Graph Traversal History Backtracking Stack (`History []int`)**: Full back-stack tracking of every slide visited during a non-linear talk. Pressing `Backspace` or `H` pops from the traversal history to return along the presenter's exact path, providing foolproof audience detour navigation.
- **Interactive Graph Map & DAG Explorer Modal (`M` key)**: Visual terminal topology modal displaying the full presentation structure, active slide marker (`●`), laser cursor selection (`▶`), outgoing branch/convergence edges, and a real-time breadcrumb traversal path (`Path: [01] ──► [02] ──► [04]`). Select any node and press `Enter` to jump instantly.
- **Edge Convergence & Directives (`::next <slug>`, `::prev <slug>`)**: Allows disparate branches to seamlessly merge back into a shared conclusion or benchmark slide, overriding default linear index progression.
- **Slide Identifiers & Markdown Header Slugs**: Define target slide identifiers with directives (`::id <slug>`) or standard markdown heading attributes (`# Title {#slug}`). Lookup resolution supports exact IDs, slugs, 1-based slide indices, and case-insensitive title substrings.
- **Terminal ASCII Topology Map CLI (`--graph`)**: Command-line flag rendering clean, colorized Unicode/ASCII DAG diagrams showing all slides, tags, outgoing edges, and unreachable orphan slide detection directly in the terminal.
- **Mermaid Diagram Export CLI (`--mermaid`)**: Emits standard GitHub-flavored Mermaid `graph LR` diagram syntax mapping the entire presentation topology for technical documentation, RFCs, and README embeds.
- **Interactive Offline HTML Branching**: Exported standalone HTML decks render branch fork cards as clickable interactive elements, with full support for numeric shortcuts (`1-9`), back history (`Backspace`), and non-linear `data-next` edge convergence.
- Expanded test suite to **over 90 automated unit tests** maintaining **90.8% statement coverage** in `deck/internal` with zero regressions and zero lint warnings.

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
- Standalone Offline HTML Deck Export (`--export-html <file.deck.md>` / `E` key): Zero-dependency, single-file HTML presentation export with responsive CSS layout matching active terminal theme palette, embedded base64 image encoding, keyboard slide navigation (`←`/`→`, `Space`, `h`/`l`, `g`/`G`, `f`), mobile swipe gesture support, and strict exclusion of private speaker notes for effortless offline deck distribution to colleagues and conference attendees.
- Talk Statistics & Sprint Velocity Metrics (`S` key and `--stats` CLI flag): Presentation intelligence computing total words, speaking time estimation at standard delivery speed (130 WPM), technical density breakdowns (code blocks, lines, tables, callout cards), sprint checklist progress bar (`[████████░░] N/M tasks (X%)`), and focused slide metrics, accessible interactively via `S` modal or non-interactively via `--stats` CLI option for terminal reporting and CI workflows.
- Auto-Advance & Rehearsal Pacing Mode (`A` key / `-a, --autoplay <sec>` flag): Hands-free auto-advance designed for lightning talks, Ignite/PechaKucha rehearsals, and conference booth ambient displays. Features per-second countdown status badge (`[▶ auto: 5s (3s)]`), configurable interval seconds, continuous deck looping, and manual navigation protection that resets the countdown so speakers are never interrupted mid-sentence.
- Test suite expanded to **88 automated unit tests** achieving **91.1% statement coverage** in `deck/internal` with zero regressions and clean `go vet`/`gofmt`.

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
