package backtest

import "math"

func SMA(xs []float64, n int) []float64 {
	out := make([]float64, len(xs))
	for i := range out {
		out[i] = math.NaN()
	}
	if n <= 0 || len(xs) < n {
		return out
	}
	sum := 0.0
	for i, v := range xs {
		sum += v
		if i >= n {
			sum -= xs[i-n]
		}
		if i >= n-1 {
			out[i] = sum / float64(n)
		}
	}
	return out
}
