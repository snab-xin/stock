package backtest

import "math"

func commission(amount, rate, min float64) float64 {
	if amount <= 0 {
		return 0
	}
	c := amount * rate
	if c < min {
		return min
	}
	return c
}

func buyFillPrice(open, slippage float64) float64 {
	return open * (1 + slippage)
}

func sellFillPrice(open, slippage float64) float64 {
	p := open * (1 - slippage)
	if p < 0 {
		return 0
	}
	return p
}

func maxAffordableShares(cash, price, commRate, commMin float64, lot int64) int64 {
	if cash <= 0 || price <= 0 || lot <= 0 {
		return 0
	}
	maxLots := int64(math.Floor(cash / price / float64(lot)))
	for lots := maxLots; lots >= 1; lots-- {
		sh := lots * lot
		cost := price * float64(sh)
		comm := commission(cost, commRate, commMin)
		if cost+comm <= cash+1e-9 {
			return sh
		}
	}
	return 0
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
