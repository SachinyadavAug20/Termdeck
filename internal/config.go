package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds termdeck configuration settings.
type Config struct {
	Theme         string
	StartAt       int
	Track         string
	Route         string
	Watch         bool
	Autoplay      bool
	AutoplaySec   int
	ListThemes    bool
	ExportHTML    bool
	ExportOutPath string
	ShowGraph     bool
	ShowMermaid   bool
	ShowDOT       bool
	TestCode      bool
	RunSlide      int
	ShowStats     bool
	ShowRadar     bool
	ShowGraphMap  bool
	LintDAG       bool
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Theme:   "termdeck",
		StartAt: 1,
	}
}

// LoadConfig loads configuration from a config file and environment variables.
// Config file locations (in order of precedence):
//  1. $HOME/.termdeck/config.toml
//  2. ./termdeck.toml
//  3. ./termdeck.yml
func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	// Try loading from config file
	configPaths := []string{
		homeDir(".termdeck/config.toml"),
		filepath.Join(".", "termdeck.toml"),
		filepath.Join(".", "termdeck.yml"),
	}

	for _, path := range configPaths {
		if err := loadConfigFile(path, cfg); err == nil {
			// Found a config file, no need to try others
			break
		}
	}

	// Apply environment variable overrides (env vars take precedence)
	cfg.applyEnvOverrides()

	return cfg, nil
}

func homeDir(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, path)
}

// loadConfigFile attempts to parse a TOML or YAML config file.
func loadConfigFile(path string, cfg Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)

	// Simple key: value parser supporting basic types
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			val = strings.Trim(val, "\"")

			switch strings.ToLower(key) {
			case "theme":
				cfg.Theme = val
			case "start-at", "start_at":
				n, err := fmt.Sscanf(val, "%d", &cfg.StartAt)
				if err != nil || n != 1 {
					return fmt.Errorf("invalid start-at value: %s", val)
				}
			case "track":
				cfg.Track = val
			case "route":
				cfg.Route = val
			case "watch":
				cfg.Watch = strings.ToLower(val) == "true" || val == "1"
			case "autoplay", "autoplay_sec":
				n, err := fmt.Sscanf(val, "%d", &cfg.AutoplaySec)
				if err != nil || n != 1 {
					return fmt.Errorf("invalid autoplay value: %s", val)
				}
				cfg.Autoplay = true
			case "list-themes", "list_themes":
				cfg.ListThemes = true
			case "export-html", "export_html":
				cfg.ExportHTML = true
			case "export-out-path", "export_out_path":
				cfg.ExportOutPath = val
			case "show-graph", "show_graph":
				cfg.ShowGraph = true
			case "show-mermaid", "show_mermaid":
				cfg.ShowMermaid = true
			case "test-code", "test_code":
				cfg.TestCode = true
			case "run-slide", "run_slide":
				n, err := fmt.Sscanf(val, "%d", &cfg.RunSlide)
				if err != nil || n != 1 {
					return fmt.Errorf("invalid run-slide value: %s", val)
				}
			case "show-stats", "show_stats":
				cfg.ShowStats = true
			case "show-radar", "show_radar":
				cfg.ShowRadar = true
			case "show-graph-map", "show_graph_map":
				cfg.ShowGraphMap = true
			case "show-dot", "show_dot":
				cfg.ShowDOT = true
			case "lint-dag", "lint_dag":
				cfg.LintDAG = true
			}
		}
	}
	return nil
}

// applyEnvOverrides applies environment variable overrides to the config.
// Environment variable format: TERMDECK_KEY=value
func (c *Config) applyEnvOverrides() {
	envMap := map[string]func(string){
		"TERMDECK_THEME": func(v string) { c.Theme = v },
		"TERMDECK_START_AT": func(v string) {
			if n, _ := fmt.Sscanf(v, "%d", &c.StartAt); n != 1 {
				c.StartAt = 1
			}
		},
		"TERMDECK_TRACK": func(v string) { c.Track = v },
		"TERMDECK_ROUTE": func(v string) { c.Route = v },
		"TERMDECK_WATCH": func(v string) {
			c.Watch = v != ""
		},
		"TERMDECK_AUTOPLAY": func(v string) {
			c.Autoplay = true
			if n, _ := fmt.Sscanf(v, "%d", &c.AutoplaySec); n != 1 {
				c.AutoplaySec = DefaultAutoplaySec
			}
		},
		"TERMDECK_EXPORT_HTML":    func(v string) { c.ExportHTML = true },
		"TERMDECK_SHOW_GRAPH":     func(v string) { c.ShowGraph = true },
		"TERMDECK_SHOW_MERMAID":   func(v string) { c.ShowMermaid = true },
		"TERMDECK_SHOW_DOT":       func(v string) { c.ShowDOT = true },
		"TERMDECK_TEST_CODE":      func(v string) { c.TestCode = true },
		"TERMDECK_SHOW_STATS":     func(v string) { c.ShowStats = true },
		"TERMDECK_SHOW_RADAR":     func(v string) { c.ShowRadar = true },
		"TERMDECK_SHOW_GRAPH_MAP": func(v string) { c.ShowGraphMap = true },
		"TERMDECK_LINT_DAG":       func(v string) { c.LintDAG = true },
	}

	for key, setter := range envMap {
		if val := os.Getenv(key); val != "" {
			setter(val)
		}
	}
}

// SaveConfig saves the configuration to a file.
func SaveConfig(cfg Config, path string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "theme: %s\n", cfg.Theme)
	if cfg.StartAt > 0 {
		fmt.Fprintf(&b, "start-at: %d\n", cfg.StartAt)
	}
	if cfg.Track != "" {
		fmt.Fprintf(&b, "track: %s\n", cfg.Track)
	}
	if cfg.Route != "" {
		fmt.Fprintf(&b, "route: %s\n", cfg.Route)
	}
	if cfg.Watch {
		fmt.Fprintf(&b, "watch: true\n")
	}
	if cfg.Autoplay {
		fmt.Fprintf(&b, "autoplay: %d\n", cfg.AutoplaySec)
	}
	if cfg.ExportHTML {
		fmt.Fprintf(&b, "export-html: true\n")
	}
	if cfg.ExportOutPath != "" {
		fmt.Fprintf(&b, "export-out-path: %s\n", cfg.ExportOutPath)
	}
	if cfg.ShowGraph {
		fmt.Fprintf(&b, "show-graph: true\n")
	}
	if cfg.ShowMermaid {
		fmt.Fprintf(&b, "show-mermaid: true\n")
	}
	if cfg.TestCode {
		fmt.Fprintf(&b, "test-code: true\n")
	}
	if cfg.RunSlide > 0 {
		fmt.Fprintf(&b, "run-slide: %d\n", cfg.RunSlide)
	}
	if cfg.ShowStats {
		fmt.Fprintf(&b, "show-stats: true\n")
	}
	if cfg.ShowRadar {
		fmt.Fprintf(&b, "show-radar: true\n")
	}
	if cfg.ShowGraphMap {
		fmt.Fprintf(&b, "show-graph-map: true\n")
	}
	if cfg.LintDAG {
		fmt.Fprintf(&b, "lint-dag: true\n")
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}
