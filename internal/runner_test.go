package internal

import (
	"context"
	"os/exec"
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

	res := ExecuteBlock(blk, 2*time.Second, nil)
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

	res := ExecuteBlock(blk, 2*time.Second, nil)
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

	res := ExecuteBlock(blk, 100*time.Millisecond, nil)
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

	res := ExecuteBlock(blk, 1*time.Second, nil)
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
	res := ExecuteBlock(blk, 1*time.Second, nil)
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

	res := ExecuteBlock(blk, 1*time.Second, nil)
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

	res := ExecuteBlock(blk, 10*time.Second, nil)
	// On systems with go installed, this runs and prints to stderr/stdout
	if res.ExitCode != 0 && res.Error != "" && !strings.Contains(res.Error, "executable file not found") {
		t.Logf("Go execution result: %+v", res)
	}
}

func TestExecuteBlockPython(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "python",
		Text: `print(f"python result: {20 + 22}")`,
	}
	res := ExecuteBlock(blk, 5*time.Second, nil)
	if res.ExitCode != 0 {
		t.Fatalf("expected python exit 0, got %d (err: %s)", res.ExitCode, res.Error)
	}
	if !strings.Contains(res.Stdout, "python result: 42") {
		t.Fatalf("expected python output 'python result: 42', got: %q", res.Stdout)
	}
}

func TestExecuteBlockNode(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "javascript",
		Text: `console.log("node result: " + (30 + 12));`,
	}
	res := ExecuteBlock(blk, 5*time.Second, nil)
	if res.ExitCode != 0 {
		t.Fatalf("expected node exit 0, got %d (err: %s)", res.ExitCode, res.Error)
	}
	if !strings.Contains(res.Stdout, "node result: 42") {
		t.Fatalf("expected node output 'node result: 42', got: %q", res.Stdout)
	}
}

func TestExecuteBlockEnv(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "bash",
		Text: `echo "CUSTOM_VAR=$MY_CUSTOM_TEST_VAR"`,
	}
	res := ExecuteBlock(blk, 2*time.Second, []string{"MY_CUSTOM_TEST_VAR=termdeck_rocks"})
	if res.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d", res.ExitCode)
	}
	if !strings.Contains(res.Stdout, "CUSTOM_VAR=termdeck_rocks") {
		t.Fatalf("expected env variable in output, got: %q", res.Stdout)
	}
}

func TestExecuteBlockWithLines(t *testing.T) {
	blk := Block{
		Kind:  BlockCode,
		Lang:  "bash",
		Lines: []string{"A=10", "B=20", "echo $((A + B))"},
	}

	res := ExecuteBlock(blk, 2*time.Second, nil)
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

	cmd := ExecuteCodeCmd(blk, 2*time.Second, 1, 2, nil)
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

func TestExecuteBlockRust(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "rust",
		Text: `println!("termdeck rust runner: {}", 1 + 2);`,
	}
	res := ExecuteBlock(blk, 10*time.Second, nil)
	if res.ExitCode != 0 {
		t.Fatalf("expected rust exit 0, got %d (err: %s, stderr: %s)", res.ExitCode, res.Error, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "termdeck rust runner: 3") {
		t.Fatalf("expected stdout to contain computed value, got: %q", res.Stdout)
	}
}

func TestExecuteBlockRust_CompileError(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "rust",
		Text: `this is not valid rust code syntax!`,
	}
	res := ExecuteBlock(blk, 10*time.Second, nil)
	if res.ExitCode == 0 {
		t.Fatalf("expected compile error exit code != 0, got %d", res.ExitCode)
	}
	if res.Error != "compilation failed" {
		t.Fatalf("expected 'compilation failed' error, got: %s", res.Error)
	}
}

func TestExecuteBlockC(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "c",
		Text: `printf("termdeck c runner: %d\n", 40 + 2);`,
	}
	res := ExecuteBlock(blk, 10*time.Second, nil)
	if res.ExitCode != 0 {
		t.Fatalf("expected C exit 0, got %d (err: %s, stderr: %s)", res.ExitCode, res.Error, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "termdeck c runner: 42") {
		t.Fatalf("expected stdout to contain computed value, got: %q", res.Stdout)
	}
}

func TestExecuteBlockCpp(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "cpp",
		Text: `std::cout << "termdeck cpp runner: " << (5 * 5) << std::endl;`,
	}
	res := ExecuteBlock(blk, 10*time.Second, nil)
	if res.ExitCode != 0 {
		t.Fatalf("expected C++ exit 0, got %d (err: %s, stderr: %s)", res.ExitCode, res.Error, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "termdeck cpp runner: 25") {
		t.Fatalf("expected stdout to contain computed value, got: %q", res.Stdout)
	}
}

func TestExecuteBlockLua(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "lua",
		Text: `print("termdeck lua: " .. (10 + 5))`,
	}
	res := ExecuteBlock(blk, 5*time.Second, nil)
	if res.ExitCode != 0 {
		t.Fatalf("expected lua exit 0, got %d (err: %s, stderr: %s)", res.ExitCode, res.Error, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "termdeck lua: 15") {
		t.Fatalf("expected stdout to contain 'termdeck lua: 15', got: %q", res.Stdout)
	}
}

func TestExecuteBlockPerl(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "perl",
		Text: `print "termdeck perl: " . (8 * 2) . "\n";`,
	}
	res := ExecuteBlock(blk, 5*time.Second, nil)
	if res.ExitCode != 0 {
		t.Fatalf("expected perl exit 0, got %d (err: %s, stderr: %s)", res.ExitCode, res.Error, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "termdeck perl: 16") {
		t.Fatalf("expected stdout to contain 'termdeck perl: 16', got: %q", res.Stdout)
	}
}

func TestExecuteBlockTypeScript(t *testing.T) {
	blk := Block{
		Kind: BlockCode,
		Lang: "ts",
		Text: `const score: number = 100; console.log("ts score:", score);`,
	}
	res := ExecuteBlock(blk, 5*time.Second, nil)
	if res.ExitCode != 0 {
		t.Logf("TypeScript runtime result: %+v", res)
	} else if !strings.Contains(res.Stdout, "ts score: 100") {
		t.Fatalf("expected stdout 'ts score: 100', got: %q", res.Stdout)
	}
}

type MockZigRunner struct{}

func (m *MockZigRunner) Supports(lang string) bool {
	return lang == "zig"
}

func (m *MockZigRunner) BuildCommand(ctx context.Context, code string) (*exec.Cmd, []string, string, error) {
	return exec.CommandContext(ctx, "echo", "zig runner: "+code), nil, "", nil
}

func TestCustomLanguageRunner_OCP(t *testing.T) {
	RegisterRunner(&MockZigRunner{})
	if !IsExecutableLanguage("zig") {
		t.Fatalf("expected zig to be recognized as executable language after registration")
	}
	blk := Block{
		Kind: BlockCode,
		Lang: "zig",
		Text: "const x = 5;",
	}
	res := ExecuteBlock(blk, 2*time.Second, nil)
	if res.ExitCode != 0 || !strings.Contains(res.Stdout, "zig runner: const x = 5;") {
		t.Fatalf("expected mock zig runner output, got: %+v", res)
	}
}


