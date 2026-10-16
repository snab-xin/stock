import { useEffect, useState } from 'react'
import { type SortingState } from '@tanstack/react-table'
import { getRouteApi } from '@tanstack/react-router'
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
import { fetchStockList, type StockListItem } from '@/lib/api'
import { StockTable } from './components/stock-table'
import { resolveStockListSearch, toStockListSearchParams } from './list-search'

const routeApi = getRouteApi('/_authenticated/stocks/')
const limit = 50

export function StockList() {
  const search = routeApi.useSearch()
  const navigate = routeApi.useNavigate()
  const resolved = resolveStockListSearch(search)

  const [query, setQuery] = useState(resolved.q || '')
  const [items, setItems] = useState<StockListItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const sorting: SortingState = [
    { id: resolved.sortBy, desc: resolved.sortDir === 'desc' },
  ]

  function patchSearch(next: {
    q?: string
    listedOnly?: boolean
    sortBy?: string
    sortDir?: 'asc' | 'desc'
    offset?: number
  }) {
    void navigate({
      search: toStockListSearchParams({
        q: next.q !== undefined ? next.q : resolved.q,
        listedOnly: next.listedOnly !== undefined ? next.listedOnly : resolved.listedOnly,
        sortBy: next.sortBy !== undefined ? next.sortBy : resolved.sortBy,
        sortDir: next.sortDir !== undefined ? next.sortDir : resolved.sortDir,
        offset: next.offset !== undefined ? next.offset : resolved.offset,
      }),
      replace: true,
    })
  }

  useEffect(() => {
    setQuery(resolved.q || '')
  }, [resolved.q])

  useEffect(() => {
    const t = window.setTimeout(() => {
      const q = query.trim()
      if (q !== (resolved.q || '')) {
        patchSearch({ q, offset: 0 })
        return
      }
      setLoading(true)
      setError('')
      void fetchStockList({
        q: resolved.q,
        limit,
        offset: resolved.offset,
        listedOnly: resolved.listedOnly,
        sortBy: resolved.sortBy,
        sortDir: resolved.sortDir,
      })
        .then((res) => {
          setItems(res.items)
          setTotal(res.total)
        })
        .catch((e) => {
          setError(e instanceof Error ? e.message : '加载失败')
          setItems([])
          setTotal(0)
        })
        .finally(() => setLoading(false))
    }, 250)
    return () => window.clearTimeout(t)
    // eslint-disable-next-line react-hooks/exhaustive-deps -- patch via navigate; load on resolved
  }, [query, resolved.q, resolved.listedOnly, resolved.sortBy, resolved.sortDir, resolved.offset])

  const quoteHint = items.find((it) => it.quote_date)?.quote_date

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
          <h1 className='text-2xl font-bold tracking-tight'>股票</h1>
          <p className='text-sm text-muted-foreground'>
            列表展示最新日线收盘与涨跌（涨红跌绿）。点击代码进入行情质检。需先同步证券列表与日线。
            {quoteHint ? ` 行情日期示例：${quoteHint}` : ''}
          </p>
        </div>

        <div className='flex flex-wrap items-center gap-2'>
          <Input
            className='w-72'
            value={query}
            placeholder='代码 / 简称'
            onChange={(e) => setQuery(e.target.value)}
          />
          <Button
            variant={resolved.listedOnly ? 'default' : 'outline'}
            size='sm'
            onClick={() => patchSearch({ listedOnly: true, offset: 0 })}
          >
            上市中
          </Button>
          <Button
            variant={!resolved.listedOnly ? 'default' : 'outline'}
            size='sm'
            onClick={() => patchSearch({ listedOnly: false, offset: 0 })}
          >
            全部
          </Button>
          <Badge variant='outline'>{loading ? '加载中…' : `共 ${total} 只`}</Badge>
        </div>

        {error ? (
          <p className='text-sm text-destructive'>
            {error}。若无行情请先：ingest instruments，再 ingest daily。
          </p>
        ) : null}

        <Card>
          <CardHeader>
            <CardTitle>列表</CardTitle>
          </CardHeader>
          <CardContent>
            <StockTable
              data={items}
              loading={loading}
              listSearch={toStockListSearchParams({
                q: resolved.q,
                listedOnly: resolved.listedOnly,
                sortBy: resolved.sortBy,
                sortDir: resolved.sortDir,
                offset: resolved.offset,
              })}
              sorting={sorting}
              onSortingChange={(updater) => {
                const next = typeof updater === 'function' ? updater(sorting) : updater
                const sort = next[0]
                patchSearch({
                  sortBy: sort?.id || 'pct_chg',
                  sortDir: sort?.desc === false ? 'asc' : 'desc',
                  offset: 0,
                })
              }}
            />

            <div className='mt-4 flex items-center gap-2'>
              <Button
                variant='outline'
                size='sm'
                disabled={loading || resolved.offset <= 0}
                onClick={() =>
                  patchSearch({ offset: Math.max(0, resolved.offset - limit) })
                }
              >
                上一页
              </Button>
              <Button
                variant='outline'
                size='sm'
                disabled={loading || resolved.offset + limit >= total}
                onClick={() => patchSearch({ offset: resolved.offset + limit })}
              >
                下一页
              </Button>
              <span className='text-xs text-muted-foreground'>
                {total === 0 ? 0 : resolved.offset + 1}–
                {Math.min(resolved.offset + limit, total)} / {total}
              </span>
            </div>
          </CardContent>
        </Card>
      </Main>
    </>
  )
}
