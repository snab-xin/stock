import {
  memo,
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { useNavigate, useRouterState } from '@tanstack/react-router'
import { X } from 'lucide-react'
import { Backtest } from '@/features/backtest'
import { CBList } from '@/features/cb'
import { CBDetail } from '@/features/cb/detail'
import { parseCBListSearch, type CBListSearch } from '@/features/cb/list-search'
import { Ingest } from '@/features/ingest'
import {
  defaultInspectCode,
  Inspect,
  type InspectSearch,
} from '@/features/inspect'
import { NotesPage } from '@/features/notes'
import { StockList } from '@/features/stocks'
import {
  parseStockListSearch,
  type StockListSearch,
} from '@/features/stocks/list-search'
import { cn } from '@/lib/utils'

const MAX_TABS = 12

type InspectTab = {
  id: string
  kind: 'inspect'
  title: string
  code: string
  search: InspectSearch
  lastActive: number
}

type StocksTab = {
  id: string
  kind: 'stocks'
  title: string
  search: StockListSearch
  lastActive: number
}

type CBListTab = {
  id: string
  kind: 'cb-list'
  title: string
  search: CBListSearch
  lastActive: number
}

type CBTab = {
  id: string
  kind: 'cb'
  title: string
  code: string
  search: CBListSearch
  lastActive: number
}

type SimpleTab = {
  id: string
  kind: 'ingest' | 'backtest' | 'notes'
  title: string
  lastActive: number
}

type WorkspaceTab = InspectTab | StocksTab | CBListTab | CBTab | SimpleTab

function inspectSearchOf(raw: Record<string, unknown>): InspectSearch {
  const list = parseStockListSearch(raw)
  const code =
    typeof raw.code === 'string' && raw.code.trim() ? raw.code.trim() : undefined
  return { ...list, code }
}

function normalizePath(pathname: string) {
  if (pathname.length > 1 && pathname.endsWith('/')) return pathname.slice(0, -1)
  return pathname
}

function describe(pathname: string, raw: Record<string, unknown>): WorkspaceTab | null {
  const path = normalizePath(pathname)
  if (path === '/') {
    const search = inspectSearchOf(raw)
    const code = search.code?.trim() || defaultInspectCode
    return {
      id: `inspect:${code}`,
      kind: 'inspect',
      title: code,
      code,
      search: { ...search, code },
      lastActive: 0,
    }
  }
  if (path === '/stocks') {
    return {
      id: 'stocks',
      kind: 'stocks',
      title: '股票',
      search: parseStockListSearch(raw),
      lastActive: 0,
    }
  }
  if (path === '/cb') {
    return {
      id: 'cb-list',
      kind: 'cb-list',
      title: '可转债',
      search: parseCBListSearch(raw),
      lastActive: 0,
    }
  }
  const matched = path.match(/^\/cb\/([^/]+)$/)
  if (matched) {
    const code = decodeURIComponent(matched[1])
    if (!code || code === '$code') return null
    return {
      id: `cb:${code}`,
      kind: 'cb',
      title: code,
      code,
      search: parseCBListSearch(raw),
      lastActive: 0,
    }
  }
  if (path === '/ingest') {
    return { id: 'ingest', kind: 'ingest', title: '采集任务', lastActive: 0 }
  }
  if (path === '/backtest') {
    return { id: 'backtest', kind: 'backtest', title: '最小回测', lastActive: 0 }
  }
  if (path === '/notes') {
    return { id: 'notes', kind: 'notes', title: '笔记', lastActive: 0 }
  }
  return null
}

function tabSearchEqual(a: WorkspaceTab, b: WorkspaceTab) {
  if (a.kind !== b.kind || a.id !== b.id) return false
  if (a.kind === 'ingest' || a.kind === 'backtest' || a.kind === 'notes') return true
  if (a.kind === 'inspect' && b.kind === 'inspect') {
    return a.code === b.code && JSON.stringify(a.search) === JSON.stringify(b.search)
  }
  if ('search' in a && 'search' in b) {
    return JSON.stringify(a.search) === JSON.stringify(b.search)
  }
  return true
}

function upsert(prev: WorkspaceTab[], incoming: WorkspaceTab) {
  const now = Date.now()
  const idx = prev.findIndex((tab) => tab.id === incoming.id)
  if (idx >= 0) {
    const current = prev[idx]
    if (tabSearchEqual(current, incoming)) return prev
    const next = prev.slice()
    next[idx] = {
      ...current,
      ...incoming,
      title: current.title,
      lastActive: now,
    }
    return next
  }
  let next = [...prev, { ...incoming, lastActive: now }]
  while (next.length > MAX_TABS) {
    let drop = -1
    for (let i = 0; i < next.length; i++) {
      if (next[i].id === incoming.id) continue
      if (drop < 0 || next[i].lastActive < next[drop].lastActive) drop = i
    }
    if (drop < 0) break
    next = next.filter((_, i) => i !== drop)
  }
  return next
}

function tabPrefix(tab: WorkspaceTab) {
  if (tab.kind === 'inspect') return '质检 '
  if (tab.kind === 'cb') return '转债 '
  return ''
}

const InspectPane = memo(function InspectPane({
  tab,
  onSearchChange,
  onTitle,
}: {
  tab: InspectTab
  onSearchChange: (id: string, search: InspectSearch) => void
  onTitle: (id: string, title: string) => void
}) {
  return (
    <Inspect
      search={tab.search}
      onSearchChange={(search) => onSearchChange(tab.id, search)}
      onTitle={(title) => onTitle(tab.id, title)}
    />
  )
})

const CBDetailPane = memo(function CBDetailPane({
  tab,
  onTitle,
}: {
  tab: CBTab
  onTitle: (id: string, title: string) => void
}) {
  return (
    <CBDetail
      code={tab.code}
      search={tab.search}
      onTitle={(title) => onTitle(tab.id, title)}
    />
  )
})

const IdlePane = memo(function IdlePane({
  kind,
}: {
  kind: 'ingest' | 'backtest' | 'notes'
}) {
  if (kind === 'ingest') return <Ingest />
  if (kind === 'backtest') return <Backtest />
  return <NotesPage />
})

function touch(prev: WorkspaceTab[], id: string) {
  const now = Date.now()
  return prev.map((tab) => (tab.id === id ? { ...tab, lastActive: now } : tab))
}

/** 详情页里切换上一个/下一个时，替换当前 tab，不另开。 */
function replaceDetail(prev: WorkspaceTab[], fromId: string, incoming: WorkspaceTab) {
  const idx = prev.findIndex((tab) => tab.id === fromId)
  if (idx < 0) return touch(upsert(prev, incoming), incoming.id)
  const next = prev.slice()
  next[idx] = { ...incoming, lastActive: Date.now() }
  return next
}

export function WorkspaceHost({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const pathname = useRouterState({ select: (state) => state.location.pathname })
  const searchStr = useRouterState({
    select: (state) => JSON.stringify(state.location.search ?? {}),
  })
  const [tabs, setTabs] = useState<WorkspaceTab[]>([])
  const [activeId, setActiveId] = useState<string | null>(null)
  const closingId = useRef<string | null>(null)
  const tabsRef = useRef(tabs)
  const scrollerRef = useRef<HTMLDivElement>(null)
  const scrollPos = useRef(new Map<string, number>())
  const prevScrollKey = useRef<string | null>(null)
  const syncedKey = useRef('')

  tabsRef.current = tabs

  const described = describe(pathname, JSON.parse(searchStr) as Record<string, unknown>)
  const onWorkspace = described !== null
  const scrollKey = onWorkspace && described ? described.id : pathname

  const focusTab = useCallback(
    (tab: WorkspaceTab, replace = false) => {
      if (tab.kind === 'inspect') {
        void navigate({ to: '/', search: tab.search, replace })
        return
      }
      if (tab.kind === 'stocks') {
        void navigate({ to: '/stocks', search: tab.search, replace })
        return
      }
      if (tab.kind === 'cb-list') {
        void navigate({ to: '/cb', search: tab.search, replace })
        return
      }
      if (tab.kind === 'cb') {
        void navigate({
          to: '/cb/$code',
          params: { code: tab.code },
          search: tab.search,
          replace,
        })
        return
      }
      if (tab.kind === 'ingest') {
        void navigate({ to: '/ingest', replace })
        return
      }
      if (tab.kind === 'backtest') {
        void navigate({ to: '/backtest', replace })
        return
      }
      void navigate({ to: '/notes', replace })
    },
    [navigate],
  )

  useLayoutEffect(() => {
    const next = describe(pathname, JSON.parse(searchStr) as Record<string, unknown>)
    if (!next) {
      setActiveId(null)
      return
    }
    if (closingId.current === next.id) return
    closingId.current = null

    const key = `${pathname}?${searchStr}`
    const from = tabsRef.current.find((tab) => tab.id === activeId)
    const existing = tabsRef.current.find((tab) => tab.id === next.id)
    const continueDetail =
      from != null &&
      !existing &&
      (from.kind === 'cb' || from.kind === 'inspect') &&
      from.kind === next.kind &&
      from.id !== next.id

    if (continueDetail) {
      const fromId = from.id
      setActiveId(next.id)
      syncedKey.current = key
      setTabs((prev) => replaceDetail(prev, fromId, next))
      return
    }

    const switchingToOpenTab = Boolean(existing) && activeId !== next.id

    setActiveId(next.id)

    if (switchingToOpenTab && existing) {
      setTabs((prev) => touch(prev, next.id))
      if (!tabSearchEqual(existing, next)) {
        syncedKey.current = ''
        focusTab(existing, true)
        return
      }
      syncedKey.current = key
      return
    }

    if (syncedKey.current === key) return
    syncedKey.current = key
    setTabs((prev) => touch(upsert(prev, next), next.id))
  }, [pathname, searchStr, activeId, focusTab])

  useLayoutEffect(() => {
    const el = scrollerRef.current
    if (!el) return
    if (prevScrollKey.current && prevScrollKey.current !== scrollKey) {
      scrollPos.current.set(prevScrollKey.current, el.scrollTop)
    }
    el.scrollTop = scrollPos.current.get(scrollKey) ?? 0
    prevScrollKey.current = scrollKey
  }, [scrollKey])

  useLayoutEffect(() => {
    const id = window.requestAnimationFrame(() => {
      window.dispatchEvent(new Event('resize'))
    })
    return () => window.cancelAnimationFrame(id)
  }, [activeId, onWorkspace])

  const setTitle = useCallback((id: string, title: string) => {
    setTabs((prev) => {
      const current = prev.find((tab) => tab.id === id)
      if (!current || current.title === title) return prev
      return prev.map((tab) => (tab.id === id ? { ...tab, title } : tab))
    })
  }, [])

  const patchInspectSearch = useCallback(
    (id: string, search: InspectSearch) => {
      const code = search.code?.trim() || defaultInspectCode
      const next = { ...search, code }
      if (`inspect:${code}` !== id) {
        void navigate({ to: '/', search: next })
        return
      }
      setTabs((prev) =>
        prev.map((tab) =>
          tab.id === id && tab.kind === 'inspect' ? { ...tab, search: next, code } : tab,
        ),
      )
      void navigate({ to: '/', search: next, replace: true })
    },
    [navigate],
  )

  const patchStocksSearch = useCallback(
    (search: StockListSearch) => {
      setTabs((prev) =>
        prev.map((tab) => (tab.kind === 'stocks' ? { ...tab, search } : tab)),
      )
      void navigate({ to: '/stocks', search, replace: true })
    },
    [navigate],
  )

  const patchCBListSearch = useCallback(
    (search: CBListSearch) => {
      setTabs((prev) =>
        prev.map((tab) => (tab.kind === 'cb-list' ? { ...tab, search } : tab)),
      )
      void navigate({ to: '/cb', search, replace: true })
    },
    [navigate],
  )

  function closeTab(tab: WorkspaceTab) {
    const rest = tabs.filter((item) => item.id !== tab.id)
    closingId.current = tab.id
    setTabs(rest)
    if (described?.id !== tab.id) return
    const fallback = rest[rest.length - 1]
    if (!fallback) {
      if (tab.kind === 'inspect' || tab.kind === 'stocks') {
        void navigate({ to: '/cb' })
        return
      }
      void navigate({ to: '/stocks' })
      return
    }
    setActiveId(fallback.id)
    focusTab(fallback)
  }

  return (
    <div className='flex min-h-0 flex-1 flex-col'>
      {tabs.length > 0 ? (
        <div
          className='flex h-9 shrink-0 items-center gap-1 overflow-x-auto border-b bg-muted/40 px-2'
          role='tablist'
          aria-label='已打开的页面'
        >
          {tabs.map((tab) => {
            const selected = onWorkspace && tab.id === (activeId ?? described?.id)
            return (
              <div
                key={tab.id}
                className={cn(
                  'flex h-7 max-w-52 shrink-0 items-center rounded-md',
                  selected ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground',
                )}
              >
                <button
                  type='button'
                  role='tab'
                  aria-selected={selected}
                  className='truncate px-2 text-xs'
                  onClick={() => focusTab(tab)}
                >
                  {tabPrefix(tab)}
                  {tab.title}
                </button>
                <button
                  type='button'
                  className='me-1 rounded-sm p-0.5 hover:bg-accent'
                  aria-label={`关闭 ${tab.title}`}
                  onClick={() => closeTab(tab)}
                >
                  <X className='size-3' />
                </button>
              </div>
            )
          })}
        </div>
      ) : null}
      <div ref={scrollerRef} data-workspace-scroll className='min-h-0 flex-1 overflow-auto'>
        {tabs.map((tab) => {
          const visible = onWorkspace && tab.id === (activeId ?? described?.id)
          return (
            <div key={tab.id} className={visible ? undefined : 'hidden'} inert={!visible}>
              {tab.kind === 'inspect' ? (
                <InspectPane tab={tab} onSearchChange={patchInspectSearch} onTitle={setTitle} />
              ) : null}
              {tab.kind === 'stocks' ? (
                <StockList search={tab.search} onSearchChange={patchStocksSearch} />
              ) : null}
              {tab.kind === 'cb-list' ? (
                <CBList search={tab.search} onSearchChange={patchCBListSearch} />
              ) : null}
              {tab.kind === 'cb' ? <CBDetailPane tab={tab} onTitle={setTitle} /> : null}
              {tab.kind === 'ingest' || tab.kind === 'backtest' || tab.kind === 'notes' ? (
                <IdlePane kind={tab.kind} />
              ) : null}
            </div>
          )
        })}
        <div className={onWorkspace ? 'hidden' : undefined}>{children}</div>
      </div>
    </div>
  )
}
