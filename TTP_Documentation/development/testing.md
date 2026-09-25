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
│   ├── editor_test.go       # Navigation, edit mode, undo/redo, block mutations, DAG history, Focus Mode
│   ├── export_test.go       # HTML export formatting, CSS/JS bundling, interactive branching, live code runner
│   ├── graph_test.go        # DAG topology construction, Mermaid export, shortest path, breadcrumbs
│   ├── runner_test.go       # Live code runner subprocess execution, timeouts, dedent, CI assertions
│   ├── view_test.go         # Syntax highlighting, inline styles, DAG modal, runner card, Focus Mode
│   ├── model_test.go        # Parser edge cases, branching syntax, round-trip serialization, no-eval flags
│   ├── image_test.go        # Image path resolution, format probing, ASCII cards
│   └── stats_test.go        # Talk statistics, progress bar, CLI format tests
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
- `TestPrintHelp`: Verifies CLI help message and key controls formatting.
- `TestPrintHelpExportHTML`: Asserts `--export-html` is documented in CLI help output.
- `TestPrintHelpStats`: Asserts `--stats` and `S` control are documented in CLI help output.
- `TestPrintHelpAutoplay`: Asserts `-a` / `--autoplay` and `A` control are documented in CLI help output.
- `TestPrintHelpGraphAndMermaid`: Asserts `--graph` and `--mermaid` CLI options are documented in help output.
- `TestPrintHelpLiveRunnerAndFocus`: Asserts --test-code, --run-slide, X, and f shortcuts are documented in help output.
- `TestPrintHelpTrack`: Asserts `-k, --track`, `K`, and `[ / ]` audience track options are documented in help output.
- `TestPrintHelpRoute`: Asserts `--route`, `P`, and `0` route controls are documented in help output.
- `TestPrintHelpLint`: Asserts `--lint` and `--lint-graph` CI/CD topology checker flags are documented in help output.
- `TestModelExecFinishedMsg`: Verifies `ExecFinishedMsg` delivers results cleanly to editor state without crashing.
- `TestThemeFlagAndListThemes`: Tests `--theme <name>` and `--list-themes` CLI flags.
- `TestModelUpdateTickMsg`: Verifies Bubble Tea model dispatches `TickCmd()` only when timer is enabled, avoiding background polling overhead.
- `TestModelInitWatchMode`: Asserts `WatchCmd()` is initialized when `-w` / `--watch` CLI flag is set.
- `TestModelUpdateWatchMsg`: Validates background disk polling and auto-reload on file modification.
- `TestModelAutoplay`: Verifies automated slide advancement and countdown ticking on `TickMsg`.

### B. Editor State Machine — `internal/editor_test.go`
Tests the navigation, editing, jumping, and toggling state machine:
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
- `TestAutoSaveFeatures`: Verifies immediate disk persistence when toggling alignment (`Tab`/`Ctrl+A`), auto-saving edited text on confirm (`Enter`), and auto-saving dirty buffers on exit (`q`/`Ctrl+C`).
- `TestSpeakerNotesEditor`: Ensures block cursor skips hidden speaker notes blocks during up/down movement.
- `TestEditorHelpModal`: Tests opening and dismissing keyboard shortcuts modal (`?`, `F1`, `Esc`).
- `TestEditorCycleTheme`: Validates forward and reverse theme cycling (`t`, `T`, `F2`) and auto-save persistence.
- `TestEditorCycleThemeInEditMode`: Asserts theme cycling works seamlessly even while typing inside edit mode.
- `TestEditorZenMode`: Verifies toggling distraction-free Zen mode (`z`).
- `TestEditorQuickJumpPrompt`: Verifies `/` prompt activation, typing query/numbers, slide jump execution on Enter, and cancellation on Esc.
- `TestEditorToggleTask`: Validates interactive toggling of task checklists (`- [ ]` $\leftrightarrow$ `- [x]`) with `x` key and auto-save to disk.
- `TestEditorToggleLineNumbers`: Validates pressing `L` toggles code line numbers mode on and off with status message updates.
- `TestEditorToggleTimer`: Validates toggling timer on/off with `c`, starting ticker cmd, and resetting timer to 00:00 with `C`.
- `TestEditorReload`: Tests manual deck reload from disk, index clamping, and dirty session preservation.
- `TestEditorReloadKeyNav`: Asserts pressing `r` / `R` in navigation mode triggers reload and updates status bar.
- `TestEditorSlideOverview`: Validates opening overview modal (`o`/`O`), 2D navigation (arrows/hjkl), Enter-to-jump, and Esc cancellation.
- `TestEditorYankAndBlankScreen`: Asserts `y`/`Y` extracts focused text to OSC 52 sequence and `b`/`B` blanks presentation screen.
- `TestEditorExportHTMLKeyNav`: Verifies pressing `E` in navigation mode invokes HTML exporter and sets status bar confirmation.
- `TestEditorAutoplay`: Validates toggling autoplay with `A`, per-second countdown ticking, automatic slide advancing, looping to start, and manual navigation reset.
- `TestEditorBranchAndGraphNavigation`: Validates numeric branch jumping (`1`..`9`), `enter` on focused branch cards, `Backspace` / `H` history backtracking stack, and `M` graph map modal cursor navigation and jumping.
- `TestEditorHistoryModal`: Validates opening traversal history modal (`H`), navigating steps (`j`/`k`/`g`/`G`), rewinding directly to any step (`Enter`), quick numeric jump (`1`..`9`), clearing history (`c`), and dismissing (`Esc`/`q`/`H`).
- `TestEditorBranchHUDModal`: Validates opening branch decision HUD (`J`), navigating options (arrows/jk), live preview switching, executing jump via Enter or numeric `1`..`9` keys, terminal slide handling, and dismiss keys.
- `TestEditorFocusModeAndLiveRunner`: Validates toggling Focus Mode (`f`), line numbering, vertical scrolling (`j`/`k`), running code via Bubble Tea command (`X`/`x`), yanking runner output (`y`), and dismissing runner card (`Esc`).

### C. View & Syntax Highlighter — `internal/view_test.go`
Tests visual layout, card rendering, and terminal text styling:
- `TestHighlightLine`: Syntax highlighter lexer covering:
  - Comments (`//`, `#`, `--`)
  - String literals (`"..."`, `'...'`, `` `...` ``)
  - Numeric literals (`42`, `3.14`)
  - Keywords for Go, Rust, TypeScript, Python, SQL, and Shell
  - Diff chunks (`+`, `-`, `@@`, `---`, `+++`)
  - PascalCase types
- `TestHighlightCode`: Multiline code and diff syntax highlighting with optional line numbers.
- `TestInlineStyle`: Markdown inline styling (`**bold**`, `*italic*`, `` `code` `` spans).
- `TestRenderHeadingLevels`: Hierarchy rendering for H1 (pink underline) and H2–H6 (stepped white opacity).
- `TestRenderLaserPointer`: Verifies the laser pointer marker (`▶ ` in `#FF2A55`) renders on the active line.
- `TestRenderImageCard`: Formatted image card with dimension metadata and open hints.
- `TestRenderBlockVariants`: Renders headings, paragraphs, code blocks with language tags, image cards, list items, and directives in both display and edit modes.
- `TestStatusBars`: Validates `navStatus` (slide position, alignment badge, dirty indicator) and `editStatus` (cursor column position, messages).
- `TestViewDimensionsAndModes`: Validates fallback on zero dimensions, edit mode rendering, and alignment padding.
- `TestSpeakerNotesView`: Verifies notes overlay rendering at bottom of screen.
- `TestTableAndHelpModalView`: Verifies Markdown table border styling and help modal formatting.
- `TestRenderProgressLine`: Tests full-width hairline progress line math and styling.
- `TestViewZenMode`: Asserts status bars are omitted in Zen mode while progress line remains.
- `TestRenderCallouts`: Asserts rounded borders and icons for `[!TIP]`, `[!NOTE]`, `[!WARNING]`, `[!IMPORTANT]`, `[!CAUTION]`, and quotes.
- `TestRenderJumpModal`: Tests layout of quick slide jump modal with matching results and cursor.
- `TestRenderListItem`: Verifies styled checkmarks (`✔`), unchecked circles (`○`), and standard bullets (`•`).
- `TestRenderDivider`: Verifies centered hairline horizontal section dividers (`***`, `___`, `::hr`).
- `TestCodeLineNumbersView`: Validates code block line numbering rendering, vertical bar separator (`1 │`), and `[L: lines]` status bar badge.
- `TestTimerView`: Validates presentation stopwatch rendering in status bar (`[⏱ 05:23]` and hour formatting `[⏱ 1:12:04]`).
- `TestWatchModeView`: Asserts `[watch]` live reload badge renders in status bar when watch mode is active.
- `TestOverviewModalView`: Verifies multi-column grid layout, slide cards, cursor highlighting, and badges.
- `TestBlankScreenView`: Asserts blackout presentation screen rendering with resume prompt.
- `TestAutoplayView`: Validates status bar `[▶ auto: 8s (4s)]` badge, `A auto` status hint, and help modal documentation.
- `TestRenderBranchBlock`: Verifies rendering of `BlockBranch` cards with key badge (`[1]`), bold label, arrow (`──►`), target `#id`, and cursor highlighting.
- `TestRenderGraphModal`: Tests interactive presentation DAG topology map rendering, active slide indicator, and visited path breadcrumbs (`Path: [01] ──► [02]`).
- `TestRenderHistoryModal`: Validates rendering of the interactive traversal history reflog modal (`H`), step badges `[1]`, `(N steps back)`, `● CURRENT` marker, and list truncation.
- `TestRenderBranchHUDModal`: Validates rendering of the branch decision HUD modal (`J`), outgoing choices cards, downstream slide count and duration badges, track match badges, destination live preview card, and help modal shortcuts.
- `TestNavStatusForkAndHistory`: Verifies `[fork: N paths]`, `[history: N]`, and `M map` badges in the status bar.
- `TestHelpModalGraphShortcuts`: Asserts `1 - 9`, `Backspace / H`, and `M` shortcuts are documented in help modal.
- `TestRenderRunnerCardAndView`: Verifies rendering of live code execution output cards with exit code badges, execution duration, and stdout/stderr blocks.
- `TestRenderFocusModeView`: Asserts full viewport focus mode layout, line numbering, vertical scrolling offsets, and integrated execution output drawer.
- `BenchmarkRenderView`: Measures frames-per-second rendering efficiency.

### D. Model & Parser — `internal/model_test.go`
Tests markdown AST parsing and serialization:
- `TestParseDeck`: Standard presentation parsing with YAML frontmatter.
- `TestSerializeRoundTrip`: AST to Markdown string serialization consistency.
- `TestParseDirective`: Directives extraction (`::code lang=python`, `::image path.png`).
- `TestParseAlignment`: Slide-level and deck-level alignment specifications.
- `TestParseDeckEdgeCases`: Empty decks, missing frontmatter, numbered lists, bullet points, closed code blocks, and custom directives.
- `TestSerializeDeckEdgeCases`: Serialization with alignments and multiple block types.
- `TestSpeakerNotesModel`: Multi-line accumulator under `::notes` directive.
- `TestStandardMarkdownFeatures`: Native Markdown tables, fenced code blocks (```` ```lang ````), and standard images (`![alt](path)`).
- `TestCalloutBlocks`: Parsing `> [!TIP]`, `> [!NOTE]`, `> [!WARNING]`, `> [!IMPORTANT]`, `> [!CAUTION]`, and quotes.
- `TestBlockDivider`: Parsing `***`, `___`, and `::hr` into `BlockDivider`.
- `TestBranchParsingAndModel`: Parsing `::branch`, `::fork`, markdown arrows (`->`, `=>`), `# Title {#slug}`, tags, and branch key lookup.
- `TestFindSlideByID`: Multi-pass slide lookup by exact ID, title slug, 1-based index, and title substring.
- `TestBranchSerialization`: Round-trip preservation of `::id`, `::next`, `::prev`, `::tags`, and `::branch`.
- `BenchmarkParseDeck`: Measures markdown parser throughput.

### E. Theme Engine — `internal/theme_test.go`
Tests theme definitions, resolution, and persistence:
- `TestAvailableThemes`: Verifies 9 curated themes and color fields.
- `TestResolveTheme`: Tests case-insensitive theme resolution.
- `TestCustomHexTheme`: Tests custom hex color strings (`#3b82f6`).
- `TestNextThemeCycle`: Tests cycling forward/backward through theme list.
- `TestThemeAppliedInView`: Tests that theme accents and background styles affect rendering.
- `TestThemeFrontmatterRoundTrip`: Verifies frontmatter `theme: dracula` serialization and preservation.

### F. Terminal Images — `internal/image_test.go`
Tests image resolution and ANSI rendering:
- `TestResolveImagePath`: Path lookup relative to `.deck.md` and current working directory.
- `TestRenderImageMissing`: Missing file fallback placeholder.
- `TestRenderImageOutput`: ANSI half-block (`▀`) truecolor character emission.
- `TestRenderRealImages`: Rendering test fixtures.
- `TestGetImageInfo`: Decoding image dimensions and format types.

### G. Standalone HTML Export — `internal/export_test.go`
Tests single-file self-contained HTML presentation generation:
- `TestFormatInlineHTML`: Escapes HTML entities and translates markdown bold, italic, and code spans to semantic HTML tags.
- `TestExportHTML`: Validates HTML assembly, embedded CSS theme palette variables, block rendering (headings, paragraphs, code blocks, diffs, tables, callouts, task checklists), and embedded vanilla JS runner.
- `TestExportHTMLFile`: Hermetic disk test asserting file creation, `.deck.md` trimming, and base64 asset encoding.
- `TestEditorExportHTML`: Asserts `editor.ExportHTML()` writes file to expected path with correct status update.
- `TestExportHTMLBranching`: Asserts branch cards render with `branch-fork-card`, `data-key`, `data-target`, and `jumpToBranch()` script.

### H. Presentation Statistics & Pacing — `internal/stats_test.go`
Tests presentation metrics calculation, speaking pacing estimates, and sprint velocity reporting:
- `TestCalculateStats`: Asserts metrics accuracy, block tallies, word counting, task completion, speaking duration, and index boundary clamping.
- `TestRenderProgressBar`: Verifies Unicode progress bar formatting at 0%, 50%, 100%, and boundary values.
- `TestFormatStatsCLI`: Asserts CLI report contains all section headers, values, and handles empty tasks.
- `TestEditorStatsModalAndKeyNav`: Validates opening modal with `S`, closing with `Esc`/`q`/`S`.
- `TestRenderStatsModalView`: Asserts modal card styling and full-view rendering.

### I. Directed Graph Topology — `internal/graph_test.go`
Tests presentation graph construction, traversal analysis, and diagram export:
- `TestBuildGraphLinear`: Validates sequential edge generation, cycle absence, and full node reachability.
- `TestBuildGraphBranching`: Tests multi-branch forks, convergence edges (`::next`), Mermaid diagram generation, and ASCII CLI format.
- `TestGraphCycleAndOrphans`: Validates cycle detection algorithms and unreferenced detached slide detection with warnings.
- `TestReachableNodesOutOfBounds`: Asserts safety against out-of-bounds start indices.
- `TestShortestPathAndBreadcrumbs`: Validates BFS shortest path calculation on presentation DAG and real-time breadcrumb trail formatting.
- `TestLintGraph`: Tests presentation DAG topology diagnostics, catching broken branch targets, broken routes, duplicate IDs, orphan slides, dead ends, and CLI diagnostic formatting.
- `TestGetForkOptions`: Tests extraction of outgoing fork choices, downstream reachability counts, talk time estimation from word counts, code block counting, track matching, and route step alignment.

### J. Live Code Runner Subsystem — `internal/runner_test.go`
Tests subprocess execution, execution timeouts, output truncation, language whitelisting, and CI deck testing:
- `TestExecuteBlockSh`: Executes simple POSIX shell command (`echo "termdeck"`) and asserts stdout and exit code 0.
- `TestExecuteBlockNonZeroExit`: Asserts non-zero exit codes (exit 2) are captured and marked as failures.
- `TestExecuteBlockTimeout`: Verifies context timeouts terminate hanging subprocesses (`sleep 2`) cleanly.
- `TestExecuteBlockNonCode`: Asserts non-code blocks return error without executing.
- `TestIsExecutableLanguage`: Validates executable whitelist (`sh`, `bash`, `python`, `go`, `node`, `ruby`) vs display types (`diff`, `sql`, `yaml`, `text`).
- `TestExecuteBlockEmpty`: Asserts empty code blocks return error.
- `TestExecuteBlockGo`: Validates compiling and running temporary Go snippets.
- `TestExecuteBlockWithLines`: Validates multiline Python execution with indentation stripping (`dedent()`).
- `TestTruncateOutput`: Validates memory safeguard truncation at 16KB and 300 lines limit.
- `TestExecuteCodeCmd`: Asserts `ExecCodeCmd` Bubble Tea command wraps execution in `ExecFinishedMsg`.
- `TestTestAllDeckCode`: Tests deck-wide CI code validation and execution summary formatting.

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

Statement coverage across packages (150 unit tests):

| Package | Statement Coverage | Status |
|---|---|---|
| `deck` (root) | 18.2% | Covers model, update loop, flags, live watch loop, runner messages (excluding `main()` process exit) |
| `deck/internal` | 90.5% | Meets $\ge 90\%$ target across all core modules |
| **Total Project** | **85.9%** | **PASSED** |

---

## 5. Performance Benchmarks

Performance targets ensure real-time terminal responsiveness:

```bash
$ make bench
BenchmarkParseDeck-8       387848       3567 ns/op        4678 B/op      24 allocs/op
BenchmarkRenderView-8        4365     268766 ns/op       78581 B/op     650 allocs/op
```

- **Markdown Parsing**: ~3.5 microseconds per slide deck.
- **Terminal Rendering**: ~0.26 milliseconds per full terminal frame (equivalent to >3700 FPS rendering capability).

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
