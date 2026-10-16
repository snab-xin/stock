import { memo, useEffect, useMemo, useState, useTransition } from 'react'
import { type SortingState } from '@tanstack/react-table'
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
import { fetchCBList, type CBListItem } from '@/lib/api'
import { CBTable } from './components/cb-table'
import {
  resolveCBListSearch,
  toCBListSearchParams,
  type CBListSearch,
} from './list-search'

const limit = 50

export const CBList = memo(function CBList({
  search,
  onSearchChange,
}: {
  search: CBListSearch
  onSearchChange: (search: CBListSearch) => void
}) {
  const resolved = resolveCBListSearch(search)
  const [, startTransition] = useTransition()

  const [query, setQuery] = useState(resolved.q || '')
  const [items, setItems] = useState<CBListItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const sorting = useMemo<SortingState>(
    () => [{ id: resolved.sortBy, desc: resolved.sortDir === 'desc' }],
    [resolved.sortBy, resolved.sortDir],
  )
  const listSearch = useMemo(
    () =>
      toCBListSearchParams({
        q: resolved.q,
        listedOnly: resolved.listedOnly,
        sortBy: resolved.sortBy,
        sortDir: resolved.sortDir,
        offset: resolved.offset,
      }),
    [resolved.q, resolved.listedOnly, resolved.sortBy, resolved.sortDir, resolved.offset],
  )

  function patchSearch(next: {
    q?: string
    listedOnly?: boolean
    sortBy?: string
    sortDir?: 'asc' | 'desc'
    offset?: number
  }) {
    const nextSearch = toCBListSearchParams({
      q: next.q !== undefined ? next.q : resolved.q,
      listedOnly: next.listedOnly !== undefined ? next.listedOnly : resolved.listedOnly,
      sortBy: next.sortBy !== undefined ? next.sortBy : resolved.sortBy,
      sortDir: next.sortDir !== undefined ? next.sortDir : resolved.sortDir,
      offset: next.offset !== undefined ? next.offset : resolved.offset,
    })
    startTransition(() => onSearchChange(nextSearch))
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
      void fetchCBList({
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
    // eslint-disable-next-line react-hooks/exhaustive-deps -- patch via onSearchChange; load on resolved
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
          <h1 className='text-2xl font-bold tracking-tight'>可转债</h1>
          <p className='text-sm text-muted-foreground'>
            列表展示最新日线收盘与涨跌（涨红跌绿）。点击表头排序会请求后端。需先同步「转债基础」和「转债行情」。
            {quoteHint ? ` 行情日期示例：${quoteHint}` : ''}
          </p>
        </div>

        <div className='flex flex-wrap items-center gap-2'>
          <Input
            className='w-72'
            value={query}
            placeholder='代码 / 转债简称 / 正股'
            onChange={(e) => setQuery(e.target.value)}
          />
          <Button
            variant={resolved.listedOnly ? 'default' : 'outline'}
            size='sm'
            onClick={() => patchSearch({ listedOnly: true, offset: 0 })}
          >
            未摘牌
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
            {error}。若无行情请执行 migrate 后：ingest cb-basic，再 ingest cb-daily -date
            YYYYMMDD。
          </p>
        ) : null}

        <Card className='gap-0 py-0'>
          <CardHeader className='px-4 py-3'>
            <CardTitle>列表</CardTitle>
          </CardHeader>
          <CardContent className='px-0'>
            <CBTable
              data={items}
              loading={loading}
              listSearch={listSearch}
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

            <div className='flex items-center gap-2 px-4 py-3'>
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
})
