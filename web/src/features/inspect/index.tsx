import { useEffect, useMemo, useRef, useState } from 'react'
import { Link } from '@tanstack/react-router'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  KlineChart,
  type ZoomRange,
} from '@/components/charts/kline-chart'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { MarkdownView } from '@/components/markdown-editor'
import { Input } from '@/components/ui/input'
import {
  fetchBars,
  fetchGaps,
  fetchIndicators,
  fetchNote,
  fetchNoteMarks,
  fetchStockNeighbors,
  searchInstruments,
  type BarPeriod,
  type BarsResponse,
  type GapsResponse,
  type Instrument,
  type NoteItem,
  type NoteMark,
  type StockNeighborsResponse,
} from '@/lib/api'
import {
  resolveStockListSearch,
  toStockListSearchParams,
  type StockListSearch,
} from '@/features/stocks/list-search'

export type InspectSearch = StockListSearch & {
  code?: string
}

export const defaultInspectCode = '000001.SZ'
const pageSize = 250
const defaultZoom: ZoomRange = { start: 60, end: 100 }

const periodLabel: Record<BarPeriod, string> = {
  day: '日线',
  week: '周线',
  month: '月线',
}

export function Inspect({
  search,
  onSearchChange,
  onTitle,
}: {
  search: InspectSearch
  onSearchChange: (search: InspectSearch) => void
  onTitle?: (title: string) => void
}) {
  const urlCode = search.code?.trim() || defaultInspectCode
  const listQuery = resolveStockListSearch(search)
  const listSearch = toStockListSearchParams({
    q: listQuery.q,
    listedOnly: listQuery.listedOnly,
    sortBy: listQuery.sortBy,
    sortDir: listQuery.sortDir,
  })

  const [query, setQuery] = useState(urlCode)
  const [hints, setHints] = useState<Instrument[]>([])
  const [code, setCode] = useState(urlCode)
  const [qfq, setQfq] = useState(true)
  const [period, setPeriod] = useState<BarPeriod>('day')
  const [loading, setLoading] = useState(false)
  const [loadingOlder, setLoadingOlder] = useState(false)
  const [hasMoreOlder, setHasMoreOlder] = useState(true)
  const [zoom, setZoom] = useState<ZoomRange>(defaultZoom)
  const [error, setError] = useState('')
  const [data, setData] = useState<BarsResponse | null>(null)
  const [gaps, setGaps] = useState<GapsResponse | null>(null)
  const [neighbors, setNeighbors] = useState<StockNeighborsResponse | null>(null)
  const [noteMarks, setNoteMarks] = useState<NoteMark[]>([])
  const [noteOpen, setNoteOpen] = useState(false)
  const [noteDetail, setNoteDetail] = useState<NoteItem | null>(null)
  const loadingOlderRef = useRef(false)
  const dataRef = useRef(data)
  const periodRef = useRef(period)
  dataRef.current = data
  periodRef.current = period

  function searchWithCode(nextCode: string) {
    return {
      ...listSearch,
      code: nextCode,
    }
  }

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

  async function load(
    nextCode: string,
    nextQfq: boolean,
    nextPeriod: BarPeriod,
    syncUrl = false,
  ) {
    const sameTab = !syncUrl || nextCode.trim() === urlCode
    if (sameTab) {
      setLoading(true)
      setZoom(defaultZoom)
      setHasMoreOlder(true)
    }
    setError('')
    try {
      const barReq = fetchBars(nextCode, nextQfq, {
        limit: pageSize,
        period: nextPeriod,
      })
      const gapReq =
        nextPeriod === 'day'
          ? fetchGaps(nextCode)
          : Promise.resolve<GapsResponse | null>(null)
      const [barRes, gapRes] = await Promise.all([barReq, gapReq])
      const resolved = barRes.instrument.ts_code || nextCode
      if (syncUrl && resolved !== urlCode) {
        onSearchChange(searchWithCode(resolved))
        setQuery(urlCode)
        return
      }
      setData(barRes)
      setGaps(gapRes)
      setHasMoreOlder(barRes.has_more !== false && barRes.bars.length >= pageSize)
      setCode(resolved)
      setQuery(resolved)
      setQfq(nextQfq)
      setPeriod(nextPeriod)
      if (syncUrl) onSearchChange(searchWithCode(resolved))
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载失败')
      setData(null)
      setGaps(null)
      setHasMoreOlder(false)
    } finally {
      setLoading(false)
    }
  }

  async function loadOlder() {
    const cur = dataRef.current
    const curPeriod = periodRef.current
    if (!cur?.bars.length || loadingOlderRef.current || !hasMoreOlder) return
    const before = cur.bars[0]?.trade_date
    if (!before) return

    loadingOlderRef.current = true
    setLoadingOlder(true)
    try {
      const older = await fetchBars(code, qfq, {
        before,
        limit: pageSize,
        period: curPeriod,
      })
      if (older.bars.length === 0) {
        setHasMoreOlder(false)
        return
      }
      if (older.bars.length < pageSize || older.has_more === false) {
        setHasMoreOlder(false)
      }

      const seen = new Set(cur.bars.map((b) => b.trade_date))
      const prepend = older.bars.filter((b) => !seen.has(b.trade_date))
      if (prepend.length === 0) {
        setHasMoreOlder(false)
        return
      }

      const oldLen = cur.bars.length
      const merged = [...prepend, ...cur.bars]
      const indicators = await fetchIndicators(merged.map((b) => b.close))
      const added = prepend.length

      setZoom((z) => ({
        start: ((added + (oldLen * z.start) / 100) / merged.length) * 100,
        end: ((added + (oldLen * z.end) / 100) / merged.length) * 100,
      }))
      setData({
        ...cur,
        bars: merged,
        indicators,
        has_more: older.has_more,
        period: curPeriod,
      })
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载更早行情失败')
    } finally {
      loadingOlderRef.current = false
      setLoadingOlder(false)
    }
  }

  useEffect(() => {
    if (period !== 'day' || !code) {
      setNoteMarks([])
      return
    }
    let cancelled = false
    void fetchNoteMarks(code)
      .then((res) => {
        if (!cancelled) setNoteMarks(res.items ?? [])
      })
      .catch(() => {
        if (!cancelled) setNoteMarks([])
      })
    return () => {
      cancelled = true
    }
  }, [code, period])

  async function openNote(id: number) {
    try {
      const note = await fetchNote(id)
      setNoteDetail(note)
      setNoteOpen(true)
    } catch (e) {
      setError(e instanceof Error ? e.message : '笔记加载失败')
    }
  }

  useEffect(() => {
    setQuery(urlCode)
    setCode(urlCode)
    void load(urlCode, qfq, period, false)
    // eslint-disable-next-line react-hooks/exhaustive-deps -- only reload when URL code changes
  }, [urlCode])

  useEffect(() => {
    let cancelled = false
    void fetchStockNeighbors(urlCode, {
      q: listQuery.q,
      listedOnly: listQuery.listedOnly,
      sortBy: listQuery.sortBy,
      sortDir: listQuery.sortDir,
    })
      .then((res) => {
        if (!cancelled) setNeighbors(res)
      })
      .catch(() => {
        if (!cancelled) setNeighbors(null)
      })
    return () => {
      cancelled = true
    }
  }, [
    urlCode,
    listQuery.q,
    listQuery.listedOnly,
    listQuery.sortBy,
    listQuery.sortDir,
  ])

  const title = useMemo(() => {
    if (!data?.instrument) return code
    const it = data.instrument
    return [it.ts_code, it.name].filter(Boolean).join(' ')
  }, [code, data])

  useEffect(() => {
    onTitle?.(title)
  }, [onTitle, title])

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
        <div className='sticky top-16 z-40 -mx-4 border-b bg-background/95 px-4 py-2.5 backdrop-blur supports-backdrop-filter:bg-background/80'>
          <div className='flex flex-wrap items-center gap-3'>
            <Link
              to='/stocks'
              search={listSearch}
              className='text-sm text-muted-foreground hover:text-foreground'
            >
              ← 股票列表
            </Link>
            <div className='flex flex-wrap items-center gap-1.5'>
              {neighbors?.prev ? (
                <Button size='sm' variant='outline' asChild>
                  <Link to='/' search={searchWithCode(neighbors.prev)}>
                    上一个
                  </Link>
                </Button>
              ) : (
                <Button size='sm' variant='outline' disabled>
                  上一个
                </Button>
              )}
              {neighbors?.next ? (
                <Button size='sm' variant='outline' asChild>
                  <Link to='/' search={searchWithCode(neighbors.next)}>
                    下一个
                  </Link>
                </Button>
              ) : (
                <Button size='sm' variant='outline' disabled>
                  下一个
                </Button>
              )}
              {neighbors && neighbors.index > 0 ? (
                <span className='text-xs text-muted-foreground'>
                  {neighbors.index} / {neighbors.total}
                </span>
              ) : null}
            </div>
            {data?.instrument ? (
              <h1 className='text-lg font-bold tracking-tight sm:text-xl'>
                {title}
              </h1>
            ) : (
              <h1 className='text-lg font-bold tracking-tight'>行情质检</h1>
            )}
            {loading ? <Badge variant='outline'>加载中…</Badge> : null}
          </div>
        </div>

        <p className='text-sm text-muted-foreground' style={{ display: 'none' }}>
          核对复权、涨跌停、停牌和缺日。周/月线由日线在库内聚合。右滑可加载更早 K
          线。上一个/下一个按股票列表当前筛选与排序。
        </p>

        <div className='relative flex flex-wrap items-center gap-2'>
          <Input
            className='w-64'
            value={query}
            placeholder='代码 / 简称，如 000001 或 平安'
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') void load(query, qfq, period, true)
            }}
          />
          <Button
            onClick={() => void load(query, qfq, period, true)}
            disabled={loading}
          >
            {loading ? '加载中…' : '查看'}
          </Button>
          {(['day', 'week', 'month'] as const).map((p) => (
            <Button
              key={p}
              variant={period === p ? 'default' : 'outline'}
              size='sm'
              onClick={() => void load(code, qfq, p, false)}
              disabled={loading}
            >
              {periodLabel[p]}
            </Button>
          ))}
          <Button
            variant={qfq ? 'default' : 'outline'}
            onClick={() => void load(code, true, period, false)}
          >
            前复权
          </Button>
          <Button
            variant={!qfq ? 'default' : 'outline'}
            onClick={() => void load(code, false, period, false)}
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
                    void load(it.ts_code, qfq, period, true)
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
            {error}。请先 migrate，并启动：go run ./cmd/ashare web
          </p>
        ) : null}

        {data && (
          <>
            <div className='flex flex-wrap items-center gap-2 text-sm'>
              <Badge variant='secondary'>{periodLabel[period]}</Badge>
              <Badge variant='secondary'>{qfq ? '前复权' : '未复权'}</Badge>
              <Badge variant='outline'>{data.bars.length} 根 K 线</Badge>
              {!hasMoreOlder ? (
                <Badge variant='outline'>已到最早</Badge>
              ) : null}
              {period === 'day' && limitUp > 0 ? (
                <Badge variant='warning'>涨停 {limitUp}</Badge>
              ) : null}
              {period === 'day' && limitDown > 0 ? (
                <Badge variant='success'>跌停 {limitDown}</Badge>
              ) : null}
              {paused > 0 ? <Badge variant='outline'>停牌 {paused}</Badge> : null}
              {period === 'day' ? (
                gaps && (gaps.missing?.length ?? 0) > 0 ? (
                  <Badge variant='warning'>缺日 {gaps.missing.length}</Badge>
                ) : (
                  <Badge variant='success'>日历对齐</Badge>
                )
              ) : null}
            </div>

            <Card>
              <CardContent className='pt-4'>
                {data.bars.length === 0 ? (
                  <p className='py-16 text-center text-sm text-muted-foreground'>
                    没有日线。先执行 ingest daily。
                  </p>
                ) : (
                  <KlineChart
                    bars={data.bars}
                    indicators={data.indicators}
                    zoom={zoom}
                    onZoomChange={setZoom}
                    onLoadOlder={() => void loadOlder()}
                    hasMoreOlder={hasMoreOlder}
                    loadingOlder={loadingOlder}
                    noteMarks={period === 'day' ? noteMarks : []}
                    onNoteClick={(id) => void openNote(id)}
                  />
                )}
              </CardContent>
            </Card>

            {period === 'day' ? (
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
            ) : null}

            <Dialog open={noteOpen} onOpenChange={setNoteOpen}>
              <DialogContent className='max-h-[80vh] overflow-y-auto sm:max-w-2xl'>
                <DialogHeader>
                  <DialogTitle>
                    {noteDetail
                      ? `${noteDetail.trade_date} ${noteDetail.summary}`
                      : '笔记'}
                  </DialogTitle>
                </DialogHeader>
                {noteDetail?.ts_code ? (
                  <p className='text-xs text-muted-foreground'>
                    {noteDetail.target_type} {noteDetail.ts_code}
                  </p>
                ) : null}
                <MarkdownView source={noteDetail?.content ?? ''} />
              </DialogContent>
            </Dialog>
          </>
        )}
      </Main>
    </>
  )
}
