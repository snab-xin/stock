import * as echarts from 'echarts'
import { useEffect, useRef } from 'react'
import type { EquityPoint } from '@/lib/api'

function iso(d: string) {
  return d.slice(0, 10)
}

export function EquityChart({ points }: { points: EquityPoint[] }) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const chart = echarts.init(el)
    const dates = points.map((p) => iso(p.date))
    const equity = points.map((p) => p.equity)
    chart.setOption({
      animation: false,
      tooltip: { trigger: 'axis' },
      grid: { left: 56, right: 16, top: 24, bottom: 40 },
      xAxis: { type: 'category', data: dates, axisLabel: { fontSize: 11 } },
      yAxis: { scale: true, splitLine: { lineStyle: { opacity: 0.25 } } },
      dataZoom: [
        { type: 'inside', start: 0, end: 100 },
        { type: 'slider', start: 0, end: 100, height: 18, bottom: 8 },
      ],
      series: [
        {
          name: '净值',
          type: 'line',
          data: equity,
          showSymbol: false,
          lineStyle: { width: 1.5, color: '#2563eb' },
        },
      ],
    })
    const onResize = () => chart.resize()
    window.addEventListener('resize', onResize)
    return () => {
      window.removeEventListener('resize', onResize)
      chart.dispose()
    }
  }, [points])

  return <div ref={ref} className="h-[360px] w-full" />
}
