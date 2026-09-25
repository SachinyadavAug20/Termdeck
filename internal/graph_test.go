package internal

import (
	"strings"
	"testing"
)

func TestBuildGraphLinear(t *testing.T) {
	src := `---
title: Linear Deck
---

# Slide 1
Intro

---

# Slide 2
Body

---

# Slide 3
Conclusion`

	deck := ParseDeck(src)
	g := BuildGraph(deck)

	if len(g.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(g.Nodes))
	}
	if len(g.Edges) != 2 {
		t.Fatalf("expected 2 linear edges, got %d", len(g.Edges))
	}

	if g.Edges[0].Kind != EdgeLinear || g.Edges[0].FromIndex != 0 || g.Edges[0].ToIndex != 1 {
		t.Errorf("unexpected edge 0: %+v", g.Edges[0])
	}
	if g.Edges[1].Kind != EdgeLinear || g.Edges[1].FromIndex != 1 || g.Edges[1].ToIndex != 2 {
		t.Errorf("unexpected edge 1: %+v", g.Edges[1])
	}

	if g.HasCycles() {
		t.Errorf("expected no cycles in linear deck")
	}

	reachable := g.ReachableNodes(0)
	if len(reachable) != 3 {
		t.Errorf("expected all 3 nodes reachable, got %d", len(reachable))
	}
}

func TestBuildGraphBranching(t *testing.T) {
	src := `---
title: Branching Deck
---

# Overview
::branch [1] Deep Dive A -> dive-a
::branch [2] Deep Dive B -> dive-b

---

::id dive-a
# Deep Dive A
::next summary
Details on A.

---

::id dive-b
# Deep Dive B
::next summary
Details on B.

---

::id summary
# Summary
Wrap up.`

	deck := ParseDeck(src)
	g := BuildGraph(deck)

	if len(g.Nodes) != 4 {
		t.Fatalf("expected 4 nodes, got %d", len(g.Nodes))
	}

	// Slide 0 has 2 branch edges
	if len(g.Nodes[0].OutEdges) != 2 {
		t.Fatalf("expected 2 out-edges on node 0, got %d", len(g.Nodes[0].OutEdges))
	}
	if g.Nodes[0].OutEdges[0].Key != "1" || g.Nodes[0].OutEdges[0].ToIndex != 1 {
		t.Errorf("unexpected out-edge 0 on node 0: %+v", g.Nodes[0].OutEdges[0])
	}
	if g.Nodes[0].OutEdges[1].Key != "2" || g.Nodes[0].OutEdges[1].ToIndex != 2 {
		t.Errorf("unexpected out-edge 1 on node 0: %+v", g.Nodes[0].OutEdges[1])
	}

	// Slide 1 has next edge to Slide 3
	if len(g.Nodes[1].OutEdges) != 1 || g.Nodes[1].OutEdges[0].ToIndex != 3 {
		t.Errorf("unexpected out-edge on node 1: %+v", g.Nodes[1].OutEdges)
	}

	// Slide 2 has next edge to Slide 3
	if len(g.Nodes[2].OutEdges) != 1 || g.Nodes[2].OutEdges[0].ToIndex != 3 {
		t.Errorf("unexpected out-edge on node 2: %+v", g.Nodes[2].OutEdges)
	}

	// Slide 3 has 2 incoming next edges
	if len(g.Nodes[3].InEdges) != 2 {
		t.Errorf("expected 2 in-edges on summary slide, got %d", len(g.Nodes[3].InEdges))
	}

	// Test Mermaid generation
	mermaid := g.ToMermaid()
	if !strings.Contains(mermaid, "graph LR") {
		t.Errorf("expected 'graph LR' in mermaid, got %q", mermaid)
	}
	if !strings.Contains(mermaid, "node0 -- \"[1] Deep Dive A\" --> node1") {
		t.Errorf("expected branch link in mermaid, got:\n%s", mermaid)
	}
	if !strings.Contains(mermaid, "node1 -. \"next\" .-> node3") {
		t.Errorf("expected next link in mermaid, got:\n%s", mermaid)
	}

	// Test FormatGraphCLI
	theme := ResolveTheme("termdeck")
	cliOutput := FormatGraphCLI(deck, theme)
	if !strings.Contains(cliOutput, "Presentation Topology Map") {
		t.Errorf("expected title in CLI output, got:\n%s", cliOutput)
	}
	if !strings.Contains(cliOutput, "Deep Dive A") {
		t.Errorf("expected Deep Dive A in CLI output, got:\n%s", cliOutput)
	}
}

func TestGraphCycleAndOrphans(t *testing.T) {
	// Deck with cycle: 1 -> 2 -> 1
	srcCycle := `---
title: Cycle Deck
---

# Slide 1
::next slide-2

---

::id slide-2
# Slide 2
::next slide-1
`
	deckCycle := ParseDeck(srcCycle)
	gCycle := BuildGraph(deckCycle)
	if !gCycle.HasCycles() {
		t.Errorf("expected cycle detected in circular deck")
	}

	// Deck with orphan slide
	srcOrphan := `---
title: Orphan Deck
---

# Slide 1
::next slide-3

---

::id slide-2
# Detached Slide

---

::id slide-3
# Slide 3
End`

	deckOrphan := ParseDeck(srcOrphan)
	gOrphan := BuildGraph(deckOrphan)
	reachable := gOrphan.ReachableNodes(0)
	// Reachable should be 0 and 2 (slide 1 and slide 3)
	if len(reachable) != 2 {
		t.Errorf("expected 2 reachable nodes, got %d", len(reachable))
	}
	cliOutput := FormatGraphCLI(deckOrphan, ResolveTheme("nord"))
	if !strings.Contains(cliOutput, "unreferenced or detached slide") {
		t.Errorf("expected orphan warning in CLI output, got:\n%s", cliOutput)
	}

	// Empty deck
	emptyDeck := Deck{}
	cliEmpty := FormatGraphCLI(emptyDeck, ResolveTheme("dracula"))
	if !strings.Contains(cliEmpty, "Empty deck") {
		t.Errorf("expected empty deck notice, got:\n%s", cliEmpty)
	}
}

func TestReachableNodesOutOfBounds(t *testing.T) {
	g := DeckGraph{}
	if r := g.ReachableNodes(-1); r != nil {
		t.Errorf("expected nil for -1, got %+v", r)
	}
	if r := g.ReachableNodes(5); r != nil {
		t.Errorf("expected nil for out-of-bounds, got %+v", r)
	}
}

func TestShortestPathAndBreadcrumbs(t *testing.T) {
	src := `---
title: Path Test
---

::id intro
# Intro
::branch [1] Opt A -> deep-a
::branch [2] Opt B -> deep-b

---

::id deep-a
# Deep A
::next summary

---

::id deep-b
# Deep B
::next summary

---

::id summary
# Summary
`
	deck := ParseDeck(src)
	g := BuildGraph(deck)

	// Test ShortestPath from 0 (intro) to 3 (summary)
	pathA := g.ShortestPath(0, 3)
	if len(pathA) != 3 {
		t.Fatalf("expected path length 3, got %+v", pathA)
	}
	if pathA[0] != 0 || pathA[2] != 3 {
		t.Fatalf("expected path 0 -> 1|2 -> 3, got %+v", pathA)
	}

	// Invalid indices
	if p := g.ShortestPath(-1, 2); p != nil {
		t.Fatalf("expected nil for -1, got %+v", p)
	}
	if p := g.ShortestPath(0, 10); p != nil {
		t.Fatalf("expected nil for 10, got %+v", p)
	}
	if p := g.ShortestPath(2, 2); len(p) != 1 || p[0] != 2 {
		t.Fatalf("expected [2] for self-path, got %+v", p)
	}

	// BreadcrumbTrail tests
	emptyTrail := BreadcrumbTrail(nil, 0, Deck{})
	if emptyTrail != "" {
		t.Fatalf("expected empty trail, got %q", emptyTrail)
	}

	trail := BreadcrumbTrail([]int{0, 1}, 3, deck)
	if !strings.Contains(trail, "[01:intro]") || !strings.Contains(trail, "[04:summary]") {
		t.Fatalf("unexpected trail: %s", trail)
	}

	// Long history truncation
	longHistory := []int{0, 1, 2, 0, 1}
	longTrail := BreadcrumbTrail(longHistory, 3, deck)
	if !strings.HasPrefix(longTrail, "...") {
		t.Fatalf("expected '...' prefix in long trail, got: %s", longTrail)
	}

	// BranchSummary tests
	bs := BranchSummary(deck.Slides[0])
	if !strings.Contains(bs, "[1] Opt A") || !strings.Contains(bs, "[2] Opt B") {
		t.Fatalf("unexpected branch summary: %s", bs)
	}

	noBs := BranchSummary(deck.Slides[3])
	if noBs != "" {
		t.Fatalf("expected empty branch summary for slide without branches, got %q", noBs)
	}
}

func TestGraphFilterByTagAndTracks(t *testing.T) {
	src := `---
title: Track Topology Test
---

::id intro
::tags backend,intro
# Intro
::branch [1] Backend Dive -> backend-dive
::branch [2] Frontend Dive -> frontend-dive

---

::id backend-dive
::tags backend
::next summary
# Backend Deep Dive

---

::id frontend-dive
::tags frontend
::next summary
# Frontend Deep Dive

---

::id summary
::tags backend,summary
# Summary
`
	deck := ParseDeck(src)
	g := BuildGraph(deck)

	// FilterByTag
	filtered := g.FilterByTag("backend")
	if len(filtered.Nodes) != 3 {
		t.Errorf("expected 3 nodes in backend filtered graph, got %d", len(filtered.Nodes))
	}
	allFiltered := g.FilterByTag("")
	if len(allFiltered.Nodes) != 4 {
		t.Errorf("expected 4 nodes when filter is empty, got %d", len(allFiltered.Nodes))
	}
	noneFiltered := g.FilterByTag("devops")
	if len(noneFiltered.Nodes) != 0 {
		t.Errorf("expected 0 nodes for devops filter, got %d", len(noneFiltered.Nodes))
	}

	// ToMermaidWithTrack
	mmdWithTrack := g.ToMermaidWithTrack("backend")
	if !strings.Contains(mmdWithTrack, "classDef trackNode") {
		t.Errorf("expected classDef trackNode in mermaid output, got:\n%s", mmdWithTrack)
	}
	if !strings.Contains(mmdWithTrack, "trackNode;") {
		t.Errorf("expected node styling with trackNode in mermaid output, got:\n%s", mmdWithTrack)
	}
	mmdAll := g.ToMermaidWithTrack("")
	if strings.Contains(mmdAll, "classDef trackNode") {
		t.Errorf("expected no trackNode class when track is empty, got:\n%s", mmdAll)
	}

	// FormatGraphCLIWithTrack
	theme := ResolveTheme("tokyo-night")
	cliOut := FormatGraphCLIWithTrack(deck, theme, "backend")
	if !strings.Contains(cliOut, "[Track: backend]") {
		t.Errorf("expected [Track: backend] in CLI output, got:\n%s", cliOut)
	}
	if !strings.Contains(cliOut, "★ backend") {
		t.Errorf("expected '★ backend' badge in CLI output, got:\n%s", cliOut)
	}
}

func TestGraphWithRoute(t *testing.T) {
	src := `---
title: Graph Route Test
routes:
  talk: intro -> outro
---

::id intro
# Intro

---

::id middle
# Middle

---

::id outro
# Outro
`
	deck := ParseDeck(src)
	g := BuildGraph(deck)
	theme := ResolveTheme("tokyo-night")

	// 1. ToMermaidWithRoute
	mmdRoute := g.ToMermaidWithRoute("talk", deck)
	if !strings.Contains(mmdRoute, "classDef routeNode") {
		t.Fatalf("expected classDef routeNode in mermaid, got:\n%s", mmdRoute)
	}
	if !strings.Contains(mmdRoute, "class node0,node2 routeNode;") {
		t.Fatalf("expected node0 and node2 styled with routeNode, got:\n%s", mmdRoute)
	}

	mmdEmpty := g.ToMermaidWithRoute("", deck)
	if strings.Contains(mmdEmpty, "classDef routeNode") {
		t.Fatalf("expected no routeNode class for empty route, got:\n%s", mmdEmpty)
	}

	// 2. FormatGraphCLIWithRoute
	cliRoute := FormatGraphCLIWithRoute(deck, theme, "talk")
	if !strings.Contains(cliRoute, "[Route: talk") {
		t.Fatalf("expected [Route: talk header in CLI, got:\n%s", cliRoute)
	}
	if !strings.Contains(cliRoute, "⚡ step 1") || !strings.Contains(cliRoute, "⚡ step 2") {
		t.Fatalf("expected step badges in CLI output, got:\n%s", cliRoute)
	}

	cliNoRoute := FormatGraphCLIWithRoute(deck, theme, "")
	if strings.Contains(cliNoRoute, "⚡ step") {
		t.Fatalf("expected no step badges when route is empty, got:\n%s", cliNoRoute)
	}
}

func TestLintGraph(t *testing.T) {
	theme := ResolveTheme("tokyo-night")

	// 1. Empty deck
	emptyDeck := Deck{}
	emptyIssues := LintGraph(emptyDeck)
	if len(emptyIssues) != 1 || emptyIssues[0].Severity != SeverityError {
		t.Fatalf("expected 1 error for empty deck, got %+v", emptyIssues)
	}

	// 2. Sound deck
	soundSrc := `---
title: Sound Deck
routes:
  quick: intro -> end
---

::id intro
# Intro
::next end

---

::id end
# End
`
	soundDeck := ParseDeck(soundSrc)
	soundIssues := LintGraph(soundDeck)
	if len(soundIssues) != 0 {
		t.Fatalf("expected 0 issues for sound deck, got %+v", soundIssues)
	}
	soundOut, soundErrCount := FormatLintCLI(soundIssues, theme, "sound.deck.md")
	if soundErrCount != 0 || !strings.Contains(soundOut, "topology is sound") {
		t.Fatalf("unexpected FormatLintCLI output for sound deck: %s", soundOut)
	}

	// 3. Problematic deck with duplicate IDs, broken branch, broken next, broken prev, broken route, unreachable, dead end
	badSrc := `---
title: Broken Deck
routes:
  missing-step: intro -> ghost
  empty-route:
---

::id intro
# Intro
::branch [1] Broken Fork -> non-existent
::next bad-next
::prev bad-prev

---

::id intro
# Duplicate ID Intro

---

::id orphan
# Unreachable Slide
::next none

---

::id end
# Final Slide
`
	badDeck := ParseDeck(badSrc)
	// Add an empty branch target explicitly
	badDeck.Slides[0].Blocks = append(badDeck.Slides[0].Blocks, Block{
		Kind:         BlockBranch,
		BranchKey:    "2",
		BranchTarget: "",
	})
	// Make sure empty route has 0 slugs
	badDeck.Routes["empty-route"] = []string{}

	badIssues := LintGraph(badDeck)
	if len(badIssues) == 0 {
		t.Fatalf("expected multiple issues for broken deck, got 0")
	}

	hasDuplicateID := false
	hasBrokenBranch := false
	hasEmptyBranch := false
	hasBrokenNext := false
	hasBrokenPrev := false
	hasBrokenRoute := false
	hasEmptyRoute := false
	hasUnreachable := false

	for _, issue := range badIssues {
		if strings.Contains(issue.Message, "duplicate slide id") {
			hasDuplicateID = true
		}
		if strings.Contains(issue.Message, "nonexistent target \"non-existent\"") {
			hasBrokenBranch = true
		}
		if strings.Contains(issue.Message, "empty target") {
			hasEmptyBranch = true
		}
		if strings.Contains(issue.Message, "::next points to nonexistent") {
			hasBrokenNext = true
		}
		if strings.Contains(issue.Message, "::prev points to nonexistent") {
			hasBrokenPrev = true
		}
		if strings.Contains(issue.Message, "route \"missing-step\" step 2") {
			hasBrokenRoute = true
		}
		if strings.Contains(issue.Message, "route \"empty-route\" contains no target") {
			hasEmptyRoute = true
		}
		if strings.Contains(issue.Message, "unreachable from the opening slide") {
			hasUnreachable = true
		}
	}

	if !hasDuplicateID || !hasBrokenBranch || !hasEmptyBranch || !hasBrokenNext || !hasBrokenPrev || !hasBrokenRoute || !hasEmptyRoute || !hasUnreachable {
		t.Fatalf("missing expected diagnostics in badIssues: %+v", badIssues)
	}

	badOut, badErrCount := FormatLintCLI(badIssues, theme, "bad.deck.md")
	if badErrCount == 0 || !strings.Contains(badOut, "✖ ERROR") || !strings.Contains(badOut, "⚠ WARN") {
		t.Fatalf("expected formatted error output with badges: %s", badOut)
	}

	// 4. Test warnings only output formatting
	warnOnlyIssues := []LintIssue{
		{Severity: SeverityWarning, SlideIdx: 1, Title: "Warn", Message: "advisory warning"},
	}
	warnOut, warnErrCount := FormatLintCLI(warnOnlyIssues, theme, "warn.deck.md")
	if warnErrCount != 0 || !strings.Contains(warnOut, "No fatal DAG errors") {
		t.Fatalf("unexpected warning-only FormatLintCLI output: %s", warnOut)
	}
}

func TestGetForkOptions(t *testing.T) {
	// 1. Boundary cases
	if opts := GetForkOptions(-1, Deck{}, "", ""); opts != nil {
		t.Fatalf("expected nil for negative index")
	}
	if opts := GetForkOptions(0, Deck{}, "", ""); opts != nil {
		t.Fatalf("expected nil for empty deck")
	}

	src := `---
title: Branching Deck
routes:
  demo: hub -> target-a
---

::id hub
# Hub Slide
::branch Option A -> target-a
::branch [2] Option B -> target-b

---

::id target-a
::tags devops
# Target A
Here is some architecture explanation text with several words.
` + "```go\nfunc main() {}\n```" + `

---

::id target-b
::tags backend
# Target B
Another branch with code.
` + "```python\nprint(1)\n```" + `

---

::id terminal
# Terminal Slide
`
	d := ParseDeck(src)

	// Hub slide (slide 0) with 2 branches
	hubOpts := GetForkOptions(0, d, "devops", "demo")
	if len(hubOpts) != 2 {
		t.Fatalf("expected 2 fork options for hub slide, got %d", len(hubOpts))
	}

	optA := hubOpts[0]
	if optA.Key != "1" || optA.Label != "Option A" || optA.TargetID != "target-a" || optA.TargetIndex != 1 {
		t.Fatalf("unexpected optA: %+v", optA)
	}
	if optA.TargetTitle != "Target A" {
		t.Fatalf("expected TargetTitle 'Target A', got %q", optA.TargetTitle)
	}
	if optA.CodeBlocksCount != 2 {
		t.Fatalf("expected 2 code blocks in downstream for optA, got %d", optA.CodeBlocksCount)
	}
	if !optA.IsTrackMatch {
		t.Fatalf("expected IsTrackMatch true for devops track")
	}
	if !optA.IsRouteMatch {
		t.Fatalf("expected IsRouteMatch true for demo route")
	}
	if optA.DownstreamCount <= 0 || optA.EstimatedMin <= 0 {
		t.Fatalf("expected positive DownstreamCount and EstimatedMin: %+v", optA)
	}

	optB := hubOpts[1]
	if optB.Key != "2" || optB.Label != "Option B" || optB.TargetIndex != 2 {
		t.Fatalf("unexpected optB: %+v", optB)
	}
	if optB.CodeBlocksCount != 1 {
		t.Fatalf("expected 1 code block in downstream for optB, got %d", optB.CodeBlocksCount)
	}
	if optB.IsTrackMatch {
		t.Fatalf("expected IsTrackMatch false for backend target on devops track")
	}
	if optB.IsRouteMatch {
		t.Fatalf("expected IsRouteMatch false for target-b on demo route")
	}

	// Linear slide with sequential fallthrough (slide 1)
	linearOpts := GetForkOptions(1, d, "", "")
	if len(linearOpts) != 1 || linearOpts[0].Key != "→" || linearOpts[0].TargetIndex != 2 {
		t.Fatalf("expected linear next option for slide 1: %+v", linearOpts)
	}

	// Terminal slide (slide 3) with no branches and no next
	termOpts := GetForkOptions(3, d, "", "")
	if len(termOpts) != 0 {
		t.Fatalf("expected 0 fork options on terminal slide, got %d", len(termOpts))
	}
}

func TestWaypointPathfinder(t *testing.T) {
	// 1. ExplainPath boundary cases
	g := DeckGraph{}
	if g.ExplainPath(nil) != nil || g.ExplainPath([]int{0}) != nil {
		t.Fatalf("expected nil for paths < 2 nodes")
	}

	src := `---
title: Pathfinder Deck
---

::id s1
# Slide 1
::branch [1] To Step 2 -> s2

---

::id s2
::tags arch,core
# Slide 2
` + "```go\nfunc step2() {}\n```" + `
::next s3

---

::id s3
# Slide 3
The End
::next s1

---

::id isolated
# Isolated Slide
`
	d := ParseDeck(src)
	bg := BuildGraph(d)

	// Explain valid path [0, 1, 2]
	path := []int{0, 1, 2}
	details := bg.ExplainPath(path)
	if len(details) != 2 {
		t.Fatalf("expected 2 details, got %d", len(details))
	}
	if details[0].FromIndex != 0 || details[0].ToIndex != 1 || details[0].Key != "1" {
		t.Fatalf("unexpected detail[0]: %+v", details[0])
	}
	if details[1].FromIndex != 1 || details[1].ToIndex != 2 || details[1].EdgeKind != EdgeNext {
		t.Fatalf("unexpected detail[1]: %+v", details[1])
	}

	// 2. FindWaypointCandidates boundaries
	if cands := FindWaypointCandidates(-1, d, ""); cands != nil {
		t.Fatalf("expected nil for negative origin")
	}
	if cands := FindWaypointCandidates(0, Deck{}, ""); cands != nil {
		t.Fatalf("expected nil for empty deck")
	}

	// 3. Find candidates from slide 0
	allCands := FindWaypointCandidates(0, d, "")
	if len(allCands) != 3 {
		t.Fatalf("expected 3 candidates (excluding slide 0), got %d", len(allCands))
	}

	// Reachable ones must be sorted first
	if !allCands[0].Reachable || allCands[0].SlideIndex != 1 {
		t.Fatalf("expected slide 1 (1 hop) to be first candidate, got %+v", allCands[0])
	}
	if allCands[0].HopCount != 1 || allCands[0].EstMin <= 0 {
		t.Fatalf("unexpected candidate 0: %+v", allCands[0])
	}

	if !allCands[1].Reachable || allCands[1].SlideIndex != 2 {
		t.Fatalf("expected slide 2 (2 hops) to be second candidate, got %+v", allCands[1])
	}
	if allCands[1].HopCount != 2 {
		t.Fatalf("expected 2 hops for slide 2, got %d", allCands[1].HopCount)
	}

	// Isolated slide is unreachable
	if allCands[2].Reachable || allCands[2].SlideIndex != 3 {
		t.Fatalf("expected slide 3 to be unreachable, got %+v", allCands[2])
	}

	// 4. Query filtering (by title, tag, ID, slide number)
	tagCands := FindWaypointCandidates(0, d, "core")
	if len(tagCands) != 1 || tagCands[0].SlideIndex != 1 {
		t.Fatalf("expected 1 candidate matching tag 'core', got %+v", tagCands)
	}

	numCands := FindWaypointCandidates(0, d, "3")
	if len(numCands) != 1 || numCands[0].SlideIndex != 2 {
		t.Fatalf("expected 1 candidate matching slide '3', got %+v", numCands)
	}

	noMatchCands := FindWaypointCandidates(0, d, "nonexistent-query")
	if len(noMatchCands) != 0 {
		t.Fatalf("expected 0 candidates for nonexistent-query, got %d", len(noMatchCands))
	}
}
