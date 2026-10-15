package tushare

import (
	"context"
	"testing"
	"time"
)

func TestRateLimitEvenSpacing(t *testing.T) {
	c := NewClient("x", "", 60) // 每秒最多 1 次
	if c.MinGap() != time.Second {
		t.Fatalf("minGap=%v", c.MinGap())
	}
	ctx := context.Background()
	start := time.Now()
	if err := c.wait(ctx); err != nil {
		t.Fatal(err)
	}
	if err := c.wait(ctx); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Fatalf("两次请求间隔过短: %v", elapsed)
	}
}

func TestRateLimitUsesConfiguredRPM(t *testing.T) {
	c := NewClient("x", "", 50)
	if c.RPM() != 50 {
		t.Fatalf("rpm=%d", c.RPM())
	}
	want := time.Minute / 50
	if c.MinGap() != want {
		t.Fatalf("minGap=%v want=%v", c.MinGap(), want)
	}
}
