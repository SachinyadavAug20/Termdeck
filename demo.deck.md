---
author: Sachin
format: 0.1
title: termdeck demo
---

::align left
# termdeck

A terminal-native presentation format

- plain **text** files, git-friendly

- AI agents can *write* decks directly

- present from any ssh session

- `deck demo.deck.md` and go

::image demo.png

---

::align left
# Why terminal?

- no font or video codec hell

- the canvas is just a character grid

- works over **SSH**, in **tmux**, anywhere

::code lang=bash
  curl -sSL https://example.com/deck.sh | bash
  deck slides.deck.md

---

::align left
# Styling

**bold text** and *italic text* and `inline code`

mix them: **bold *with italic*** inside

- **git-friendly** — plain text diffs

- **AI-native** — agents write `.deck.md` directly

- *zero dependencies* — just a binary

---

::align left
# Code highlighting

Syntax highlighting works in `::code` blocks:

::code lang=go
  // keyword, string, comment highlighting
  func hello(name string) string {
      return "Hello, " + name  // strings are green
  }

::code lang=bash
  # comments are dimmed
  echo "keywords are highlighted"

---

::align left
# Code blocks

::code lang=go
  package main

  import "fmt"

  func main() {
      fmt.Println("Hello, termdeck!")
  }

::code lang=python
  print("Hello, termdeck!")

---

::align left
# Headings

## this is h2

### this is h3

#### this is h4

##### this is h5

###### this is h6

---

::align left
# Navigation

- `←/→` or `j/k` — next/prev slide

- `Tab` / `Ctrl+A` — cycle alignment (left / center / right)

- `G` — last slide

- `g` — first slide

- `q` or `Ctrl+C` — quit

---

::align left
# Editor

Press `i` to enter edit mode on any block.

- `↑/↓` or `k/j` — move between blocks

- `Esc` — exit edit mode

- `Enter` — confirm edit

- `Ctrl+N` — add new block

- `Ctrl+D` — delete block

- `Ctrl+K/J` — reorder blocks

- `Ctrl+S` — save file

- `u` / `Ctrl+R` — undo / redo

- `p` — open image in system viewer

---

::align left
# Terminal Images

::image 1.png

::code lang=text
  +-----------+     +----------+
  | .deck.md  | --> |  deck    |
  +-----------+     +----------+
                         |
                    +----v-----+
                    | terminal  |
                    +----------+

---

::align left
# Notes (hidden in presentation)

::notes

These notes are not shown to the audience.

Use them for speaker prep.
