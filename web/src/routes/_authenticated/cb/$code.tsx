import { createFileRoute } from '@tanstack/react-router'
import { CBDetail } from '@/features/cb/detail'
import { parseCBListSearch } from '@/features/cb/list-search'

export const Route = createFileRoute('/_authenticated/cb/$code')({
  validateSearch: (search: Record<string, unknown>) => parseCBListSearch(search),
  component: CBDetail,
})
