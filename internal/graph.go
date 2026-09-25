package internal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type GraphEdgeKind int

const (
	EdgeLinear GraphEdgeKind = iota
	EdgeBranch
	EdgeNext
	EdgePrev
)

type GraphEdge struct {
	FromIndex int
	ToIndex   int
	Kind      GraphEdgeKind
	Key       string
	Label     string
	TargetID  string
}

type GraphNode struct {
	Index    int
	ID       string
	Slug     string
	Title    string
	Tags     []string
	HasFork  bool
	OutEdges []GraphEdge
	InEdges  []GraphEdge
}

type DeckGraph struct {
	Nodes []GraphNode
	Edges []GraphEdge
}

func BuildGraph(d Deck) DeckGraph {
	g := DeckGraph{
		Nodes: make([]GraphNode, len(d.Slides)),
	}

	for i, s := range d.Slides {
		g.Nodes[i] = GraphNode{
			Index:   i,
			ID:      s.ID,
			Slug:    s.Slug(),
			Title:   s.Title(),
			Tags:    s.Tags,
			HasFork: len(s.Branches()) > 0,
		}
	}

	for i, s := range d.Slides {
		branches := s.Branches()
		if len(branches) > 0 {
			for _, b := range branches {
				targetIdx := d.FindSlideByID(b.Target)
				edge := GraphEdge{
					FromIndex: i,
					ToIndex:   targetIdx,
					Kind:      EdgeBranch,
					Key:       b.Key,
					Label:     b.Label,
					TargetID:  b.Target,
				}
				g.Edges = append(g.Edges, edge)
				g.Nodes[i].OutEdges = append(g.Nodes[i].OutEdges, edge)
				if targetIdx >= 0 && targetIdx < len(g.Nodes) {
					g.Nodes[targetIdx].InEdges = append(g.Nodes[targetIdx].InEdges, edge)
				}
			}
		}

		if s.NextID != "" {
			targetIdx := d.FindSlideByID(s.NextID)
			edge := GraphEdge{
				FromIndex: i,
				ToIndex:   targetIdx,
				Kind:      EdgeNext,
				TargetID:  s.NextID,
			}
			g.Edges = append(g.Edges, edge)
			g.Nodes[i].OutEdges = append(g.Nodes[i].OutEdges, edge)
			if targetIdx >= 0 && targetIdx < len(g.Nodes) {
				g.Nodes[targetIdx].InEdges = append(g.Nodes[targetIdx].InEdges, edge)
			}
		} else if len(branches) == 0 && i+1 < len(d.Slides) {
			edge := GraphEdge{
				FromIndex: i,
				ToIndex:   i + 1,
				Kind:      EdgeLinear,
			}
			g.Edges = append(g.Edges, edge)
			g.Nodes[i].OutEdges = append(g.Nodes[i].OutEdges, edge)
			g.Nodes[i+1].InEdges = append(g.Nodes[i+1].InEdges, edge)
		}

		if s.PrevID != "" {
			targetIdx := d.FindSlideByID(s.PrevID)
			edge := GraphEdge{
				FromIndex: i,
				ToIndex:   targetIdx,
				Kind:      EdgePrev,
				TargetID:  s.PrevID,
			}
			g.Edges = append(g.Edges, edge)
			g.Nodes[i].OutEdges = append(g.Nodes[i].OutEdges, edge)
			if targetIdx >= 0 && targetIdx < len(g.Nodes) {
				g.Nodes[targetIdx].InEdges = append(g.Nodes[targetIdx].InEdges, edge)
			}
		}
	}

	return g
}

func (g DeckGraph) FilterByTag(tag string) DeckGraph {
	if tag == "" || strings.EqualFold(tag, "all") {
		return g
	}
	tagLower := strings.ToLower(strings.TrimSpace(tag))
	matches := make(map[int]bool)
	for i, n := range g.Nodes {
		for _, t := range n.Tags {
			if strings.ToLower(strings.TrimSpace(t)) == tagLower {
				matches[i] = true
				break
			}
		}
	}

	var filteredNodes []GraphNode
	for i, n := range g.Nodes {
		if matches[i] {
			filteredNodes = append(filteredNodes, n)
		}
	}

	var filteredEdges []GraphEdge
	for _, e := range g.Edges {
		if matches[e.FromIndex] && matches[e.ToIndex] {
			filteredEdges = append(filteredEdges, e)
		}
	}

	return DeckGraph{
		Nodes: filteredNodes,
		Edges: filteredEdges,
	}
}

func (g DeckGraph) ToMermaidWithTrack(track string) string {
	var b strings.Builder
	b.WriteString("graph LR\n")

	for _, n := range g.Nodes {
		cleanTitle := strings.ReplaceAll(n.Title, "\"", "'")
		if n.ID != "" {
			fmt.Fprintf(&b, "  node%d[\"[%02d] %s<br/>#%s\"]\n", n.Index, n.Index+1, cleanTitle, n.ID)
		} else {
			fmt.Fprintf(&b, "  node%d[\"[%02d] %s\"]\n", n.Index, n.Index+1, cleanTitle)
		}
	}

	for _, e := range g.Edges {
		if e.ToIndex < 0 || e.ToIndex >= len(g.Nodes) {
			continue
		}
		switch e.Kind {
		case EdgeLinear:
			fmt.Fprintf(&b, "  node%d --> node%d\n", e.FromIndex, e.ToIndex)
		case EdgeBranch:
			label := e.Label
			if e.Key != "" {
				label = fmt.Sprintf("[%s] %s", e.Key, label)
			}
			label = strings.ReplaceAll(label, "\"", "'")
			fmt.Fprintf(&b, "  node%d -- \"%s\" --> node%d\n", e.FromIndex, label, e.ToIndex)
		case EdgeNext:
			fmt.Fprintf(&b, "  node%d -. \"next\" .-> node%d\n", e.FromIndex, e.ToIndex)
		case EdgePrev:
			fmt.Fprintf(&b, "  node%d -. \"prev\" .-> node%d\n", e.FromIndex, e.ToIndex)
		}
	}

	if track != "" {
		trackLower := strings.ToLower(strings.TrimSpace(track))
		var trackNodes []string
		for _, n := range g.Nodes {
			for _, t := range n.Tags {
				if strings.ToLower(strings.TrimSpace(t)) == trackLower {
					trackNodes = append(trackNodes, fmt.Sprintf("node%d", n.Index))
					break
				}
			}
		}
		if len(trackNodes) > 0 {
			b.WriteString("\n  classDef trackNode fill:#22c55e,stroke:#16a34a,stroke-width:2px,color:#ffffff;\n")
			fmt.Fprintf(&b, "  class %s trackNode;\n", strings.Join(trackNodes, ","))
		}
	}

	return b.String()
}

func (g DeckGraph) ToMermaid() string {
	return g.ToMermaidWithTrack("")
}

func (g DeckGraph) ToMermaidWithRoute(routeName string, d Deck) string {
	var b strings.Builder
	b.WriteString("graph LR\n")

	for _, n := range g.Nodes {
		cleanTitle := strings.ReplaceAll(n.Title, "\"", "'")
		if n.ID != "" {
			fmt.Fprintf(&b, "  node%d[\"[%02d] %s<br/>#%s\"]\n", n.Index, n.Index+1, cleanTitle, n.ID)
		} else {
			fmt.Fprintf(&b, "  node%d[\"[%02d] %s\"]\n", n.Index, n.Index+1, cleanTitle)
		}
	}

	for _, e := range g.Edges {
		if e.ToIndex < 0 || e.ToIndex >= len(g.Nodes) {
			continue
		}
		switch e.Kind {
		case EdgeLinear:
			fmt.Fprintf(&b, "  node%d --> node%d\n", e.FromIndex, e.ToIndex)
		case EdgeBranch:
			label := e.Label
			if e.Key != "" {
				label = fmt.Sprintf("[%s] %s", e.Key, label)
			}
			label = strings.ReplaceAll(label, "\"", "'")
			fmt.Fprintf(&b, "  node%d -- \"%s\" --> node%d\n", e.FromIndex, label, e.ToIndex)
		case EdgeNext:
			fmt.Fprintf(&b, "  node%d -. \"next\" .-> node%d\n", e.FromIndex, e.ToIndex)
		case EdgePrev:
			fmt.Fprintf(&b, "  node%d -. \"prev\" .-> node%d\n", e.FromIndex, e.ToIndex)
		}
	}

	if routeName != "" {
		indices := d.RouteSlideIndices(routeName)
		var routeNodes []string
		for _, sIdx := range indices {
			if sIdx >= 0 && sIdx < len(g.Nodes) {
				routeNodes = append(routeNodes, fmt.Sprintf("node%d", sIdx))
			}
		}
		if len(routeNodes) > 0 {
			b.WriteString("\n  classDef routeNode fill:#f59e0b,stroke:#d97706,stroke-width:2px,color:#ffffff;\n")
			fmt.Fprintf(&b, "  class %s routeNode;\n", strings.Join(routeNodes, ","))
		}
	}

	return b.String()
}

func (g DeckGraph) ReachableNodes(startIndex int) []int {
	if startIndex < 0 || startIndex >= len(g.Nodes) {
		return nil
	}
	visited := make(map[int]bool)
	queue := []int{startIndex}
	visited[startIndex] = true

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, e := range g.Nodes[cur].OutEdges {
			if e.ToIndex >= 0 && e.ToIndex < len(g.Nodes) && !visited[e.ToIndex] {
				visited[e.ToIndex] = true
				queue = append(queue, e.ToIndex)
			}
		}
	}

	var result []int
	for i := range g.Nodes {
		if visited[i] {
			result = append(result, i)
		}
	}
	return result
}

func (g DeckGraph) HasCycles() bool {
	visited := make([]int, len(g.Nodes)) // 0: unvisited, 1: visiting, 2: visited

	var dfs func(u int) bool
	dfs = func(u int) bool {
		visited[u] = 1
		for _, e := range g.Nodes[u].OutEdges {
			v := e.ToIndex
			if v < 0 || v >= len(g.Nodes) {
				continue
			}
			if visited[v] == 1 {
				return true
			}
			if visited[v] == 0 && dfs(v) {
				return true
			}
		}
		visited[u] = 2
		return false
	}

	for i := range g.Nodes {
		if visited[i] == 0 {
			if dfs(i) {
				return true
			}
		}
	}
	return false
}

func FormatGraphCLIWithTrack(d Deck, theme Theme, track string) string {
	g := BuildGraph(d)
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Accent))

	idxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Secondary)).
		Bold(true)

	arrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Muted))

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Accent))

	targetStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Secondary))

	tagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Comment))

	trackBadgeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color(theme.Accent)).
		Padding(0, 1)

	header := "Termdeck Presentation Topology Map (DAG)"
	if track != "" {
		header += fmt.Sprintf(" [Track: %s]", track)
	}
	b.WriteString(titleStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(arrowStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	if len(g.Nodes) == 0 {
		b.WriteString(tagStyle.Render("Empty deck (0 slides)\n"))
		return b.String()
	}

	trackLower := strings.ToLower(strings.TrimSpace(track))

	for _, n := range g.Nodes {
		isTrackNode := false
		if trackLower != "" {
			for _, t := range n.Tags {
				if strings.ToLower(strings.TrimSpace(t)) == trackLower {
					isTrackNode = true
					break
				}
			}
		}

		b.WriteString(idxStyle.Render(fmt.Sprintf("[%02d]", n.Index+1)))
		b.WriteString(" ")
		if isTrackNode {
			b.WriteString(titleStyle.Render(n.Title) + " " + trackBadgeStyle.Render("★ "+track))
		} else {
			b.WriteString(titleStyle.Render(n.Title))
		}

		if n.ID != "" {
			b.WriteString(" " + tagStyle.Render("#"+n.ID))
		}
		if len(n.Tags) > 0 {
			b.WriteString(" " + tagStyle.Render("["+strings.Join(n.Tags, ",")+"]"))
		}
		b.WriteString("\n")

		if len(n.OutEdges) == 0 {
			if n.Index+1 == len(g.Nodes) {
				b.WriteString("     " + arrowStyle.Render("└──► (terminal slide)"))
				b.WriteString("\n")
			}
		} else {
			for j, e := range n.OutEdges {
				prefix := "├──►"
				if j == len(n.OutEdges)-1 {
					prefix = "└──►"
				}

				switch e.Kind {
				case EdgeLinear:
					targetTitle := "End"
					if e.ToIndex >= 0 && e.ToIndex < len(g.Nodes) {
						targetTitle = g.Nodes[e.ToIndex].Title
					}
					b.WriteString(fmt.Sprintf("     %s %s [%02d] %s\n",
						arrowStyle.Render(prefix),
						tagStyle.Render("(linear) ──►"),
						e.ToIndex+1,
						targetTitle,
					))
				case EdgeBranch:
					targetTitle := e.TargetID
					if e.ToIndex >= 0 && e.ToIndex < len(g.Nodes) {
						targetTitle = fmt.Sprintf("[%02d] %s", e.ToIndex+1, g.Nodes[e.ToIndex].Title)
					}
					b.WriteString(fmt.Sprintf("     %s %s %s %s %s\n",
						arrowStyle.Render(prefix),
						keyStyle.Render(fmt.Sprintf("[%s]", e.Key)),
						e.Label,
						arrowStyle.Render("──►"),
						targetStyle.Render(targetTitle),
					))
				case EdgeNext:
					targetTitle := e.TargetID
					if e.ToIndex >= 0 && e.ToIndex < len(g.Nodes) {
						targetTitle = fmt.Sprintf("[%02d] %s", e.ToIndex+1, g.Nodes[e.ToIndex].Title)
					}
					b.WriteString(fmt.Sprintf("     %s %s %s %s\n",
						arrowStyle.Render(prefix),
						tagStyle.Render("(next)"),
						arrowStyle.Render("──►"),
						targetStyle.Render(targetTitle),
					))
				case EdgePrev:
					targetTitle := e.TargetID
					if e.ToIndex >= 0 && e.ToIndex < len(g.Nodes) {
						targetTitle = fmt.Sprintf("[%02d] %s", e.ToIndex+1, g.Nodes[e.ToIndex].Title)
					}
					b.WriteString(fmt.Sprintf("     %s %s %s %s\n",
						arrowStyle.Render(prefix),
						tagStyle.Render("(prev)"),
						arrowStyle.Render("──►"),
						targetStyle.Render(targetTitle),
					))
				}
			}
		}
		b.WriteString("\n")
	}

	reachable := g.ReachableNodes(0)
	if len(reachable) < len(g.Nodes) {
		orphans := len(g.Nodes) - len(reachable)
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent)).Render(
			fmt.Sprintf("Note: %d unreferenced or detached slide(s) found in topology.\n", orphans),
		))
	}

	return b.String()
}

func FormatGraphCLI(d Deck, theme Theme) string {
	return FormatGraphCLIWithTrack(d, theme, "")
}

func FormatGraphCLIWithRoute(d Deck, theme Theme, routeName string) string {
	g := BuildGraph(d)
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Accent))

	idxStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Secondary)).
		Bold(true)

	arrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Muted))

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(theme.Accent))

	targetStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Secondary))

	tagStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(theme.Comment))

	routeBadgeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#f59e0b")).
		Padding(0, 1)

	header := "Termdeck Presentation Topology Map (DAG)"
	if routeName != "" {
		header += fmt.Sprintf(" [Route: %s]", routeName)
	}
	b.WriteString(titleStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(arrowStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	if len(g.Nodes) == 0 {
		b.WriteString(tagStyle.Render("Empty deck (0 slides)\n"))
		return b.String()
	}

	routeIndices := d.RouteSlideIndices(routeName)
	routeStepMap := make(map[int]int)
	for sIdx, nodeIdx := range routeIndices {
		routeStepMap[nodeIdx] = sIdx + 1
	}

	for _, n := range g.Nodes {
		step, isRouteNode := routeStepMap[n.Index]

		b.WriteString(idxStyle.Render(fmt.Sprintf("[%02d]", n.Index+1)))
		b.WriteString(" ")
		if isRouteNode {
			b.WriteString(titleStyle.Render(n.Title) + " " + routeBadgeStyle.Render(fmt.Sprintf("⚡ step %d", step)))
		} else {
			b.WriteString(titleStyle.Render(n.Title))
		}

		if n.ID != "" {
			b.WriteString(" " + tagStyle.Render("#"+n.ID))
		}
		if len(n.Tags) > 0 {
			b.WriteString(" " + tagStyle.Render("["+strings.Join(n.Tags, ",")+"]"))
		}
		b.WriteString("\n")

		if len(n.OutEdges) == 0 {
			b.WriteString(arrowStyle.Render("     └──► ") + tagStyle.Render("(terminal slide)") + "\n\n")
			continue
		}

		for eIdx, edge := range n.OutEdges {
			isLast := eIdx == len(n.OutEdges)-1
			prefix := "     ├──► "
			if isLast {
				prefix = "     └──► "
			}

			switch edge.Kind {
			case EdgeLinear:
				targetTitle := "End of deck"
				if edge.ToIndex < len(g.Nodes) {
					targetTitle = fmt.Sprintf("[%02d] %s", edge.ToIndex+1, g.Nodes[edge.ToIndex].Title)
				}
				b.WriteString(arrowStyle.Render(prefix) + tagStyle.Render("(linear)") + arrowStyle.Render(" ──► ") + targetStyle.Render(targetTitle) + "\n")

			case EdgeBranch:
				targetTitle := edge.TargetID
				if edge.ToIndex >= 0 && edge.ToIndex < len(g.Nodes) {
					targetTitle = fmt.Sprintf("[%02d] %s", edge.ToIndex+1, g.Nodes[edge.ToIndex].Title)
				}
				keyPart := ""
				if edge.Key != "" {
					keyPart = fmt.Sprintf("[%s] ", edge.Key)
				}
				b.WriteString(arrowStyle.Render(prefix) + keyStyle.Render(keyPart+edge.Label) + arrowStyle.Render(" ──► ") + targetStyle.Render(targetTitle) + "\n")

			case EdgeNext:
				targetTitle := edge.TargetID
				if edge.ToIndex >= 0 && edge.ToIndex < len(g.Nodes) {
					targetTitle = fmt.Sprintf("[%02d] %s", edge.ToIndex+1, g.Nodes[edge.ToIndex].Title)
				}
				b.WriteString(arrowStyle.Render(prefix) + tagStyle.Render("(next)") + arrowStyle.Render(" ──► ") + targetStyle.Render(targetTitle) + "\n")

			case EdgePrev:
				targetTitle := edge.TargetID
				if edge.ToIndex >= 0 && edge.ToIndex < len(g.Nodes) {
					targetTitle = fmt.Sprintf("[%02d] %s", edge.ToIndex+1, g.Nodes[edge.ToIndex].Title)
				}
				b.WriteString(arrowStyle.Render(prefix) + tagStyle.Render("(prev)") + arrowStyle.Render(" ──► ") + targetStyle.Render(targetTitle) + "\n")
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ShortestPath computes the sequence of slide indices representing the shortest path from 'from' to 'to' in the DAG using BFS.
// Returns nil if no path exists or indices are invalid.
func (g *DeckGraph) ShortestPath(from, to int) []int {
	if from < 0 || from >= len(g.Nodes) || to < 0 || to >= len(g.Nodes) {
		return nil
	}
	if from == to {
		return []int{from}
	}

	queue := []int{from}
	parent := make(map[int]int)
	visited := make(map[int]bool)
	visited[from] = true

	found := false
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr == to {
			found = true
			break
		}

		for _, edge := range g.Nodes[curr].OutEdges {
			target := edge.ToIndex
			if target >= 0 && target < len(g.Nodes) && !visited[target] {
				visited[target] = true
				parent[target] = curr
				queue = append(queue, target)
			}
		}
	}

	if !found {
		return nil
	}

	var path []int
	curr := to
	for curr != from {
		path = append([]int{curr}, path...)
		curr = parent[curr]
	}
	path = append([]int{from}, path...)
	return path
}

// PathStepDetail contains transition metadata for a single hop along a graph path.
type PathStepDetail struct {
	FromIndex int
	ToIndex   int
	EdgeKind  GraphEdgeKind
	Key       string
	Label     string
}

// ExplainPath extracts edge transition metadata for each hop along a path sequence.
func (g *DeckGraph) ExplainPath(path []int) []PathStepDetail {
	if len(path) < 2 {
		return nil
	}
	details := make([]PathStepDetail, len(path)-1)
	for i := 0; i < len(path)-1; i++ {
		u := path[i]
		v := path[i+1]
		detail := PathStepDetail{
			FromIndex: u,
			ToIndex:   v,
			EdgeKind:  EdgeLinear,
		}
		if u >= 0 && u < len(g.Nodes) {
			for _, edge := range g.Nodes[u].OutEdges {
				if edge.ToIndex == v {
					detail.EdgeKind = edge.Kind
					detail.Key = edge.Key
					detail.Label = edge.Label
					break
				}
			}
		}
		details[i] = detail
	}
	return details
}

// WaypointCandidate represents a candidate destination for graph pathfinding.
type WaypointCandidate struct {
	SlideIndex int
	Title      string
	ID         string
	Tags       []string
	Path       []int
	Reachable  bool
	EstMin     int
	HopCount   int
	Details    []PathStepDetail
}

// FindWaypointCandidates computes shortest paths and duration metrics to all matching candidate destinations.
func FindWaypointCandidates(originIdx int, d Deck, query string) []WaypointCandidate {
	if len(d.Slides) == 0 || originIdx < 0 || originIdx >= len(d.Slides) {
		return nil
	}

	g := BuildGraph(d)
	queryLower := strings.ToLower(strings.TrimSpace(query))

	var candidates []WaypointCandidate

	for i, s := range d.Slides {
		if i == originIdx {
			continue // skip self
		}

		titleLower := strings.ToLower(s.Title())
		idLower := strings.ToLower(s.ID)
		idxStr := fmt.Sprintf("%d", i+1)

		matchesQuery := true
		if queryLower != "" {
			matchesQuery = strings.Contains(titleLower, queryLower) ||
				strings.Contains(idLower, queryLower) ||
				idxStr == queryLower
			if !matchesQuery {
				for _, t := range s.Tags {
					if strings.Contains(strings.ToLower(t), queryLower) {
						matchesQuery = true
						break
					}
				}
			}
		}

		if !matchesQuery {
			continue
		}

		path := g.ShortestPath(originIdx, i)
		cand := WaypointCandidate{
			SlideIndex: i,
			Title:      s.Title(),
			ID:         s.ID,
			Tags:       s.Tags,
			Path:       path,
			Reachable:  len(path) > 1,
		}

		if cand.Reachable {
			cand.HopCount = len(path) - 1
			cand.Details = g.ExplainPath(path)

			words := 0
			for _, nodeIdx := range path {
				if nodeIdx >= 0 && nodeIdx < len(d.Slides) {
					slide := d.Slides[nodeIdx]
					for _, b := range slide.Blocks {
						if b.Kind == BlockCode {
							words += countWords(strings.Join(b.Lines, " "))
						} else {
							words += countWords(b.Text)
							for _, l := range b.Lines {
								words += countWords(l)
							}
						}
					}
				}
			}
			est := (words + 129) / 130
			if est == 0 && words > 0 {
				est = 1
			}
			cand.EstMin = est
		}

		candidates = append(candidates, cand)
	}

	// Sort candidates: reachable first, then by HopCount ascending, then by SlideIndex
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Reachable != candidates[j].Reachable {
			return candidates[i].Reachable
		}
		if candidates[i].Reachable && candidates[i].HopCount != candidates[j].HopCount {
			return candidates[i].HopCount < candidates[j].HopCount
		}
		return candidates[i].SlideIndex < candidates[j].SlideIndex
	})

	return candidates
}

// BreadcrumbTrail builds a compact breadcrumb representation of the path traversed so far.
func BreadcrumbTrail(history []int, current int, d Deck) string {
	if len(d.Slides) == 0 {
		return ""
	}
	var sequence []int
	sequence = append(sequence, history...)
	sequence = append(sequence, current)

	if len(sequence) == 0 {
		return ""
	}

	maxItems := 4
	var parts []string
	startIdx := 0
	if len(sequence) > maxItems {
		startIdx = len(sequence) - maxItems
		parts = append(parts, "...")
	}

	for _, idx := range sequence[startIdx:] {
		if idx >= 0 && idx < len(d.Slides) {
			s := d.Slides[idx]
			name := s.ID
			if name == "" {
				name = s.Slug()
			}
			if len(name) > 10 {
				name = name[:10]
			}
			parts = append(parts, fmt.Sprintf("[%02d:%s]", idx+1, name))
		}
	}

	return strings.Join(parts, " ──► ")
}

// BranchSummary formats the available branch options on a slide for fast status display.
func BranchSummary(slide Slide) string {
	branches := slide.Branches()
	if len(branches) == 0 {
		return ""
	}
	var items []string
	for _, b := range branches {
		key := b.Key
		if key == "" {
			key = "→"
		}
		label := b.Label
		if len(label) > 12 {
			label = label[:12]
		}
		items = append(items, fmt.Sprintf("[%s] %s", key, label))
	}
	return strings.Join(items, "  ")
}

// LintSeverity indicates whether a graph issue is fatal or advisory.
type LintSeverity int

const (
	SeverityError LintSeverity = iota
	SeverityWarning
)

// LintIssue represents a single DAG topology diagnostic issue.
type LintIssue struct {
	Severity LintSeverity
	SlideIdx int // 0-indexed slide, or -1 for deck-level issue
	Title    string
	Message  string
}

// LintGraph analyzes deck graph topology for broken links, unreachable slides, dead ends, and route errors.
func LintGraph(d Deck) []LintIssue {
	var issues []LintIssue
	if len(d.Slides) == 0 {
		issues = append(issues, LintIssue{
			Severity: SeverityError,
			SlideIdx: -1,
			Message:  "deck contains no slides",
		})
		return issues
	}

	// 1. Duplicate IDs
	seenIDs := make(map[string]int)
	for i, s := range d.Slides {
		if s.ID != "" {
			idLower := strings.ToLower(strings.TrimSpace(s.ID))
			if prevIdx, exists := seenIDs[idLower]; exists {
				issues = append(issues, LintIssue{
					Severity: SeverityError,
					SlideIdx: i,
					Title:    s.Title(),
					Message:  fmt.Sprintf("duplicate slide id %q (already defined on slide %d)", s.ID, prevIdx+1),
				})
			} else {
				seenIDs[idLower] = i
			}
		}
	}

	// 2. Broken branch targets, broken next/prev
	for i, s := range d.Slides {
		for _, b := range s.Branches() {
			if strings.TrimSpace(b.Target) == "" {
				issues = append(issues, LintIssue{
					Severity: SeverityError,
					SlideIdx: i,
					Title:    s.Title(),
					Message:  fmt.Sprintf("branch [%s] has empty target", b.Key),
				})
			} else if d.FindSlideByID(b.Target) == -1 {
				issues = append(issues, LintIssue{
					Severity: SeverityError,
					SlideIdx: i,
					Title:    s.Title(),
					Message:  fmt.Sprintf("branch [%s] points to nonexistent target %q", b.Key, b.Target),
				})
			}
		}

		if s.NextID != "" && !strings.EqualFold(s.NextID, "none") && !strings.EqualFold(s.NextID, "end") {
			if d.FindSlideByID(s.NextID) == -1 {
				issues = append(issues, LintIssue{
					Severity: SeverityError,
					SlideIdx: i,
					Title:    s.Title(),
					Message:  fmt.Sprintf("::next points to nonexistent target %q", s.NextID),
				})
			}
		}

		if s.PrevID != "" && !strings.EqualFold(s.PrevID, "none") {
			if d.FindSlideByID(s.PrevID) == -1 {
				issues = append(issues, LintIssue{
					Severity: SeverityError,
					SlideIdx: i,
					Title:    s.Title(),
					Message:  fmt.Sprintf("::prev points to nonexistent target %q", s.PrevID),
				})
			}
		}
	}

	// 3. Check routes
	for routeName, slugs := range d.Routes {
		if len(slugs) == 0 {
			issues = append(issues, LintIssue{
				Severity: SeverityWarning,
				SlideIdx: -1,
				Message:  fmt.Sprintf("route %q contains no target slides", routeName),
			})
			continue
		}
		for stepIdx, slug := range slugs {
			if d.FindSlideByID(slug) == -1 {
				issues = append(issues, LintIssue{
					Severity: SeverityError,
					SlideIdx: -1,
					Message:  fmt.Sprintf("route %q step %d points to nonexistent target %q", routeName, stepIdx+1, slug),
				})
			}
		}
	}

	// 4. Reachability from root slide 0
	g := BuildGraph(d)
	reachableSlice := g.ReachableNodes(0)
	reachable := make(map[int]bool, len(reachableSlice))
	for _, idx := range reachableSlice {
		reachable[idx] = true
	}
	for i, s := range d.Slides {
		if i > 0 && !reachable[i] {
			issues = append(issues, LintIssue{
				Severity: SeverityWarning,
				SlideIdx: i,
				Title:    s.Title(),
				Message:  "slide is unreachable from the opening slide",
			})
		}
	}

	// 5. Dead-ends: reachable slides before the last slide that have 0 outgoing edges
	for i := 0; i < len(d.Slides)-1; i++ {
		s := d.Slides[i]
		if reachable[i] && len(g.Nodes[i].OutEdges) == 0 {
			issues = append(issues, LintIssue{
				Severity: SeverityWarning,
				SlideIdx: i,
				Title:    s.Title(),
				Message:  "slide is a dead end with no outgoing edges before presentation conclusion",
			})
		}
	}

	return issues
}

// FormatLintCLI formats graph diagnostics into a clean compiler-style terminal report.
func FormatLintCLI(issues []LintIssue, theme Theme, filePath string) (string, int) {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Accent))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted))
	errBadge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#ef4444")).Padding(0, 1).Render("✖ ERROR")
	warnBadge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#f59e0b")).Padding(0, 1).Render("⚠ WARN ")
	slideBadge := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Secondary))
	successStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10b981"))

	b.WriteString(titleStyle.Render("Termdeck Presentation Graph Diagnostics (DAG Linter)") + "\n")
	if filePath != "" {
		b.WriteString(dimStyle.Render("Target: "+filePath) + "\n")
	}
	b.WriteString(dimStyle.Render("──────────────────────────────────────────────────") + "\n\n")

	errCount := 0
	warnCount := 0

	for _, issue := range issues {
		if issue.Severity == SeverityError {
			errCount++
			b.WriteString(errBadge)
		} else {
			warnCount++
			b.WriteString(warnBadge)
		}

		if issue.SlideIdx >= 0 {
			b.WriteString("  " + slideBadge.Render(fmt.Sprintf("[slide %02d]", issue.SlideIdx+1)))
			if issue.Title != "" {
				b.WriteString(dimStyle.Render(" (" + issue.Title + ")"))
			}
		} else {
			b.WriteString("  " + slideBadge.Render("[deck]"))
		}
		b.WriteString(" " + issue.Message + "\n")
	}

	b.WriteString("\n")
	if errCount == 0 && warnCount == 0 {
		b.WriteString(successStyle.Render("✔ Presentation DAG topology is sound! 0 errors, 0 warnings.") + "\n")
	} else if errCount == 0 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#f59e0b")).Render(fmt.Sprintf("✔ No fatal DAG errors found (%d warning%s).", warnCount, plural(warnCount))) + "\n")
	} else {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ef4444")).Render(fmt.Sprintf("✖ Failed DAG verification: %d error%s, %d warning%s.", errCount, plural(errCount), warnCount, plural(warnCount))) + "\n")
	}

	return b.String(), errCount
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// BranchForkOption represents an interactive decision choice from the current slide.
type BranchForkOption struct {
	Key             string   // e.g. "1", "2", "→"
	Label           string   // e.g. "Kubernetes Operator"
	TargetID        string   // e.g. "arch-k8s"
	TargetIndex     int      // 0-indexed slide in deck (-1 if unresolved)
	TargetTitle     string   // Title of target slide
	TargetTags      []string // Tags of target slide
	DownstreamCount int      // Total reachable slides from target
	EstimatedMin    int      // Estimated talk time in minutes for downstream
	CodeBlocksCount int      // Number of code blocks in downstream path
	IsTrackMatch    bool     // Does target match active audience track?
	IsRouteMatch    bool     // Is target the next step in active route?
}

// GetForkOptions returns all outgoing choices for the given slide index,
// calculating downstream reachability, word counts, talk time, and track matches.
func GetForkOptions(slideIdx int, d Deck, activeTrack string, activeRoute string) []BranchForkOption {
	if slideIdx < 0 || slideIdx >= len(d.Slides) {
		return nil
	}

	g := BuildGraph(d)
	curSlide := d.Slides[slideIdx]
	trackLower := strings.ToLower(strings.TrimSpace(activeTrack))

	// Find what the next slide in activeRoute would be
	routeNextIdx := -1
	if activeRoute != "" {
		routeIndices := d.RouteSlideIndices(activeRoute)
		for i, idx := range routeIndices {
			if idx == slideIdx && i+1 < len(routeIndices) {
				routeNextIdx = routeIndices[i+1]
				break
			}
		}
	}

	buildOption := func(key, label, targetID string, targetIdx int) BranchForkOption {
		opt := BranchForkOption{
			Key:         key,
			Label:       label,
			TargetID:    targetID,
			TargetIndex: targetIdx,
		}

		if targetIdx >= 0 && targetIdx < len(d.Slides) {
			targetSlide := d.Slides[targetIdx]
			opt.TargetTitle = targetSlide.Title()
			opt.TargetTags = targetSlide.Tags

			// Reachable downstream nodes from target
			downstream := g.ReachableNodes(targetIdx)
			opt.DownstreamCount = len(downstream)

			totalWords := 0
			codeBlocks := 0
			for _, nodeIdx := range downstream {
				if nodeIdx >= 0 && nodeIdx < len(d.Slides) {
					s := d.Slides[nodeIdx]
					for _, b := range s.Blocks {
						if b.Kind == BlockCode {
							codeBlocks++
							totalWords += countWords(strings.Join(b.Lines, " "))
						} else {
							totalWords += countWords(b.Text)
							for _, l := range b.Lines {
								totalWords += countWords(l)
							}
						}
					}
				}
			}
			opt.CodeBlocksCount = codeBlocks
			est := (totalWords + 129) / 130
			if est == 0 && totalWords > 0 {
				est = 1
			}
			opt.EstimatedMin = est

			if trackLower != "" {
				for _, t := range targetSlide.Tags {
					if strings.ToLower(strings.TrimSpace(t)) == trackLower {
						opt.IsTrackMatch = true
						break
					}
				}
			}

			if routeNextIdx >= 0 && targetIdx == routeNextIdx {
				opt.IsRouteMatch = true
			}
		}
		return opt
	}

	var options []BranchForkOption

	branches := curSlide.Branches()
	if len(branches) > 0 {
		for _, b := range branches {
			targetIdx := d.FindSlideByID(b.Target)
			key := b.Key
			if key == "" {
				key = fmt.Sprintf("%d", len(options)+1)
			}
			options = append(options, buildOption(key, b.Label, b.Target, targetIdx))
		}
	}

	if curSlide.NextID != "" {
		targetIdx := d.FindSlideByID(curSlide.NextID)
		options = append(options, buildOption("→", "Next Slide", curSlide.NextID, targetIdx))
	} else if len(branches) == 0 && slideIdx+1 < len(d.Slides) {
		options = append(options, buildOption("→", "Next Slide", "", slideIdx+1))
	}

	return options
}

// BranchCoverageItem represents exploration status of a specific branch in the presentation DAG.
type BranchCoverageItem struct {
	ForkSlideIdx   int
	ForkTitle      string
	ForkID         string
	BranchKey      string
	BranchLabel    string
	TargetIdx      int
	TargetTitle    string
	TargetID       string
	SubtreeTotal   int
	SubtreeVisited int
	SubtreeEstMin  int
	IsComplete     bool
	IsUnvisited    bool
	IsCurrent      bool
}

// GraphRadarStats contains holistic presentation DAG exploration telemetry and completion metrics.
type GraphRadarStats struct {
	TotalSlides       int
	VisitedSlides     int
	CoveragePct       float64
	TotalDeckEstMin   int
	VisitedEstMin     int
	UnvisitedEstMin   int
	ForkCount         int
	CompletedBranches int
	TotalBranches     int
	Items             []BranchCoverageItem
}

// CalculateRadarStats computes graph coverage and per-branch completion metrics.
func (g DeckGraph) CalculateRadarStats(d Deck, visitedMap map[int]bool, currentIdx int) GraphRadarStats {
	stats := GraphRadarStats{
		TotalSlides: len(d.Slides),
	}
	if len(d.Slides) == 0 {
		return stats
	}

	vMap := make(map[int]bool)
	for k, v := range visitedMap {
		if v && k >= 0 && k < len(d.Slides) {
			vMap[k] = true
		}
	}
	if currentIdx >= 0 && currentIdx < len(d.Slides) {
		vMap[currentIdx] = true
	}
	if len(vMap) == 0 {
		vMap[0] = true
	}

	stats.VisitedSlides = len(vMap)
	if stats.TotalSlides > 0 {
		stats.CoveragePct = (float64(stats.VisitedSlides) / float64(stats.TotalSlides)) * 100.0
	}

	totalWords := 0
	visitedWords := 0
	for i, s := range d.Slides {
		sWords := 0
		for _, b := range s.Blocks {
			switch b.Kind {
			case BlockHeading, BlockParagraph, BlockCallout, BlockList:
				sWords += len(strings.Fields(b.Text))
			case BlockCode, BlockTable:
				for _, line := range b.Lines {
					sWords += len(strings.Fields(line))
				}
			}
		}
		totalWords += sWords
		if vMap[i] {
			visitedWords += sWords
		}
	}

	if totalWords > 0 {
		stats.TotalDeckEstMin = (totalWords + 129) / 130
	} else if len(d.Slides) > 0 {
		stats.TotalDeckEstMin = 1
	}
	if visitedWords > 0 {
		stats.VisitedEstMin = (visitedWords + 129) / 130
	}
	stats.UnvisitedEstMin = stats.TotalDeckEstMin - stats.VisitedEstMin
	if stats.UnvisitedEstMin < 0 {
		stats.UnvisitedEstMin = 0
	}

	for forkIdx, forkSlide := range d.Slides {
		branches := forkSlide.Branches()
		if len(branches) <= 1 {
			continue
		}
		stats.ForkCount++

		type branchReach struct {
			branch    Branch
			targetIdx int
			reachSet  map[int]bool
		}
		var bReaches []branchReach
		for _, b := range branches {
			tIdx := d.FindSlideByID(b.Target)
			if tIdx >= 0 && tIdx < len(d.Slides) {
				subNodes := g.ReachableNodes(tIdx)
				rSet := make(map[int]bool)
				for _, n := range subNodes {
					rSet[n] = true
				}
				bReaches = append(bReaches, branchReach{branch: b, targetIdx: tIdx, reachSet: rSet})
			}
		}

		for bIdx, br := range bReaches {
			stats.TotalBranches++
			var exclusive []int
			for node := range br.reachSet {
				isShared := false
				for oIdx, other := range bReaches {
					if oIdx != bIdx && other.reachSet[node] {
						isShared = true
						break
					}
				}
				if !isShared {
					exclusive = append(exclusive, node)
				}
			}
			if len(exclusive) == 0 {
				exclusive = []int{br.targetIdx}
			}

			subVisited := 0
			unvisitedWords := 0
			for _, nIdx := range exclusive {
				if vMap[nIdx] {
					subVisited++
				} else if nIdx >= 0 && nIdx < len(d.Slides) {
					for _, b := range d.Slides[nIdx].Blocks {
						switch b.Kind {
						case BlockHeading, BlockParagraph:
							unvisitedWords += len(strings.Fields(b.Text))
						case BlockCode:
							for _, line := range b.Lines {
								unvisitedWords += len(strings.Fields(line))
							}
						}
					}
				}
			}

			subEstMin := 0
			if unvisitedWords > 0 {
				subEstMin = (unvisitedWords + 129) / 130
			} else if subVisited < len(exclusive) {
				subEstMin = 1
			}

			targetSlide := d.Slides[br.targetIdx]
			item := BranchCoverageItem{
				ForkSlideIdx:   forkIdx,
				ForkTitle:      forkSlide.Title(),
				ForkID:         forkSlide.ID,
				BranchKey:      br.branch.Key,
				BranchLabel:    br.branch.Label,
				TargetIdx:      br.targetIdx,
				TargetTitle:    targetSlide.Title(),
				TargetID:       targetSlide.ID,
				SubtreeTotal:   len(exclusive),
				SubtreeVisited: subVisited,
				SubtreeEstMin:  subEstMin,
				IsComplete:     subVisited == len(exclusive),
				IsUnvisited:    subVisited == 0,
				IsCurrent:      br.targetIdx == currentIdx,
			}
			if item.BranchKey == "" {
				item.BranchKey = fmt.Sprintf("%d", bIdx+1)
			}
			if item.IsComplete {
				stats.CompletedBranches++
			}
			stats.Items = append(stats.Items, item)
		}
	}

	return stats
}

// FormatRadarCLI formats presentation graph exploration telemetry as an ASCII CLI report.
func FormatRadarCLI(d Deck, visitedMap map[int]bool, currentIdx int) string {
	bg := BuildGraph(d)
	stats := bg.CalculateRadarStats(d, visitedMap, currentIdx)

	var sb strings.Builder
	sb.WriteString("Termdeck Presentation Graph Exploration Radar\n")
	title := ""
	if d.Meta != nil && d.Meta["title"] != "" {
		title = d.Meta["title"]
	}
	if title != "" {
		sb.WriteString(fmt.Sprintf("Title: %s\n", title))
	}
	sb.WriteString("──────────────────────────────────────────────────\n")

	barLen := 20
	filled := int(float64(barLen) * (stats.CoveragePct / 100.0))
	if filled > barLen {
		filled = barLen
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)

	inProgress := 0
	unvisited := 0
	for _, item := range stats.Items {
		if item.IsUnvisited {
			unvisited++
		} else if !item.IsComplete {
			inProgress++
		}
	}

	sb.WriteString(fmt.Sprintf("Graph Coverage:  %s  %.1f%% (%d/%d slides)\n", bar, stats.CoveragePct, stats.VisitedSlides, stats.TotalSlides))
	sb.WriteString(fmt.Sprintf("Speaking Time:   ~%dm visited / ~%dm total (~%dm unvisited)\n", stats.VisitedEstMin, stats.TotalDeckEstMin, stats.UnvisitedEstMin))
	sb.WriteString(fmt.Sprintf("Decision Forks:  %d fork%s · %d branches (%d complete, %d in-progress, %d unvisited)\n\n",
		stats.ForkCount, plural(stats.ForkCount), stats.TotalBranches, stats.CompletedBranches, inProgress, unvisited))

	if len(stats.Items) == 0 {
		sb.WriteString("No decision forks found in presentation (linear deck).\n")
	} else {
		sb.WriteString("Branch Completion Matrix:\n")
		curFork := -1
		for _, item := range stats.Items {
			if item.ForkSlideIdx != curFork {
				curFork = item.ForkSlideIdx
				forkSlug := ""
				if item.ForkID != "" {
					forkSlug = " #" + item.ForkID
				}
				sb.WriteString(fmt.Sprintf("  Fork [%02d]%s: %s\n", item.ForkSlideIdx+1, forkSlug, item.ForkTitle))
			}
			marker := "○"
			statusText := fmt.Sprintf("%d/%d slides · Unvisited · ~%dm", item.SubtreeVisited, item.SubtreeTotal, item.SubtreeEstMin)
			if item.IsComplete {
				marker = "✔"
				statusText = fmt.Sprintf("%d/%d slides · 100%% complete", item.SubtreeVisited, item.SubtreeTotal)
			} else if !item.IsUnvisited {
				pct := int((float64(item.SubtreeVisited) / float64(item.SubtreeTotal)) * 100.0)
				marker = "◐"
				statusText = fmt.Sprintf("%d/%d slides · %d%% in-progress · ~%dm left", item.SubtreeVisited, item.SubtreeTotal, pct, item.SubtreeEstMin)
			}
			targetSlug := ""
			if item.TargetID != "" {
				targetSlug = " ──► #" + item.TargetID
			}
			sb.WriteString(fmt.Sprintf("    %s [%s] %s%s (%s)\n", marker, item.BranchKey, item.BranchLabel, targetSlug, statusText))
		}
	}

	return sb.String()
}
