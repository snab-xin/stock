import * as echarts from 'echarts'
import { useEffect, useRef } from 'react'
import type { Bar, BarIndicators, NoteMark } from '@/lib/api'

export type ZoomRange = { start: number; end: number }

type Props = {
  bars: Bar[]
  indicators?: BarIndicators
  zoom?: ZoomRange
  onZoomChange?: (range: ZoomRange) => void
  onLoadOlder?: () => void
  hasMoreOlder?: boolean
  loadingOlder?: boolean
  noteMarks?: NoteMark[]
  onNoteClick?: (id: number) => void
}

function roundPrice(n: number) {
  return Math.round(n * 1000) / 1000
}

function fmtPrice(n: number) {
  return roundPrice(n).toFixed(3)
}

function fmtRsi(n: number) {
  return n.toFixed(1)
}

function fmtMacd(n: number) {
  return n.toFixed(3)
}

function asNum(v: number | null | undefined): number | null {
  return v == null || Number.isNaN(v) ? null : v
}

/** 简单移动平均；预热段为 null。 */
function sma(closes: number[], n: number): (number | null)[] {
  const out: (number | null)[] = Array(closes.length).fill(null)
  if (n <= 0 || closes.length < n) return out
  let sum = 0
  for (let i = 0; i < closes.length; i++) {
    sum += closes[i]
    if (i >= n) sum -= closes[i - n]
    if (i >= n - 1) out[i] = roundPrice(sum / n)
  }
  return out
}

const MA_PERIODS = [5, 10, 20, 60, 120] as const
const MA_COLORS: Record<(typeof MA_PERIODS)[number], string> = {
  5: '#eab308',
  10: '#a855f7',
  20: '#3b82f6',
  60: '#f97316',
  120: '#64748b',
}

export function KlineChart({
  bars,
  indicators,
  zoom,
  onZoomChange,
  onLoadOlder,
  hasMoreOlder,
  loadingOlder,
  noteMarks,
  onNoteClick,
}: Props) {
  const ref = useRef<HTMLDivElement>(null)
  const loadArmed = useRef(true)
  const onLoadOlderRef = useRef(onLoadOlder)
  const onZoomChangeRef = useRef(onZoomChange)
  const hasMoreOlderRef = useRef(hasMoreOlder)
  const loadingOlderRef = useRef(loadingOlder)
  const onNoteClickRef = useRef(onNoteClick)
  onLoadOlderRef.current = onLoadOlder
  onZoomChangeRef.current = onZoomChange
  hasMoreOlderRef.current = hasMoreOlder
  loadingOlderRef.current = loadingOlder
  onNoteClickRef.current = onNoteClick

  useEffect(() => {
    loadArmed.current = true
  }, [bars.length, hasMoreOlder])

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const chart = echarts.init(el)
    // 仅在 bars 变化时应用父组件传入的缩放（加载更早后校正视口）
    const zoomStart = zoom?.start ?? 60
    const zoomEnd = zoom?.end ?? 100
    const up = '#dc2626'
    const down = '#16a34a'
    // 涨停 / 跌停用更醒目的填充+描边，与普通涨跌区分
    const limitUpFill = '#ef4444'
    const limitUpBorder = '#f59e0b'
    const limitDownFill = '#22c55e'
    const limitDownBorder = '#0ea5e9'

    const dates = bars.map((b) => b.trade_date)
    const ohlc = bars.map((b) => {
      const value = [
        roundPrice(b.open),
        roundPrice(b.close),
        roundPrice(b.low),
        roundPrice(b.high),
      ]
      if (b.is_limit_up) {
        return {
          value,
          itemStyle: {
            color: limitUpFill,
            color0: limitUpFill,
            borderColor: limitUpBorder,
            borderColor0: limitUpBorder,
            borderWidth: 2,
          },
        }
      }
      if (b.is_limit_down) {
        return {
          value,
          itemStyle: {
            color: limitDownFill,
            color0: limitDownFill,
            borderColor: limitDownBorder,
            borderColor0: limitDownBorder,
            borderWidth: 2,
          },
        }
      }
      return value
    })
    const vols = bars.map((b) => b.volume)
    const closes = bars.map((b) => b.close)
    const maSeries = MA_PERIODS.map((n) => ({
      name: `MA${n}`,
      data: sma(closes, n),
      color: MA_COLORS[n],
    }))
    const rsi = bars.map((_, i) => asNum(indicators?.rsi?.[i]))
    const dif = bars.map((_, i) => asNum(indicators?.macd_dif?.[i]))
    const dea = bars.map((_, i) => asNum(indicators?.macd_dea?.[i]))
    const hist = bars.map((_, i) => asNum(indicators?.macd_hist?.[i]))

    const markPoints = bars.flatMap((b) => {
      const items: {
        name: string
        coord: [string, number]
        value: string
        itemStyle: { color: string }
        noteId?: number
      }[] = []
      if (b.is_limit_up) {
        items.push({
          name: '涨停',
          coord: [b.trade_date, roundPrice(b.high)],
          value: '涨停',
          itemStyle: { color: limitUpBorder },
        })
      }
      if (b.is_limit_down) {
        items.push({
          name: '跌停',
          coord: [b.trade_date, roundPrice(b.low)],
          value: '跌停',
          itemStyle: { color: limitDownBorder },
        })
      }
      if (b.is_paused) {
        items.push({
          name: '停牌',
          coord: [b.trade_date, roundPrice(b.close)],
          value: '停',
          itemStyle: { color: '#a16207' },
        })
      }
      return items
    })

    const byDate = new Map(bars.map((b) => [b.trade_date, b]))
    for (const n of noteMarks ?? []) {
      const bar = byDate.get(n.trade_date)
      if (!bar) continue
      markPoints.push({
        name: n.summary || '笔记',
        coord: [n.trade_date, roundPrice(bar.high)],
        value: '记',
        itemStyle: { color: '#2563eb' },
        noteId: n.id,
      })
    }

    const xAxisIdx = [0, 1, 2, 3]

    chart.setOption({
      animation: false,
      legend: {
        data: [
          'K线',
          ...maSeries.map((m) => m.name),
          '成交量',
          'RSI',
          'DIF',
          'DEA',
          'MACD',
        ],
        top: 0,
        type: 'scroll',
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross' },
      },
      axisPointer: { link: [{ xAxisIndex: 'all' }] },
      grid: [
        { left: 56, right: 16, top: 36, height: '36%' },
        { left: 56, right: 16, top: '48%', height: '10%' },
        { left: 56, right: 16, top: '62%', height: '12%' },
        { left: 56, right: 16, top: '78%', height: '12%' },
      ],
      xAxis: [
        {
          type: 'category',
          data: dates,
          gridIndex: 0,
          axisLabel: { show: false },
          axisTick: { show: false },
        },
        {
          type: 'category',
          data: dates,
          gridIndex: 1,
          axisLabel: { show: false },
          axisTick: { show: false },
        },
        {
          type: 'category',
          data: dates,
          gridIndex: 2,
          axisLabel: { show: false },
          axisTick: { show: false },
        },
        {
          type: 'category',
          data: dates,
          gridIndex: 3,
          axisLabel: { fontSize: 11 },
        },
      ],
      yAxis: [
        {
          scale: true,
          gridIndex: 0,
          splitLine: { lineStyle: { opacity: 0.25 } },
          axisLabel: { formatter: (v: number) => fmtPrice(v) },
          axisPointer: {
            label: {
              formatter: (p: { value?: unknown }) =>
                typeof p.value === 'number' ? fmtPrice(p.value) : String(p.value ?? ''),
            },
          },
        },
        {
          gridIndex: 1,
          splitNumber: 2,
          axisLabel: { show: false },
          splitLine: { show: false },
        },
        {
          scale: true,
          gridIndex: 2,
          min: 0,
          max: 100,
          splitNumber: 2,
          axisLabel: { formatter: (v: number) => fmtRsi(v), fontSize: 10 },
          splitLine: { lineStyle: { opacity: 0.2 } },
        },
        {
          scale: true,
          gridIndex: 3,
          splitNumber: 2,
          axisLabel: { formatter: (v: number) => fmtMacd(v), fontSize: 10 },
          splitLine: { lineStyle: { opacity: 0.2 } },
        },
      ],
      dataZoom: [
        {
          type: 'inside',
          xAxisIndex: xAxisIdx,
          start: zoomStart,
          end: zoomEnd,
        },
        {
          type: 'slider',
          xAxisIndex: xAxisIdx,
          start: zoomStart,
          end: zoomEnd,
          height: 16,
          bottom: 4,
        },
      ],
      series: [
        {
          name: 'K线',
          type: 'candlestick',
          data: ohlc,
          xAxisIndex: 0,
          yAxisIndex: 0,
          itemStyle: {
            color: up,
            color0: down,
            borderColor: up,
            borderColor0: down,
          },
          markPoint: {
            symbol: 'pin',
            symbolSize: 30,
            label: { fontSize: 10, color: '#fff' },
            data: markPoints,
          },
          tooltip: {
            valueFormatter: (v: unknown) =>
              typeof v === 'number' ? fmtPrice(v) : String(v ?? ''),
          },
        },
        ...maSeries.map((m) => ({
          name: m.name,
          type: 'line' as const,
          xAxisIndex: 0,
          yAxisIndex: 0,
          data: m.data,
          showSymbol: false,
          lineStyle: { width: 1.2, color: m.color },
          itemStyle: { color: m.color },
          tooltip: {
            valueFormatter: (v: unknown) =>
              typeof v === 'number' ? fmtPrice(v) : '—',
          },
        })),
        {
          name: '成交量',
          type: 'bar',
          xAxisIndex: 1,
          yAxisIndex: 1,
          data: vols.map((v, i) => {
            const b = bars[i]
            let color = b.close >= b.open ? up : down
            if (b.is_limit_up) color = limitUpBorder
            else if (b.is_limit_down) color = limitDownBorder
            return { value: v, itemStyle: { color } }
          }),
          tooltip: {
            valueFormatter: (v: unknown) =>
              typeof v === 'number' ? v.toLocaleString('zh-CN') : String(v ?? ''),
          },
        },
        {
          name: 'RSI',
          type: 'line',
          xAxisIndex: 2,
          yAxisIndex: 2,
          data: rsi,
          showSymbol: false,
          lineStyle: { width: 1.5, color: '#2563eb' },
          markLine: {
            silent: true,
            symbol: 'none',
            lineStyle: { type: 'dashed', opacity: 0.5 },
            data: [{ yAxis: 30 }, { yAxis: 70 }],
            label: { formatter: '{c}', fontSize: 10 },
          },
          tooltip: {
            valueFormatter: (v: unknown) =>
              typeof v === 'number' ? fmtRsi(v) : '—',
          },
        },
        {
          name: 'MACD',
          type: 'bar',
          xAxisIndex: 3,
          yAxisIndex: 3,
          data: hist.map((v) =>
            v == null
              ? null
              : {
                  value: v,
                  itemStyle: {
                    color: v >= 0 ? '#dc2626' : '#16a34a',
                  },
                },
          ),
          tooltip: {
            valueFormatter: (v: unknown) =>
              typeof v === 'number' ? fmtMacd(v) : '—',
          },
        },
        {
          name: 'DIF',
          type: 'line',
          xAxisIndex: 3,
          yAxisIndex: 3,
          data: dif,
          showSymbol: false,
          lineStyle: { width: 1.2, color: '#ca8a04' },
          tooltip: {
            valueFormatter: (v: unknown) =>
              typeof v === 'number' ? fmtMacd(v) : '—',
          },
        },
        {
          name: 'DEA',
          type: 'line',
          xAxisIndex: 3,
          yAxisIndex: 3,
          data: dea,
          showSymbol: false,
          lineStyle: { width: 1.2, color: '#7c3aed' },
          tooltip: {
            valueFormatter: (v: unknown) =>
              typeof v === 'number' ? fmtMacd(v) : '—',
          },
        },
      ],
    })

    const onResize = () => chart.resize()
    window.addEventListener('resize', onResize)

    chart.on('click', (params) => {
      const data = params.data
      if (!data || typeof data !== 'object' || Array.isArray(data)) return
      const id = (data as { noteId?: number }).noteId
      if (id) onNoteClickRef.current?.(id)
    })

    chart.on('datazoom', () => {
      const opt = chart.getOption() as {
        dataZoom?: Array<{ start?: number; end?: number }>
      }
      const dz = opt.dataZoom?.[0]
      const start = dz?.start ?? zoomStart
      const end = dz?.end ?? zoomEnd
      onZoomChangeRef.current?.({ start, end })
      // 右滑看更早：窗口贴到左侧时继续拉历史
      if (
        hasMoreOlderRef.current &&
        !loadingOlderRef.current &&
        loadArmed.current &&
        start <= 8
      ) {
        loadArmed.current = false
        onLoadOlderRef.current?.()
      }
    })

    return () => {
      window.removeEventListener('resize', onResize)
      chart.dispose()
    }
    // zoom 故意不进依赖：拖动缩放只更新父状态，避免整图重建
    // eslint-disable-next-line react-hooks/exhaustive-deps -- apply zoom only when bars/indicators change
  }, [bars, indicators, noteMarks])

  return (
    <div className='relative'>
      {loadingOlder ? (
        <div className='absolute left-2 top-8 z-10 rounded bg-background/80 px-2 py-1 text-xs text-muted-foreground'>
          加载更早行情…
        </div>
      ) : null}
      <div ref={ref} className='h-[720px] w-full' />
    </div>
  )
}
