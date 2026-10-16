import {
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { useNavigate, useRouterState } from '@tanstack/react-router'
import { X } from 'lucide-react'
import { CBDetail } from '@/features/cb/detail'
import { parseCBListSearch, type CBListSearch } from '@/features/cb/list-search'
import {
  defaultInspectCode,
  Inspect,
  type InspectSearch,
} from '@/features/inspect'
import { parseStockListSearch } from '@/features/stocks/list-search'
import { cn } from '@/lib/utils'

const MAX_TABS = 8

type InspectTab = {
  id: string
  kind: 'inspect'
  code: string
  title: string
  search: InspectSearch
  lastActive: number
}

type CBTab = {
  id: string
  kind: 'cb'
  code: string
  title: string
  search: CBListSearch
  lastActive: number
}

type WorkspaceTab = InspectTab | CBTab

function inspectSearchOf(raw: Record<string, unknown>): InspectSearch {
  const list = parseStockListSearch(raw)
  const code =
    typeof raw.code === 'string' && raw.code.trim() ? raw.code.trim() : undefined
  return { ...list, code }
}

function describe(pathname: string, raw: Record<string, unknown>): WorkspaceTab | null {
  if (pathname === '/') {
    const search = inspectSearchOf(raw)
    const code = search.code?.trim() || defaultInspectCode
    return {
      id: `inspect:${code}`,
      kind: 'inspect',
      code,
      title: code,
      search: { ...search, code },
      lastActive: 0,
    }
  }
  const matched = pathname.match(/^\/cb\/([^/]+)$/)
  if (!matched) return null
  const code = decodeURIComponent(matched[1])
  if (!code || code === '$code') return null
  return {
    id: `cb:${code}`,
    kind: 'cb',
    code,
    title: code,
    search: parseCBListSearch(raw),
    lastActive: 0,
  }
}

function sameSearch(a: InspectSearch | CBListSearch, b: InspectSearch | CBListSearch) {
  return JSON.stringify(a) === JSON.stringify(b)
}

function upsert(prev: WorkspaceTab[], incoming: WorkspaceTab) {
  const now = Date.now()
  const idx = prev.findIndex((tab) => tab.id === incoming.id)
  if (idx >= 0) {
    const current = prev[idx]
    if (sameSearch(current.search, incoming.search) && current.code === incoming.code) {
      return prev
    }
    const next = prev.slice()
    next[idx] = {
      ...current,
      code: incoming.code,
      search: incoming.search,
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

export function WorkspaceHost({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const pathname = useRouterState({ select: (state) => state.location.pathname })
  const searchStr = useRouterState({
    select: (state) => JSON.stringify(state.location.search ?? {}),
  })
  const [tabs, setTabs] = useState<WorkspaceTab[]>([])
  const [activeId, setActiveId] = useState<string | null>(null)
  const closingId = useRef<string | null>(null)
  const scrollerRef = useRef<HTMLDivElement>(null)
  const scrollPos = useRef(new Map<string, number>())
  const prevScrollKey = useRef<string | null>(null)

  const described = describe(pathname, JSON.parse(searchStr) as Record<string, unknown>)
  const onWorkspace = described !== null
  const scrollKey = onWorkspace && described ? described.id : pathname

  const syncedKey = useRef('')

  useLayoutEffect(() => {
    if (!described) {
      setActiveId(null)
      return
    }
    if (closingId.current === described.id) return
    closingId.current = null
    setActiveId(described.id)
    const key = `${pathname}?${searchStr}`
    if (syncedKey.current === key) return
    syncedKey.current = key
    setTabs((prev) => {
      const next = upsert(prev, described)
      return next.map((tab) =>
        tab.id === described.id ? { ...tab, lastActive: Date.now() } : tab,
      )
    })
  }, [pathname, searchStr])

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

  const focusTab = useCallback(
    (tab: WorkspaceTab) => {
      if (tab.kind === 'inspect') {
        void navigate({ to: '/', search: tab.search })
        return
      }
      void navigate({
        to: '/cb/$code',
        params: { code: tab.code },
        search: tab.search,
      })
    },
    [navigate],
  )

  const patchInspectSearch = useCallback(
    (id: string, search: InspectSearch) => {
      const code = search.code?.trim() || defaultInspectCode
      const next = { ...search, code }
      if (`inspect:${code}` !== id) {
        void navigate({ to: '/', search: next })
        return
      }
      setTabs((prev) =>
        prev.map((tab) => (tab.id === id && tab.kind === 'inspect' ? { ...tab, search: next, code } : tab)),
      )
      void navigate({ to: '/', search: next, replace: true })
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
      void navigate({ to: tab.kind === 'cb' ? '/cb' : '/stocks' })
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
                  {tab.kind === 'cb' ? '转债 ' : '质检 '}
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
            <div key={tab.id} className={visible ? undefined : 'hidden'}>
              {tab.kind === 'inspect' ? (
                <Inspect
                  search={tab.search}
                  onSearchChange={(search) => patchInspectSearch(tab.id, search)}
                  onTitle={(title) => setTitle(tab.id, title)}
                />
              ) : (
                <CBDetail
                  code={tab.code}
                  search={tab.search}
                  onTitle={(title) => setTitle(tab.id, title)}
                />
              )}
            </div>
          )
        })}
        <div className={onWorkspace ? 'hidden' : undefined}>{children}</div>
      </div>
    </div>
  )
}
