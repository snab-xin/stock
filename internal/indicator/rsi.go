package indicator

import "math"

// RSI 相对强弱指数（Wilder 平滑）。
// 首个 RSI 用前 n 根涨跌幅的简单均值；其后 avgGain/avgLoss 按
// (prev*(n-1)+cur)/n 递推。输出长度与 closes 一致，预热期为 NaN。
func RSI(closes []float64, n int) []float64 {
	out := nanSlice(len(closes))
	if n <= 0 || len(closes) < n+1 {
		return out
	}

	var sumGain, sumLoss float64
	for i := 1; i <= n; i++ {
		d := closes[i] - closes[i-1]
		if d >= 0 {
			sumGain += d
		} else {
			sumLoss -= d
		}
	}
	avgGain := sumGain / float64(n)
	avgLoss := sumLoss / float64(n)
	out[n] = rsiFromAvgs(avgGain, avgLoss)

	for i := n + 1; i < len(closes); i++ {
		d := closes[i] - closes[i-1]
		var gain, loss float64
		if d >= 0 {
			gain = d
		} else {
			loss = -d
		}
		avgGain = (avgGain*float64(n-1) + gain) / float64(n)
		avgLoss = (avgLoss*float64(n-1) + loss) / float64(n)
		out[i] = rsiFromAvgs(avgGain, avgLoss)
	}
	return out
}

func rsiFromAvgs(avgGain, avgLoss float64) float64 {
	if avgLoss == 0 {
		if avgGain == 0 {
			return 50
		}
		return 100
	}
	rs := avgGain / avgLoss
	return 100 - 100/(1+rs)
}

// RSI14 默认 14 周期。
func RSI14(closes []float64) []float64 {
	return RSI(closes, 14)
}

// Finite 把 NaN/Inf 标成 false，便于序列化或测试。
func Finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
