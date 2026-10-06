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
	// Env contains the environment variables that were used during execution (for debugging)
	Env []string
}

// ExecFinishedMsg is delivered to Bubble Tea when asynchronous code execution completes.
type ExecFinishedMsg struct {
	Result ExecResult
}

const (
	MaxOutputBytes     = 16 * 1024 // 16 KB max stdout/stderr capture
	MaxOutputLines     = 300       // 300 lines max
	DefaultTimeout     = 5 * time.Second
	DefaultAutoplaySec = 5
)

// LanguageRunner abstracts execution and compilation for a specific language (SRP & ISP).
type LanguageRunner interface {
	Supports(lang string) bool
	BuildCommand(ctx context.Context, code string) (cmd *exec.Cmd, cleanupFiles []string, compileStderr string, err error)
}

// RunnerRegistry manages language runners, enabling the Open-Closed Principle (OCP).
type RunnerRegistry struct {
	runners []LanguageRunner
}

// NewRunnerRegistry constructs a registry with default supported language runners.
func NewRunnerRegistry() *RunnerRegistry {
	r := &RunnerRegistry{}
	r.Register(&ShellRunner{})
	r.Register(&PythonRunner{})
	r.Register(&NodeRunner{})
	r.Register(&TypeScriptRunner{})
	r.Register(&RubyRunner{})
	r.Register(&GoRunner{})
	r.Register(&RustRunner{})
	r.Register(&CRunner{})
	r.Register(&CppRunner{})
	r.Register(&LuaRunner{})
	r.Register(&PerlRunner{})
	r.Register(&PhpRunner{})
	return r
}

// Register registers a new language runner without modifying existing runner logic (OCP).
func (r *RunnerRegistry) Register(runner LanguageRunner) {
	r.runners = append(r.runners, runner)
}

// Find retrieves the language runner responsible for the specified language.
func (r *RunnerRegistry) Find(lang string) (LanguageRunner, bool) {
	lang = strings.ToLower(strings.TrimSpace(lang))
	for _, runner := range r.runners {
		if runner.Supports(lang) {
			return runner, true
		}
	}
	return nil, false
}

// DefaultRegistry is the global language runner registry (DIP).
var DefaultRegistry = NewRunnerRegistry()

// RegisterRunner adds a new language runner to the global registry.
func RegisterRunner(runner LanguageRunner) {
	DefaultRegistry.Register(runner)
}

// ShellRunner handles bash, sh, zsh, and shell scripts.
type ShellRunner struct{}

func (r *ShellRunner) Supports(lang string) bool {
	switch lang {
	case "bash", "sh", "zsh", "shell", "":
		return true
	default:
		return false
	}
}

func (r *ShellRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	shBin := "/bin/sh"
	if _, err := exec.LookPath("bash"); err == nil {
		shBin = "bash"
	}
	return exec.CommandContext(ctx, shBin, "-c", code), nil, "", nil
}

// PythonRunner handles python, py, and python3 scripts.
type PythonRunner struct{}

func (r *PythonRunner) Supports(lang string) bool {
	switch lang {
	case "python", "py", "python3":
		return true
	default:
		return false
	}
}

func (r *PythonRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	pyBin := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err2 := exec.LookPath("python"); err2 == nil {
			pyBin = "python"
		}
	}
	return exec.CommandContext(ctx, pyBin, "-c", code), nil, "", nil
}

// NodeRunner handles JavaScript via Node.js (with bun/deno fallback).
type NodeRunner struct{}

func (r *NodeRunner) Supports(lang string) bool {
	switch lang {
	case "node", "js", "javascript":
		return true
	default:
		return false
	}
}

func (r *NodeRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	if nodeBin, err := exec.LookPath("node"); err == nil {
		return exec.CommandContext(ctx, nodeBin, "-e", code), nil, "", nil
	}
	if bunBin, err := exec.LookPath("bun"); err == nil {
		return exec.CommandContext(ctx, bunBin, "run", "-e", code), nil, "", nil
	}
	if denoBin, err := exec.LookPath("deno"); err == nil {
		return exec.CommandContext(ctx, denoBin, "eval", code), nil, "", nil
	}
	return exec.CommandContext(ctx, "node", "-e", code), nil, "", nil
}

// TypeScriptRunner handles TypeScript via Deno, Bun, tsx, or ts-node.
type TypeScriptRunner struct{}

func (r *TypeScriptRunner) Supports(lang string) bool {
	switch lang {
	case "ts", "typescript":
		return true
	default:
		return false
	}
}

func (r *TypeScriptRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	if denoBin, err := exec.LookPath("deno"); err == nil {
		return exec.CommandContext(ctx, denoBin, "eval", "--ext=ts", code), nil, "", nil
	}
	if bunBin, err := exec.LookPath("bun"); err == nil {
		return exec.CommandContext(ctx, bunBin, "run", "-e", code), nil, "", nil
	}
	if tsxBin, err := exec.LookPath("tsx"); err == nil {
		return exec.CommandContext(ctx, tsxBin, "-e", code), nil, "", nil
	}
	if tsNodeBin, err := exec.LookPath("ts-node"); err == nil {
		return exec.CommandContext(ctx, tsNodeBin, "-e", code), nil, "", nil
	}
	return nil, nil, "", fmt.Errorf("no TypeScript runtime found (install deno, bun, tsx, or ts-node)")
}

// RubyRunner handles Ruby scripts.
type RubyRunner struct{}

func (r *RubyRunner) Supports(lang string) bool {
	return lang == "ruby" || lang == "rb"
}

func (r *RubyRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	return exec.CommandContext(ctx, "ruby", "-e", code), nil, "", nil
}

// GoRunner compiles and runs Go snippets.
type GoRunner struct{}

func (r *GoRunner) Supports(lang string) bool {
	return lang == "go" || lang == "golang"
}

func (r *GoRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	goCode := code
	if !strings.Contains(code, "package ") {
		goCode = fmt.Sprintf("package main\n\nimport \"fmt\"\n\nfunc main() {\n\t%s\n}\n", strings.ReplaceAll(code, "\n", "\n\t"))
	}
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, fmt.Sprintf("termdeck_run_%d.go", time.Now().UnixNano()))
	if err := os.WriteFile(tmpFile, []byte(goCode), 0600); err != nil {
		return nil, nil, "", fmt.Errorf("failed to create temp file: %v", err)
	}
	return exec.CommandContext(ctx, "go", "run", tmpFile), []string{tmpFile}, "", nil
}

// RustRunner compiles and executes Rust snippets.
type RustRunner struct{}

func (r *RustRunner) Supports(lang string) bool {
	return lang == "rust" || lang == "rs"
}

func (r *RustRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	rustcBin, err := exec.LookPath("rustc")
	if err != nil {
		return nil, nil, "", fmt.Errorf("rust compiler not found (install rustc)")
	}
	rustCode := code
	if !strings.Contains(code, "fn main") {
		rustCode = fmt.Sprintf("fn main() {\n    %s\n}\n", strings.ReplaceAll(code, "\n", "\n    "))
	}
	tmpDir := os.TempDir()
	nano := time.Now().UnixNano()
	srcFile := filepath.Join(tmpDir, fmt.Sprintf("termdeck_run_%d.rs", nano))
	binFile := filepath.Join(tmpDir, fmt.Sprintf("termdeck_run_%d.bin", nano))
	if err := os.WriteFile(srcFile, []byte(rustCode), 0600); err != nil {
		return nil, nil, "", fmt.Errorf("failed to create temp file: %v", err)
	}
	cleanup := []string{srcFile, binFile}

	var compileStderr bytes.Buffer
	compileCmd := exec.CommandContext(ctx, rustcBin, srcFile, "-o", binFile)
	compileCmd.Stderr = &compileStderr
	if err := compileCmd.Run(); err != nil {
		return nil, cleanup, compileStderr.String(), fmt.Errorf("compilation failed")
	}
	return exec.CommandContext(ctx, binFile), cleanup, "", nil
}

// CRunner compiles and executes C snippets.
type CRunner struct{}

func (r *CRunner) Supports(lang string) bool {
	return lang == "c"
}

func (r *CRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	ccBin, err := exec.LookPath("gcc")
	if err != nil {
		ccBin, err = exec.LookPath("clang")
	}
	if err != nil {
		return nil, nil, "", fmt.Errorf("C compiler not found (install gcc or clang)")
	}
	cCode := code
	if !strings.Contains(code, "main(") && !strings.Contains(code, "main (") {
		cCode = fmt.Sprintf("#include <stdio.h>\n#include <stdlib.h>\n\nint main(void) {\n    %s\n    return 0;\n}\n", strings.ReplaceAll(code, "\n", "\n    "))
	}
	tmpDir := os.TempDir()
	nano := time.Now().UnixNano()
	srcFile := filepath.Join(tmpDir, fmt.Sprintf("termdeck_run_%d.c", nano))
	binFile := filepath.Join(tmpDir, fmt.Sprintf("termdeck_run_%d.bin", nano))
	if err := os.WriteFile(srcFile, []byte(cCode), 0600); err != nil {
		return nil, nil, "", fmt.Errorf("failed to create temp file: %v", err)
	}
	cleanup := []string{srcFile, binFile}

	var compileStderr bytes.Buffer
	compileCmd := exec.CommandContext(ctx, ccBin, srcFile, "-o", binFile, "-lm")
	compileCmd.Stderr = &compileStderr
	if err := compileCmd.Run(); err != nil {
		return nil, cleanup, compileStderr.String(), fmt.Errorf("compilation failed")
	}
	return exec.CommandContext(ctx, binFile), cleanup, "", nil
}

// CppRunner compiles and executes C++ snippets.
type CppRunner struct{}

func (r *CppRunner) Supports(lang string) bool {
	return lang == "cpp" || lang == "c++" || lang == "cc"
}

func (r *CppRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	cxxBin, err := exec.LookPath("g++")
	if err != nil {
		cxxBin, err = exec.LookPath("clang++")
	}
	if err != nil {
		return nil, nil, "", fmt.Errorf("C++ compiler not found (install g++ or clang++)")
	}
	cppCode := code
	if !strings.Contains(code, "main(") && !strings.Contains(code, "main (") {
		cppCode = fmt.Sprintf("#include <iostream>\n\nint main() {\n    %s\n    return 0;\n}\n", strings.ReplaceAll(code, "\n", "\n    "))
	}
	tmpDir := os.TempDir()
	nano := time.Now().UnixNano()
	srcFile := filepath.Join(tmpDir, fmt.Sprintf("termdeck_run_%d.cpp", nano))
	binFile := filepath.Join(tmpDir, fmt.Sprintf("termdeck_run_%d.bin", nano))
	if err := os.WriteFile(srcFile, []byte(cppCode), 0600); err != nil {
		return nil, nil, "", fmt.Errorf("failed to create temp file: %v", err)
	}
	cleanup := []string{srcFile, binFile}

	var compileStderr bytes.Buffer
	compileCmd := exec.CommandContext(ctx, cxxBin, srcFile, "-o", binFile)
	compileCmd.Stderr = &compileStderr
	if err := compileCmd.Run(); err != nil {
		return nil, cleanup, compileStderr.String(), fmt.Errorf("compilation failed")
	}
	return exec.CommandContext(ctx, binFile), cleanup, "", nil
}

// LuaRunner handles Lua scripts.
type LuaRunner struct{}

func (r *LuaRunner) Supports(lang string) bool {
	return lang == "lua" || lang == "luajit"
}

func (r *LuaRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	luaBin := "lua"
	if _, err := exec.LookPath("lua"); err != nil {
		if _, err2 := exec.LookPath("luajit"); err2 == nil {
			luaBin = "luajit"
		}
	}
	return exec.CommandContext(ctx, luaBin, "-e", code), nil, "", nil
}

// PerlRunner handles Perl scripts.
type PerlRunner struct{}

func (r *PerlRunner) Supports(lang string) bool {
	return lang == "perl" || lang == "pl"
}

func (r *PerlRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	return exec.CommandContext(ctx, "perl", "-e", code), nil, "", nil
}

// PhpRunner handles PHP scripts.
type PhpRunner struct{}

func (r *PhpRunner) Supports(lang string) bool {
	return lang == "php"
}

func (r *PhpRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	return exec.CommandContext(ctx, "php", "-r", code), nil, "", nil
}

// IsExecutableLanguage returns true for languages supported by the live runner.
func IsExecutableLanguage(lang string) bool {
	_, found := DefaultRegistry.Find(lang)
	return found
}

// ExecuteBlock executes the code within a Block using host interpreters and compilers.
func ExecuteBlock(blk Block, timeout time.Duration, env []string) ExecResult {
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
	runner, found := DefaultRegistry.Find(lang)
	if !found && lang != "" {
		res.Error = fmt.Sprintf("language %q is not executable (supported: bash, sh, python, go, rust, c, cpp, ts, node, ruby, lua, perl, php)", lang)
		res.ExitCode = 1
		res.Duration = time.Since(start)
		return res
	}
	if !found {
		runner, _ = DefaultRegistry.Find("sh")
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

	cmd, filesToRemove, compileStderr, prepErr := runner.BuildCommand(ctx, code)
	defer func() {
		for _, f := range filesToRemove {
			_ = os.Remove(f)
		}
	}()

	if prepErr != nil {
		res.Duration = time.Since(start)
		res.Stderr = compileStderr
		res.Error = prepErr.Error()
		res.ExitCode = 1
		return res
	}

	// Apply environment variables if specified (overrides parent env)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
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

	res.Env = env

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
func ExecuteCodeCmd(blk Block, timeout time.Duration, slideNum, blockNum int, env []string) tea.Cmd {
	return func() tea.Msg {
		res := ExecuteBlock(blk, timeout, env)
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
				res := ExecuteBlock(blk, timeout, nil)
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
							res := ExecuteBlock(inner, timeout, nil)
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
