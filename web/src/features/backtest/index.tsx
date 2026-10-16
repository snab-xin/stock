import { useEffect, useMemo, useState } from 'react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

import { EquityChart } from '@/components/charts/equity-chart'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  fetchAppConfig,
  runBacktest,
  searchInstruments,
  type BacktestResult,
  type Instrument,
} from '@/lib/api'

function todayYYYYMMDD() {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}${m}${day}`
}

function pct(n: number) {
  const sign = n > 0 ? '+' : ''
  return `${sign}${(n * 100).toFixed(2)}%`
}

function money(n: number) {
  return n.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

function iso(d: string) {
  return d.slice(0, 10)
}

export function Backtest() {
  const [query, setQuery] = useState('000001.SZ')
  const [hints, setHints] = useState<Instrument[]>([])
  const [start, setStart] = useState('20260101')
  const [end, setEnd] = useState(todayYYYYMMDD())
  const [fast, setFast] = useState(5)
  const [slow, setSlow] = useState(20)
  const [cash, setCash] = useState(100000)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState<BacktestResult | null>(null)

  useEffect(() => {
    void fetchAppConfig()
      .then((cfg) => {
        if (!cfg.backtest) return
        setFast(cfg.backtest.fast)
        setSlow(cfg.backtest.slow)
        setCash(cfg.backtest.init_cash)
      })
      .catch(() => undefined)
  }, [])

  useEffect(() => {
    const q = query.trim()
    if (q.length < 2) {
      setHints([])
      return
    }
    const t = window.setTimeout(() => {
      void searchInstruments(q)
        .then(setHints)
        .catch(() => setHints([]))
    }, 250)
    return () => window.clearTimeout(t)
  }, [query])

  async function run() {
    setLoading(true)
    setError('')
    try {
      const res = await runBacktest({
        code: query,
        start,
        end,
        fast,
        slow,
        cash,
      })
      setResult(res)
    } catch (e) {
      setError(e instanceof Error ? e.message : '回测失败')
      setResult(null)
    } finally {
      setLoading(false)
    }
  }

  const fills = useMemo(() => result?.trades.filter((t) => t.side === 'buy' || t.side === 'sell') ?? [], [result])
  const skips = useMemo(() => result?.trades.filter((t) => t.side === 'skip') ?? [], [result])

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4'>
      <div>
        <h1 className="text-2xl font-bold tracking-tight">最小回测</h1>
        <p className="text-sm text-muted-foreground">
          双均线：T 日收盘用前复权算信号，T+1 未复权开盘成交。涨停买不进，跌停卖不出，停牌不成交。
        </p>
      </div>

      <div className="relative flex flex-wrap items-end gap-2">
        <label className="space-y-1 text-xs text-muted-foreground">
          代码
          <Input
            className="w-44"
            value={query}
            placeholder="000001.SZ"
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void run()
            }}
          />
        </label>
        <label className="space-y-1 text-xs text-muted-foreground">
          开始
          <Input className="w-32" value={start} onChange={(e) => setStart(e.target.value)} />
        </label>
        <label className="space-y-1 text-xs text-muted-foreground">
          结束
          <Input className="w-32" value={end} onChange={(e) => setEnd(e.target.value)} />
        </label>
        <label className="space-y-1 text-xs text-muted-foreground">
          快线
          <Input className="w-20" type="number" value={fast} onChange={(e) => setFast(Number(e.target.value))} />
        </label>
        <label className="space-y-1 text-xs text-muted-foreground">
          慢线
          <Input className="w-20" type="number" value={slow} onChange={(e) => setSlow(Number(e.target.value))} />
        </label>
        <label className="space-y-1 text-xs text-muted-foreground">
          资金
          <Input className="w-28" type="number" value={cash} onChange={(e) => setCash(Number(e.target.value))} />
        </label>
        <Button onClick={() => void run()} disabled={loading}>
          {loading ? '计算中…' : '运行'}
        </Button>
        {hints.length > 0 && (
          <div className="absolute top-16 z-10 w-80 overflow-hidden rounded-md border bg-popover shadow-md">
            {hints.map((it) => (
              <button
                key={it.ts_code}
                className="flex w-full items-center justify-between px-3 py-2 text-left text-sm hover:bg-accent"
                onClick={() => {
                  setQuery(it.ts_code)
                  setHints([])
                }}
              >
                <span>
                  {it.ts_code} {it.name}
                </span>
                {it.is_st ? <Badge variant="warning">ST</Badge> : null}
              </button>
            ))}
          </div>
        )}
      </div>

      {error ? (
        <p className="text-sm text-destructive">
          {error}。请先启动 PostgreSQL 和：go run ./cmd/ashare web
        </p>
      ) : null}

      {result && (
        <>
          <div className="flex flex-wrap items-center gap-2 text-sm">
            <span className="font-medium">
              {result.code} {result.name}
            </span>
            <Badge variant="secondary">
              双均线 {result.fast}/{result.slow}
            </Badge>
            <Badge variant="outline">
              {result.start} ~ {result.end}
            </Badge>
            <Badge variant="outline">{result.bar_count} 根</Badge>
          </div>

          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <Metric label="期末资金" value={money(result.end_equity)} />
            <Metric label="策略收益" value={pct(result.total_return)} />
            <Metric label="最大回撤" value={pct(result.max_drawdown)} />
            <Metric label="买入持有" value={pct(result.buy_hold_return)} hint="同期前复权" />
            <Metric label="年化" value={pct(result.ann_return)} />
            <Metric label="成交笔数" value={String(result.trade_count)} />
            <Metric label="回合胜率" value={`${(result.win_rate * 100).toFixed(0)}%`} />
            <Metric
              label="费用"
              value={`${money(result.commission + result.stamp_tax + result.slippage_cost)}`}
              hint={`佣金 ${money(result.commission)} / 印花 ${money(result.stamp_tax)} / 滑点 ${money(result.slippage_cost)}`}
            />
          </div>

          <Card>
            <CardHeader>
              <CardTitle>净值（现金 + 持股×未复权收盘）</CardTitle>
            </CardHeader>
            <CardContent>
              {result.equity.length === 0 ? (
                <p className="py-12 text-center text-sm text-muted-foreground">没有净值序列</p>
              ) : (
                <EquityChart points={result.equity} />
              )}
            </CardContent>
          </Card>

          <Card className="gap-0 py-0">
            <CardHeader className="px-4 py-3">
              <CardTitle>成交明细</CardTitle>
            </CardHeader>
            <CardContent className="px-0">
              {fills.length === 0 ? (
                <p className="px-4 pb-4 text-sm text-muted-foreground">这段区间没有成交。</p>
              ) : (
                <div className="overflow-x-auto border-t">
                  <table className="w-full text-left text-sm [&_td:first-child]:pl-4 [&_td:last-child]:pr-4 [&_th:first-child]:pl-4 [&_th:last-child]:pr-4">
                    <thead className="text-muted-foreground">
                      <tr>
                        <th className="py-1 pr-3 font-medium">日期</th>
                        <th className="py-1 pr-3 font-medium">方向</th>
                        <th className="py-1 pr-3 font-medium">价格</th>
                        <th className="py-1 pr-3 font-medium">股数</th>
                        <th className="py-1 pr-3 font-medium">金额</th>
                        <th className="py-1 pr-3 font-medium">佣金</th>
                        <th className="py-1 pr-3 font-medium">印花税</th>
                        <th className="py-1 font-medium">说明</th>
                      </tr>
                    </thead>
                    <tbody>
                      {fills.map((t, i) => (
                        <tr key={`${t.date}-${t.side}-${i}`} className="border-t">
                          <td className="py-1.5 pr-3">{iso(t.date)}</td>
                          <td className="py-1.5 pr-3">{t.side === 'buy' ? '买' : '卖'}</td>
                          <td className="py-1.5 pr-3">{t.price.toFixed(2)}</td>
                          <td className="py-1.5 pr-3">{t.shares}</td>
                          <td className="py-1.5 pr-3">{money(t.amount)}</td>
                          <td className="py-1.5 pr-3">{t.commission.toFixed(2)}</td>
                          <td className="py-1.5 pr-3">{t.stamp_tax.toFixed(2)}</td>
                          <td className="py-1.5 text-muted-foreground">{t.reason}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </CardContent>
          </Card>

          {skips.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle>未成交（每种原因只记一次连续段）</CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="space-y-1 text-sm text-muted-foreground">
                  {skips.map((t, i) => (
                    <li key={`${t.date}-${t.reason}-${i}`}>
                      {iso(t.date)} · {t.reason}
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          )}

          <Card>
            <CardHeader>
              <CardTitle>还没覆盖的规则</CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="list-disc space-y-1 pl-5 text-sm text-muted-foreground">
                {result.limitations.map((line) => (
                  <li key={line}>{line}</li>
                ))}
              </ul>
            </CardContent>
          </Card>
        </>
      )}
      </Main>
    </>
  )
}

function Metric({ label, value, hint }: { label: string; value: string; hint?: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-xs font-medium text-muted-foreground">{label}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-lg font-semibold">{value}</div>
        {hint ? <div className="mt-1 text-xs text-muted-foreground">{hint}</div> : null}
      </CardContent>
    </Card>
  )
}
