package internal

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Tick messages for presentation timer ---

type TickMsg time.Time

func TickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type WatchMsg time.Time

func WatchCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return WatchMsg(t)
	})
}

// --- Editor modes ---

type EditorMode int

const (
	ModeNav EditorMode = iota
	ModeEdit
	ModePrompt
)

// --- Editor state ---

type Editor struct {
	Mode              EditorMode
	SlideIdx          int
	BlockIdx          int
	CursorCol         int
	Draft             string
	UndoStack         []string
	RedoStack         []string
	FilePath          string
	Dirty             bool
	Message           string
	ShowNotes         bool
	ShowHelp          bool
	ZenMode           bool
	ShowLineNumbers   bool
	ShowTimer         bool
	TimerStart        time.Time
	WatchMode         bool
	Theme             string
	ShowOverview      bool
	OverviewCursor    int
	OverviewCols      int
	ScreenBlank       bool
	ShowStats         bool
	Autoplay          bool
	AutoplayInterval  int
	AutoplayCountdown int
	AutoplayLoop      bool
	History           []int
	ShowGraphMap      bool
	GraphMapCursor    int
}

func NewEditor(filePath string) Editor {
	return Editor{
		Mode:              ModeNav,
		FilePath:          filePath,
		OverviewCols:      3,
		AutoplayInterval:  5,
		AutoplayCountdown: 5,
		AutoplayLoop:      true,
	}
}

func (e *Editor) CycleTheme(d *Deck) {
	currentID := e.Theme
	if currentID == "" && d != nil {
		currentID = d.Theme
	}
	next := NextTheme(currentID)
	e.Theme = next.ID
	if d != nil {
		d.Theme = next.ID
		if d.Meta == nil {
			d.Meta = make(map[string]string)
		}
		d.Meta["theme"] = next.ID
		e.Save(*d)
	}
	e.Message = fmt.Sprintf("theme: %s", next.Name)
}

// --- Undo / Redo ---

func (e *Editor) SaveUndo(d Deck) {
	state := SerializeDeck(d)
	e.UndoStack = append(e.UndoStack, state)
	e.RedoStack = nil
	if len(e.UndoStack) > 100 {
		e.UndoStack = e.UndoStack[1:]
	}
}

// --- Block operations ---

func (e *Editor) currentBlock(d *Deck) *Block {
	if e.SlideIdx >= len(d.Slides) {
		return nil
	}
	slide := &d.Slides[e.SlideIdx]
	if e.BlockIdx >= len(slide.Blocks) {
		return nil
	}
	return &slide.Blocks[e.BlockIdx]
}

func (e *Editor) ToggleAutoplay(interval ...int) tea.Cmd {
	e.Autoplay = !e.Autoplay
	if e.Autoplay {
		if len(interval) > 0 && interval[0] > 0 {
			e.AutoplayInterval = interval[0]
		}
		if e.AutoplayInterval <= 0 {
			e.AutoplayInterval = 5
		}
		e.AutoplayCountdown = e.AutoplayInterval
		e.Message = fmt.Sprintf("autoplay: on (%ds interval)", e.AutoplayInterval)
		return TickCmd()
	}
	e.Message = "autoplay: off"
	return nil
}

func (e *Editor) TickAutoplay(d *Deck) bool {
	if !e.Autoplay || d == nil || len(d.Slides) == 0 {
		return false
	}
	if e.AutoplayCountdown > 1 {
		e.AutoplayCountdown--
		return false
	}

	// Countdown expired -> advance slide
	e.AutoplayCountdown = e.AutoplayInterval
	if e.SlideIdx < len(d.Slides)-1 {
		e.SlideIdx++
		e.BlockIdx = 0
		e.ClampBlockIdx(d)
		return true
	} else if e.AutoplayLoop {
		e.SlideIdx = 0
		e.BlockIdx = 0
		e.ClampBlockIdx(d)
		return true
	}
	return false
}

func (e *Editor) FollowBranch(target string, d *Deck) bool {
	if d == nil {
		return false
	}
	targetIdx := d.FindSlideByID(target)
	if targetIdx < 0 || targetIdx >= len(d.Slides) {
		return false
	}
	e.History = append(e.History, e.SlideIdx)
	e.SlideIdx = targetIdx
	e.BlockIdx = 0
	e.ClampBlockIdx(d)
	return true
}

func (e *Editor) BackHistory(d *Deck) bool {
	if len(e.History) == 0 {
		return false
	}
	prevIdx := e.History[len(e.History)-1]
	e.History = e.History[:len(e.History)-1]
	if d != nil && prevIdx >= 0 && prevIdx < len(d.Slides) {
		e.SlideIdx = prevIdx
		e.BlockIdx = 0
		e.ClampBlockIdx(d)
		return true
	}
	return false
}

func (e *Editor) ClampBlockIdx(d *Deck) {
	if e.SlideIdx >= len(d.Slides) {
		e.BlockIdx = 0
		return
	}
	vis := d.Slides[e.SlideIdx].VisibleBlockIndices()
	if len(vis) == 0 {
		e.BlockIdx = 0
		return
	}
	for _, idx := range vis {
		if idx == e.BlockIdx {
			return
		}
	}
	e.BlockIdx = vis[0]
}

func (e *Editor) MoveUp(d *Deck) {
	if e.SlideIdx >= len(d.Slides) {
		return
	}
	vis := d.Slides[e.SlideIdx].VisibleBlockIndices()
	if len(vis) == 0 {
		return
	}
	for i, idx := range vis {
		if idx == e.BlockIdx {
			if i > 0 {
				e.BlockIdx = vis[i-1]
				e.CursorCol = 0
			}
			return
		}
	}
	e.BlockIdx = vis[0]
	e.CursorCol = 0
}

func (e *Editor) MoveDown(d *Deck) {
	if e.SlideIdx >= len(d.Slides) {
		return
	}
	vis := d.Slides[e.SlideIdx].VisibleBlockIndices()
	if len(vis) == 0 {
		return
	}
	for i, idx := range vis {
		if idx == e.BlockIdx {
			if i < len(vis)-1 {
				e.BlockIdx = vis[i+1]
				e.CursorCol = 0
			}
			return
		}
	}
	e.BlockIdx = vis[0]
	e.CursorCol = 0
}

func (e *Editor) EnterEdit(d *Deck) {
	blk := e.currentBlock(d)
	if blk == nil {
		return
	}
	e.Mode = ModeEdit
	switch blk.Kind {
	case BlockHeading:
		e.Draft = blk.Text
		e.CursorCol = len(blk.Text)
	case BlockParagraph:
		e.Draft = blk.Text
		e.CursorCol = len(blk.Text)
	case BlockCode:
		e.Draft = strings.Join(blk.Lines, "\n")
		e.CursorCol = len(e.Draft)
	case BlockList:
		e.Draft = blk.Text
		e.CursorCol = len(blk.Text)
	case BlockImage:
		e.Draft = blk.Src
		e.CursorCol = len(blk.Src)
	default:
		e.Draft = blk.Text
		e.CursorCol = len(e.Draft)
	}
}

func (e *Editor) ExitEdit(d *Deck) {
	blk := e.currentBlock(d)
	if blk == nil {
		e.Mode = ModeNav
		return
	}
	e.SaveUndo(*d)
	switch blk.Kind {
	case BlockHeading:
		blk.Text = e.Draft
	case BlockParagraph:
		blk.Text = e.Draft
	case BlockCode:
		blk.Lines = strings.Split(e.Draft, "\n")
	case BlockList:
		blk.Text = e.Draft
	case BlockImage:
		blk.Src = strings.TrimSpace(e.Draft)
	case BlockCallout:
		blk.Text = e.Draft
		blk.Lines = strings.Split(e.Draft, "\n")
	case BlockDivider:
		if e.Draft != "***" && e.Draft != "___" && e.Draft != "::hr" {
			blk.Kind = BlockParagraph
			blk.Text = e.Draft
		}
	}
	e.Mode = ModeNav
	e.Dirty = true
	e.Message = ""
}

func (e *Editor) CancelEdit() {
	e.Mode = ModeNav
	e.Draft = ""
	e.Message = ""
}

func (e *Editor) AddBlock(d *Deck) {
	if e.SlideIdx >= len(d.Slides) {
		return
	}
	e.SaveUndo(*d)
	slide := &d.Slides[e.SlideIdx]
	newBlock := Block{Kind: BlockParagraph, Text: "new block"}
	pos := e.BlockIdx + 1
	slide.Blocks = append(slide.Blocks, Block{})
	copy(slide.Blocks[pos+1:], slide.Blocks[pos:])
	slide.Blocks[pos] = newBlock
	e.BlockIdx = pos
	e.Dirty = true
	e.Message = "block added"
}

func (e *Editor) DeleteBlock(d *Deck) {
	if e.SlideIdx >= len(d.Slides) {
		return
	}
	slide := &d.Slides[e.SlideIdx]
	vis := slide.VisibleBlockIndices()
	if len(vis) <= 1 {
		e.Message = "can't delete last block"
		return
	}
	if e.BlockIdx >= len(slide.Blocks) {
		return
	}
	e.SaveUndo(*d)
	slide.Blocks = append(slide.Blocks[:e.BlockIdx], slide.Blocks[e.BlockIdx+1:]...)
	if e.BlockIdx >= len(slide.Blocks) {
		e.BlockIdx = len(slide.Blocks) - 1
	}
	e.ClampBlockIdx(d)
	e.Dirty = true
	e.Message = "block deleted"
}

func (e *Editor) MoveBlockUp(d *Deck) {
	if e.SlideIdx >= len(d.Slides) {
		return
	}
	slide := &d.Slides[e.SlideIdx]
	vis := slide.VisibleBlockIndices()
	currPos := -1
	for i, idx := range vis {
		if idx == e.BlockIdx {
			currPos = i
			break
		}
	}
	if currPos <= 0 {
		return
	}
	prevIdx := vis[currPos-1]
	e.SaveUndo(*d)
	slide.Blocks[e.BlockIdx], slide.Blocks[prevIdx] = slide.Blocks[prevIdx], slide.Blocks[e.BlockIdx]
	e.BlockIdx = prevIdx
	e.Dirty = true
}

func (e *Editor) MoveBlockDown(d *Deck) {
	if e.SlideIdx >= len(d.Slides) {
		return
	}
	slide := &d.Slides[e.SlideIdx]
	vis := slide.VisibleBlockIndices()
	currPos := -1
	for i, idx := range vis {
		if idx == e.BlockIdx {
			currPos = i
			break
		}
	}
	if currPos == -1 || currPos >= len(vis)-1 {
		return
	}
	nextIdx := vis[currPos+1]
	e.SaveUndo(*d)
	slide.Blocks[e.BlockIdx], slide.Blocks[nextIdx] = slide.Blocks[nextIdx], slide.Blocks[e.BlockIdx]
	e.BlockIdx = nextIdx
	e.Dirty = true
}

func (e *Editor) AddSlide(d *Deck) {
	e.SaveUndo(*d)
	newSlide := Slide{Blocks: []Block{{Kind: BlockParagraph, Text: "new slide"}}}
	pos := e.SlideIdx + 1
	d.Slides = append(d.Slides, Slide{})
	copy(d.Slides[pos+1:], d.Slides[pos:])
	d.Slides[pos] = newSlide
	e.SlideIdx = pos
	e.BlockIdx = 0
	e.Dirty = true
	e.Message = "slide added"
}

func (e *Editor) DeleteSlide(d *Deck) {
	if len(d.Slides) <= 1 {
		e.Message = "can't delete last slide"
		return
	}
	e.SaveUndo(*d)
	d.Slides = append(d.Slides[:e.SlideIdx], d.Slides[e.SlideIdx+1:]...)
	if e.SlideIdx >= len(d.Slides) {
		e.SlideIdx = len(d.Slides) - 1
	}
	e.BlockIdx = 0
	e.Dirty = true
	e.Message = "slide deleted"
}

// --- Save ---

func (e *Editor) Save(d Deck) {
	data := SerializeDeck(d)
	if err := os.WriteFile(e.FilePath, []byte(data), 0644); err != nil {
		e.Message = fmt.Sprintf("save error: %v", err)
		return
	}
	e.Dirty = false
	e.Message = "saved"
}

// --- Reload from disk ---

func (e *Editor) Reload(d *Deck) error {
	if e.FilePath == "" {
		return fmt.Errorf("no file path")
	}
	content, err := os.ReadFile(e.FilePath)
	if err != nil {
		return err
	}
	newDeck := ParseDeck(string(content))
	if len(newDeck.Slides) == 0 {
		return fmt.Errorf("file contains no slides")
	}
	baseDir := d.BaseDir
	deckTheme := d.Theme
	deckAlign := d.Align
	newDeck.BaseDir = baseDir
	if newDeck.Theme == "" {
		newDeck.Theme = deckTheme
	}
	if newDeck.Align == "" {
		newDeck.Align = deckAlign
	}
	*d = newDeck
	if e.Theme != "" && newDeck.Theme != "" {
		e.Theme = newDeck.Theme
		SetCurrentTheme(e.Theme)
	}
	if e.SlideIdx >= len(d.Slides) {
		e.SlideIdx = len(d.Slides) - 1
		if e.SlideIdx < 0 {
			e.SlideIdx = 0
		}
	}
	e.ClampBlockIdx(d)
	e.Dirty = false
	e.Message = "reloaded from disk"
	return nil
}

// --- Slide alignment ---

func (e *Editor) ToggleAlign(d *Deck) {
	if e.SlideIdx >= len(d.Slides) {
		return
	}
	e.SaveUndo(*d)
	cur := d.Slides[e.SlideIdx].Align
	if cur == "" {
		cur = d.Align
		if cur == "" {
			cur = AlignCenter
		}
	}
	var next AlignKind
	switch cur {
	case AlignLeft:
		next = AlignCenter
	case AlignCenter:
		next = AlignRight
	case AlignRight:
		next = AlignLeft
	default:
		next = AlignCenter
	}
	d.Slides[e.SlideIdx].Align = next
	e.Dirty = true
	e.Message = "align: " + string(next)
	if e.FilePath != "" {
		if _, err := os.Stat(e.FilePath); err == nil {
			e.Save(*d)
			e.Message = "align: " + string(next) + " (saved)"
		}
	}
}

// --- Task toggle ---

func (e *Editor) ToggleTask(d *Deck) {
	blk := e.currentBlock(d)
	if blk == nil || blk.Kind != BlockList {
		return
	}
	trimmed := strings.TrimSpace(blk.Text)
	prefix := ""
	if strings.HasPrefix(trimmed, "- ") {
		prefix = "- "
	} else if strings.HasPrefix(trimmed, "* ") {
		prefix = "* "
	}
	if prefix == "" {
		return
	}
	content := trimmed[len(prefix):]
	e.SaveUndo(*d)

	if strings.HasPrefix(content, "[ ] ") {
		blk.Text = prefix + "[x] " + content[4:]
		e.Dirty = true
		e.Message = "task: complete"
	} else if strings.HasPrefix(content, "[x] ") || strings.HasPrefix(content, "[X] ") {
		blk.Text = prefix + "[ ] " + content[4:]
		e.Dirty = true
		e.Message = "task: pending"
	} else {
		blk.Text = prefix + "[x] " + content
		e.Dirty = true
		e.Message = "task: converted"
	}

	if e.FilePath != "" {
		if _, err := os.Stat(e.FilePath); err == nil {
			e.Save(*d)
		}
	}
}

// --- Input handling ---

func (e *Editor) HandleKey(msg tea.KeyMsg, d *Deck) tea.Cmd {
	key := msg.String()

	switch e.Mode {
	case ModeNav:
		return e.handleNav(key, d)
	case ModeEdit:
		return e.handleEdit(key, d)
	case ModePrompt:
		return e.handlePrompt(key, d)
	}
	return nil
}

func (e *Editor) handleNav(key string, d *Deck) tea.Cmd {
	if e.ScreenBlank {
		e.ScreenBlank = false
		e.Message = "screen resumed"
		return nil
	}

	if e.ShowOverview {
		cols := e.OverviewCols
		if cols <= 0 {
			cols = 3
		}
		totalSlides := len(d.Slides)
		switch key {
		case "esc", "o", "O":
			e.ShowOverview = false
			e.Message = ""
			return nil
		case "left", "h":
			if e.OverviewCursor > 0 {
				e.OverviewCursor--
			}
			return nil
		case "right", "l":
			if e.OverviewCursor < totalSlides-1 {
				e.OverviewCursor++
			}
			return nil
		case "up", "k":
			if e.OverviewCursor-cols >= 0 {
				e.OverviewCursor -= cols
			}
			return nil
		case "down", "j":
			if e.OverviewCursor+cols < totalSlides {
				e.OverviewCursor += cols
			} else if e.OverviewCursor < totalSlides-1 {
				e.OverviewCursor = totalSlides - 1
			}
			return nil
		case "enter", " ":
			if e.OverviewCursor >= 0 && e.OverviewCursor < totalSlides {
				e.SlideIdx = e.OverviewCursor
				e.BlockIdx = 0
				e.ClampBlockIdx(d)
				e.Message = fmt.Sprintf("jumped to slide %d/%d", e.SlideIdx+1, totalSlides)
			}
			e.ShowOverview = false
			return nil
		case "g", "home":
			e.OverviewCursor = 0
			return nil
		case "G", "end":
			if totalSlides > 0 {
				e.OverviewCursor = totalSlides - 1
			}
			return nil
		case "q", "ctrl+c":
			e.ShowOverview = false
			return nil
		}
		return nil
	}

	if e.ShowStats {
		switch key {
		case "esc", "S", "q", "ctrl+c":
			e.ShowStats = false
			return nil
		}
		return nil
	}

	if e.ShowGraphMap {
		totalSlides := len(d.Slides)
		switch key {
		case "esc", "M", "m":
			e.ShowGraphMap = false
			e.Message = ""
			return nil
		case "up", "k":
			if e.GraphMapCursor > 0 {
				e.GraphMapCursor--
			}
			return nil
		case "down", "j":
			if e.GraphMapCursor < totalSlides-1 {
				e.GraphMapCursor++
			}
			return nil
		case "g", "home":
			e.GraphMapCursor = 0
			return nil
		case "G", "end":
			if totalSlides > 0 {
				e.GraphMapCursor = totalSlides - 1
			}
			return nil
		case "enter", " ":
			if e.GraphMapCursor >= 0 && e.GraphMapCursor < totalSlides {
				if e.GraphMapCursor != e.SlideIdx {
					e.History = append(e.History, e.SlideIdx)
					e.SlideIdx = e.GraphMapCursor
					e.BlockIdx = 0
					e.ClampBlockIdx(d)
					e.Message = fmt.Sprintf("jumped to slide %d/%d", e.SlideIdx+1, totalSlides)
				}
			}
			e.ShowGraphMap = false
			return nil
		case "q", "ctrl+c":
			e.ShowGraphMap = false
			return nil
		}
		return nil
	}

	switch key {
	case "q", "ctrl+c":
		if e.Dirty && e.FilePath != "" {
			e.Save(*d)
		}
		return tea.Quit

	case "right", "l", " ", "enter", "pgdown":
		if e.Autoplay {
			e.AutoplayCountdown = e.AutoplayInterval
		}
		if key == "enter" {
			blk := e.currentBlock(d)
			if blk != nil && blk.Kind == BlockBranch && blk.BranchTarget != "" {
				if e.FollowBranch(blk.BranchTarget, d) {
					e.Message = fmt.Sprintf("branch [%s] ──► %s", blk.BranchKey, blk.Text)
					return nil
				}
			}
		}
		if d != nil && e.SlideIdx < len(d.Slides) && d.Slides[e.SlideIdx].NextID != "" {
			nextIdx := d.FindSlideByID(d.Slides[e.SlideIdx].NextID)
			if nextIdx >= 0 && nextIdx < len(d.Slides) {
				e.History = append(e.History, e.SlideIdx)
				e.SlideIdx = nextIdx
				e.BlockIdx = 0
				e.ClampBlockIdx(d)
				return nil
			}
		}
		if d != nil && e.SlideIdx < len(d.Slides)-1 {
			e.SlideIdx++
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
		}

	case "left", "h", "pgup":
		if e.Autoplay {
			e.AutoplayCountdown = e.AutoplayInterval
		}
		if d != nil && e.SlideIdx < len(d.Slides) && d.Slides[e.SlideIdx].PrevID != "" {
			prevIdx := d.FindSlideByID(d.Slides[e.SlideIdx].PrevID)
			if prevIdx >= 0 && prevIdx < len(d.Slides) {
				e.SlideIdx = prevIdx
				e.BlockIdx = 0
				e.ClampBlockIdx(d)
				return nil
			}
		}
		if e.SlideIdx > 0 {
			e.SlideIdx--
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
		}

	case "backspace":
		if e.Autoplay {
			e.AutoplayCountdown = e.AutoplayInterval
		}
		if e.BackHistory(d) {
			e.Message = fmt.Sprintf("back to slide %d/%d", e.SlideIdx+1, len(d.Slides))
			return nil
		}
		if d != nil && e.SlideIdx < len(d.Slides) && d.Slides[e.SlideIdx].PrevID != "" {
			prevIdx := d.FindSlideByID(d.Slides[e.SlideIdx].PrevID)
			if prevIdx >= 0 && prevIdx < len(d.Slides) {
				e.SlideIdx = prevIdx
				e.BlockIdx = 0
				e.ClampBlockIdx(d)
				return nil
			}
		}
		if e.SlideIdx > 0 {
			e.SlideIdx--
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
		}

	case "H":
		if e.BackHistory(d) {
			e.Message = fmt.Sprintf("back to slide %d/%d", e.SlideIdx+1, len(d.Slides))
		}

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if d != nil && e.SlideIdx < len(d.Slides) {
			branch := d.Slides[e.SlideIdx].FindBranchByKey(key)
			if branch != nil {
				if e.FollowBranch(branch.Target, d) {
					e.Message = fmt.Sprintf("branch [%s] ──► %s", branch.Key, branch.Label)
					return nil
				}
			}
		}

	case "M":
		e.ShowGraphMap = !e.ShowGraphMap
		if e.ShowGraphMap {
			e.GraphMapCursor = e.SlideIdx
			e.Message = "presentation graph map (arrows/jk to select, enter to jump, M/esc to close)"
		} else {
			e.Message = "graph map closed"
		}

	case "down", "j":
		e.MoveDown(d)
	case "up", "k":
		e.MoveUp(d)

	case "/":
		e.Mode = ModePrompt
		e.Draft = ""
		e.CursorCol = 0
		e.Message = "jump: enter slide number or search"

	case "g":
		e.SlideIdx = 0
		e.BlockIdx = 0
		e.ClampBlockIdx(d)
	case "G":
		e.SlideIdx = len(d.Slides) - 1
		e.BlockIdx = 0
		if e.SlideIdx < 0 {
			e.SlideIdx = 0
		}
		e.ClampBlockIdx(d)

	case "n":
		e.ShowNotes = !e.ShowNotes
		if e.ShowNotes {
			e.Message = "notes open (press 'n' to hide)"
		} else {
			e.Message = "notes closed"
		}

	case "z":
		e.ZenMode = !e.ZenMode

	case "c":
		e.ShowTimer = !e.ShowTimer
		if e.ShowTimer {
			if e.TimerStart.IsZero() {
				e.TimerStart = time.Now()
			}
			e.Message = "timer: on (press 'C' to reset)"
			return TickCmd()
		}
		e.Message = "timer: off"

	case "C":
		e.TimerStart = time.Now()
		e.ShowTimer = true
		e.Message = "timer: reset to 00:00"
		return TickCmd()

	case "r", "R":
		if err := e.Reload(d); err == nil {
			e.Message = "reloaded from disk"
		} else {
			e.Message = "reload error: " + err.Error()
		}

	case "L":
		e.ShowLineNumbers = !e.ShowLineNumbers
		if e.ShowLineNumbers {
			e.Message = "line numbers: on"
		} else {
			e.Message = "line numbers: off"
		}

	case "x":
		e.ToggleTask(d)

	case "o", "O":
		e.ShowOverview = true
		e.OverviewCursor = e.SlideIdx
		if e.OverviewCols <= 0 {
			e.OverviewCols = 3
		}
		e.Message = "slide overview (arrows/hjkl to navigate, enter to jump, esc/o to close)"

	case "b", "B":
		e.ScreenBlank = !e.ScreenBlank
		if e.ScreenBlank {
			e.Message = "screen blanked (press any key to resume)"
		} else {
			e.Message = "screen resumed"
		}

	case "y", "Y":
		text, err := e.YankBlock(d)
		if err != nil {
			e.Message = "yank: " + err.Error()
			return nil
		}
		blk := e.currentBlock(d)
		if blk != nil && blk.Kind == BlockCode {
			e.Message = fmt.Sprintf("yanked %d code lines to clipboard", len(blk.Lines))
		} else {
			e.Message = fmt.Sprintf("yanked %d chars to clipboard", len(text))
		}
		return tea.Printf("%s", OSC52Copy(text))

	case "E":
		_, _ = e.ExportHTML(d)

	case "S":
		e.ShowStats = !e.ShowStats
		if e.ShowStats {
			e.Message = "talk statistics (press 'S', 'q', or 'esc' to close)"
		} else {
			e.Message = "statistics closed"
		}

	case "A":
		return e.ToggleAutoplay()

	case "i", "a", "I":
		e.EnterEdit(d)

	case "ctrl+n":
		e.AddBlock(d)
	case "ctrl+d":
		e.DeleteBlock(d)

	case "ctrl+k":
		e.MoveBlockUp(d)
	case "ctrl+j":
		e.MoveBlockDown(d)

	case "ctrl+N":
		e.AddSlide(d)
	case "ctrl+D":
		e.DeleteSlide(d)

	case "u":
		if len(e.UndoStack) > 0 {
			currentState := SerializeDeck(*d)
			e.RedoStack = append(e.RedoStack, currentState)
			state := e.UndoStack[len(e.UndoStack)-1]
			e.UndoStack = e.UndoStack[:len(e.UndoStack)-1]
			baseDir := d.BaseDir
			deckAlign := d.Align
			deckTheme := d.Theme
			*d = ParseDeck(state)
			d.BaseDir = baseDir
			if d.Align == "" {
				d.Align = deckAlign
			}
			if d.Theme == "" {
				d.Theme = deckTheme
			}
			e.Theme = d.Theme
			e.Message = "undo"
		}
	case "ctrl+r":
		if len(e.RedoStack) > 0 {
			currentState := SerializeDeck(*d)
			e.UndoStack = append(e.UndoStack, currentState)
			state := e.RedoStack[len(e.RedoStack)-1]
			e.RedoStack = e.RedoStack[:len(e.RedoStack)-1]
			baseDir := d.BaseDir
			deckAlign := d.Align
			deckTheme := d.Theme
			*d = ParseDeck(state)
			d.BaseDir = baseDir
			if d.Align == "" {
				d.Align = deckAlign
			}
			if d.Theme == "" {
				d.Theme = deckTheme
			}
			e.Theme = d.Theme
			e.Message = "redo"
		}

	case "ctrl+s":
		e.Save(*d)

	case "tab", "ctrl+a":
		e.ToggleAlign(d)

	case "t", "T", "ctrl+t", "f2":
		e.CycleTheme(d)

	case "p":
		blk := e.currentBlock(d)
		if blk != nil && blk.Kind == BlockImage {
			fullPath, found := ResolveImagePath(blk.Src, d.BaseDir)
			if found {
				if err := openFile(fullPath); err == nil {
					e.Message = "opened " + filepath.Base(fullPath)
				} else {
					e.Message = fmt.Sprintf("open error: %v", err)
				}
			} else {
				e.Message = "image not found: " + blk.Src
			}
		}

	case "?", "f1":
		e.ShowHelp = !e.ShowHelp

	case "esc":
		if e.ShowHelp {
			e.ShowHelp = false
			return nil
		}
		if e.ShowOverview {
			e.ShowOverview = false
			return nil
		}
		if e.ShowGraphMap {
			e.ShowGraphMap = false
			return nil
		}
		if e.ShowStats {
			e.ShowStats = false
			return nil
		}
		e.Message = ""
	}

	return nil
}

func openFile(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// OSC52Copy constructs the terminal OSC 52 sequence for copying text to the host system clipboard.
func OSC52Copy(text string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("\x1b]52;c;%s\x07", b64)
}

// CopyToSystemClipboard attempts to dispatch text to standard host clipboard utilities.
func CopyToSystemClipboard(text string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		if _, err := exec.LookPath("wl-copy"); err == nil {
			cmd = exec.Command("wl-copy")
		} else if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		}
	}
	if cmd != nil {
		cmd.Stdin = strings.NewReader(text)
		_ = cmd.Start()
	}
}

// YankBlock extracts the text/code content of the focused block and dispatches it to the system clipboard.
func (e *Editor) YankBlock(d *Deck) (string, error) {
	blk := e.currentBlock(d)
	if blk == nil {
		return "", fmt.Errorf("no block selected")
	}
	var text string
	switch blk.Kind {
	case BlockCode:
		text = strings.Join(blk.Lines, "\n")
	case BlockTable:
		text = strings.Join(blk.Lines, "\n")
	case BlockCallout:
		if len(blk.Lines) > 0 {
			text = strings.Join(blk.Lines, "\n")
		} else {
			text = blk.Text
		}
	default:
		text = blk.Text
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("block is empty")
	}
	CopyToSystemClipboard(text)
	return text, nil
}

func (e *Editor) handleEdit(key string, d *Deck) tea.Cmd {
	switch key {
	case "esc", "ctrl+c":
		e.CancelEdit()
		return nil

	case "enter":
		e.ExitEdit(d)
		if e.FilePath != "" {
			if _, err := os.Stat(e.FilePath); err == nil {
				e.Save(*d)
			}
		}
		return nil

	case "ctrl+t", "f2":
		e.CycleTheme(d)
		return nil

	case "backspace":
		if e.CursorCol > 0 && len(e.Draft) > 0 {
			e.Draft = e.Draft[:e.CursorCol-1] + e.Draft[e.CursorCol:]
			e.CursorCol--
		}
		return nil

	case "delete":
		if e.CursorCol < len(e.Draft) {
			e.Draft = e.Draft[:e.CursorCol] + e.Draft[e.CursorCol+1:]
		}
		return nil

	case "left":
		if e.CursorCol > 0 {
			e.CursorCol--
		}
		return nil

	case "right":
		if e.CursorCol < len(e.Draft) {
			e.CursorCol++
		}
		return nil

	case "home", "0":
		e.CursorCol = 0
		return nil

	case "end", "$":
		e.CursorCol = len(e.Draft)
		return nil

	default:
		if len(key) == 1 && key[0] >= 32 {
			e.Draft = e.Draft[:e.CursorCol] + key + e.Draft[e.CursorCol:]
			e.CursorCol++
		}
	}

	return nil
}

func (e *Editor) handlePrompt(key string, d *Deck) tea.Cmd {
	switch key {
	case "esc", "ctrl+c":
		e.Mode = ModeNav
		e.Draft = ""
		e.Message = ""
		return nil

	case "enter":
		query := strings.TrimSpace(e.Draft)
		e.Mode = ModeNav
		e.Draft = ""
		if query == "" {
			e.Message = ""
			return nil
		}

		var slideNum int
		if n, err := fmt.Sscanf(query, "%d", &slideNum); err == nil && n == 1 {
			target := slideNum - 1
			if target < 0 {
				target = 0
			}
			if target >= len(d.Slides) {
				target = len(d.Slides) - 1
			}
			e.SlideIdx = target
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
			e.Message = fmt.Sprintf("jumped to slide %d/%d", target+1, len(d.Slides))
			return nil
		}

		lowerQuery := strings.ToLower(query)
		foundIdx := -1
		for idx, slide := range d.Slides {
			title := strings.ToLower(slide.Title())
			if strings.Contains(title, lowerQuery) {
				foundIdx = idx
				break
			}
			foundInBlock := false
			for _, b := range slide.Blocks {
				if strings.Contains(strings.ToLower(b.Text), lowerQuery) {
					foundInBlock = true
					break
				}
				for _, l := range b.Lines {
					if strings.Contains(strings.ToLower(l), lowerQuery) {
						foundInBlock = true
						break
					}
				}
				if foundInBlock {
					break
				}
			}
			if foundInBlock {
				foundIdx = idx
				break
			}
		}

		if foundIdx != -1 {
			e.SlideIdx = foundIdx
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
			e.Message = fmt.Sprintf("jumped to slide %d: %s", foundIdx+1, d.Slides[foundIdx].Title())
		} else {
			e.Message = fmt.Sprintf("no slide matching %q", query)
		}
		return nil

	case "backspace":
		if len(e.Draft) > 0 {
			e.Draft = e.Draft[:len(e.Draft)-1]
			e.CursorCol = len(e.Draft)
		}
		return nil

	default:
		if len(key) == 1 && key[0] >= 32 {
			e.Draft += key
			e.CursorCol = len(e.Draft)
		}
		return nil
	}
}
