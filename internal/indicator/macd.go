package indicator

// MACDResult DIF / DEA / HIST，与 closes 等长，预热期为 NaN。
// HIST = DIF − DEA（不用 2× 放大；常见软件亦有用 2× 的版本，此处采用差值本身）。
type MACDResult struct {
	DIF  []float64
	DEA  []float64
	HIST []float64
}

// MACD 指数平滑异同移动平均。fast/slow/signal 为 EMA 周期。
func MACD(closes []float64, fast, slow, signal int) MACDResult {
	n := len(closes)
	out := MACDResult{
		DIF:  nanSlice(n),
		DEA:  nanSlice(n),
		HIST: nanSlice(n),
	}
	if fast <= 0 || slow <= 0 || signal <= 0 || fast >= slow || n < slow {
		return out
	}

	emaFast := EMA(closes, fast)
	emaSlow := EMA(closes, slow)
	for i := 0; i < n; i++ {
		if Finite(emaFast[i]) && Finite(emaSlow[i]) {
			out.DIF[i] = emaFast[i] - emaSlow[i]
		}
	}

	// DEA 是 DIF 的 EMA；DIF 前 slow-1 为 NaN，从第一个有效 DIF 起算 EMA。
	out.DEA = emaOnSeries(out.DIF, signal)
	for i := 0; i < n; i++ {
		if Finite(out.DIF[i]) && Finite(out.DEA[i]) {
			out.HIST[i] = out.DIF[i] - out.DEA[i]
		}
	}
	return out
}

// emaOnSeries 对可能含 NaN 前缀的序列做 EMA：从第一个有限值起，
// 用其后 signal 个有限值的 SMA 作种子，再递推。
func emaOnSeries(xs []float64, n int) []float64 {
	out := nanSlice(len(xs))
	if n <= 0 {
		return out
	}
	start := -1
	for i, v := range xs {
		if Finite(v) {
			start = i
			break
		}
	}
	if start < 0 || start+n > len(xs) {
		return out
	}
	for i := start; i < start+n; i++ {
		if !Finite(xs[i]) {
			return out
		}
	}
	sum := 0.0
	for i := start; i < start+n; i++ {
		sum += xs[i]
	}
	seedIdx := start + n - 1
	out[seedIdx] = sum / float64(n)
	alpha := 2.0 / float64(n+1)
	for i := seedIdx + 1; i < len(xs); i++ {
		if !Finite(xs[i]) {
			continue
		}
		out[i] = alpha*xs[i] + (1-alpha)*out[i-1]
	}
	return out
}

// MACDDefault 标准参数 12/26/9。
func MACDDefault(closes []float64) MACDResult {
	return MACD(closes, 12, 26, 9)
}
