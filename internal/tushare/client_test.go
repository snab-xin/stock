package tushare

import (
	"context"
	"testing"
	"time"
)

func TestRateLimitEvenSpacing(t *testing.T) {
	c := NewClient("x", "", 60, nil) // 每秒最多 1 次
	if c.MinGap() != time.Second {
		t.Fatalf("minGap=%v", c.MinGap())
	}
	ctx := context.Background()
	start := time.Now()
	if err := c.wait(ctx, "daily"); err != nil {
		t.Fatal(err)
	}
	if err := c.wait(ctx, "daily"); err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)
	if elapsed < 900*time.Millisecond {
		t.Fatalf("两次请求间隔过短: %v", elapsed)
	}
}

func TestRateLimitUsesConfiguredRPM(t *testing.T) {
	c := NewClient("x", "", 50, map[string]int{"top10_cb_holders": 30})
	if c.RPM() != 50 {
		t.Fatalf("rpm=%d", c.RPM())
	}
	if c.RPMFor("daily") != 50 {
		t.Fatalf("default api rpm=%d", c.RPMFor("daily"))
	}
	if c.RPMFor("top10_cb_holders") != 30 {
		t.Fatalf("override rpm=%d", c.RPMFor("top10_cb_holders"))
	}
	want := time.Minute / 30
	if c.MinGapFor("top10_cb_holders") != want {
		t.Fatalf("minGap=%v want=%v", c.MinGapFor("top10_cb_holders"), want)
	}
}

func TestRateLimitPerAPIIndependent(t *testing.T) {
	c := NewClient("x", "", 60, nil)
	ctx := context.Background()
	if err := c.wait(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	if err := c.wait(ctx, "b"); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("不同接口不应互相阻塞: %v", time.Since(start))
	}
}
