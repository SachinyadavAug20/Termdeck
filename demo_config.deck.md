---
title: Config Demo Presentation
theme: tokyo-night
routes:
  quick: intro -> interactive -> conclusion
  full: intro -> interactive -> branching -> code -> conclusion
tracks:
  demo: intro -> interactive -> code -> conclusion
---

# Introduction {#intro}
::tags: intro, demo

### Welcome to Config-Driven Presentations

This presentation demonstrates the new configuration file features in Termdeck.

> Press Space or Enter to advance

---

# Interactive {#interactive}
::tags: interactive, demo

Everything you need to navigate and present effortlessly:

- Command Palette (`:` / `Ctrl+P`): Search and execute any presentation command
- Audience Tracks (`K`): Filter presentation paths for demo audiences
- Live Code Execution (`X`): Run code snippets on the fly

---

# Code Demonstration {#code}
::tags: code, demo

::code lang="go" eval=false
fmt.Println("Hello from Termdeck with config!")
::

> Press X or Ctrl+X to execute this block directly in the terminal!

---

# Conclusion {#conclusion}
::tags: summary

You have now seen the configuration capabilities of Termdeck.

- Edit this file directly in markdown or press `i` in Termdeck
- Export standalone HTML slides using `E` or `deck --export-html`
- Check your presentation graph topology with `deck --lint`
- Use `--config` to load presentation settings
