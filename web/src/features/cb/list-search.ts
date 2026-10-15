/** 列表与详情共用的检索条件（同步到 URL search）。 */
export type CBListSearch = {
  q?: string
  listed_only?: boolean
  sort_by?: string
  sort_dir?: 'asc' | 'desc'
  offset?: number
}

export const defaultCBListSearch = {
  listed_only: true,
  sort_by: 'pct_chg',
  sort_dir: 'desc' as const,
  offset: 0,
}

export function parseCBListSearch(raw: Record<string, unknown>): CBListSearch {
  const q = typeof raw.q === 'string' && raw.q.trim() ? raw.q : undefined
  let listed_only: boolean | undefined
  if (raw.listed_only === false || raw.listed_only === '0' || raw.listed_only === 0) {
    listed_only = false
  } else if (raw.listed_only === true || raw.listed_only === '1' || raw.listed_only === 1) {
    listed_only = true
  }
  const sort_by = typeof raw.sort_by === 'string' && raw.sort_by ? raw.sort_by : undefined
  const sort_dir = raw.sort_dir === 'asc' || raw.sort_dir === 'desc' ? raw.sort_dir : undefined
  let offset: number | undefined
  if (typeof raw.offset === 'number' && Number.isFinite(raw.offset) && raw.offset > 0) {
    offset = Math.floor(raw.offset)
  } else if (typeof raw.offset === 'string' && raw.offset) {
    const n = Number(raw.offset)
    if (Number.isFinite(n) && n > 0) offset = Math.floor(n)
  }
  return { q, listed_only, sort_by, sort_dir, offset }
}

/** 解析后填入默认值，供 API / UI 使用。 */
export function resolveCBListSearch(s: CBListSearch) {
  return {
    q: s.q?.trim() || undefined,
    listedOnly: s.listed_only ?? defaultCBListSearch.listed_only,
    sortBy: s.sort_by || defaultCBListSearch.sort_by,
    sortDir: (s.sort_dir || defaultCBListSearch.sort_dir) as 'asc' | 'desc',
    offset: s.offset && s.offset > 0 ? s.offset : 0,
  }
}

/** 写入 URL 时省略默认值，保持链接干净。 */
export function toCBListSearchParams(input: {
  q?: string
  listedOnly: boolean
  sortBy?: string
  sortDir?: 'asc' | 'desc'
  offset?: number
}): CBListSearch {
  const out: CBListSearch = {}
  if (input.q?.trim()) out.q = input.q.trim()
  if (!input.listedOnly) out.listed_only = false
  else out.listed_only = true
  if (input.sortBy && input.sortBy !== defaultCBListSearch.sort_by) out.sort_by = input.sortBy
  else if (input.sortBy) out.sort_by = input.sortBy
  if (input.sortDir && input.sortDir !== defaultCBListSearch.sort_dir) out.sort_dir = input.sortDir
  else if (input.sortDir) out.sort_dir = input.sortDir
  // 始终带上 sort，方便详情页邻居接口
  if (!out.sort_by) out.sort_by = defaultCBListSearch.sort_by
  if (!out.sort_dir) out.sort_dir = defaultCBListSearch.sort_dir
  if (input.offset && input.offset > 0) out.offset = input.offset
  return out
}
