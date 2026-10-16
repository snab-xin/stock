package indicator

import "math"

func nanSlice(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}
	return out
}

// SMA 简单移动平均。预热期为 NaN。
func SMA(xs []float64, n int) []float64 {
	out := nanSlice(len(xs))
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

// EMA 指数移动平均。第一个有效值取前 n 根 SMA，其后 α=2/(n+1)。预热期为 NaN。
func EMA(xs []float64, n int) []float64 {
	out := nanSlice(len(xs))
	if n <= 0 || len(xs) < n {
		return out
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += xs[i]
	}
	out[n-1] = sum / float64(n)
	alpha := 2.0 / float64(n+1)
	for i := n; i < len(xs); i++ {
		out[i] = alpha*xs[i] + (1-alpha)*out[i-1]
	}
	return out
}
