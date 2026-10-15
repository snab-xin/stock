import { useEffect, useMemo, useState } from 'react'
import { KlineChart } from '@/components/charts/kline-chart'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  fetchBars,
  fetchGaps,
  searchInstruments,
  type BarsResponse,
  type GapsResponse,
  type Instrument,
} from '@/lib/api'

export function Inspect() {
  const [query, setQuery] = useState('000001.SZ')
  const [hints, setHints] = useState<Instrument[]>([])
  const [code, setCode] = useState('000001.SZ')
  const [qfq, setQfq] = useState(true)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [data, setData] = useState<BarsResponse | null>(null)
  const [gaps, setGaps] = useState<GapsResponse | null>(null)

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

  async function load(nextCode: string, nextQfq: boolean) {
    setLoading(true)
    setError('')
    try {
      const [barRes, gapRes] = await Promise.all([
        fetchBars(nextCode, nextQfq),
        fetchGaps(nextCode),
      ])
      setData(barRes)
      setGaps(gapRes)
      setCode(barRes.instrument.ts_code || nextCode)
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载失败')
      setData(null)
      setGaps(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load('000001.SZ', true)
  }, [])

  const title = useMemo(() => {
    if (!data?.instrument) return code
    const it = data.instrument
    return [it.ts_code, it.name].filter(Boolean).join(' ')
  }, [code, data])

  const limitUp = data?.bars.filter((b) => b.is_limit_up).length ?? 0
  const limitDown = data?.bars.filter((b) => b.is_limit_down).length ?? 0
  const paused = data?.bars.filter((b) => b.is_paused).length ?? 0

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
          <h1 className='text-2xl font-bold tracking-tight'>行情质检</h1>
          <p className='text-sm text-muted-foreground'>
            核对复权、涨跌停、停牌和缺日，而不是做交易终端。
          </p>
        </div>

        <div className='relative flex flex-wrap items-center gap-2'>
          <Input
            className='w-64'
            value={query}
            placeholder='代码 / 简称，如 000001 或 平安'
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void load(query, qfq)
            }}
          />
          <Button onClick={() => void load(query, qfq)} disabled={loading}>
            {loading ? '加载中…' : '查看'}
          </Button>
          <Button
            variant={qfq ? 'default' : 'outline'}
            onClick={() => {
              setQfq(true)
              void load(code, true)
            }}
          >
            前复权
          </Button>
          <Button
            variant={!qfq ? 'default' : 'outline'}
            onClick={() => {
              setQfq(false)
              void load(code, false)
            }}
          >
            未复权
          </Button>
          {hints.length > 0 && (
            <div className='absolute top-11 z-10 w-80 overflow-hidden rounded-md border bg-popover shadow-md'>
              {hints.map((it) => (
                <button
                  key={it.ts_code}
                  className='flex w-full items-center justify-between px-3 py-2 text-left text-sm hover:bg-accent'
                  onClick={() => {
                    setQuery(it.ts_code)
                    setHints([])
                    void load(it.ts_code, qfq)
                  }}
                >
                  <span>
                    {it.ts_code} {it.name}
                  </span>
                  {it.is_st ? <Badge variant='warning'>ST</Badge> : null}
                </button>
              ))}
            </div>
          )}
        </div>

        {error ? (
          <p className='text-sm text-destructive'>
            {error}。请先启动 PostgreSQL 和：go run ./cmd/ashare web
          </p>
        ) : null}

        {data && (
          <>
            <div className='flex flex-wrap items-center gap-2 text-sm'>
              <span className='font-medium'>{title}</span>
              <Badge variant='secondary'>{qfq ? '前复权' : '未复权'}</Badge>
              <Badge variant='outline'>{data.bars.length} 根 K 线</Badge>
              {limitUp > 0 ? <Badge variant='warning'>涨停 {limitUp}</Badge> : null}
              {limitDown > 0 ? <Badge variant='success'>跌停 {limitDown}</Badge> : null}
              {paused > 0 ? <Badge variant='outline'>停牌 {paused}</Badge> : null}
              {gaps && (gaps.missing?.length ?? 0) > 0 ? (
                <Badge variant='warning'>缺日 {gaps.missing.length}</Badge>
              ) : (
                <Badge variant='success'>日历对齐</Badge>
              )}
            </div>

            <Card>
              <CardContent className='pt-4'>
                {data.bars.length === 0 ? (
                  <p className='py-16 text-center text-sm text-muted-foreground'>
                    没有日线。先执行 ingest daily。
                  </p>
                ) : (
                  <KlineChart bars={data.bars} />
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>缺日（开市但无 K 线）</CardTitle>
              </CardHeader>
              <CardContent>
                {!gaps || (gaps.missing?.length ?? 0) === 0 ? (
                  <p className='text-sm text-muted-foreground'>
                    当前区间没有缺日，或还没有交易日历。
                  </p>
                ) : (
                  <div className='flex flex-wrap gap-2'>
                    {gaps.missing.map((d) => (
                      <Badge key={d} variant='outline'>
                        {d}
                      </Badge>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </>
        )}
      </Main>
    </>
  )
}
