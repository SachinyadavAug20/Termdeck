# VHS tape for generating Termdeck demo GIF
# Install vhs: go install github.com/charmbracelet/vhs@latest
# Run: vhs demo.tape

Output demo.gif
Output demo.mp4

Set FontSize 16
Set Width 1280
Set Height 720
Set Theme "TokyoNight"
Set Padding 20
Set Framerate 60

# Launch Termdeck presentation
Type "./deck demo.deck.md"
Enter
Sleep 2s

# Advance a couple slides
Space
Sleep 1.5s
Space
Sleep 1.5s

# Run live code snippet with X
Type "X"
Sleep 2.5s

# Advance to multi-column slide
Space
Sleep 2s

# Navigate to Non-Linear Branching Hub
Type "g"
Sleep 1s
Type "/15"
Enter
Sleep 2s

# Open Branch Decision Fork HUD (J)
Type "J"
Sleep 1.5s
Down
Sleep 1s
Down
Sleep 1s

# Jump to Radar branch
Enter
Sleep 2s

# Open Graph Exploration Radar (V)
Type "V"
Sleep 2.5s
Escape
Sleep 1s

# Open Waypoint Pathfinder (W)
Type "W"
Sleep 1s
Type "loop"
Sleep 1.5s
Enter
Sleep 2.5s

# Execute loop pass on TDD slide
Space
Sleep 1.5s
Space
Sleep 1.5s
Space
Sleep 1.5s

# Advance to conclusion
Space
Sleep 2.5s

# Exit presentation
Type "q"
Sleep 1s
