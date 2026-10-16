import { createFileRoute } from '@tanstack/react-router'
import { StockList } from '@/features/stocks'
import { parseStockListSearch } from '@/features/stocks/list-search'

export const Route = createFileRoute('/_authenticated/stocks/')({
  validateSearch: (search: Record<string, unknown>) => parseStockListSearch(search),
  component: StockList,
})
