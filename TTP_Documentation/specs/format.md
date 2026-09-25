# termdeck format v0.1

A plain-text slide format for terminal presentations.

## File extension

`.deck.md`

## Structure

A deck is a Markdown file with three sections:

```
---              ← opening delimiter
format: 0.1      ← YAML frontmatter (key: value)
title: My Deck
---              ← closing delimiter

# Slide 1        ← slide content

---

# Slide 2
...
```

### Frontmatter

Optional YAML block between `---` delimiters. Keys:

| Key | Required | Description |
|-----|----------|-------------|
| `format` | yes | Version string, e.g. `0.1` |
| `title` | no | Deck title |
| `author` | no | Author name |
| `align` | no | Default alignment: `left`, `center`, `right` |
| `theme` | no | Color theme: `tokyo-night`, `dracula`, `catppuccin`, `nord`, `gruvbox`, `monokai`, `solarized`, `cyberpunk`, `termdeck` (or `#hex`) |

### Slides

Separated by `---` on its own line. The first slide starts after the frontmatter closing `---`.

## Block types

### Headings

`#` through `######` — rendered with decreasing visual weight.

```
# Title
## Section
### Subsection
```

### Text

Plain text lines. Supports inline formatting:

| Syntax | Rendered as |
|--------|-------------|
| `**bold**` | **bold** |
| `*italic*` | *italic* |
| `` `code` `` | `inline code` |

### Lists & Task Checklists

Lines starting with `-` or `*` (unordered) or `1.` (ordered).
- **Checklists**: `- [ ]` renders as an unchecked item (`○ `), and `- [x]` renders with a bold green checkmark (`✔ `) with dimmed completed text. Press `x` in navigation mode to toggle task state (`[ ]` ⇄ `[x]`) and auto-save immediately to file.
- **Bullets**: Rendered with clean dot glyphs (`• `) with theme accenting.

### Directives

Lines starting with `::` — special instructions for the viewer.

| Directive | Syntax | Description |
|-----------|--------|-------------|
| `::align` | `::align left/center/right` | Slide alignment (also `::left`, `::center`, `::right`) |
| `::code` | `::code lang=X` | Code block (until next `::` or `---`) |
| `::image` | `::image src=X` | Image card (or `::image filename.png`) |
| `::hr` | `::hr` | Horizontal divider hairline within slide (also `***`, `___`) |
| `::id` | `::id <slug>` | Custom slide identifier (also `# Title {#slug}`) |
| `::next` | `::next <slug>` | Override forward transition to converge branches |
| `::prev` | `::prev <slug>` | Override backward transition |
| `::branch` | `::branch [key] label -> target` | Interactive decision branch fork (also `::fork`, `-> [Label](target)`) |
| `::tags` | `::tags backend,perf` | Slide classification tags for topology and search |
| `::notes` | `::notes` | Speaker notes (hidden in presentation) |

### Non-Linear Branching & Directed Graph (DAG)

Termdeck supports non-linear presentation topologies. Instead of rigid $1 \to 2 \to 3$ sequence, slides can branch dynamically based on audience interaction:

#### Decision Branches
Author forks using directives or native markdown arrows:
```markdown
::branch [1] Deep Dive: Core Architecture -> arch-deepdive
::branch [2] Deep Dive: Performance Pacing -> perf-deepdive
-> [Concurrency Patterns](concurrency)
=> [Memory Optimizations](memory)
```
- **Keys**: `[1]`, `[2]`, ... or unkeyed (auto-assigned `1`, `2`, ...).
- **Navigation**: Pressing `1`..`9` on the keyboard follows the corresponding branch immediately. Presenters can also navigate down to any branch card with the laser pointer (`▶`) and press `Enter`.
- **Backtracking**: Pressing `Backspace` or `H` pops from `History []int` to reverse along the presenter's exact path.

#### Slide Identifiers & Convergence
```markdown
# Storage Engine {#storage}
::next conclusion
::tags backend,storage

LSM tree design...
```
- `::id <slug>` or `{#slug}`: Sets slide identifier. Target resolution matches exact IDs, slugs, 1-based slide numbers, or title substrings.
- `::next <slug>`: When advancing with `→` / `Space` / `Enter`, jumps directly to the target slide, allowing multiple deep dives to converge cleanly into a shared conclusion.
- `::prev <slug>`: When going back with `←` / `h`, jumps to the specified previous slide.

### Horizontal Dividers

Lines containing `***`, `___`, or `::hr` render a clean horizontal hairline (`────────`) in theme border styling to separate concepts within a single slide.

#### `::align`

```
::align left
```

Sets the text alignment for the slide. Supported values: `left`, `center`, `right`. Shorthand directives `::left`, `::center`, and `::right` are also supported.

#### `::code` / Standard Code Fences

```
::code lang=python
  def hello():
      print("world")
```
Or standard markdown syntax:
````markdown
```bash
echo "Hello from Termdeck runner"
```
````

Standard Markdown code fences (```` ```lang ````) and `::code` blocks are natively supported with syntax highlighting for Go, Python, TypeScript, Rust, Shell, SQL, and `diff`/`patch` (with green additions and red deletions). Content between code fences is rendered as a bordered code block with language labeling and dynamic box sizing.

##### Live Code Execution & Disabling Eval
Executable languages (`bash`, `sh`, `zsh`, `python`, `go`, `node`, `ruby`) can be run live during presentations with `X`, `ctrl+x`, or `x`. Terminal output is displayed in an attached output drawer with exit code badges and execution duration.

To mark illustrative, non-runnable, or destructive snippets that should **not** execute during `--test-code` CI runs or interactive presentations, append `eval=false`, `no-eval`, `no_run`, or `noexec`:
````markdown
```bash no-eval
rm -rf /tmp/scratch-build-cache
```
````
Or with directives:
```markdown
::code lang=python eval=false
```

##### Element Zoom & Focus Mode (`f` / `F`)
Press `f` or `F` while focused on any code block, diagram, or table to expand it to the full terminal canvas. Focus Mode provides vertical scrolling (`j`/`k`/arrow keys), line numbers, and integrated live code execution (`X` or `x`). Press `Esc` or `f` to exit Focus Mode.

#### `::image` / Standard Markdown Images

```
::image screenshot.png
```
Or standard markdown syntax:
```
![Alt Text](screenshot.png)
```

Rendered as a clean presentation card with file information, dimensions, and format. Press `p` in navigation mode to open the high-resolution image directly in your system viewer (`xdg-open` / `open`).

**Path Resolution Order**:
1. Absolute path (if path begins with `/`)
2. Relative to the `.deck.md` file directory
3. Relative to `images/` or `assets/` subdirectories next to the `.deck.md` file
4. Relative to current working directory

If the file is not found, a styled placeholder `[ image not found: <filename> ]` is displayed.

#### Markdown Tables

```
| Feature | SVN | Git |
|---|---|---|
| Model | Delta | Snapshot |
| Branching | Slow | Instant |
```

Standard GitHub-flavored Markdown tables are parsed into `BlockTable` blocks and rendered with styled borders, padded columns, and bold highlighted headers.

#### Callouts & Admonitions

```markdown
> [!TIP]
> Keep database transactions short to minimize lock contention.

> [!NOTE]
> Backward compatible with v1 API.

> [!WARNING]
> Breaking change in v2.

> "Simplicity is prerequisite for reliability."
> — Edsger W. Dijkstra
```

Callouts are parsed into `BlockCallout` blocks and rendered with styled borders matching theme accent, warning, and success colors, prefixed with clear icons (`💡 TIP`, `ℹ NOTE`, `⚠ WARNING`, `🚨 IMPORTANT`, `🛑 CAUTION`, `❝ QUOTE`).

#### Multi-Column Split Grids (`:::columns`)

Multi-column side-by-side layouts are authored using container directives:

```markdown
:::columns
### Problem: Monolith
- Tight coupling
- Long compile times
:::col
### Solution: Micro-Engines
- Modular boundaries
- Sub-millisecond frames
:::
```

Alternative syntax: `::columns` or `::split` containers with `::col` column separators. Inside each column, any Markdown element (headings, paragraphs, code blocks, diffs, tables, callouts, lists, images) is parsed recursively. Column widths are dynamically balanced based on terminal viewport width with configured horizontal gaps.

#### `::tags`

```markdown
::tags backend,arch,demo
```

Comma-separated tags associated with the slide. Tags power **Audience Tracks**, allowing presenters to tailor presentation flow dynamically for different audiences:
- Press `K` to open the Track Selection Modal.
- Press `[` and `]` to hop directly backward or forward along slides matching the active track.
- Pass `-k, --track <name>` on the CLI to filter TUI startup or CLI DAG diagrams (`--graph`).

#### `::route` & `routes:` Frontmatter

Preset Graph Routes define guided, pre-planned walks through a non-linear DAG presentation:

**Frontmatter syntax:**
```yaml
---
title: System Architecture
routes:
  lightning: intro -> why -> summary
  deepdive: intro -> arch -> benchmarks -> summary
---
```

**Inline directive syntax:**
```markdown
::route workshop: intro -> setup -> demo -> hands-on -> wrapup
```

Both arrow syntax (`->`, `=>`) and comma-separated syntax (`intro, setup, demo`) are supported.
Presenters can press `P` to view available routes in the Route Switcher Modal, see estimated speaking duration and breadcrumbs, and quick-select with `1`–`9` (or `0` to clear).

#### `::notes`

```
::notes
  These are speaker notes.
  Not visible to the audience canvas.
```

Speaker notes are completely omitted from the audience canvas by default. The block cursor and laser pointer (`▶ `) skip notes blocks entirely. Presenters can press `n` at any time to toggle a styled speaker notes overlay box at the bottom of the terminal. When notes exist on the current slide, the status bar displays an `[n: notes]` indicator.

## Slide rendering

- Slides are vertically centered in the terminal with configurable horizontal alignment (`left`, `center`, `right`).
- Footer shows: `slide N/M (align) · blocks B · [n: notes] · ? help · tab align · t theme · n notes · i edit · ^n add · ^d del · ^s save · u undo · q quit`
- Subtle hairline progress line (`─`) across the bottom edge of the terminal.
- Content fills the available height minus footer (and notes overlay when active).

## Keyboard shortcuts (viewer)

| Key | Action |
|-----|--------|
| `→`, `l`, `Space`, `Enter`, `PageDown` | Next slide / advance directed edge / follow active route |
| `←`, `h`, `PageUp` | Previous slide / previous route slide |
| `1` – `9` | Jump directly along numbered branch option |
| `J` | Open interactive Branch Decision Fork HUD with live target previews & metrics |
| `W` | Open Waypoint Pathfinder & Shortest-Path Graph Router (Enter to route, w to step 1 hop) |
| `Backspace` | Pop back 1 slide along traversal history |
| `H` | Open Traversal History & Graph Reflog modal (`1`-`9` or Enter to rewind) |
| `P` | Open Preset Graph Routes & Guided Paths modal (`0` clears, `1`-`9` activates) |
| `K` | Open Audience Tracks & Subgraph Filtering modal (`0`-`9` quick select) |
| `[`, `]` | Hop backward / forward along slides matching active audience track |
| `M` | Open presentation graph map & DAG explorer modal |
| `X`, `Ctrl+X` | Run focused code block live in background & show output card |
| `f`, `F` | Toggle Element Zoom & Focus Mode (full-viewport view with `j`/`k` scroll) |
| `↓`, `j` | Move block cursor / laser pointer down (or scroll in Focus Mode) |
| `↑`, `k` | Move block cursor / laser pointer up (or scroll in Focus Mode) |
| `/` | Quick Jump to slide (by number or title search) |
| `o`, `O` | Slide Overview & 2D Grid Sorter (navigate cards, Enter to jump) |
| `?`, `F1` | Toggle in-app keyboard shortcuts help modal |
| `t`, `T`, `F2` | Cycle color theme (`tokyo-night`, `dracula`, `nord`, etc.) & auto-save |
| `z` | Toggle distraction-free zen mode (hides status bar) |
| `c`, `C` | Toggle presentation stopwatch (`c`) / Reset timer to 00:00 (`C`) |
| `A` | Toggle auto-advance slides & rehearsal pacing |
| `r`, `R` | Reload deck file from disk |
| `y`, `Y` | Copy focused code/block or runner execution output to clipboard |
| `b`, `B` | Blank/blackout presentation screen (any key resumes) |
| `E` | Export deck to standalone HTML presentation |
| `S` | Talk statistics & sprint deck metrics modal |
| `L` | Toggle code block line numbers |
| `x` | Run focused code block live (or toggle task checklist `[ ]` ⇄ `[x]`) |
| `n` | Toggle speaker notes overlay box |
| `Tab`, `Ctrl+A` | Cycle alignment (`left` → `center` → `right`) & auto-save |
| `p` | Open focused image in system viewer |
| `g`, `Home` | First slide |
| `G`, `End` | Last slide |
| `q`, `Ctrl+C` | Quit (auto-saves any unsaved changes) |
| `Esc` | Dismiss runner card / exit focus mode / close modals / clear status |

## Example

See `demo.deck.md`.
