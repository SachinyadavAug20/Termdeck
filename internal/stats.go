package internal

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DeckStats contains calculated metrics and technical density statistics for a deck.
type DeckStats struct {
	TotalSlides        int
	TotalBlocks        int
	VisibleBlocks      int
	TotalWords         int
	NotesWords         int
	EstDurationMin     int
	EstDurationSec     int
	CodeBlocks         int
	CodeLines          int
	TableBlocks        int
	CalloutBlocks      int
	ImageBlocks        int
	DividerBlocks      int
	TaskTotal          int
	TaskCompleted      int
	TaskPercent        float64
	CurrentSlideIdx    int
	CurrentSlideTitle  string
	CurrentSlideWords  int
	CurrentSlideBlocks int
}

// countWords returns the number of whitespace-separated tokens in a string.
func countWords(s string) int {
	fields := strings.Fields(s)
	return len(fields)
}

// CalculateStats computes comprehensive presentation metrics for a deck and active slide.
func CalculateStats(d *Deck, activeSlideIdx int) DeckStats {
	stats := DeckStats{
		TotalSlides:     len(d.Slides),
		CurrentSlideIdx: activeSlideIdx,
	}

	if stats.TotalSlides == 0 {
		return stats
	}

	if stats.CurrentSlideIdx < 0 {
		stats.CurrentSlideIdx = 0
	}
	if stats.CurrentSlideIdx >= stats.TotalSlides {
		stats.CurrentSlideIdx = stats.TotalSlides - 1
	}

	for sIdx, slide := range d.Slides {
		slideWords := 0
		slideBlocks := len(slide.VisibleBlockIndices())

		for _, blk := range slide.Blocks {
			stats.TotalBlocks++

			switch blk.Kind {
			case BlockHeading:
				w := countWords(blk.Text)
				stats.TotalWords += w
				slideWords += w

			case BlockParagraph:
				w := countWords(blk.Text)
				stats.TotalWords += w
				slideWords += w

			case BlockCode:
				stats.CodeBlocks++
				stats.CodeLines += len(blk.Lines)
				w := 0
				for _, l := range blk.Lines {
					w += countWords(l)
				}
				stats.TotalWords += w
				slideWords += w

			case BlockTable:
				stats.TableBlocks++
				w := 0
				for _, l := range blk.Lines {
					w += countWords(l)
				}
				stats.TotalWords += w
				slideWords += w

			case BlockCallout:
				stats.CalloutBlocks++
				w := countWords(blk.Text)
				stats.TotalWords += w
				slideWords += w

			case BlockImage:
				stats.ImageBlocks++

			case BlockDivider:
				stats.DividerBlocks++

			case BlockList:
				w := countWords(blk.Text)
				stats.TotalWords += w
				slideWords += w

				trimmed := strings.TrimSpace(blk.Text)
				prefix := ""
				if strings.HasPrefix(trimmed, "- ") {
					prefix = "- "
				} else if strings.HasPrefix(trimmed, "* ") {
					prefix = "* "
				}
				if prefix != "" {
					content := trimmed[len(prefix):]
					if strings.HasPrefix(content, "[x] ") || strings.HasPrefix(content, "[X] ") {
						stats.TaskTotal++
						stats.TaskCompleted++
					} else if strings.HasPrefix(content, "[ ] ") {
						stats.TaskTotal++
					}
				}

			case BlockDirective:
				if strings.HasPrefix(blk.Directive, "::notes") {
					nw := 0
					if len(blk.Lines) > 0 {
						for _, l := range blk.Lines {
							nw += countWords(l)
						}
					} else {
						nw = countWords(blk.Text)
					}
					stats.NotesWords += nw
				}
			}
		}

		stats.VisibleBlocks += slideBlocks

		if sIdx == stats.CurrentSlideIdx {
			stats.CurrentSlideTitle = slide.Title()
			stats.CurrentSlideBlocks = slideBlocks
			stats.CurrentSlideWords = slideWords
		}
	}

	if stats.TaskTotal > 0 {
		stats.TaskPercent = (float64(stats.TaskCompleted) / float64(stats.TaskTotal)) * 100.0
	}

	// Speaking rate: 130 words per minute
	totalSpeakingWords := stats.TotalWords + stats.NotesWords
	if totalSpeakingWords > 0 {
		totalSeconds := int(math.Round(float64(totalSpeakingWords) / 130.0 * 60.0))
		if totalSeconds < 30 && totalSpeakingWords > 0 {
			totalSeconds = 30
		}
		stats.EstDurationMin = totalSeconds / 60
		stats.EstDurationSec = totalSeconds % 60
	}

	return stats
}

// RenderProgressBar generates a styled ASCII/Unicode progress bar.
func RenderProgressBar(percent float64, barWidth int, filledColor, emptyColor string) string {
	if barWidth <= 0 {
		barWidth = 10
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filledChars := int(math.Round((percent / 100.0) * float64(barWidth)))
	if filledChars > barWidth {
		filledChars = barWidth
	}
	emptyChars := barWidth - filledChars

	fStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(filledColor))
	eStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(emptyColor))

	var sb strings.Builder
	sb.WriteString("[")
	if filledChars > 0 {
		sb.WriteString(fStyle.Render(strings.Repeat("█", filledChars)))
	}
	if emptyChars > 0 {
		sb.WriteString(eStyle.Render(strings.Repeat("░", emptyChars)))
	}
	sb.WriteString("]")
	return sb.String()
}

// FormatStatsCLI formats DeckStats into a clean terminal report.
func FormatStatsCLI(stats DeckStats, theme Theme) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Laser))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(theme.Accent))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Muted))
	valueStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	barStyle := RenderProgressBar(stats.TaskPercent, 16, theme.Success, theme.DimTrack)

	sb.WriteString(titleStyle.Render("📊 Termdeck — Presentation Statistics & Deck Metrics") + "\n\n")

	sb.WriteString(headerStyle.Render("  DECK OVERVIEW") + "\n")
	sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Total Slides:"), valueStyle.Render(fmt.Sprintf("%d", stats.TotalSlides))))
	sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Visible Blocks:"), valueStyle.Render(fmt.Sprintf("%d (avg %.1f / slide)", stats.VisibleBlocks, float64(stats.VisibleBlocks)/math.Max(1, float64(stats.TotalSlides))))))
	sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Total Words:"), valueStyle.Render(fmt.Sprintf("%d words", stats.TotalWords))))
	if stats.NotesWords > 0 {
		sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Speaker Notes Words:"), valueStyle.Render(fmt.Sprintf("%d words", stats.NotesWords))))
	}
	durStr := fmt.Sprintf("~%d min %02d sec (at 130 WPM)", stats.EstDurationMin, stats.EstDurationSec)
	sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Est. Talk Duration:"), valueStyle.Render(durStr)))

	sb.WriteString("\n" + headerStyle.Render("  TECHNICAL DENSITY") + "\n")
	sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Code Blocks:"), valueStyle.Render(fmt.Sprintf("%d blocks (%d lines)", stats.CodeBlocks, stats.CodeLines))))
	sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Tables & Callouts:"), valueStyle.Render(fmt.Sprintf("%d tables · %d callout boxes", stats.TableBlocks, stats.CalloutBlocks))))
	sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Images & Dividers:"), valueStyle.Render(fmt.Sprintf("%d images · %d dividers", stats.ImageBlocks, stats.DividerBlocks))))

	sb.WriteString("\n" + headerStyle.Render("  SPRINT TASK COMPLETION") + "\n")
	if stats.TaskTotal > 0 {
		taskStr := fmt.Sprintf("%s  %d/%d completed (%.1f%%)", barStyle, stats.TaskCompleted, stats.TaskTotal, stats.TaskPercent)
		sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Checklist Velocity:"), valueStyle.Render(taskStr)))
	} else {
		sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Checklist Velocity:"), labelStyle.Render("no task checklists found")))
	}

	if stats.TotalSlides > 0 {
		sb.WriteString("\n" + headerStyle.Render("  FOCUSED SLIDE") + "\n")
		title := stats.CurrentSlideTitle
		if len(title) > 36 {
			title = title[:36] + "..."
		}
		sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Active Slide:"), valueStyle.Render(fmt.Sprintf("#%d: %s", stats.CurrentSlideIdx+1, title))))
		sb.WriteString(fmt.Sprintf("  %-24s %s\n", labelStyle.Render("Slide Metrics:"), valueStyle.Render(fmt.Sprintf("%d blocks · %d words", stats.CurrentSlideBlocks, stats.CurrentSlideWords))))
	}

	return sb.String()
}
