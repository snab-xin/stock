import { createFileRoute } from '@tanstack/react-router'
import { parseStockListSearch, type StockListSearch } from '@/features/stocks/list-search'

export type InspectSearch = StockListSearch & {
  code?: string
}

function parseInspectSearch(raw: Record<string, unknown>): InspectSearch {
  const list = parseStockListSearch(raw)
  const code =
    typeof raw.code === 'string' && raw.code.trim() ? raw.code.trim() : undefined
  return { ...list, code }
}

export const Route = createFileRoute('/_authenticated/')({
  validateSearch: (search: Record<string, unknown>) => parseInspectSearch(search),
  component: () => null,
})
