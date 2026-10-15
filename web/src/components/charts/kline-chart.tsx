import * as echarts from 'echarts'
import { useEffect, useRef } from 'react'
import type { Bar } from '@/lib/api'

type Props = { bars: Bar[] }

export function KlineChart({ bars }: Props) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const chart = echarts.init(el)
    const dates = bars.map((b) => b.trade_date)
    const ohlc = bars.map((b) => [b.open, b.close, b.low, b.high])
    const vols = bars.map((b) => b.volume)
    const markPoints = bars.flatMap((b) => {
      const items: { name: string; coord: [string, number]; value: string; itemStyle: { color: string } }[] = []
      if (b.is_limit_up) {
        items.push({
          name: '涨停',
          coord: [b.trade_date, b.high],
          value: '涨',
          itemStyle: { color: '#dc2626' },
        })
      }
      if (b.is_limit_down) {
        items.push({
          name: '跌停',
          coord: [b.trade_date, b.low],
          value: '跌',
          itemStyle: { color: '#16a34a' },
        })
      }
      if (b.is_paused) {
        items.push({
          name: '停牌',
          coord: [b.trade_date, b.close],
          value: '停',
          itemStyle: { color: '#a16207' },
        })
      }
      return items
    })

    chart.setOption({
      animation: false,
      legend: { data: ['K线', '成交量'] },
      tooltip: { trigger: 'axis', axisPointer: { type: 'cross' } },
      axisPointer: { link: [{ xAxisIndex: 'all' }] },
      grid: [
        { left: 56, right: 16, top: 32, height: '58%' },
        { left: 56, right: 16, top: '76%', height: '16%' },
      ],
      xAxis: [
        { type: 'category', data: dates, gridIndex: 0, axisLabel: { show: false }, axisTick: { show: false } },
        { type: 'category', data: dates, gridIndex: 1, axisLabel: { fontSize: 11 } },
      ],
      yAxis: [
        { scale: true, gridIndex: 0, splitLine: { lineStyle: { opacity: 0.25 } } },
        { gridIndex: 1, splitNumber: 2, axisLabel: { show: false }, splitLine: { show: false } },
      ],
      dataZoom: [
        { type: 'inside', xAxisIndex: [0, 1], start: 60, end: 100 },
        { type: 'slider', xAxisIndex: [0, 1], start: 60, end: 100, height: 18, bottom: 8 },
      ],
      series: [
        {
          name: 'K线',
          type: 'candlestick',
          data: ohlc,
          itemStyle: {
            color: '#dc2626',
            color0: '#16a34a',
            borderColor: '#dc2626',
            borderColor0: '#16a34a',
          },
          markPoint: {
            symbol: 'pin',
            symbolSize: 28,
            data: markPoints,
          },
        },
        {
          name: '成交量',
          type: 'bar',
          xAxisIndex: 1,
          yAxisIndex: 1,
          data: vols.map((v, i) => ({
            value: v,
            itemStyle: {
              color: bars[i].close >= bars[i].open ? '#dc2626' : '#16a34a',
            },
          })),
        },
      ],
    })

    const onResize = () => chart.resize()
    window.addEventListener('resize', onResize)
    return () => {
      window.removeEventListener('resize', onResize)
      chart.dispose()
    }
  }, [bars])

  return <div ref={ref} className="h-[520px] w-full" />
}

