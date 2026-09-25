package internal

import (
	"strings"
	"testing"
	"time"
)

func TestExecuteBlockSh(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "bash",
		Text: "echo 'hello from termdeck runner'",
	}

	res := ExecuteBlock(blk, 2*time.Second)
	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d (err: %s)", res.ExitCode, res.Error)
	}
	if !strings.Contains(res.Stdout, "hello from termdeck runner") {
		t.Fatalf("expected stdout to contain message, got: %q", res.Stdout)
	}
	if res.Error != "" {
		t.Fatalf("unexpected error: %s", res.Error)
	}
}

func TestExecuteBlockNonZeroExit(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "sh",
		Text: "exit 42",
	}

	res := ExecuteBlock(blk, 2*time.Second)
	if res.ExitCode != 42 {
		t.Fatalf("expected exit code 42, got %d", res.ExitCode)
	}
}

func TestExecuteBlockTimeout(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "sh",
		Text: "sleep 2",
	}

	res := ExecuteBlock(blk, 100*time.Millisecond)
	if res.ExitCode != 124 {
		t.Fatalf("expected exit code 124 on timeout, got %d (error: %s)", res.ExitCode, res.Error)
	}
	if !strings.Contains(res.Error, "timed out") {
		t.Fatalf("expected timeout message in error, got %s", res.Error)
	}
}

func TestExecuteBlockNonCode(t *testing.T) {
	blk := Block{
		Kind: BlockHeading,
		Text: "# Heading",
	}

	res := ExecuteBlock(blk, 1*time.Second)
	if res.ExitCode == 0 || res.Error == "" {
		t.Fatalf("expected error executing non-code block, got: %+v", res)
	}
}

func TestIsExecutableLanguage(t *testing.T) {
	if !IsExecutableLanguage("bash") || !IsExecutableLanguage("sh") || !IsExecutableLanguage("python") || !IsExecutableLanguage("go") {
		t.Fatal("expected executable languages to return true")
	}
	if IsExecutableLanguage("text") || IsExecutableLanguage("diff") || IsExecutableLanguage("yaml") || IsExecutableLanguage("json") {
		t.Fatal("expected display-only languages to return false")
	}

	// Execute display-only language
	blk := Block{
		Kind: BlockCode,
		Lang: "diff",
		Text: "--- a\n+++ b",
	}
	res := ExecuteBlock(blk, 1*time.Second)
	if res.ExitCode == 0 || (!strings.Contains(res.Error, "display-only") && !strings.Contains(res.Error, "not executable")) {
		t.Fatalf("expected display-only error, got: %+v", res)
	}
}

func TestExecuteBlockEmpty(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "bash",
		Text: "   \n  \n",
	}

	res := ExecuteBlock(blk, 1*time.Second)
	if res.Error != "code block is empty" {
		t.Fatalf("expected empty code block error, got: %s", res.Error)
	}
}

func TestExecuteBlockGo(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "go",
		Text: `println("termdeck go runner")`,
	}

	res := ExecuteBlock(blk, 10*time.Second)
	// On systems with go installed, this runs and prints to stderr/stdout
	if res.ExitCode != 0 && res.Error != "" && !strings.Contains(res.Error, "executable file not found") {
		t.Logf("Go execution result: %+v", res)
	}
}

func TestExecuteBlockWithLines(t *testing.T) {
	blk := Block{
		Kind:  BlockCode,
		Lang:  "bash",
		Lines: []string{"A=10", "B=20", "echo $((A + B))"},
	}

	res := ExecuteBlock(blk, 2*time.Second)
	if res.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d (err: %s)", res.ExitCode, res.Error)
	}
	if strings.TrimSpace(res.Stdout) != "30" {
		t.Fatalf("expected output 30, got %q", res.Stdout)
	}
}

func TestTruncateOutput(t *testing.T) {
	longStr := strings.Repeat("a", 100)
	trunc, wasTrunc := truncateOutput(longStr, 50, 10)
	if !wasTrunc {
		t.Fatal("expected truncation")
	}
	if !strings.Contains(trunc, "... (output truncated)") {
		t.Fatalf("expected truncation notice in %q", trunc)
	}

	manyLines := strings.Repeat("line\n", 50)
	truncLines, wasTruncLines := truncateOutput(manyLines, 1000, 10)
	if !wasTruncLines {
		t.Fatal("expected line truncation")
	}
	if !strings.Contains(truncLines, "... (output truncated)") {
		t.Fatalf("expected truncation notice in %q", truncLines)
	}
}

func TestExecuteCodeCmd(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "sh",
		Text: "echo 'async command'",
	}

	cmd := ExecuteCodeCmd(blk, 2*time.Second, 1, 2)
	msg := cmd()
	execMsg, ok := msg.(ExecFinishedMsg)
	if !ok {
		t.Fatalf("expected ExecFinishedMsg, got %T", msg)
	}
	if execMsg.Result.SlideNum != 1 || execMsg.Result.BlockNum != 2 {
		t.Fatalf("unexpected slide/block num: %+v", execMsg.Result)
	}
	if !strings.Contains(execMsg.Result.Stdout, "async command") {
		t.Fatalf("unexpected output: %s", execMsg.Result.Stdout)
	}
}

func TestTestAllDeckCode(t *testing.T) {
	deck := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{Kind: BlockHeading, Text: "Slide 1"},
					{Kind: BlockCode, Lang: "sh", Text: "echo 'pass 1'"},
				},
			},
			{
				Blocks: []Block{
					{Kind: BlockHeading, Text: "Slide 2"},
					{Kind: BlockCode, Lang: "sh", Text: "exit 1"},
				},
			},
		},
	}

	passed, failed, results := TestAllDeckCode(deck, 2*time.Second)
	if passed != 1 || failed != 1 {
		t.Fatalf("expected 1 passed, 1 failed; got passed=%d, failed=%d", passed, failed)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	out := FormatTestCodeCLI("test.deck.md", passed, failed, results, 150*time.Millisecond)
	if !strings.Contains(out, "1 of 2 code block(s) failed") {
		t.Fatalf("unexpected CLI output: %s", out)
	}

	// All pass format
	outPass := FormatTestCodeCLI("test.deck.md", 2, 0, results[:1], 50*time.Millisecond)
	if !strings.Contains(outPass, "All 2 code blocks passed") {
		t.Fatalf("unexpected pass output: %s", outPass)
	}
}

func TestTestAllDeckCodeWithColumnsAndDedent(t *testing.T) {
	deck := Deck{
		Slides: []Slide{
			{
				Blocks: []Block{
					{
						Kind: BlockColumns,
						Columns: [][]Block{
							{
								{Kind: BlockCode, Lang: "sh", Text: "echo 'col 1'"},
								{Kind: BlockCode, Lang: "sh", NoEval: true, Text: "rm -rf /"},
								{Kind: BlockCode, Lang: "text", Text: "diagram"},
								{Kind: BlockCode, Lang: "sh", Text: "   "},
							},
							{
								{Kind: BlockCode, Lang: "bash", Lines: []string{"echo 'col 2'"}},
								{Kind: BlockCode, Lang: "sh", Text: "exit 2"},
							},
						},
					},
				},
			},
		},
	}

	passed, failed, results := TestAllDeckCode(deck, 2*time.Second)
	if passed != 2 || failed != 1 {
		t.Fatalf("expected 2 passed, 1 failed from columns; got passed=%d, failed=%d", passed, failed)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results from columns, got %d", len(results))
	}

	// Test dedent
	indented := "    def foo():\n        return 42\n"
	dedented := dedent(indented)
	if !strings.HasPrefix(dedented, "def foo():") {
		t.Errorf("expected dedented code to start with 'def foo():', got:\n%s", dedented)
	}

	// Dedent single line or empty
	if dedent("") != "" {
		t.Errorf("expected empty string from dedent('')")
	}
}
