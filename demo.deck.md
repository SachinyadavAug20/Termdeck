---
author: Sachin
format: 0.1
title: termdeck demo
theme: catppuccin
---

::route deepdive: intro -> arch-comparison -> branching-hub -> arch-deepdive -> columns-deepdive -> routes-deepdive -> history-deepdive -> fork-hud-deepdive -> waypoint-deepdive -> radar-deepdive -> conclusion
::route lightning: intro -> why-terminal -> arch-comparison -> conclusion
::route live-demo: branching-hub -> runner-deepdive -> columns-deepdive -> routes-deepdive -> history-deepdive -> fork-hud-deepdive -> waypoint-deepdive -> radar-deepdive -> conclusion

::id intro
::tags intro, demo
::align left
# termdeck

A terminal-native presentation format designed by developers, for developers.

- plain **text** files, git-friendly

- AI agents can *write* decks directly

- present from any ssh session or tmux pane

- zero dependencies — just a single binary

::code lang=bash eval=false
  # run demo deck
  deck demo.deck.md

---

::id why-terminal
::tags intro, demo
::align left
# Why Terminal?

Present technical ideas in a simple, clean, and distraction-free medium:

- no font or video codec hell

- the canvas is just a character grid

- works over **SSH**, in **tmux**, anywhere

- switch between editing and presenting instantly

::code lang=bash eval=false
  curl -sSL https://example.com/deck.sh | bash
  deck slides.deck.md

> [!TIP]
> Press `y` anytime to copy this code snippet directly to your system clipboard!

---

::align left
# Styling & Dividers

Support for inline formatting and horizontal section dividers:

**bold text** and *italic text* and `inline code` spans.

Mix styles seamlessly: **bold with *italic* text** inside.

***

- **git-friendly** — readable line-by-line diffs

- **AI-native** — agents generate `.deck.md` cleanly

- *zero bloat* — fast launch, sub-millisecond rendering

---

::id arch-comparison
::tags arch, demo
::align left
# Architecture & Comparison

Structured technical comparison with native Markdown tables:

| Dimension | Traditional Slides | Termdeck |
|---|---|---|
| Source Format | Binary / Bloated XML | Git-friendly Markdown |
| Rendering | Heavy GUI / Web View | Native ANSI Character Grid |
| Version Control | Unreadable binary diffs | Clean line-by-line git diffs |
| Workflow | Context-switch to browser | In-terminal with your code |
| Latency | Slow launch, high RAM | Instant (<0.2ms per frame) |

---

::align left
# Callouts & Admonitions

Highlight crucial technical insights, warnings, and architectural decisions:

> [!TIP]
> Keep database transactions short to minimize lock contention.

> [!NOTE]
> Termdeck stores all slide content directly in human-readable Markdown files.

> [!WARNING]
> Breaking change in v2: verify client migration before deploying.

> "Simplicity is prerequisite for reliability."
> — Edsger W. Dijkstra

---

::align left
# Syntax Highlighting

Native syntax highlighting for modern backend and systems languages (press `L` to toggle line numbers):

::code lang=go eval=false
  // Concurrent worker pool
  func worker(ctx context.Context, jobs <-chan Job) {
      for job := range jobs {
          process(ctx, job) // strings & comments highlighted
      }
  }

::code lang=bash eval=false
  # automated deployment pipeline
  git pull origin main && make test && make build

---

::align left
# Multi-Language Code

Highlighting Rust, Python, TypeScript, and SQL queries:

::code lang=rust
  // Memory safety without garbage collection
  fn compute_hash(input: &str) -> String {
      let digest = sha256::digest(input.as_bytes());
      format!("0x{}", digest)
  }

::code lang=sql
  -- Fast analytical queries
  SELECT service, AVG(latency_ms) AS p99_latency
  FROM telemetry_events
  GROUP BY service ORDER BY p99_latency DESC;

---

::align left
# Code Diffs

Present migrations, code reviews, and refactors cleanly:

::code lang=diff
@@ -1,4 +1,4 @@
- func getUser(id int) (*User, error)
+ func getUser(ctx context.Context, id int) (*User, error)
  {
-     return db.QueryRow("SELECT * FROM users WHERE id = ?", id)
+     return db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", id)
  }

---

::align left
# Task Checklists

Track technical sprint milestones in real time. Press `x` to toggle:

- [x] Design RFC & API schema

- [x] Implement read replicas

- [ ] Migrate caching layer

- [ ] Decommission legacy monolith

> [!TIP]
> Toggling tasks with `x` automatically saves changes directly to disk.

---

::align left
# Themes & Zen Mode

Tailored for terminal aesthetics and distraction-free presenting:

- **9 Curated Themes** — Tokyo Night, Dracula, Nord, Catppuccin, Gruvbox, Monokai, Solarized, Cyberpunk, Termdeck Pink

- Press `t` / `T` / `F2` to cycle color themes on the fly

- Press `z` to enter **Zen Mode** (hiding status bars for clean screen sharing)

- Hairline slide progress indicator at the bottom edge

***

::notes
Speaker reminder: Demonstrate pressing 'z' to toggle Zen Mode on and off.

---

::align left
# Navigation & Shortcuts

Full keyboard control designed for efficient presenting:

- `←/→` or `j/k` — next/prev slide or block

- `/` — quick slide jump (number or live title search)

- `o` / `O` — slide overview & 2D grid sorter (visual deck map)

- `y` / `Y` — copy/yank focused code or block to system clipboard (OSC 52)

- `E` — export deck to standalone offline HTML presentation

- `S` — talk statistics & sprint deck metrics modal (`--stats` CLI flag)

- `A` — toggle auto-advance & rehearsal pacing (`-a <sec>` / `--autoplay` CLI flag)

- `b` / `B` — blank/blackout presentation screen (any key resumes)

- `c` / `C` — toggle presentation stopwatch / reset timer

- `r` / `R` — reload deck file from disk (`-w` flag for auto-watch)

- `L` — toggle code block line numbers

- `x` — toggle task checklist item ([ ] ⇄ [x])

- `z` — toggle distraction-free zen mode

- `t` / `T` / `F2` — cycle color themes

- `Tab` / `Ctrl+A` — cycle alignment (left / center / right)

- `n` — toggle speaker notes overlay

- `?` or `F1` — open in-app keyboard shortcuts help modal

- `q` or `Ctrl+C` — quit (with auto-save)

---

::align left
# Live Block Editor

Press `i` to enter edit mode on any block. Edit slides directly:

- `↑/↓` or `k/j` — move between blocks

- `Esc` — exit edit mode

- `Enter` — confirm edit (auto-saves to disk)

- `Ctrl+N` — add new block

- `Ctrl+D` — delete block

- `Ctrl+K/J` — reorder blocks up / down

- `Ctrl+S` — save file manually

- `u` / `Ctrl+R` — undo / redo changes

- `p` — open image in system viewer

---

::align left
# Terminal Images & Diagrams

Native presentation image cards with dimensions and viewer integration:

::image demo.png

::code lang=text
  +-----------+     +----------+     +---------------+
  | .deck.md  | --> |  deck    | --> | Terminal ANSI |
  +-----------+     +----------+     +---------------+
                          |
                     [ press 'p' to open image ]

---

::align left
# Speaker Notes (Hidden)

::notes

These notes are strictly omitted from the audience canvas.
Use them for speaker cues, talk timing, and Q&A prep.
Press 'n' during presentation to toggle this overlay on/off.

Speaker notes are isolated from audience view.

- Status bar displays `[n: notes]` when notes exist on slide
- Press `n` to toggle private speaker notes overlay
- Laser pointer cursor (`▶`) automatically skips hidden notes
- Safe for screen shares and projector mirroring

---

::id branching-hub
::tags arch, demo, tracks
::align left
# Non-Linear Branching & DAG Engine

Termdeck transforms static linear slides into an interactive directed graph (DAG).

Presenters dynamically steer the presentation based on audience choice:

::branch [1] Deep Dive: Core Architecture & Topology -> arch-deepdive

::branch [2] Deep Dive: Rehearsal & Autoplay Pacing -> autoplay-deepdive

::branch [3] Deep Dive: Offline HTML & Diagram Exports -> export-deepdive

::branch [4] Deep Dive: Live Code Runner & Zoom Focus -> runner-deepdive

::branch [5] Deep Dive: Multi-Column Split & Audience Tracks -> columns-deepdive

::branch [6] Deep Dive: Preset Routes & Guided Paths -> routes-deepdive

::branch [7] Deep Dive: Traversal History & DAG Linter -> history-deepdive

::branch [8] Deep Dive: Decision Fork HUD & Previews -> fork-hud-deepdive

::branch [9] Deep Dive: Waypoint Pathfinder & Router -> waypoint-deepdive

::branch [v] Deep Dive: Graph Exploration Radar & Coverage -> radar-deepdive

::branch [l] Deep Dive: Bounded Graph Cycles & Loop Engine -> loop-deepdive

> [!TIP]
> Press `1`-`9`/`v`/`l` to branch, `J` for Fork HUD, `W` for Waypoints, `V` for Radar, `U` to return to fork, or `M` for topology!

---

::id arch-deepdive
::next conclusion
::tags arch, demo
::align left
# Core Architecture & DAG Topology

Termdeck presentations can converge and fork anywhere:

- **Directives**: `::id <slug>`, `::next <slug>`, `::prev <slug>`

- **Arrows**: `-> [Label](target)` or `::branch [key] Label -> target`

- **History Stack**: Press `Backspace` or `H` to return along your path

- **Convergence**: Slides with `::next` automatically merge branches back

::code lang=text
  [Overview] ──► [Branch 1: Architecture] ──► [Conclusion]
             └──► [Branch 2: Autoplay]     ──► [Conclusion]

---

::id autoplay-deepdive
::next conclusion
::tags demo
::align left
# Rehearsal & Autoplay Pacing

Prepare talk pacing with hands-free automated rehearsal:

- Press `A` to toggle auto-advance mode on/off

- Default 5-second interval or pass `--autoplay [sec]` CLI flag

- Manual arrow key presses reset the countdown timer

- Rehearsal loops automatically for unattended booth displays

> [!NOTE]
> Press `Backspace` to return to the branching hub, or advance to converge to conclusion.

---

::id export-deepdive
::next conclusion
::tags demo, export
::align left
# Offline HTML & Diagram Exports

Share technical ideas beyond the terminal:

- **HTML Export**: Press `E` or run `deck --export-html demo.deck.md`

- **Mermaid Export**: Run `deck --mermaid demo.deck.md` for GitHub markdown diagrams

- **ASCII DAG**: Run `deck --graph demo.deck.md` for terminal topology visualization

::code lang=bash eval=false
  # inspect presentation topology map
  deck --graph demo.deck.md
  deck --mermaid demo.deck.md > topology.mmd

---

::id runner-deepdive
::next conclusion
::tags dev, demo
::align left
# Live Code Runner & Zoom Focus Mode

Termdeck lets developers execute live code directly during presentations:

- **Live Runner**: Press `X` or `ctrl+x` (or `x` on a code block) to execute

- **Ephemeral Output**: Exit code, elapsed time, stdout/stderr rendered in runner card

- **Element Zoom**: Press `f` or `F` to maximize focused code, table, or diagram to full screen!

- **CI/CD Verification**: Run `deck --test-code demo.deck.md` to verify all slide code snippets

::code lang=bash
# Live bash execution test
uname -s -m
echo "Termdeck presentation engine: 100% operational"

::code lang=python
# Verified live calculation
import math
print(f"Verified live prime count up to 50: {len([p for p in range(2, 50) if all(p%d!=0 for d in range(2, int(p**0.5)+1))])}")

> [!TIP]
> Focus on either code block above and press `X` to run, or press `f` to enter Zoom Focus mode!

---

::id columns-deepdive
::next conclusion
::tags arch, demo, tracks
::align left
# Multi-Column Split & Audience Tracks

Present side-by-side technical content and tailor presentations to audience tracks:

:::columns
:::col
### Backend / Arch Track

- Systems design & internals

- Zero-alloc rendering loops

- Deterministic DAG traversal

- Tag with `::tags arch,backend`
:::col
### Developer Experience Track

- Live terminal execution (`X`)

- Instant element zoom (`f`)

- Non-linear branches (`1-9`, `M`)

- Press `K` for audience tracks
:::

> [!TIP]
> Press `K` to filter presentation flow by audience track, or use `[` / `]` to hop along track slides!

---

::id routes-deepdive
::next conclusion
::tags arch, demo, routes
::align left
# Preset Graph Routes & Guided Paths

Pre-configure tailored presentation paths across your non-linear DAG:

:::columns
:::col
### Authoring Routes

- In frontmatter:

`routes:`

`lightning: intro -> summary`

- Inline: `::route name: s1 -> s2 -> s3`

- Arrow (`->`) or comma (`s1, s2`) syntax
:::col
### Presenter Controls

- Press `P` to open Route Switcher modal

- `1`-`9` selects a route instantly

- `0` clears back to standard navigation

- Estimated talk duration based on WPM

- CLI startup: `deck --route=lightning`
:::

> [!TIP]
> Press `P` right now to preview and activate preset talk routes, or press `M` to see the route highlighted on the DAG!

---

::id history-deepdive
::next conclusion
::tags arch, demo, history
::align left
# Traversal History & Graph Reflog

Inspect and rewind your presentation journey through the non-linear DAG:

:::columns
:::col
### Visual Journey Stack (`H`)

- Press `H` to open Traversal History modal

- Lists all visited slides in exact path order

- Shows distance back (`(3 steps back)`)

- Rewind to any step with `1`-`9` or `Enter`

- Press `c` to reset traversal history
:::col
### DAG Linter & CI/CD (`--lint`)

- Validates presentation graph topology

- Detects broken `::branch` / `::next` targets

- Catches duplicate slide identifiers

- Identifies unreachable slides & dead ends

- CLI validation: `deck --lint demo.deck.md`
:::

> [!TIP]
> Press `H` anytime to see your reflog trail, or run `deck --lint` before your talk to guarantee zero broken links!

---

::id fork-hud-deepdive
::next conclusion
::tags arch, demo, branching
::align left
# Decision Fork HUD & Subgraph Metrics

Preview destinations and inspect downstream path depth before branching:

:::columns
:::col
### Interactive Fork HUD (`J`)

- Press `J` on any branching slide

- Real-time **code & content preview** of destinations

- Displays keys `[1]`, `[2]`, `[→]` with target IDs

- `j`/`k` to select target, `Enter` to jump

- Direct numeric execution `1`-`9`
:::col
### Downstream Metrics & Budgets

- **Slide depth**: Reachable slides in subgraph

- **Time estimate**: Duration based on ~130 WPM

- **Code snippets**: Tally of executable blocks

- **Track match**: `★ Track Match` indicators

- **Route sync**: Highlights `⚡ Route Step`
:::

> [!TIP]
> Press `J` whenever you reach a fork to see live previews and time estimates for every branch!

---

::id waypoint-deepdive
::next conclusion
::tags arch, demo, routing
::align left
# Waypoint Pathfinder & Shortest-Path Graph Router

Dynamically navigate complex presentation graphs using real-time graph pathfinding:

:::columns
:::col
### Interactive Pathfinder (`W`)

- Press `W` anywhere to summon the Pathfinder modal

- Live shortest-path routing via BFS graph traversal

- Real-time search by title, slide `#id`, or `::tags`

- Detailed hop breadcrumb trail (`[01:hub] ──[9]──► [10:waypoint]`)

- Talk duration calculation (~130 WPM) per path
:::col
### Dynamic Route Execution

- `Enter`: Locks optimal path into a dynamic presentation route

- `w`: Steps immediately 1 hop along the shortest path

- Reachability checking prevents routing to dead ends

- Unreachable slides flagged with rewind recommendations

- Zero presentation spoilers or awkward backtracks
:::

> [!TIP]
> Press `W` anytime to find and navigate the shortest path to any slide in your deck!

---

::id radar-deepdive
::next conclusion
::tags arch, demo, radar
::align left
# Graph Exploration Radar & Upstream Fork Return

Monitor graph coverage, track unvisited branches, and fast-return to decision hubs:

:::columns
:::col
### Exploration Radar (`V`)

- Press `V` anywhere to open the Exploration Radar modal

- Global coverage bar: `Visited 8 / 18 unique slides (44.4%)`

- Time budget tracking: `~12m visited / ~25m total deck talk time`

- Per-branch completion status:

- `✔` Completed branches (`3/3 slides · 100%`)

- `◐` In-progress branches (`1/2 slides · 50%`)

- `○` Unvisited branches (`0/2 slides · Unvisited · ~2m`)

- Direct jumping: `Enter` jumps to selected branch
:::col
### Upstream Fork Fast-Return (`U`)

- Press `U` anywhere inside a deep sub-branch

- Teleports speaker directly back to the upstream decision fork

- Zero tedious `Backspace` tapping or reflog math

- Perfect for Q&A: explore a sub-branch, then `U` back to the hub

- Status bar hints: `[U: return to fork]` & `[radar: 44% (V)]`
:::

> [!TIP]
> Press `V` to see what branches remain unvisited, or hit `U` to teleport straight back to the fork hub!

---

::id loop-deepdive
::tags arch, demo, loop
::align left
::loop [r] TDD Iteration Cycle -> loop-deepdive max=3 next=conclusion
# Bounded Graph Cycles & Loop Iteration

Model computational workflows, algorithmic iterations, and engineering design cycles:

:::columns
:::col
### Bounded Cycle Directive (`::loop`)

- `::loop [r] TDD Cycle -> loop-deepdive max=3 next=conclusion`

- Supports finite presentation iterations without infinite DAG cycles

- Tracks pass counts per loop: `pass 1/3 (2 remaining)`

- Hotkey shortcut: press `r` or `Space`/`Enter` to iterate

- Auto-exits to `next=` slide upon exhaustion
:::col
### Real-World Engineering Workflows

- **TDD Cycles**: Red ──► Green ──► Refactor (repeat N times)

- **Retry Backoff**: Request ──► Timeout ──► Exponential Delay

- **Consensus Rounds**: Propose ──► Vote ──► Commit (Raft/Paxos)

- **ML Training**: Epoch iteration over batches & validations

- In-slide loop iteration card with real-time pass progress
:::

> [!TIP]
> Press `Space` or `r` to execute loop passes. Notice the pass counter advancing! Once 3 passes complete, `Space` smoothly continues to the conclusion!

---

::id conclusion
::tags demo, summary
::align left
# Developer Summary

Simple, clean, terminal-native presentations built by a developer for developers:

- **Directed Graph Presentations** with interactive decision branches (`1-9`, `M`)

- **Live Terminal Code Runner** (`X`, `ctrl+x`, `--test-code`) for running code live

- **Zoom Focus Mode** (`f`, `F`) to maximize code/tables/diagrams distraction-free

- **Offline HTML Export** with embedded standalone navigation (`E`)

- **Distraction-Free Zen Mode** (`z`) & Presentation screen blackout (`b`)

- **Rehearsal Pacing & Auto-advance** (`A`, `-a`) & Talk stopwatch (`c`, `C`)

- **Live File Watch & Auto-reload** (`-w`, `r`) & Interactive Checklists (`x`)

- **Zero bloat**, single binary, sub-millisecond per-frame rendering

> [!IMPORTANT]
> Run `deck --graph demo.deck.md` to view the full presentation topology map!
