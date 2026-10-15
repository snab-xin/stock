import { createFileRoute } from '@tanstack/react-router'
import { CBList } from '@/features/cb'
import { parseCBListSearch } from '@/features/cb/list-search'

export const Route = createFileRoute('/_authenticated/cb/')({
  validateSearch: (search: Record<string, unknown>) => parseCBListSearch(search),
  component: CBList,
})
