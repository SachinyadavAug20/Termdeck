package internal

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ExecResult captures the outcome of executing a code block.
type ExecResult struct {
	Command   string
	Language  string
	ExitCode  int
	Duration  time.Duration
	Stdout    string
	Stderr    string
	Error     string
	SlideNum  int
	BlockNum  int
	Truncated bool
}

// ExecFinishedMsg is delivered to Bubble Tea when asynchronous code execution completes.
type ExecFinishedMsg struct {
	Result ExecResult
}

const (
	MaxOutputBytes = 16 * 1024 // 16 KB max stdout/stderr capture
	MaxOutputLines = 300       // 300 lines max
)

// IsExecutableLanguage returns true for languages supported by the live runner.
func IsExecutableLanguage(lang string) bool {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "bash", "sh", "zsh", "shell", "python", "py", "python3", "node", "js", "javascript", "ruby", "rb", "go":
		return true
	default:
		return false
	}
}

// ExecuteBlock executes the code within a Block using host interpreters.
// Supported languages:
//   - bash, sh, zsh, shell (or empty): /bin/sh -c
//   - python, py, python3: python3 -c
//   - node, js, javascript: node -e
//   - ruby, rb: ruby -e
//   - go: go run (wraps snippet in package main if needed)
func ExecuteBlock(blk Block, timeout time.Duration) ExecResult {
	start := time.Now()
	res := ExecResult{
		Language: strings.ToLower(strings.TrimSpace(blk.Lang)),
	}

	if blk.Kind != BlockCode {
		res.Error = "block is not a code block"
		res.ExitCode = 1
		res.Duration = time.Since(start)
		return res
	}

	lang := res.Language
	if !IsExecutableLanguage(lang) && lang != "" {
		res.Error = fmt.Sprintf("language %q is not executable (supported: bash, sh, python, go, node, ruby)", lang)
		res.ExitCode = 1
		res.Duration = time.Since(start)
		return res
	}

	code := blk.Text
	if len(blk.Lines) > 0 {
		code = strings.Join(blk.Lines, "\n")
	}
	code = dedent(strings.TrimRight(code, "\n\r"))
	if strings.TrimSpace(code) == "" {
		res.Error = "code block is empty"
		res.ExitCode = 0
		res.Duration = time.Since(start)
		return res
	}

	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd
	var tempFileToRemove string

	switch lang {
	case "python", "py", "python3":
		pyBin := "python3"
		if _, err := exec.LookPath("python3"); err != nil {
			if _, err2 := exec.LookPath("python"); err2 == nil {
				pyBin = "python"
			}
		}
		cmd = exec.CommandContext(ctx, pyBin, "-c", code)
		res.Command = fmt.Sprintf("%s -c ...", pyBin)

	case "node", "js", "javascript":
		cmd = exec.CommandContext(ctx, "node", "-e", code)
		res.Command = "node -e ..."

	case "ruby", "rb":
		cmd = exec.CommandContext(ctx, "ruby", "-e", code)
		res.Command = "ruby -e ..."

	case "go":
		goCode := code
		if !strings.Contains(code, "package ") {
			goCode = fmt.Sprintf("package main\n\nimport \"fmt\"\n\nfunc main() {\n\t%s\n}\n", strings.ReplaceAll(code, "\n", "\n\t"))
		}
		tmpDir := os.TempDir()
		tmpFile := filepath.Join(tmpDir, fmt.Sprintf("termdeck_run_%d.go", time.Now().UnixNano()))
		if err := os.WriteFile(tmpFile, []byte(goCode), 0600); err != nil {
			res.Error = fmt.Sprintf("failed to create temp file: %v", err)
			res.ExitCode = 1
			res.Duration = time.Since(start)
			return res
		}
		tempFileToRemove = tmpFile
		cmd = exec.CommandContext(ctx, "go", "run", tmpFile)
		res.Command = "go run ..."

	default: // bash, sh, zsh, shell or generic
		shBin := "/bin/sh"
		if _, err := exec.LookPath("bash"); err == nil {
			shBin = "bash"
		}
		cmd = exec.CommandContext(ctx, shBin, "-c", code)
		res.Command = fmt.Sprintf("%s -c ...", shBin)
	}

	if tempFileToRemove != "" {
		defer os.Remove(tempFileToRemove)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()
	res.Duration = time.Since(start)

	rawStdout := stdoutBuf.String()
	rawStderr := stderrBuf.String()

	res.Stdout, res.Truncated = truncateOutput(rawStdout, MaxOutputBytes, MaxOutputLines)
	stderrTrunc, truncErr := truncateOutput(rawStderr, MaxOutputBytes, MaxOutputLines)
	res.Stderr = stderrTrunc
	if truncErr {
		res.Truncated = true
	}

	if runErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			res.Error = fmt.Sprintf("execution timed out after %v", timeout)
			res.ExitCode = 124
		} else if exitErr, ok := runErr.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
			if res.Error == "" && res.Stderr == "" {
				res.Error = exitErr.Error()
			}
		} else {
			res.Error = runErr.Error()
			res.ExitCode = 1
		}
	} else {
		res.ExitCode = 0
	}

	return res
}

func dedent(code string) string {
	lines := strings.Split(code, "\n")
	minIndent := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		indent := 0
		for _, ch := range l {
			if ch == ' ' {
				indent++
			} else {
				break
			}
		}
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return code
	}
	var res []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			res = append(res, "")
			continue
		}
		if len(l) >= minIndent {
			res = append(res, l[minIndent:])
		} else {
			res = append(res, strings.TrimLeft(l, " "))
		}
	}
	return strings.Join(res, "\n")
}

func truncateOutput(s string, maxBytes, maxLines int) (string, bool) {
	truncated := false
	if len(s) > maxBytes {
		s = s[:maxBytes]
		truncated = true
	}
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		s = strings.Join(lines[:maxLines], "\n")
		truncated = true
	}
	if truncated {
		s += "\n... (output truncated)"
	}
	return s, truncated
}

// ExecuteCodeCmd creates a Bubble Tea Cmd that runs ExecuteBlock asynchronously.
func ExecuteCodeCmd(blk Block, timeout time.Duration, slideNum, blockNum int) tea.Cmd {
	return func() tea.Msg {
		res := ExecuteBlock(blk, timeout)
		res.SlideNum = slideNum
		res.BlockNum = blockNum
		return ExecFinishedMsg{Result: res}
	}
}

// TestAllDeckCode iterates through all slides in a deck and executes all code blocks.
// Returns count of passed, failed, and detailed execution results.
func TestAllDeckCode(d Deck, timeout time.Duration) (int, int, []ExecResult) {
	var results []ExecResult
	passed := 0
	failed := 0

	for sIdx, slide := range d.Slides {
		for bIdx, blk := range slide.Blocks {
			if blk.Kind == BlockCode {
				if blk.NoEval || !IsExecutableLanguage(blk.Lang) {
					continue
				}
				code := blk.Text
				if len(blk.Lines) > 0 {
					code = strings.Join(blk.Lines, "\n")
				}
				if strings.TrimSpace(code) == "" {
					continue
				}
				res := ExecuteBlock(blk, timeout)
				res.SlideNum = sIdx + 1
				res.BlockNum = bIdx + 1
				results = append(results, res)
				if res.ExitCode == 0 && res.Error == "" {
					passed++
				} else {
					failed++
				}
			} else if blk.Kind == BlockColumns {
				for _, col := range blk.Columns {
					for _, inner := range col {
						if inner.Kind == BlockCode {
							if inner.NoEval || !IsExecutableLanguage(inner.Lang) {
								continue
							}
							code := inner.Text
							if len(inner.Lines) > 0 {
								code = strings.Join(inner.Lines, "\n")
							}
							if strings.TrimSpace(code) == "" {
								continue
							}
							res := ExecuteBlock(inner, timeout)
							res.SlideNum = sIdx + 1
							res.BlockNum = bIdx + 1
							results = append(results, res)
							if res.ExitCode == 0 && res.Error == "" {
								passed++
							} else {
								failed++
							}
						}
					}
				}
			}
		}
	}

	return passed, failed, results
}

// FormatTestCodeCLI formats the output of --test-code for terminal display.
func FormatTestCodeCLI(filePath string, passed, failed int, results []ExecResult, totalDur time.Duration) string {
	var sb strings.Builder
	total := passed + failed

	sb.WriteString(boldStyle.Render(fmt.Sprintf("⚡ Testing %d code block(s) in %s:\n\n", total, filepath.Base(filePath))))

	for _, res := range results {
		lang := res.Language
		if lang == "" {
			lang = "sh"
		}
		durStr := fmt.Sprintf("%v", res.Duration.Round(time.Millisecond))
		if res.ExitCode == 0 && res.Error == "" {
			checkMark := lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true).Render("✔")
			sb.WriteString(fmt.Sprintf("  %s Slide %02d [blk %d, %s]: exit 0 · %s\n", checkMark, res.SlideNum, res.BlockNum, lang, durStr))
		} else {
			crossMark := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true).Render("✖")
			errMsg := res.Error
			if errMsg == "" && res.Stderr != "" {
				firstLine := strings.Split(strings.TrimSpace(res.Stderr), "\n")[0]
				errMsg = firstLine
			}
			sb.WriteString(fmt.Sprintf("  %s Slide %02d [blk %d, %s]: exit %d · %s (%s)\n", crossMark, res.SlideNum, res.BlockNum, lang, res.ExitCode, durStr, errMsg))
		}
	}

	sb.WriteString("\n")
	if failed == 0 {
		summary := lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true).
			Render(fmt.Sprintf("All %d code blocks passed (%v total).", total, totalDur.Round(time.Millisecond)))
		sb.WriteString(summary + "\n")
	} else {
		summary := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true).
			Render(fmt.Sprintf("%d of %d code block(s) failed (%v total).", failed, total, totalDur.Round(time.Millisecond)))
		sb.WriteString(summary + "\n")
	}

	return sb.String()
}
