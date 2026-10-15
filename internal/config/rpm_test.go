package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeYAMLRPMObject(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`
tushare:
  rpm:
    default: 300
    apis:
      top10_cb_holders: 100
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := Config{TushareRPM: DefaultTushareRPM, TushareAPIs: map[string]int{}}
	if err := mergeYAML(path, &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.TushareRPM != 300 {
		t.Fatalf("rpm=%d", cfg.TushareRPM)
	}
	if cfg.RPMFor("top10_cb_holders") != 100 {
		t.Fatalf("holders=%d", cfg.RPMFor("top10_cb_holders"))
	}
	if cfg.RPMFor("cb_daily") != 300 {
		t.Fatalf("daily fallback=%d", cfg.RPMFor("cb_daily"))
	}
}
