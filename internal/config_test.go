package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Theme != "termdeck" {
		t.Errorf("expected default theme 'termdeck', got %q", cfg.Theme)
	}
	if cfg.StartAt != 1 {
		t.Errorf("expected default startAt 1, got %d", cfg.StartAt)
	}
}

func TestLoadConfigFromPath_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test_config.toml")

	content := `# Termdeck Configuration
theme: "nord"
start-at: 3
track: "deep-dive"
route: "fast-track"
watch: true
autoplay: 8
list-themes: true
export-html: true
export-out-path: "./dist/slides.html"
show-graph: true
show-mermaid: true
show-dot: true
test-code: true
run-slide: 4
show-stats: true
show-radar: true
show-graph-map: true
lint-dag: true
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfigFromPath(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Theme != "nord" {
		t.Errorf("expected theme 'nord', got %q", cfg.Theme)
	}
	if cfg.StartAt != 3 {
		t.Errorf("expected startAt 3, got %d", cfg.StartAt)
	}
	if cfg.Track != "deep-dive" {
		t.Errorf("expected track 'deep-dive', got %q", cfg.Track)
	}
	if cfg.Route != "fast-track" {
		t.Errorf("expected route 'fast-track', got %q", cfg.Route)
	}
	if !cfg.Watch {
		t.Errorf("expected watch true")
	}
	if !cfg.Autoplay || cfg.AutoplaySec != 8 {
		t.Errorf("expected autoplay true with 8s, got %v (%d)", cfg.Autoplay, cfg.AutoplaySec)
	}
	if !cfg.ListThemes {
		t.Errorf("expected listThemes true")
	}
	if !cfg.ExportHTML {
		t.Errorf("expected exportHTML true")
	}
	if cfg.ExportOutPath != "./dist/slides.html" {
		t.Errorf("expected exportOutPath './dist/slides.html', got %q", cfg.ExportOutPath)
	}
	if !cfg.ShowGraph {
		t.Errorf("expected showGraph true")
	}
	if !cfg.ShowMermaid {
		t.Errorf("expected showMermaid true")
	}
	if !cfg.ShowDOT {
		t.Errorf("expected showDOT true")
	}
	if !cfg.TestCode {
		t.Errorf("expected testCode true")
	}
	if cfg.RunSlide != 4 {
		t.Errorf("expected runSlide 4, got %d", cfg.RunSlide)
	}
	if !cfg.ShowStats {
		t.Errorf("expected showStats true")
	}
	if !cfg.ShowRadar {
		t.Errorf("expected showRadar true")
	}
	if !cfg.ShowGraphMap {
		t.Errorf("expected showGraphMap true")
	}
	if !cfg.LintDAG {
		t.Errorf("expected lintDAG true")
	}
}

func TestLoadConfigFromPath_SnakeCaseKeys(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test_config_snake.toml")

	content := `start_at: 5
autoplay_sec: 12
run_slide: 2
list_themes: true
export_html: true
export_out_path: "out.html"
show_graph: true
show_mermaid: true
show_dot: true
test_code: true
show_stats: true
show_radar: true
show_graph_map: true
lint_dag: true
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfigFromPath(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.StartAt != 5 {
		t.Errorf("expected startAt 5, got %d", cfg.StartAt)
	}
	if cfg.AutoplaySec != 12 || !cfg.Autoplay {
		t.Errorf("expected autoplay 12s, got %v (%d)", cfg.Autoplay, cfg.AutoplaySec)
	}
	if cfg.RunSlide != 2 {
		t.Errorf("expected runSlide 2, got %d", cfg.RunSlide)
	}
	if !cfg.ShowDOT {
		t.Errorf("expected showDOT true")
	}
}

func TestLoadConfigFromPath_InvalidValues(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Invalid start-at
	p1 := filepath.Join(tmpDir, "inv1.toml")
	_ = os.WriteFile(p1, []byte("start-at: notanumber\n"), 0644)
	if _, err := LoadConfigFromPath(p1); err == nil {
		t.Errorf("expected error for invalid start-at, got nil")
	}

	// 2. Invalid autoplay
	p2 := filepath.Join(tmpDir, "inv2.toml")
	_ = os.WriteFile(p2, []byte("autoplay: abc\n"), 0644)
	if _, err := LoadConfigFromPath(p2); err == nil {
		t.Errorf("expected error for invalid autoplay, got nil")
	}

	// 3. Invalid run-slide
	p3 := filepath.Join(tmpDir, "inv3.toml")
	_ = os.WriteFile(p3, []byte("run-slide: invalid\n"), 0644)
	if _, err := LoadConfigFromPath(p3); err == nil {
		t.Errorf("expected error for invalid run-slide, got nil")
	}

	// 4. Non-existent file
	if _, err := LoadConfigFromPath(filepath.Join(tmpDir, "non_existent.toml")); err == nil {
		t.Errorf("expected error for non-existent file, got nil")
	}
}

func TestConfigEnvOverrides(t *testing.T) {
	t.Setenv("TERMDECK_THEME", "matrix")
	t.Setenv("TERMDECK_START_AT", "7")
	t.Setenv("TERMDECK_TRACK", "backend")
	t.Setenv("TERMDECK_ROUTE", "quick")
	t.Setenv("TERMDECK_WATCH", "1")
	t.Setenv("TERMDECK_AUTOPLAY", "10")
	t.Setenv("TERMDECK_EXPORT_HTML", "1")
	t.Setenv("TERMDECK_SHOW_GRAPH", "1")
	t.Setenv("TERMDECK_SHOW_MERMAID", "1")
	t.Setenv("TERMDECK_SHOW_DOT", "1")
	t.Setenv("TERMDECK_TEST_CODE", "1")
	t.Setenv("TERMDECK_SHOW_STATS", "1")
	t.Setenv("TERMDECK_SHOW_RADAR", "1")
	t.Setenv("TERMDECK_SHOW_GRAPH_MAP", "1")
	t.Setenv("TERMDECK_LINT_DAG", "1")

	cfg := DefaultConfig()
	cfg.applyEnvOverrides()

	if cfg.Theme != "matrix" {
		t.Errorf("expected theme 'matrix', got %q", cfg.Theme)
	}
	if cfg.StartAt != 7 {
		t.Errorf("expected startAt 7, got %d", cfg.StartAt)
	}
	if cfg.Track != "backend" {
		t.Errorf("expected track 'backend', got %q", cfg.Track)
	}
	if cfg.Route != "quick" {
		t.Errorf("expected route 'quick', got %q", cfg.Route)
	}
	if !cfg.Watch {
		t.Errorf("expected watch true")
	}
	if !cfg.Autoplay || cfg.AutoplaySec != 10 {
		t.Errorf("expected autoplay true with 10s, got %v (%d)", cfg.Autoplay, cfg.AutoplaySec)
	}
	if !cfg.ExportHTML || !cfg.ShowGraph || !cfg.ShowMermaid || !cfg.ShowDOT || !cfg.TestCode || !cfg.ShowStats || !cfg.ShowRadar || !cfg.ShowGraphMap || !cfg.LintDAG {
		t.Errorf("expected all boolean env flags to be true")
	}
}

func TestConfigEnvOverrides_InvalidNumeric(t *testing.T) {
	t.Setenv("TERMDECK_START_AT", "not_a_number")
	t.Setenv("TERMDECK_AUTOPLAY", "not_a_number")

	cfg := DefaultConfig()
	cfg.applyEnvOverrides()

	if cfg.StartAt != 1 {
		t.Errorf("expected startAt fallback to 1, got %d", cfg.StartAt)
	}
	if !cfg.Autoplay || cfg.AutoplaySec != DefaultAutoplaySec {
		t.Errorf("expected autoplay fallback to DefaultAutoplaySec (%d), got %d", DefaultAutoplaySec, cfg.AutoplaySec)
	}
}

func TestSaveConfigRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	savePath := filepath.Join(tmpDir, "saved_config.toml")

	original := Config{
		Theme:         "tokyo-night",
		StartAt:       2,
		Track:         "frontend",
		Route:         "deep",
		Watch:         true,
		Autoplay:      true,
		AutoplaySec:   6,
		ExportHTML:    true,
		ExportOutPath: "out.html",
		ShowGraph:     true,
		ShowMermaid:   true,
		ShowDOT:       true,
		TestCode:      true,
		RunSlide:      3,
		ShowStats:     true,
		ShowRadar:     true,
		ShowGraphMap:  true,
		LintDAG:       true,
	}

	if err := SaveConfig(original, savePath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := LoadConfigFromPath(savePath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loaded.Theme != original.Theme {
		t.Errorf("theme mismatch: %q vs %q", loaded.Theme, original.Theme)
	}
	if loaded.StartAt != original.StartAt {
		t.Errorf("startAt mismatch: %d vs %d", loaded.StartAt, original.StartAt)
	}
	if loaded.Track != original.Track {
		t.Errorf("track mismatch: %q vs %q", loaded.Track, original.Track)
	}
	if loaded.Route != original.Route {
		t.Errorf("route mismatch: %q vs %q", loaded.Route, original.Route)
	}
	if loaded.Watch != original.Watch {
		t.Errorf("watch mismatch: %v vs %v", loaded.Watch, original.Watch)
	}
	if loaded.Autoplay != original.Autoplay || loaded.AutoplaySec != original.AutoplaySec {
		t.Errorf("autoplay mismatch: %v (%d) vs %v (%d)", loaded.Autoplay, loaded.AutoplaySec, original.Autoplay, original.AutoplaySec)
	}
	if loaded.ShowDOT != original.ShowDOT {
		t.Errorf("showDOT mismatch: %v vs %v", loaded.ShowDOT, original.ShowDOT)
	}
	if loaded.RunSlide != original.RunSlide {
		t.Errorf("runSlide mismatch: %d vs %d", loaded.RunSlide, original.RunSlide)
	}
}

func TestLoadConfig_DefaultFallback(t *testing.T) {
	// Calling LoadConfig without existing local config should succeed and return defaults
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}
	if cfg.Theme == "" {
		t.Errorf("expected non-empty theme in default config")
	}
}
