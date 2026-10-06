# Changelog

## v0.17 — 06 October 2026

### Added
- **Presentation Health Doctor (`deck doctor <file>` / `internal/doctor.go`)**:
  - Holistic pre-flight presentation diagnostics verifying everything prior to taking the stage.
  - **Graph DAG Integrity Prober**: Validates decision forks, loops, broken branch targets, and unreachable orphan slides via `LintGraph`.
  - **Asset & Image Prober**: Verifies all referenced local image assets (`![alt](path)` and column images) exist on disk relative to deck base directories and asset subfolders.
  - **Runtime Toolchain Validator**: Inspects code blocks throughout the deck and verifies that the required compilers/interpreters (`python3`, `go`, `rustc`, `gcc`/`clang`, `g++`, `deno`, `node`, `lua`, `perl`, `php`) exist in `$PATH`.
  - **Slide Viewport Fit & Readability**: Alerts on vertical overflows (>32 lines) or line lengths exceeding 100 columns that might clip or wrap awkwardly on standard terminal projectors.
  - Beautiful color-coded terminal report with severity badges (`[ERROR]`, `[WARN]`, `[INFO]`, `[PASS]`) and actionable remediation tips. Exits with code 1 on fatal errors for CI automation.
- **Canonical Presentation Markdown Formatter (`deck fmt` / `internal/format.go`)**:
  - Automated presentation code and layout formatter (`deck fmt [--check] <file>`).
  - Canonicalizes presentation directive syntax: `::tags: ...`, `::branch [key] label -> target`, `::track: ...`, `::route: ...`, `::loop: ...`, `::bg: ...`, `::align: ...`, `::note: ...`.
  - Standardizes slide separators (`---`) and collapses excessive consecutive blank lines.
  - Trims trailing whitespace on all lines while strictly preserving code fences, indentation, and code block internals.
  - Supports `--check` mode to verify formatting in CI/CD pre-commit workflows without modifying files.
- **Native Multi-Language Code Runner Expansion (`internal/runner.go`)**:
  - Expanded live code runner to natively compile and execute:
    - **Rust (`rust`, `rs`)**: Auto-wraps in `fn main()` if omitted, compiles with `rustc`, executes binary, captures stdout/stderr, and cleanly purges temp artifacts.
    - **C (`c`)**: Auto-wraps in `#include <stdio.h>\nint main(void)` if omitted, compiles with `gcc`/`clang`, executes binary, and cleans up.
    - **C++ (`cpp`, `c++`, `cc`)**: Auto-wraps in `#include <iostream>\nint main()`, compiles with `g++`/`clang++`, executes binary, and cleans up.
    - **TypeScript (`ts`, `typescript`)**: Auto-dispatches to `deno eval`, `bun run -e`, `tsx -e`, or `ts-node -e`.
    - **Lua (`lua`, `luajit`)**: Native execution via `lua -e` / `luajit -e`.
    - **Perl (`perl`, `pl`)**: Native execution via `perl -e`.
    - **PHP (`php`)**: Native execution via `php -r`.
- **Configuration Subsystem Overhaul & Bug Fixes (`internal/config.go`)**:
  - Fixed pointer receiver bug in `loadConfigFile` where configuration values parsed from TOML/YAML files were mutated on a value copy and discarded.
  - Added exported `LoadConfigFromPath(path)` and pre-parsed `--config <path>` support in `runCLI` so configuration files take precedence over default settings.
  - Added serialization of `show-dot` in `SaveConfig`.
  - Added full test suite in `internal/config_test.go` achieving 100% test coverage for default config, file parsing, environment variable overrides, and save round-trips.
- **Testable CLI Architecture & Expanded Test Suite**:
  - Refactored `main()` into `runCLI(args []string, stdout, stderr io.Writer) int`, allowing end-to-end testing of all CLI commands and flag combinations.
  - Boosted statement coverage of package `deck` from 13.6% to **84.5%**, `deck/internal` to **89.5%**, and overall repository test coverage to **89.2%**.

## v0.16 — 04 October 2026

### Added
- **Universal Graphviz DOT Export (`--dot` / `--graphviz`)**:
  - Implemented `ToGraphvizDOT()` and `ToGraphvizDOTWithTrack(track)` generating standard Graphviz DOT digraph syntax with rounded box styling, semantic color-coded borders (green for entry roots, red for terminal sinks, cyan for active track slides), dashed loop arrows, and bold exit connectors.
  - Added CLI flags `--dot` and `--graphviz` for piped diagram generation with Graphviz `dot`, OmniGraffle, PlantUML, and Obsidian pipelines.
- **Topological Analysis & Metrics Engine (`TopologyMetrics` / `AnalyzeTopology`)**:
  - Automatically calculates graph diameter and longest acyclic path sequence via Kahn's algorithm and dynamic programming to avoid cyclic traps.
  - Computes root/entry slides, terminal sinks, decision forks, convergence joins, loop cycles, cyclomatic complexity ($M = E - V + 2P$), and average branching factor.
  - Displays topology summary in CLI ASCII DAG graphs (`--graph`) and TUI Graph Map modal (`M`).
- **Interactive TUI Graph Map Modal Overhaul (`M`)**:
  - **3 View Modes (`v` to cycle)**:
    - **Mode 0: Tree Flow View**: Visual ASCII connection tree rendering branches, forks, loops (`⟳ [key]`), and exits (`exit ──►`).
    - **Mode 1: Detailed List View**: In/out degrees, tags, and real-time shortest BFS path preview from current slide (`N hops from current`).
    - **Mode 2: Topology Metrics Dashboard**: Full-screen telemetry dashboard visualizing nodes, edges, roots, sinks, forks, joins, longest path sequence, cyclomatic complexity rating, and audience tracks breakdown.
  - **Real-Time Interactive Search Filtering (`/`)**: Type search queries with live character buffering, backspace deletion, and instant list pruning across titles, tags, and slide numbers (`c` to clear).
  - **Audience Track Quick Cycling (`t` / `Tab`)**: Cycle through all audience tracks directly within the graph modal.
  - **Instant Clipboard Export Shortcuts**: Press `y` to copy the presentation Mermaid diagram or `d` to copy the Graphviz DOT digraph directly to the system clipboard via ANSI OSC 52 / host clipboard utilities.
  - **Direct Waypoint Pathfinder Launch (`w`)**: Launch the Waypoint BFS pathfinder for the selected destination slide directly from the Graph Map.
- **Loop Edge Distinction & Exit Reachability (`EdgeLoop` / `EdgeLoopExit`)**:
  - Distinguishes loop iteration edges (`EdgeLoop`) and loop exit transitions (`EdgeLoopExit`) in `BuildGraph`, resolving reachability so slides following an iteration loop are properly linked and reachable without false-positive orphan warnings.
  - Added loop exit target validation and duplicate branch shortcut key detection to DAG linter (`LintGraph`).

## v0.15 — 28 September 2026

### Added
- **Interactive Fuzzy Command Palette (`:` / `Ctrl+P`)**:
  - Solves shortcut discoverability for both novice presenters and seasoned power users.
  - Floating searchable modal indexing 30+ navigation, graph, developer tool, editor, and system commands.
  - Real-time case-insensitive filtering matching across command titles, hotkey shortcuts, functional categories (`Navigation`, `Graph`, `Tools`, `Help`, `Editor`, `System`), and detailed descriptions.
  - Ergonomic keyboard navigation: `↑` / `↓` cursor motion with wrap-around, character typing with real-time prompt cursor, `Backspace` correction, `Enter` instant execution, and `Esc` cancellation.
  - Responsive pagination: 8-item viewport window with dynamic scroll indicators (`▲ more commands above`, `▼ N more commands below`).
  - Status line integration: added `: menu` hint to global presentation status bar.
- **Starter Deck Scaffolding (`deck init [filename]`)**:
  - Seamless zero-friction onboarding: `deck init [filename.deck.md]` bootstraps a fully structured, lint-verified starter presentation template.
  - Template includes Tokyo Night color theme, multi-format talk routes (`quick`, `full`), audience track tags, decision branching links (`::branch`), and an executable live bash code block.
  - Compiler-grade safety: guarantees 0 DAG linting diagnostics on new templates, and guards against accidental file overwrites.
  - User-friendly CLI diagnostics: running `deck` without arguments immediately offers `deck init [filename.deck.md]`.
- **Audience Track & Tag Awareness in Overview Grid (`o`)**:
  - Slide overview 2D thumbnail cards now render audience track tags (`[tag1,tag2]`) directly in each card's summary line, giving presenters instant spatial orientation of filtered subgraphs.
- Expanded automated test suite to **161 unit tests** maintaining **90.5% statement coverage** in `deck/internal` with zero external runtime dependencies and sub-millisecond per-frame rendering (< 0.26ms).

## v0.14 — 27 September 2026

### Added
- **Responsive Viewport & Screen Aspect Ratio Adaptation Engine**:
  - **Stage Canvas Bounding & Horizontal Centering**: Implements standard TUI best practices for widescreen, 16:9, and 21:9 ultrawide monitors. Clamps maximum presentation canvas width to optimal reading typography widths (`maxCanvasW = 104` standard, `124` for multi-column grids) and centers the slide horizontally via `lipgloss.PlaceHorizontal`, eliminating awkward text stretching on wide monitors.
  - **Responsive Multi-Column Fallback**: In `renderColumns`, automatically stacks side-by-side columns vertically with subtle dashed dividers (`┄`) whenever terminal width is narrow or column width drops below threshold (`colW < 20 || w < 50`), preserving code and table readability without horizontal clipping.
  - **Dynamic Indentation Padding**: Automatically adjusts left/right alignment padding (`padLeft/Right = 2` on compact terminals < 70 columns, `4` on standard/wide screens).
  - **Vertical Centering Overflow Safeguard**: Retains centered vertical alignment by default, but dynamically shifts to top alignment (`lipgloss.Top`) when slide content height meets or exceeds terminal body height, ensuring titles and top blocks never get pushed offscreen.
  - **Small Terminal Dimensions Guard**: Safeguards against unreadable viewports (`width < 36 || height < 6`), displaying a clear resize prompt rather than garbled frames.
- **5-Page Interactive Learning & Help Hub (`?` / `F1`)**:
  - Replaces the single-page shortcuts popup with a full-fledged, multi-page onboarding and learning hub with tabbed navigation:
    - **Tab 1: `[1:Navigation]`**: Slide advancement, laser pointer cursor, title/number jump modal, 2D thumbnail grid overview, blackout screen, presentation talk timer, zen mode, line numbers, image card viewer.
    - **Tab 2: `[2:Graphs]`**: Non-linear DAG branching, branch decision fork preview HUD (`J`), shortest-path waypoint pathfinder (`W`), graph exploration radar & budget gauge (`V`), upstream fork fast-return (`U`), traversal history reflog modal (`H`), ASCII topology map (`M`), audience track hops (`[`/`]`), bounded cycle engine (`::loop`).
    - **Tab 3: `[3:Tools]`**: Live subprocess code runner (`X`), interactive task checklist toggle (`x`), element viewport zoom & focus mode (`f`), preset talk routes (`P`), audience tracks (`K`), standalone HTML export (`E`), speaking pace metrics & stats (`S`), hands-free auto-advance (`A`), ANSI OSC 52 clipboard copy (`y`), color theme cycles (`t`).
    - **Tab 4: `[4:Editor]`**: Live in-slide editor (`i`), confirm edit (`Enter`), cancel (`Esc`), block additions (`^n`), block deletions (`^d`), reordering (`^k`/`^j`), undo/redo (`u`/`^r`), manual save (`^s`).
    - **Tab 5: `[5:Guide]`**: Complete Markdown authoring syntax cheat sheet (`---`, `{#slug}`, `::branch`, `::loop`, `::next`, `:::columns`, `::tags`, frontmatter `routes:`, code blocks, modern callout cards `[!TIP]`, private speaker notes `::notes`).
  - **Fluid Keyboard State Machine**:
    - `Tab` / `l` / `→` / `n` / `pgdown`: Next tab with cycle wrap
    - `Shift+Tab` / `h` / `←` / `p` / `pgup`: Previous tab with cycle wrap
    - `1` – `5`: Direct jump to corresponding tab page
    - `Esc` / `q` / `?` / `F1`: Close hub
- Expanded automated unit test suite to **158 tests** maintaining **90.7% statement coverage** in `deck/internal` with zero external runtime dependencies and sub-millisecond per-frame rendering (< 0.26ms).

## v0.13 — 26 September 2026

### Added
- **Bounded Graph Cycles & Presentation Loop Iteration Engine (`::loop` / `::cycle`)**:
  - Extends Termdeck's presentation DAG graph with bounded cycle modeling for computational workflows, algorithmic cycles, engineering feedback loops, and state machine iterations (e.g. TDD Red-Green-Refactor, exponential retry backoff, Raft consensus rounds, ML training epochs).
  - **Directive Syntax**:
    - `::loop [key] Label -> target max=N next=exit_slug`
    - `::cycle [c] Label => target limit=N exit=exit_slug`
    - Supported parameters: `max=N` / `limit=N` / `passes=N` / `count=N` (defaults to 3 passes), `next=slug` / `exit=slug` / `break=slug` (automatic exit target upon completion).
  - **Runtime Iteration State Machine**:
    - `LoopCounters: map[int]int` in `Editor` tracking completed iteration passes per slide.
    - Advancing via standard presentation keys (`Space`, `Enter`, `right`, `l`, `pgdown`) automatically increments pass count and routes back to `Target` while `pass < max`.
    - Auto-Exit Guarantee: once passes reach the threshold (`pass >= max`), subsequent advance seamlessly exits the loop to `ExitTarget` (or linear next slide), preventing infinite presentation traps.
    - Hotkey execution: directly trigger loop pass via optional shortcut key (e.g. `[r]` or `[1]`).
    - Presentation restart (`g`) automatically resets all loop counters to zero.
  - **Slide Loop Card & Navigation Badging**:
    - Active loop card: displays styled box with theme border, current pass progress (`[pass 1/3] · 2 remaining`), target link (`──► #slug`), and navigation key hints.
    - Completed loop card: displays success badging (`✔ LOOP COMPLETED: Label (3/3 passes)`) and exit prompts.
    - Navigation status bar: displays real-time loop telemetry (`[⟳ loop: pass 1/3 (TDD Loop ──► tdd-red)]` or `[✔ loop: 3/3 done]`).
  - **Offline HTML Export Integration**:
    - Exported HTML presentations (`deck --export-html`) include styled `.loop-card` badges with interactive JavaScript click-to-jump.
  - **Graph Topology & Linter Integration**:
    - Loop branches integrated into `Slide.Branches()`, `BuildGraph()`, `GetForkOptions()`, and `LintGraph()`.
- Expanded automated test suite to **155 tests** maintaining **90.7% statement coverage** in `deck/internal` with zero external runtime dependencies and sub-millisecond per-frame rendering (< 0.26ms).

## v0.12 — 25 September 2026

### Added
- **Graph Exploration Radar & Branch Completion Matrix (`V` / `--radar` / `--coverage`)**:
  - Presentation DAG exploration telemetry gauge and completion matrix designed to solve non-linear presentation spatial tracking and time budgeting.
  - **Global Graph Exploration Gauge**:
    - High-contrast Unicode progress bar rendering visited unique slide percentage (`████░░░░ 44.4%`).
    - Speaking time budgeting: tracks estimated minutes delivered (~130 WPM) vs total deck speaking time and unvisited remaining budget.
    - Decision fork summary: tallies total forks, branches, completed paths, in-progress paths, and unvisited paths.
  - **Branch Completion Matrix**:
    - Evaluates every outgoing branch from presentation decision hubs and partitions branch-exclusive nodes from common downstream convergence points.
    - Status indicators: `✔` Completed (`100%`), `◐` In-progress (`X%`), and `○` Unvisited (`~Nm remaining`).
    - Destination IDs (`──► #slug`) and remaining duration estimates per branch.
  - **Interactive Radar Navigation**:
    - Browse branches with `j`/`k`/arrow keys, `g`/`G` for top/bottom.
    - `Enter`: Jumps directly into selected branch destination.
    - `u`: Jumps immediately to that branch's parent decision fork hub.
    - `1`..`9`: Instant numeric branch execution.
    - `Esc`/`q`/`V`: Dismisses radar modal.
- **Upstream Decision Fork Fast-Return (`U`)**:
  - One-key instant backtrack (`U`) that teleports the presenter back to the nearest upstream decision fork where branches diverged.
  - Eliminates tedious sequential `Backspace` tapping during live Q&A deep-dives and modular demos.
  - Status bar prompt: displays `[U: return to fork]` when inside a sub-branch.
- **CLI Graph Exploration Radar (`--radar` / `--coverage`)**:
  - Terminal ASCII diagnostics reporter printing the Graph Exploration Matrix and branch completion summary for automated verification or talk rehearsal.
- Expanded automated test suite to **152 tests** maintaining **90.9% statement coverage** in `deck/internal` with zero external runtime dependencies and sub-millisecond per-frame rendering (< 0.26ms).

## v0.11 — 25 September 2026

### Added
- **Interactive Waypoint Pathfinder & Shortest-Path Graph Router (`W`)**:
  - Dynamic BFS shortest-path graph routing engine allowing presenters and audiences to chart optimal paths across complex non-linear presentation DAGs on demand.
  - **Interactive Modal Overlay (`W`)**:
    - Real-time search and filter destinations by title, slide ID (`#slug`), audience tags (`::tags`), or slide number (`1`-`N`).
    - Displays current slide as origin (`Origin: [01] Current Title`).
    - Computes and visualizes complete multi-hop edge transitions (`[01:hub] ──[1]──► [04:arch] ──(next)──► [09:target]`).
    - Reachable candidates sorted ascending by hop distance for instant selection of closest destinations.
    - Unreachable destinations clearly flagged with rewind guidance (`(No downstream path from current slide · Rewind via Backspace / H)`).
    - Speaking time budgeting: Estimates talk duration (~130 WPM) and hop counts per route.
  - **Dual Execution Modes**:
    - `Enter`: Locks the optimal shortest path into a temporary dynamic presentation route (`waypoint-N`), enabling guided step-by-step traversal with standard keys (`Space`, `Enter`, `Left`, `Backspace`).
    - `w`: Steps immediately 1 hop along the optimal path without route locking.
    - Keyboard ergonomics: `j`/`k`/arrows to scroll, `g`/`G` for top/bottom, alphanumeric typing to filter query, `Backspace` to delete query characters or close on empty query, `Esc`/`W` to dismiss.
- **Pathfinder Integration & Navigation Synergy**:
  - Navigation status bar footer includes `W waypoint` indicator.
  - In-app Help modal (`?`/`F1`) and CLI help updated to document `W` waypoint pathfinder.
  - `demo.deck.md` updated with `waypoint-deepdive` slide, branch `[9]`, and route integrations.
- Expanded automated test suite to **151 tests** maintaining **90.8% statement coverage** in `deck/internal` with zero external runtime dependencies and sub-millisecond per-frame rendering (< 0.25ms).

## v0.10 — 25 September 2026

### Added
- **Branch Decision Fork HUD & Target Preview Picker (`J`)**:
  - Interactive floating HUD overlay invoked at any branching decision point during presentations.
  - Lists all outgoing path choices with direct jump keys (`[1]`, `[2]`, `[→]`), destination slide indices, titles, and slide IDs (`#slug`).
  - Real-time syntax-highlighted **code & text preview panel** rendered within the HUD displaying destination code snippets and markdown text before jumping.
  - Controls: `j`/`k`/arrow keys to browse choices, `Enter`/`Space` to commit jump, `1`..`9` for instant branch execution, `Esc`/`q`/`J` to dismiss.
- **Downstream Subgraph Metrics & Time Budgeting**:
  - Analyzes downstream reachable subtrees for every outgoing branch path:
    - **Reachable Slide Depth**: total number of slides downstream.
    - **Estimated Speaking Duration**: talk time in minutes computed from downstream word counts (~130 WPM).
    - **Code Snippet Counter**: total executable and display code blocks in downstream path.
    - **Track Match Indicators**: `★ Track Match` badge when destination slide matches active audience track (`-k, --track`).
    - **Route Sync Indicators**: `⚡ Route Step` badge when destination aligns with active preset route (`--route`).
- **HUD & Status Bar Ergonomics**:
  - Status bar displays `[fork: N paths (J)]` when branches exist on slide, plus quick-hint `J fork`.
  - In-app Help modal (`?`/`F1`) and CLI help updated to document `J` decision fork HUD.
- Expanded automated test suite to **150 tests** maintaining **90.5% statement coverage** in `deck/internal` with zero external runtime dependencies and sub-millisecond per-frame rendering.

## v0.9 — 25 September 2026

### Added
- **Traversal History & Graph Reflog Modal (`H`)**:
  - Visual presentation journey stack explorer allowing presenters to review every slide visited across branching decision forks and guided paths.
  - Interactive Modal:
    - Lists visited slides in exact temporal order with step numbering (`[1]`, `[2]`, ...).
    - Relative distance indicator for prior steps (`(3 steps back)`).
    - Highlights current slide with `● CURRENT` badge.
    - Rewind to any prior historical step by pressing `1`..`9` or selecting with cursor and hitting `Enter`.
    - Instant stack clear (`c`) or step pop (`Backspace`).
    - Dismiss via `Esc`, `q`, or `H`.
  - Traversal Stack Helpers (`JumpToHistory`): rewinds presenter to target slide and cleanly truncates forward stack.
- **Graph Topology Linter & CI/CD Diagnostics (`--lint` / `--lint-graph`)**:
  - Static DAG validation engine (`internal.LintGraph`) detecting:
    - Broken branch targets (`::branch ... -> missing`)
    - Broken edge targets (`::next missing`, `::prev missing`)
    - Broken route steps (`routes:` referencing nonexistent slide slugs)
    - Duplicate slide identifiers (`::id duplicate`)
    - Unreachable orphan slides (isolated from root slide 0)
    - Dead-end slides (slides before final conclusion with zero outgoing edges)
  - Compiler-style terminal diagnostics reporter (`internal.FormatLintCLI`) with `✖ ERROR` and `⚠ WARN` badges.
  - Returns exit code 0 when DAG is valid, or exit code 1 when fatal errors exist, making it ideal for automated GitHub Actions and CI pre-presentation checks.
- **Traversal Indicators & Navigation Synergy**:
  - Status bar indicator: `[history: N (H)]` and real-time breadcrumb journey trail `[path: [01:intro] ──► [15:hub] ──► [22:history]]`.
  - Dual history controls: `Backspace` pops 1 step immediately, while `H` opens the Visual Reflog Modal.
- Expanded automated test suite to **147 tests** maintaining **90.5% statement coverage** in `deck/internal` with zero external runtime dependencies and sub-millisecond per-frame rendering.

## v0.8 — 25 September 2026

### Added
- **Preset Graph Routes & Guided Paths (`P` / `--route`)**:
  - Non-destructive presentation route overlay engine enabling presenters to pre-configure tailored walkthrough paths across complex non-linear DAGs (e.g. `lightning`, `deepdive`, `workshop`).
  - Supports dual authoring syntax:
    - Deck Frontmatter: `routes:` mapping (or `route.<name>:`) with arrow (`->`, `=>`) or comma-separated slide slugs.
    - Inline Directives: `::route <name>: <slug1> -> <slug2> -> <slug3>` anywhere in the deck body.
    - Round-trip serialization (`SerializeDeck`) preserving custom `::route` directives.
- **Route Switcher Modal (`P`)**:
  - Dedicated interactive terminal modal displaying available preset routes.
  - Real-time word count calculation and speaking duration estimation (at 130 WPM) for each route.
  - Visual breadcrumb path previews (`[01:intro] ──► [04:arch] ──► [22:conclusion]`) showing exact slide sequence.
  - Navigation controls: `j`/`k`/arrow keys to select, `1`–`9` to quick-activate, `0` to clear active route and return to free graph traversal, `Enter` to apply, and `Esc`/`q`/`P` to dismiss.
- **Presenter Route Navigation Engine**:
  - Advancing with `Space` / `Enter` / `Right` / `l` / `PageDown` automatically steps forward along the active route's pre-planned slide sequence.
  - Backstepping with `Left` / `h` / `PageUp` moves backwards along the route.
  - Maintains full compatibility with interactive decision forks (`::branch` / `1`–`9`) and the traversal history stack (`Backspace` / `H`), allowing spontaneous detours without losing route state.
  - Status bar indicator: `[⚡ route: <name> (step X/Y)]` with quick-access hint (`P route`).
- **Graph Topology & Diagram Integration**:
  - Graph Modal (`M`): displays active route banner and marks route slides with step numbers (`#1`, `#2`, etc.).
  - Terminal ASCII DAG Map (`deck --route=<name> --graph`): displays route step badges (`⚡ step N`) on connected slides.
  - Mermaid Export (`deck --route=<name> --mermaid`): highlights route nodes with custom styling class (`classDef routeNode fill:#f59e0b,stroke:#d97706,...`).
- **CLI Guided Path Integration**:
  - Launch presentations directly into a preset route: `deck --route <name> presentation.deck.md` or `deck --route=<name>`.
- Expanded automated test suite to **143 tests** achieving **90.2% statement coverage** in `deck/internal` with zero external runtime dependencies and sub-millisecond per-frame rendering.

## v0.7 — 25 September 2026

### Added
- **Multi-Column Side-by-Side Split Grids (`BlockColumns`)**:
  - Author split layouts with `:::columns` (or `:::split` / `::split`) and column dividers `:::col` (or `::col`).
  - Supports recursive block parsing inside columns: any valid Markdown construct (headings, text, code blocks, diffs, tables, callout admonitions, tasks, images) works within each column.
  - Dynamically calculates column widths `(viewportWidth - totalGaps) / columnCount` with horizontal joining via `lipgloss.JoinHorizontal`.
  - Laser cursor alignment and element zoom focus mode (`f` / `F`) support for multi-column structures.
  - Live code runner execution (`X`, `x`) detects and executes code blocks situated within columns.
  - Headless CI automated testing (`--test-code`) traverses and validates code blocks nested inside multi-column grids.
  - Standalone HTML export (`ExportHTML`) emits responsive CSS Flexbox `.columns-grid` with `.column-item` cells and mobile breakpoint fallback.
- **Audience Tracks & Subgraph Filtering**:
  - Tag slides with `::tags <tag1>,<tag2>` in `.deck.md` files.
  - Interactive Track Selection Modal (`K`): dedicated terminal modal listing all unique tags in the deck with slide counts and active indicators. Quick-select tags via `0` (all slides) or numbers `1`–`9`.
  - Sequential Track Navigation (`[` and `]`): hop directly between slides tagged with the active track without altering the deck's underlying slide order.
  - Status bar indicator: displays active track badge `[★ track: <name>]`.
  - Graph Modal (`M`) track integration: displays active track summary header and highlights matching track nodes with `★` and accent badges.
  - Subgraph Extraction (`DeckGraph.FilterByTag`): filters full DAG topology to isolate only nodes belonging to a designated audience track.
  - Mermaid Export (`g.ToMermaidWithTrack`): annotates track nodes with `classDef trackNode` and `class nodeX trackNode;` styling.
  - Terminal ASCII DAG Map (`FormatGraphCLIWithTrack`): displays `[Track: <name>]` in the header and badges track nodes with `★ <name>`.
- **CLI Options**:
  - `-k, --track <name>` and `--track=<name>`: launches TUI presentation locked to an audience track, or scopes `--graph` ASCII DAG / `--mermaid` diagrams.
- Expanded automated test suite to **138 tests** achieving **90.5% statement coverage** in `deck/internal` with zero external dependencies beyond Bubble Tea and Lipgloss.

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

## v0.15.1 — 03 October 2026

### Added
- **Configuration File Support** (`deck --config`):
  - Load presentation settings from `~/.termdeck/config.toml` or `./termdeck.toml`/`.yml`.
  - INI-style `key: value` format supporting theme, autoplay delay, tracks, routes, and more.
  - Environment variables (`TERMDECK_THEME`, `TERMDECK_AUTOPLAY`, etc.) override config file values.
  - `--config <path>` CLI flag to specify custom config file location.
- **Expanded Language Support** in Live Code Runner:
  - Added 13 new executable languages: rust, java, scala, kotlin, c, c++, csharp, php, perl, r, lua, matlab.
  - Code blocks can now specify `Env` field for execution environment variables.
- **Configuration Section in Core Highlights**:
  - Documentation updated to describe config file usage and environment variable overrides.
