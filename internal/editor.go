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
	RunningCode       bool
	ShowRunner        bool
	RunnerResult      *ExecResult
	FocusMode         bool
	FocusScroll       int
	ActiveTrack       string
	ShowTrackModal    bool
	TrackCursor       int
	ActiveRoute       string
	RouteStep         int
	ShowRouteModal    bool
	RouteCursor       int
	ShowHistoryModal  bool
	HistoryCursor     int
	ShowBranchHUD     bool
	BranchHUDCursor   int
	ShowWaypointModal bool
	WaypointCursor    int
	WaypointQuery     string
	ShowRadarModal    bool
	RadarCursor       int
	LoopCounters      map[int]int
}

func NewEditor(filePath string) Editor {
	return Editor{
		Mode:              ModeNav,
		FilePath:          filePath,
		OverviewCols:      3,
		AutoplayInterval:  5,
		AutoplayCountdown: 5,
		AutoplayLoop:      true,
		LoopCounters:      make(map[int]int),
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

func (e *Editor) JumpToHistory(stepIndex int, d *Deck) bool {
	if stepIndex < 0 || stepIndex >= len(e.History) {
		return false
	}
	targetIdx := e.History[stepIndex]
	e.History = e.History[:stepIndex]
	if d != nil && targetIdx >= 0 && targetIdx < len(d.Slides) {
		e.SlideIdx = targetIdx
		e.BlockIdx = 0
		e.ClampBlockIdx(d)
		e.Message = fmt.Sprintf("rewound to slide %d/%d", e.SlideIdx+1, len(d.Slides))
		return true
	}
	return false
}

func (e *Editor) SelectTrack(track string, d *Deck) {
	track = strings.TrimSpace(track)
	if strings.EqualFold(track, "all") {
		track = ""
	}
	e.ActiveTrack = track
	if track == "" {
		e.Message = "track: all slides"
	} else {
		e.Message = "track: " + track
		if d != nil && len(d.Slides) > 0 && !d.Slides[e.SlideIdx].HasTag(track) {
			matching := d.SlideIndicesForTag(track)
			if len(matching) > 0 {
				e.History = append(e.History, e.SlideIdx)
				e.SlideIdx = matching[0]
				e.BlockIdx = 0
				e.ClampBlockIdx(d)
			}
		}
	}
}

func (e *Editor) NextTrackSlide(d *Deck) bool {
	if d == nil || len(d.Slides) == 0 {
		return false
	}
	if e.ActiveTrack == "" {
		if e.SlideIdx < len(d.Slides)-1 {
			e.SlideIdx++
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
			return true
		}
		return false
	}
	for i := 1; i <= len(d.Slides); i++ {
		targetIdx := (e.SlideIdx + i) % len(d.Slides)
		if d.Slides[targetIdx].HasTag(e.ActiveTrack) {
			if targetIdx != e.SlideIdx {
				e.History = append(e.History, e.SlideIdx)
				e.SlideIdx = targetIdx
				e.BlockIdx = 0
				e.ClampBlockIdx(d)
				return true
			}
			return false
		}
	}
	return false
}

func (e *Editor) PrevTrackSlide(d *Deck) bool {
	if d == nil || len(d.Slides) == 0 {
		return false
	}
	if e.ActiveTrack == "" {
		if e.SlideIdx > 0 {
			e.SlideIdx--
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
			return true
		}
		return false
	}
	for i := 1; i <= len(d.Slides); i++ {
		targetIdx := (e.SlideIdx - i + len(d.Slides)) % len(d.Slides)
		if d.Slides[targetIdx].HasTag(e.ActiveTrack) {
			if targetIdx != e.SlideIdx {
				e.History = append(e.History, e.SlideIdx)
				e.SlideIdx = targetIdx
				e.BlockIdx = 0
				e.ClampBlockIdx(d)
				return true
			}
			return false
		}
	}
	return false
}

func (e *Editor) SelectRoute(routeName string, d *Deck) {
	routeName = strings.TrimSpace(routeName)
	if strings.EqualFold(routeName, "none") || strings.EqualFold(routeName, "off") || strings.EqualFold(routeName, "all") {
		routeName = ""
	}
	if routeName == "" {
		e.ActiveRoute = ""
		e.RouteStep = 0
		e.Message = "route: none (free graph traversal)"
		return
	}
	if d == nil || len(d.Slides) == 0 {
		return
	}
	indices := d.RouteSlideIndices(routeName)
	if len(indices) == 0 {
		e.Message = "route " + routeName + ": no matching slides"
		return
	}
	e.ActiveRoute = routeName
	// Check if current slide is on this route
	foundStep := -1
	for step, idx := range indices {
		if idx == e.SlideIdx {
			foundStep = step
			break
		}
	}
	if foundStep >= 0 {
		e.RouteStep = foundStep
		e.Message = fmt.Sprintf("route: %s (step %d/%d: %s)", routeName, e.RouteStep+1, len(indices), d.Slides[e.SlideIdx].Title())
	} else {
		e.History = append(e.History, e.SlideIdx)
		e.RouteStep = 0
		e.SlideIdx = indices[0]
		e.BlockIdx = 0
		e.ClampBlockIdx(d)
		e.Message = fmt.Sprintf("route: %s (step 1/%d: %s)", routeName, len(indices), d.Slides[e.SlideIdx].Title())
	}
}

func (e *Editor) NextRouteSlide(d *Deck) bool {
	if d == nil || len(d.Slides) == 0 || e.ActiveRoute == "" {
		return false
	}
	indices := d.RouteSlideIndices(e.ActiveRoute)
	if len(indices) == 0 {
		return false
	}
	if e.RouteStep < len(indices)-1 {
		e.History = append(e.History, e.SlideIdx)
		e.RouteStep++
		e.SlideIdx = indices[e.RouteStep]
		e.BlockIdx = 0
		e.ClampBlockIdx(d)
		e.Message = fmt.Sprintf("route: %s (step %d/%d: %s)", e.ActiveRoute, e.RouteStep+1, len(indices), d.Slides[e.SlideIdx].Title())
		return true
	}
	e.Message = fmt.Sprintf("route: %s (completed - step %d/%d)", e.ActiveRoute, len(indices), len(indices))
	return false
}

func (e *Editor) PrevRouteSlide(d *Deck) bool {
	if d == nil || len(d.Slides) == 0 || e.ActiveRoute == "" {
		return false
	}
	indices := d.RouteSlideIndices(e.ActiveRoute)
	if len(indices) == 0 {
		return false
	}
	if e.RouteStep > 0 {
		e.History = append(e.History, e.SlideIdx)
		e.RouteStep--
		e.SlideIdx = indices[e.RouteStep]
		e.BlockIdx = 0
		e.ClampBlockIdx(d)
		e.Message = fmt.Sprintf("route: %s (step %d/%d: %s)", e.ActiveRoute, e.RouteStep+1, len(indices), d.Slides[e.SlideIdx].Title())
		return true
	}
	return false
}

func (e *Editor) ToggleBranchHUD(d *Deck) {
	if d == nil || len(d.Slides) == 0 {
		return
	}
	options := GetForkOptions(e.SlideIdx, *d, e.ActiveTrack, e.ActiveRoute)
	if len(options) == 0 {
		e.Message = "terminal slide: no outgoing branches"
		return
	}
	e.ShowBranchHUD = !e.ShowBranchHUD
	if e.ShowBranchHUD {
		e.BranchHUDCursor = 0
		e.Message = "branch HUD: select path (arrows/jk, enter/1-9 to jump, J/esc to close)"
	} else {
		e.Message = "branch HUD closed"
	}
}

func (e *Editor) JumpToForkOption(opt BranchForkOption, d *Deck) bool {
	if d == nil || opt.TargetIndex < 0 || opt.TargetIndex >= len(d.Slides) {
		return false
	}
	e.History = append(e.History, e.SlideIdx)
	e.SlideIdx = opt.TargetIndex
	e.BlockIdx = 0
	e.ClampBlockIdx(d)
	e.ShowBranchHUD = false
	if opt.Key != "" && opt.Label != "" {
		e.Message = fmt.Sprintf("branch [%s] ──► %s", opt.Key, opt.Label)
	} else {
		e.Message = fmt.Sprintf("jumped to slide %d/%d", e.SlideIdx+1, len(d.Slides))
	}
	return true
}

func (e *Editor) ToggleWaypointModal(d *Deck) {
	if d == nil || len(d.Slides) <= 1 {
		e.Message = "waypoint pathfinder: presentation must have multiple slides"
		return
	}
	e.ShowWaypointModal = !e.ShowWaypointModal
	if e.ShowWaypointModal {
		e.WaypointCursor = 0
		e.WaypointQuery = ""
		e.Message = "waypoint pathfinder: select destination (arrows/jk, enter: follow path, w: step, esc: close)"
	} else {
		e.Message = "waypoint pathfinder closed"
	}
}

func (e *Editor) ApplyWaypointPath(cand WaypointCandidate, d *Deck) bool {
	if d == nil || !cand.Reachable || len(cand.Path) < 2 {
		return false
	}
	routeName := fmt.Sprintf("waypoint-%d", cand.SlideIndex+1)
	var slugs []string
	for _, nodeIdx := range cand.Path {
		if nodeIdx >= 0 && nodeIdx < len(d.Slides) {
			s := d.Slides[nodeIdx]
			slug := s.ID
			if slug == "" {
				slug = s.Slug()
			}
			slugs = append(slugs, slug)
		}
	}
	if d.Routes == nil {
		d.Routes = make(map[string][]string)
	}
	d.Routes[routeName] = slugs
	e.SelectRoute(routeName, d)
	e.ShowWaypointModal = false
	e.Message = fmt.Sprintf("following optimal path to slide %d: %s (%d hops · ~%dm)", cand.SlideIndex+1, cand.Title, cand.HopCount, cand.EstMin)
	return true
}

func (e *Editor) StepWaypointPath(cand WaypointCandidate, d *Deck) bool {
	if d == nil || !cand.Reachable || len(cand.Path) < 2 {
		return false
	}
	nextIdx := cand.Path[1]
	e.History = append(e.History, e.SlideIdx)
	e.SlideIdx = nextIdx
	e.BlockIdx = 0
	e.ClampBlockIdx(d)
	e.ShowWaypointModal = false
	e.Message = fmt.Sprintf("stepped along path to [%02d] %s (target: [%02d] %s)", e.SlideIdx+1, d.Slides[e.SlideIdx].Title(), cand.SlideIndex+1, cand.Title)
	return true
}

func (e Editor) BuildVisitedMap() map[int]bool {
	visited := make(map[int]bool)
	visited[0] = true
	for _, sIdx := range e.History {
		visited[sIdx] = true
	}
	visited[e.SlideIdx] = true
	return visited
}

func (e *Editor) FindLastForkIndex(d *Deck) int {
	if d == nil || len(d.Slides) == 0 {
		return -1
	}
	for i := len(e.History) - 1; i >= 0; i-- {
		sIdx := e.History[i]
		if sIdx >= 0 && sIdx < len(d.Slides) && sIdx != e.SlideIdx {
			s := d.Slides[sIdx]
			if len(s.Branches()) > 1 {
				return sIdx
			}
		}
	}
	return -1
}

func (e *Editor) ReturnToUpstreamFork(d *Deck) bool {
	if d == nil || len(d.Slides) == 0 {
		return false
	}
	forkIdx := e.FindLastForkIndex(d)
	if forkIdx < 0 {
		e.Message = "no prior fork found in traversal history"
		return false
	}

	lastStep := -1
	for i := len(e.History) - 1; i >= 0; i-- {
		if e.History[i] == forkIdx {
			lastStep = i
			break
		}
	}
	if lastStep >= 0 {
		e.History = e.History[:lastStep]
	}
	e.SlideIdx = forkIdx
	e.BlockIdx = 0
	e.ClampBlockIdx(d)
	e.Message = fmt.Sprintf("⤺ returned to upstream fork: [%02d] %s", forkIdx+1, d.Slides[forkIdx].Title())
	return true
}

func (e *Editor) ToggleRadarModal(d *Deck) {
	if d == nil || len(d.Slides) == 0 {
		e.Message = "exploration radar: presentation has no slides"
		return
	}
	e.ShowRadarModal = !e.ShowRadarModal
	if e.ShowRadarModal {
		e.RadarCursor = 0
		e.ShowHelp = false
		e.ShowStats = false
		e.ShowOverview = false
		e.ShowGraphMap = false
		e.ShowTrackModal = false
		e.ShowRouteModal = false
		e.ShowHistoryModal = false
		e.ShowBranchHUD = false
		e.ShowWaypointModal = false
		e.Message = "graph exploration radar: browse branches (arrows/jk, enter: jump, u: fork, esc: close)"
	} else {
		e.Message = "exploration radar closed"
	}
}

func (e *Editor) JumpToRadarBranch(item BranchCoverageItem, d *Deck) bool {
	if d == nil || item.TargetIdx < 0 || item.TargetIdx >= len(d.Slides) {
		return false
	}
	e.History = append(e.History, e.SlideIdx)
	e.SlideIdx = item.TargetIdx
	e.BlockIdx = 0
	e.ClampBlockIdx(d)
	e.ShowRadarModal = false
	e.Message = fmt.Sprintf("jumped to branch [%s] %s", item.BranchKey, item.TargetTitle)
	return true
}

func (e Editor) CurrentLoopPass(slideIdx int, d *Deck) (pass int, maxPasses int, hasLoop bool) {
	if d == nil || slideIdx < 0 || slideIdx >= len(d.Slides) {
		return 0, 0, false
	}
	s := d.Slides[slideIdx]
	if s.Loop == nil {
		return 0, 0, false
	}
	if e.LoopCounters != nil {
		pass = e.LoopCounters[slideIdx]
	}
	maxPasses = s.Loop.MaxPasses
	if maxPasses <= 0 {
		maxPasses = 3
	}
	return pass, maxPasses, true
}

func (e *Editor) ResetLoopCounter(slideIdx int) {
	if e.LoopCounters != nil {
		delete(e.LoopCounters, slideIdx)
	}
}

func (e *Editor) ResetAllLoops() {
	e.LoopCounters = make(map[int]int)
}

func (e *Editor) AdvanceLoop(slideIdx int, d *Deck) (tookLoop bool, targetIdx int) {
	if d == nil || slideIdx < 0 || slideIdx >= len(d.Slides) {
		return false, -1
	}
	s := d.Slides[slideIdx]
	if s.Loop == nil {
		return false, -1
	}
	if e.LoopCounters == nil {
		e.LoopCounters = make(map[int]int)
	}
	pass := e.LoopCounters[slideIdx]
	maxPasses := s.Loop.MaxPasses
	if maxPasses <= 0 {
		maxPasses = 3
	}

	if pass < maxPasses {
		e.LoopCounters[slideIdx] = pass + 1
		tgtIdx := d.FindSlideByID(s.Loop.Target)
		if tgtIdx >= 0 && tgtIdx < len(d.Slides) {
			e.History = append(e.History, e.SlideIdx)
			e.SlideIdx = tgtIdx
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
			e.Message = fmt.Sprintf("⟳ loop [%s]: pass %d/%d ──► %s", s.Loop.Label, pass+1, maxPasses, s.Loop.Target)
			return true, tgtIdx
		}
	}

	// Loop pass threshold reached or target invalid -> Exit loop
	exitTarget := s.Loop.ExitTarget
	if exitTarget == "" {
		exitTarget = s.NextID
	}
	if exitTarget != "" {
		exitIdx := d.FindSlideByID(exitTarget)
		if exitIdx >= 0 && exitIdx < len(d.Slides) {
			e.History = append(e.History, e.SlideIdx)
			e.SlideIdx = exitIdx
			e.BlockIdx = 0
			e.ClampBlockIdx(d)
			e.Message = fmt.Sprintf("✔ loop [%s] completed (%d/%d) ──► %s", s.Loop.Label, maxPasses, maxPasses, exitTarget)
			return false, exitIdx
		}
	}

	if e.SlideIdx < len(d.Slides)-1 {
		e.SlideIdx++
		e.BlockIdx = 0
		e.ClampBlockIdx(d)
		e.Message = fmt.Sprintf("✔ loop [%s] completed (%d/%d)", s.Loop.Label, maxPasses, maxPasses)
		return false, e.SlideIdx
	}

	return false, -1
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

// --- Live code runner & focus mode ---

func (e *Editor) RunFocusedCode(d *Deck) tea.Cmd {
	if d == nil || e.SlideIdx >= len(d.Slides) {
		e.Message = "no slide active"
		return nil
	}
	slide := d.Slides[e.SlideIdx]
	var targetBlock *Block
	blkIdx := e.BlockIdx

	findCodeInColumns := func(cols [][]Block) *Block {
		for _, col := range cols {
			for i := range col {
				if col[i].Kind == BlockCode {
					return &col[i]
				}
			}
		}
		return nil
	}

	// 1. Check current block
	if blkIdx >= 0 && blkIdx < len(slide.Blocks) {
		if slide.Blocks[blkIdx].Kind == BlockCode {
			targetBlock = &slide.Blocks[blkIdx]
		} else if slide.Blocks[blkIdx].Kind == BlockColumns {
			targetBlock = findCodeInColumns(slide.Blocks[blkIdx].Columns)
		}
	}
	// 2. Search for first BlockCode on this slide
	if targetBlock == nil {
		for i, b := range slide.Blocks {
			if b.Kind == BlockCode {
				targetBlock = &slide.Blocks[i]
				blkIdx = i
				e.BlockIdx = i
				break
			} else if b.Kind == BlockColumns {
				if inner := findCodeInColumns(b.Columns); inner != nil {
					targetBlock = inner
					blkIdx = i
					e.BlockIdx = i
					break
				}
			}
		}
	}

	if targetBlock == nil {
		e.Message = "no code block found on this slide"
		return nil
	}

	if targetBlock.NoEval {
		e.Message = "code block is marked display-only (no-eval)"
		return nil
	}

	lang := targetBlock.Lang
	if !IsExecutableLanguage(lang) && lang != "" {
		e.Message = fmt.Sprintf("language '%s' is display-only (supported: sh, python, go, node, ruby)", lang)
		return nil
	}

	e.RunningCode = true
	e.ShowRunner = false
	if lang == "" {
		lang = "sh"
	}
	e.Message = fmt.Sprintf("executing [%s] code...", lang)
	return ExecuteCodeCmd(*targetBlock, 5*time.Second, e.SlideIdx+1, blkIdx+1)
}

func (e *Editor) ToggleFocusMode(d *Deck) {
	e.FocusMode = !e.FocusMode
	e.FocusScroll = 0
	if e.FocusMode {
		e.Message = "focus mode: on (j/k to scroll, X to run, f/esc to exit)"
	} else {
		e.Message = "focus mode: off"
	}
}

func (e *Editor) DismissRunner() {
	e.ShowRunner = false
	e.Message = "runner closed"
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

	if e.ShowTrackModal {
		tags := d.AllTags()
		totalOptions := len(tags) + 1
		switch key {
		case "esc", "q", "K":
			e.ShowTrackModal = false
			return nil
		case "up", "k":
			if e.TrackCursor > 0 {
				e.TrackCursor--
			}
			return nil
		case "down", "j":
			if e.TrackCursor < totalOptions-1 {
				e.TrackCursor++
			}
			return nil
		case "g", "home":
			e.TrackCursor = 0
			return nil
		case "G", "end":
			if totalOptions > 0 {
				e.TrackCursor = totalOptions - 1
			}
			return nil
		case "0":
			e.SelectTrack("", d)
			e.ShowTrackModal = false
			return nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(key[0] - '0')
			if idx <= len(tags) {
				e.SelectTrack(tags[idx-1], d)
				e.ShowTrackModal = false
				return nil
			}
		case "enter", " ":
			if e.TrackCursor == 0 {
				e.SelectTrack("", d)
			} else if e.TrackCursor-1 < len(tags) {
				e.SelectTrack(tags[e.TrackCursor-1], d)
			}
			e.ShowTrackModal = false
			return nil
		}
		return nil
	}

	if e.ShowRouteModal {
		routeNames := d.AllRouteNames()
		totalOptions := len(routeNames) + 1
		switch key {
		case "esc", "q", "P":
			e.ShowRouteModal = false
			return nil
		case "up", "k":
			if e.RouteCursor > 0 {
				e.RouteCursor--
			}
			return nil
		case "down", "j":
			if e.RouteCursor < totalOptions-1 {
				e.RouteCursor++
			}
			return nil
		case "g", "home":
			e.RouteCursor = 0
			return nil
		case "G", "end":
			if totalOptions > 0 {
				e.RouteCursor = totalOptions - 1
			}
			return nil
		case "0":
			e.SelectRoute("", d)
			e.ShowRouteModal = false
			return nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(key[0] - '0')
			if idx <= len(routeNames) {
				e.SelectRoute(routeNames[idx-1], d)
				e.ShowRouteModal = false
				return nil
			}
		case "enter", " ":
			if e.RouteCursor == 0 {
				e.SelectRoute("", d)
			} else if e.RouteCursor-1 < len(routeNames) {
				e.SelectRoute(routeNames[e.RouteCursor-1], d)
			}
			e.ShowRouteModal = false
			return nil
		}
		return nil
	}

	if e.ShowHistoryModal {
		totalSteps := len(e.History) + 1
		switch key {
		case "esc", "q", "H":
			e.ShowHistoryModal = false
			return nil
		case "up", "k":
			if e.HistoryCursor > 0 {
				e.HistoryCursor--
			}
			return nil
		case "down", "j":
			if e.HistoryCursor < totalSteps-1 {
				e.HistoryCursor++
			}
			return nil
		case "g", "home":
			e.HistoryCursor = 0
			return nil
		case "G", "end":
			if totalSteps > 0 {
				e.HistoryCursor = totalSteps - 1
			}
			return nil
		case "c", "C":
			e.History = nil
			e.HistoryCursor = 0
			e.Message = "traversal history cleared"
			e.ShowHistoryModal = false
			return nil
		case "backspace":
			if e.BackHistory(d) {
				if e.HistoryCursor > len(e.History) {
					e.HistoryCursor = len(e.History)
				}
				e.Message = fmt.Sprintf("back to slide %d/%d", e.SlideIdx+1, len(d.Slides))
			}
			return nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			step := int(key[0]-'0') - 1
			if step < len(e.History) {
				e.JumpToHistory(step, d)
				e.ShowHistoryModal = false
				return nil
			}
		case "enter", " ":
			if e.HistoryCursor < len(e.History) {
				e.JumpToHistory(e.HistoryCursor, d)
			}
			e.ShowHistoryModal = false
			return nil
		}
		return nil
	}

	if e.ShowBranchHUD {
		options := GetForkOptions(e.SlideIdx, *d, e.ActiveTrack, e.ActiveRoute)
		switch key {
		case "esc", "q", "J":
			e.ShowBranchHUD = false
			e.Message = "branch HUD closed"
			return nil
		case "up", "k":
			if e.BranchHUDCursor > 0 {
				e.BranchHUDCursor--
			}
			return nil
		case "down", "j":
			if e.BranchHUDCursor < len(options)-1 {
				e.BranchHUDCursor++
			}
			return nil
		case "g", "home":
			e.BranchHUDCursor = 0
			return nil
		case "G", "end":
			if len(options) > 0 {
				e.BranchHUDCursor = len(options) - 1
			}
			return nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			for _, opt := range options {
				if opt.Key == key {
					e.JumpToForkOption(opt, d)
					return nil
				}
			}
			idx := int(key[0] - '0')
			if idx <= len(options) {
				e.JumpToForkOption(options[idx-1], d)
				return nil
			}
		case "enter", " ":
			if e.BranchHUDCursor >= 0 && e.BranchHUDCursor < len(options) {
				e.JumpToForkOption(options[e.BranchHUDCursor], d)
			}
			return nil
		}
		return nil
	}

	if e.ShowWaypointModal {
		candidates := FindWaypointCandidates(e.SlideIdx, *d, e.WaypointQuery)
		switch key {
		case "esc", "W":
			e.ShowWaypointModal = false
			e.Message = "waypoint pathfinder closed"
			return nil
		case "up", "k":
			if e.WaypointCursor > 0 {
				e.WaypointCursor--
			}
			return nil
		case "down", "j":
			if e.WaypointCursor < len(candidates)-1 {
				e.WaypointCursor++
			}
			return nil
		case "g", "home":
			e.WaypointCursor = 0
			return nil
		case "G", "end":
			if len(candidates) > 0 {
				e.WaypointCursor = len(candidates) - 1
			}
			return nil
		case "backspace":
			if len(e.WaypointQuery) > 0 {
				e.WaypointQuery = e.WaypointQuery[:len(e.WaypointQuery)-1]
				e.WaypointCursor = 0
			} else {
				e.ShowWaypointModal = false
				e.Message = "waypoint pathfinder closed"
			}
			return nil
		case "enter":
			if e.WaypointCursor >= 0 && e.WaypointCursor < len(candidates) {
				cand := candidates[e.WaypointCursor]
				if cand.Reachable {
					e.ApplyWaypointPath(cand, d)
				} else {
					e.Message = fmt.Sprintf("target [%02d] %s is not reachable downstream from current slide", cand.SlideIndex+1, cand.Title)
				}
			}
			return nil
		case "w":
			if e.WaypointCursor >= 0 && e.WaypointCursor < len(candidates) {
				cand := candidates[e.WaypointCursor]
				if cand.Reachable {
					e.StepWaypointPath(cand, d)
				} else {
					e.Message = fmt.Sprintf("target [%02d] %s is not reachable downstream", cand.SlideIndex+1, cand.Title)
				}
			}
			return nil
		default:
			if len(key) == 1 && key[0] >= 32 {
				e.WaypointQuery += key
				e.WaypointCursor = 0
			}
			return nil
		}
	}

	if e.ShowRadarModal {
		bg := BuildGraph(*d)
		visitedMap := e.BuildVisitedMap()
		radar := bg.CalculateRadarStats(*d, visitedMap, e.SlideIdx)
		switch key {
		case "esc", "q", "V":
			e.ShowRadarModal = false
			e.Message = "graph exploration radar closed"
			return nil
		case "up", "k":
			if e.RadarCursor > 0 {
				e.RadarCursor--
			}
			return nil
		case "down", "j":
			if e.RadarCursor < len(radar.Items)-1 {
				e.RadarCursor++
			}
			return nil
		case "g", "home":
			e.RadarCursor = 0
			return nil
		case "G", "end":
			if len(radar.Items) > 0 {
				e.RadarCursor = len(radar.Items) - 1
			}
			return nil
		case "u", "U":
			if len(radar.Items) > 0 && e.RadarCursor >= 0 && e.RadarCursor < len(radar.Items) {
				item := radar.Items[e.RadarCursor]
				e.History = append(e.History, e.SlideIdx)
				e.SlideIdx = item.ForkSlideIdx
				e.BlockIdx = 0
				e.ClampBlockIdx(d)
				e.ShowRadarModal = false
				e.Message = fmt.Sprintf("jumped to fork: [%02d] %s", item.ForkSlideIdx+1, item.ForkTitle)
			}
			return nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			for _, item := range radar.Items {
				if item.BranchKey == key {
					e.JumpToRadarBranch(item, d)
					return nil
				}
			}
			idx := int(key[0] - '1')
			if idx >= 0 && idx < len(radar.Items) {
				e.JumpToRadarBranch(radar.Items[idx], d)
				return nil
			}
		case "enter", " ":
			if len(radar.Items) > 0 && e.RadarCursor >= 0 && e.RadarCursor < len(radar.Items) {
				e.JumpToRadarBranch(radar.Items[e.RadarCursor], d)
			}
			return nil
		}
		return nil
	}

	if e.FocusMode {
		switch key {
		case "esc", "f", "F", "q", "ctrl+c":
			e.FocusMode = false
			e.FocusScroll = 0
			e.Message = "focus mode: off"
			return nil
		case "down", "j":
			e.FocusScroll++
			return nil
		case "up", "k":
			if e.FocusScroll > 0 {
				e.FocusScroll--
			}
			return nil
		case "g", "home":
			e.FocusScroll = 0
			return nil
		case "G", "end":
			e.FocusScroll = 100
			return nil
		case "X", "ctrl+x":
			return e.RunFocusedCode(d)
		case "x":
			blk := e.currentBlock(d)
			if blk != nil && blk.Kind == BlockCode {
				return e.RunFocusedCode(d)
			}
			e.ToggleTask(d)
			return nil
		case "y", "Y":
			if e.ShowRunner && e.RunnerResult != nil {
				output := e.RunnerResult.Stdout
				if e.RunnerResult.Stderr != "" {
					if output != "" {
						output += "\n"
					}
					output += e.RunnerResult.Stderr
				}
				e.Message = fmt.Sprintf("yanked %d chars of execution output to clipboard", len(output))
				return tea.Printf("%s", OSC52Copy(output))
			}
			text, err := e.YankBlock(d)
			if err != nil {
				e.Message = "yank: " + err.Error()
				return nil
			}
			return tea.Printf("%s", OSC52Copy(text))
		case "t", "T":
			e.CycleTheme(d)
			return nil
		case "L":
			e.ShowLineNumbers = !e.ShowLineNumbers
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
		if e.ShowRunner {
			e.ShowRunner = false
		}
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
		if e.ActiveRoute != "" {
			if e.NextRouteSlide(d) {
				return nil
			}
		}
		if d != nil && e.SlideIdx >= 0 && e.SlideIdx < len(d.Slides) && d.Slides[e.SlideIdx].Loop != nil {
			pass, maxPasses, _ := e.CurrentLoopPass(e.SlideIdx, d)
			if pass < maxPasses {
				e.AdvanceLoop(e.SlideIdx, d)
				return nil
			}
			curLoop := d.Slides[e.SlideIdx].Loop
			loopLabel := curLoop.Label
			exitTarget := curLoop.ExitTarget
			if exitTarget == "" {
				exitTarget = d.Slides[e.SlideIdx].NextID
			}
			if exitTarget != "" {
				nextIdx := d.FindSlideByID(exitTarget)
				if nextIdx >= 0 && nextIdx < len(d.Slides) {
					e.History = append(e.History, e.SlideIdx)
					e.SlideIdx = nextIdx
					e.BlockIdx = 0
					e.ClampBlockIdx(d)
					e.Message = fmt.Sprintf("✔ loop [%s] completed (%d/%d) ──► %s", loopLabel, maxPasses, maxPasses, exitTarget)
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
		if e.ShowRunner {
			e.ShowRunner = false
		}
		if e.Autoplay {
			e.AutoplayCountdown = e.AutoplayInterval
		}
		if e.ActiveRoute != "" {
			if e.PrevRouteSlide(d) {
				return nil
			}
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
		if e.ShowRunner {
			e.ShowRunner = false
		}
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
		if e.ShowRunner {
			e.ShowRunner = false
		}
		e.ShowHistoryModal = !e.ShowHistoryModal
		e.HistoryCursor = len(e.History)
		return nil

	case "J":
		if e.ShowRunner {
			e.ShowRunner = false
		}
		e.ToggleBranchHUD(d)
		return nil

	case "W":
		if e.ShowRunner {
			e.ShowRunner = false
		}
		e.ToggleWaypointModal(d)
		return nil

	case "V":
		if e.ShowRunner {
			e.ShowRunner = false
		}
		e.ToggleRadarModal(d)
		return nil

	case "U":
		if e.ShowRunner {
			e.ShowRunner = false
		}
		e.ReturnToUpstreamFork(d)
		return nil

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if e.ShowRunner {
			e.ShowRunner = false
		}
		if d != nil && e.SlideIdx < len(d.Slides) {
			if d.Slides[e.SlideIdx].Loop != nil && d.Slides[e.SlideIdx].Loop.Key == key {
				e.AdvanceLoop(e.SlideIdx, d)
				return nil
			}
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
		e.ResetAllLoops()
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
		if d != nil && e.SlideIdx >= 0 && e.SlideIdx < len(d.Slides) && d.Slides[e.SlideIdx].Loop != nil && strings.EqualFold(d.Slides[e.SlideIdx].Loop.Key, "r") {
			e.AdvanceLoop(e.SlideIdx, d)
			return nil
		}
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
		blk := e.currentBlock(d)
		if blk != nil && blk.Kind == BlockCode {
			return e.RunFocusedCode(d)
		}
		e.ToggleTask(d)

	case "X", "ctrl+x":
		return e.RunFocusedCode(d)

	case "f", "F":
		e.ToggleFocusMode(d)
		return nil

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
		if e.ShowRunner && e.RunnerResult != nil {
			output := e.RunnerResult.Stdout
			if e.RunnerResult.Stderr != "" {
				if output != "" {
					output += "\n"
				}
				output += e.RunnerResult.Stderr
			}
			e.Message = fmt.Sprintf("yanked %d chars of execution output to clipboard", len(output))
			return tea.Printf("%s", OSC52Copy(output))
		}
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

	case "K":
		e.ShowTrackModal = !e.ShowTrackModal
		e.TrackCursor = 0
		e.ShowHelp = false
		e.ShowOverview = false
		e.ShowStats = false
		e.ShowGraphMap = false
		e.ShowRouteModal = false
		e.DismissRunner()
		return nil

	case "P":
		e.ShowRouteModal = !e.ShowRouteModal
		e.RouteCursor = 0
		e.ShowHelp = false
		e.ShowOverview = false
		e.ShowStats = false
		e.ShowGraphMap = false
		e.ShowTrackModal = false
		e.DismissRunner()
		return nil

	case "]", "ctrl+]":
		e.NextTrackSlide(d)
		return nil

	case "[", "ctrl+[":
		e.PrevTrackSlide(d)
		return nil

	case "?", "f1":
		e.ShowHelp = !e.ShowHelp

	case "esc":
		if e.ShowRunner {
			e.ShowRunner = false
			e.Message = "runner closed"
			return nil
		}
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
		if e.ShowTrackModal {
			e.ShowTrackModal = false
			return nil
		}
		if e.ShowRouteModal {
			e.ShowRouteModal = false
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
