package indicator

import (
	"math"
	"testing"
)

func almostEqual(a, b, eps float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return math.Abs(a-b) <= eps
}

func TestSMA(t *testing.T) {
	xs := []float64{1, 2, 3, 4}
	got := SMA(xs, 2)
	if !math.IsNaN(got[0]) {
		t.Fatal("first should be NaN")
	}
	if !almostEqual(got[1], 1.5, 1e-9) || !almostEqual(got[2], 2.5, 1e-9) {
		t.Fatalf("%v", got)
	}
}

func TestEMA(t *testing.T) {
	// SMA seed at i=2: (1+2+3)/3=2；α=2/4=0.5 → next=0.5*4+0.5*2=3
	xs := []float64{1, 2, 3, 4}
	got := EMA(xs, 3)
	if !math.IsNaN(got[0]) || !math.IsNaN(got[1]) {
		t.Fatalf("warmup: %v", got)
	}
	if !almostEqual(got[2], 2, 1e-9) || !almostEqual(got[3], 3, 1e-9) {
		t.Fatalf("%v", got)
	}
}

func TestRSI(t *testing.T) {
	// 单调上涨：n=2 时 avgLoss=0 → RSI=100
	up := []float64{10, 11, 12, 13, 14}
	got := RSI(up, 2)
	if !math.IsNaN(got[0]) || !math.IsNaN(got[1]) {
		t.Fatalf("warmup: %v", got)
	}
	if !almostEqual(got[2], 100, 1e-9) || !almostEqual(got[4], 100, 1e-9) {
		t.Fatalf("up %v", got)
	}

	// 手工：closes 10,11,10,11；n=2
	// Δ: +1,-1,+1
	// 首段 i=2（用前 2 个 Δ）：gain=1, loss=1 → avg=0.5 → RSI=50
	// i=3：Δ=+1；avgGain=(0.5+1)/2=0.75；avgLoss=(0.5+0)/2=0.25 → RS=3 → RSI=75
	xs := []float64{10, 11, 10, 11}
	got = RSI(xs, 2)
	if !almostEqual(got[2], 50, 1e-9) || !almostEqual(got[3], 75, 1e-9) {
		t.Fatalf("osc %v", got)
	}
}

func TestMACD(t *testing.T) {
	// 足够长的序列，保证 DIF/DEA 在末尾有限
	closes := make([]float64, 60)
	for i := range closes {
		closes[i] = 100 + float64(i)*0.5 + math.Sin(float64(i)/3)*2
	}
	m := MACDDefault(closes)
	if len(m.DIF) != 60 || len(m.DEA) != 60 || len(m.HIST) != 60 {
		t.Fatal("length mismatch")
	}
	// DIF 从 slow-1=25 起有效
	for i := 0; i < 25; i++ {
		if Finite(m.DIF[i]) {
			t.Fatalf("DIF[%d] should be NaN", i)
		}
	}
	if !Finite(m.DIF[25]) {
		t.Fatal("DIF[25] should be finite")
	}
	// DEA 种子在第一个有效 DIF 起第 9 根：25+8=33
	for i := 0; i < 33; i++ {
		if Finite(m.DEA[i]) {
			t.Fatalf("DEA[%d] should be NaN", i)
		}
	}
	if !Finite(m.DEA[33]) {
		t.Fatal("DEA[33] should be finite")
	}
	last := 59
	if !almostEqual(m.HIST[last], m.DIF[last]-m.DEA[last], 1e-9) {
		t.Fatalf("hist != dif-dea: %v %v %v", m.HIST[last], m.DIF[last], m.DEA[last])
	}
}

func TestMACDShortSeries(t *testing.T) {
	m := MACD([]float64{1, 2, 3}, 12, 26, 9)
	for i := range m.DIF {
		if Finite(m.DIF[i]) || Finite(m.DEA[i]) || Finite(m.HIST[i]) {
			t.Fatalf("expected all NaN, got %v", m)
		}
	}
}
