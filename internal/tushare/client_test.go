package tushare

import (
	"context"
	"testing"
	"time"
)

func TestRateLimitRPM(t *testing.T) {
	c := NewClient("x", "", 3)
	ctx := context.Background()
	start := time.Now()
	for i := 0; i < 3; i++ {
		if err := c.wait(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("前 %d 次不应被限流", 3)
	}
	done := make(chan error, 1)
	go func() {
		ctx2, cancel := context.WithTimeout(ctx, 80*time.Millisecond)
		defer cancel()
		done <- c.wait(ctx2)
	}()
	err := <-done
	if err == nil {
		t.Fatal("第 4 次应在窗口未滑出前被挡住")
	}
}
