package applog

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestInitText(t *testing.T) {
	var buf bytes.Buffer
	if err := Init(Options{Level: "info", Format: "text", Out: &buf}); err != nil {
		t.Fatal(err)
	}
	slog.Info("hello", "k", 1)
	if !strings.Contains(buf.String(), "hello") || !strings.Contains(buf.String(), "k=1") {
		t.Fatalf("log=%q", buf.String())
	}
}

func TestParseLevel(t *testing.T) {
	if _, err := ParseLevel("nope"); err == nil {
		t.Fatal("expected error")
	}
	lv, err := ParseLevel("debug")
	if err != nil || lv != slog.LevelDebug {
		t.Fatalf("%v %v", lv, err)
	}
}
