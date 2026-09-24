---
author: Sachin
format: 0.1
title: termdeck demo
---

::align left
# termdeck

A terminal-native presentation format designed by developers, for developers.

- plain **text** files, git-friendly
- AI agents can *write* decks directly
- present from any ssh session or tmux pane
- zero dependencies — just a single binary

::code lang=bash
  # run demo deck
  deck demo.deck.md

---

::align left
# Why Terminal?

Present technical ideas in a simple, clean, and distraction-free medium:

- no font or video codec hell
- the canvas is just a character grid
- works over **SSH**, in **tmux**, anywhere
- switch between editing and presenting instantly

::code lang=bash
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

::code lang=go
  // Concurrent worker pool
  func worker(ctx context.Context, jobs <-chan Job) {
      for job := range jobs {
          process(ctx, job) // strings & comments highlighted
      }
  }

::code lang=bash
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

```diff
@@ -1,4 +1,4 @@
- func getUser(id int) (*User, error)
+ func getUser(ctx context.Context, id int) (*User, error)
  {
-     return db.QueryRow("SELECT * FROM users WHERE id = ?", id)
+     return db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = ?", id)
  }
```

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

