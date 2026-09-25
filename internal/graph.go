package internal

import (
	"fmt"
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
