package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Editor modes ---

type EditorMode int

const (
	ModeNav EditorMode = iota
	ModeEdit
	ModePrompt
)

// --- Editor state ---

type Editor struct {
	Mode      EditorMode
	SlideIdx  int
	BlockIdx  int
	CursorCol int
	Draft     string
	UndoStack []string
	RedoStack []string
	FilePath  string
	Dirty     bool
	Message   string
	ShowNotes bool
	ShowHelp  bool
	ZenMode   bool
	Theme     string
}

func NewEditor(filePath string) Editor {
	return Editor{
		Mode:     ModeNav,
		FilePath: filePath,
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

// --- Input handling ---

func (e *Editor) HandleKey(msg tea.KeyMsg, d *Deck) tea.Cmd {
	key := msg.String()

	switch e.Mode {
	case ModeNav:
		return e.handleNav(key, d)
	case ModeEdit:
		return e.handleEdit(key, d)
	}
	return nil
}

func (e *Editor) handleNav(key string, d *Deck) tea.Cmd {
	switch key {
	case "q", "ctrl+c":
		if e.Dirty && e.FilePath != "" {
			e.Save(*d)
		}
		return tea.Quit

	case "right", "l", " ", "enter", "pgdown":
		if e.SlideIdx < len(d.Slides)-1 {
			e.SlideIdx++
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
		}
	case "left", "h", "pgup", "backspace":
		if e.SlideIdx > 0 {
			e.SlideIdx--
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
		}

	case "down", "j":
		e.MoveDown(d)
	case "up", "k":
		e.MoveUp(d)

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

	case "i", "a", "o", "I", "A", "O":
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
