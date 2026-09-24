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

func (g DeckGraph) ToMermaid() string {
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

func FormatGraphCLI(d Deck, theme Theme) string {
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

	b.WriteString(titleStyle.Render("Termdeck Presentation Topology Map (DAG)"))
	b.WriteString("\n")
	b.WriteString(arrowStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n\n")

	if len(g.Nodes) == 0 {
		b.WriteString(tagStyle.Render("Empty deck (0 slides)\n"))
		return b.String()
	}

	for _, n := range g.Nodes {
		b.WriteString(idxStyle.Render(fmt.Sprintf("[%02d]", n.Index+1)))
		b.WriteString(" ")
		b.WriteString(titleStyle.Render(n.Title))

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
