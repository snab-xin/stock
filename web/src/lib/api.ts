export type Instrument = {
  ts_code: string
  symbol: string
  name: string
  exchange: string
  is_st: boolean
  status: string
}

export type Bar = {
  trade_date: string
  open: number
  high: number
  low: number
  close: number
  volume: number
  amount: number
  is_limit_up: boolean | null
  is_limit_down: boolean | null
  is_paused: boolean
  high_limit: number | null
  low_limit: number | null
}

export type BarIndicators = {
  rsi: (number | null)[]
  macd_dif: (number | null)[]
  macd_dea: (number | null)[]
  macd_hist: (number | null)[]
}

export type BarPeriod = 'day' | 'week' | 'month'

export type BarsResponse = {
  instrument: Instrument
  qfq: boolean
  period?: BarPeriod
  bars: Bar[]
  indicators?: BarIndicators
  has_more?: boolean
}

export type GapsResponse = {
  instrument: Instrument
  missing: string[]
  bar_count: number
  open_days?: number
}

export type IngestRun = {
  id: number
  job_name: string
  status: string
  message: string
  rows_affected: number
  started_at: string
  finished_at?: string
  params?: Record<string, unknown>
}

export type IngestJobStatus = {
  running: boolean
  job?: string
  start?: string
  end?: string
  code?: string
  message?: string
  error?: string
  started_at?: string
  finished_at?: string
}

export type AppConfig = {
  tushare_rpm: number
  tushare_apis?: Record<string, number>
  tushare_url: string
  http_addr: string
  has_tushare_token: boolean
  backtest?: BacktestDefaults
  log?: { level: string; format: string }
}

export type BacktestDefaults = {
  init_cash: number
  fast: number
  slow: number
  commission_rate: number
  commission_min: number
  stamp_tax_rate: number
  slippage: number
  lot_size: number
}

export type BacktestRequest = {
  code: string
  start?: string
  end?: string
  fast?: number
  slow?: number
  cash?: number
}

export type BacktestTrade = {
  date: string
  side: string
  price: number
  shares: number
  amount: number
  commission: number
  stamp_tax: number
  slippage: number
  cash_after: number
  reason: string
}

export type EquityPoint = {
  date: string
  cash: number
  shares: number
  close: number
  equity: number
  target: number
}

export type BacktestResult = {
  code: string
  name?: string
  strategy: string
  fast: number
  slow: number
  init_cash: number
  end_equity: number
  total_return: number
  max_drawdown: number
  ann_return: number
  buy_hold_return: number
  trade_count: number
  win_rate: number
  commission: number
  stamp_tax: number
  slippage_cost: number
  bar_count: number
  start: string
  end: string
  trades: BacktestTrade[]
  equity: EquityPoint[]
  limitations: string[]
}

export type StartIngestRequest = {
  job:
    | 'calendar'
    | 'instruments'
    | 'daily'
    | 'sync'
    | 'cb-basic'
    | 'cb-holders'
    | 'cb-daily'
    | 'index-basic'
    | 'index-classify'
    | 'index-daily'
    | 'sw-daily'
    | 'etf-basic'
    | 'etf-daily'
    | 'etf-adj'
    | 'etf-share'
    | 'etf-sh-cons'
    | 'etf-sz-cons'
  start?: string
  end?: string
  period?: string
  date?: string
  code?: string
}

async function parseError(res: Response): Promise<string> {
  let msg = res.statusText
  try {
    const body = (await res.json()) as { error?: string }
    if (body.error) msg = body.error
  } catch {
    /* ignore */
  }
  return msg
}

async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url, { credentials: 'include' })
  if (!res.ok) throw new Error(await parseError(res))
  return (await res.json()) as T
}

async function postJSON<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(await parseError(res))
  return (await res.json()) as T
}

async function putJSON<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'PUT',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(await parseError(res))
  return (await res.json()) as T
}

async function deleteJSON<T>(url: string): Promise<T> {
  const res = await fetch(url, { method: 'DELETE', credentials: 'include' })
  if (!res.ok) throw new Error(await parseError(res))
  return (await res.json()) as T
}

export type AuthUser = {
  id: number
  username: string
  role: string
  status: string
}

export type AuthStatus = {
  needs_setup: boolean
  user: AuthUser | null
}

export function fetchAuthStatus() {
  return getJSON<AuthStatus>('/api/auth/status')
}

export function registerUser(username: string, password: string) {
  return postJSON<{ user: AuthUser }>('/api/auth/register', { username, password })
}

export function loginUser(username: string, password: string) {
  return postJSON<{ user: AuthUser }>('/api/auth/login', { username, password })
}

export function logoutUser() {
  return postJSON<{ ok: boolean }>('/api/auth/logout', {})
}

export type NoteTargetType = '' | 'stock' | 'cb' | 'index'

export type NoteItem = {
  id: number
  ts_code: string
  target_type: NoteTargetType
  trade_date: string
  summary: string
  content?: string
  updated_at?: string
}

export type NoteMark = {
  id: number
  trade_date: string
  summary: string
}

export function fetchNotes(params?: { ts_code?: string; target_type?: string }) {
  const qs = new URLSearchParams()
  if (params?.ts_code) qs.set('ts_code', params.ts_code)
  if (params?.target_type) qs.set('target_type', params.target_type)
  const suffix = qs.toString() ? `?${qs}` : ''
  return getJSON<{ items: NoteItem[] }>(`/api/notes${suffix}`)
}

export function fetchNoteMarks(tsCode: string) {
  return getJSON<{ items: NoteMark[] }>(
    `/api/notes/marks?ts_code=${encodeURIComponent(tsCode)}`,
  )
}

export function fetchNote(id: number) {
  return getJSON<NoteItem>(`/api/notes/${id}`)
}

export function createNote(body: {
  ts_code?: string
  target_type?: string
  trade_date: string
  summary: string
  content: string
}) {
  return postJSON<NoteItem>('/api/notes', body)
}

export function updateNote(
  id: number,
  body: {
    ts_code?: string
    target_type?: string
    trade_date: string
    summary: string
    content: string
  },
) {
  return putJSON<NoteItem>(`/api/notes/${id}`, body)
}

export function deleteNote(id: number) {
  return deleteJSON<{ ok: boolean }>(`/api/notes/${id}`)
}

export function searchInstruments(q: string) {
  return getJSON<Instrument[]>(`/api/instruments?q=${encodeURIComponent(q)}`)
}

export type StockListItem = {
  ts_code: string
  symbol: string
  name: string
  exchange: string
  is_st: boolean
  status: string
  list_date?: string
  delist_date?: string
  quote_date?: string
  pre_close?: number | null
  close?: number | null
  change?: number | null
  pct_chg?: number | null
  volume?: number | null
  amount?: number | null
  is_limit_up?: boolean | null
  is_limit_down?: boolean | null
  is_paused?: boolean | null
}

export type StockListResponse = {
  total: number
  items: StockListItem[]
}

export function fetchStockList(params: {
  q?: string
  limit?: number
  offset?: number
  listedOnly?: boolean
  sortBy?: string
  sortDir?: 'asc' | 'desc'
}) {
  const qs = new URLSearchParams()
  if (params.q) qs.set('q', params.q)
  if (params.limit != null) qs.set('limit', String(params.limit))
  if (params.offset != null) qs.set('offset', String(params.offset))
  if (params.listedOnly) qs.set('listed_only', '1')
  if (params.sortBy) qs.set('sort_by', params.sortBy)
  if (params.sortDir) qs.set('sort_dir', params.sortDir)
  return getJSON<StockListResponse>(`/api/stocks?${qs}`)
}

export type StockNeighborsResponse = {
  prev: string
  next: string
  index: number
  total: number
}

export function fetchStockNeighbors(
  code: string,
  params: {
    q?: string
    listedOnly?: boolean
    sortBy?: string
    sortDir?: 'asc' | 'desc'
  },
) {
  const qs = new URLSearchParams()
  if (params.q) qs.set('q', params.q)
  if (params.listedOnly) qs.set('listed_only', '1')
  if (params.sortBy) qs.set('sort_by', params.sortBy)
  if (params.sortDir) qs.set('sort_dir', params.sortDir)
  const suffix = qs.toString() ? `?${qs}` : ''
  return getJSON<StockNeighborsResponse>(
    `/api/stocks/${encodeURIComponent(code)}/neighbors${suffix}`,
  )
}

export function fetchBars(
  code: string,
  qfq: boolean,
  opts?: { before?: string; limit?: number; period?: BarPeriod },
) {
  const qs = new URLSearchParams({ code, qfq: qfq ? '1' : '0' })
  if (opts?.before) qs.set('before', opts.before)
  if (opts?.limit != null) qs.set('limit', String(opts.limit))
  if (opts?.period && opts.period !== 'day') qs.set('period', opts.period)
  return getJSON<BarsResponse>(`/api/bars?${qs}`)
}

export function fetchIndicators(closes: number[]) {
  return postJSON<BarIndicators>('/api/indicators', { closes })
}

export function fetchGaps(code: string) {
  return getJSON<GapsResponse>(`/api/gaps?code=${encodeURIComponent(code)}`)
}

export function fetchIngestRuns() {
  return getJSON<IngestRun[]>('/api/ingest-runs?limit=30')
}

export function fetchIngestStatus() {
  return getJSON<IngestJobStatus>('/api/ingest/status')
}

export function startIngest(req: StartIngestRequest) {
  return postJSON<IngestJobStatus>('/api/ingest', req)
}

export function fetchAppConfig() {
  return getJSON<AppConfig>('/api/config')
}

export function runBacktest(req: BacktestRequest) {
  return postJSON<BacktestResult>('/api/backtest', req)
}

export type CBListItem = {
  ts_code: string
  bond_short_name: string
  bond_full_name: string
  cb_type: string
  stk_code: string
  stk_short_name: string
  exchange: string
  list_date?: string
  delist_date?: string
  conv_price?: number | null
  coupon_rate?: number | null
  issue_rating?: string
  newest_rating?: string
  issue_size?: number | null
  remain_size?: number | null
  float_size?: number | null
  top10_ratio?: number | null
  quote_date?: string
  pre_close?: number | null
  close?: number | null
  change?: number | null
  pct_chg?: number | null
  volume?: number | null
  amount?: number | null
  cb_over_rate?: number | null
  maturity_call_price?: number | null
}

export type CBListResponse = {
  total: number
  items: CBListItem[]
}

export type CBDetail = {
  ts_code: string
  bond_full_name: string
  bond_short_name: string
  cb_code: string
  cb_type: string
  stk_code: string
  stk_short_name: string
  maturity?: number | null
  par?: number | null
  issue_price?: number | null
  issue_size?: number | null
  remain_size?: number | null
  value_date?: string
  maturity_date?: string
  rate_type?: string
  coupon_rate?: number | null
  add_rate?: number | null
  pay_per_year?: number | null
  list_date?: string
  delist_date?: string
  exchange: string
  conv_start_date?: string
  conv_end_date?: string
  conv_stop_date?: string
  first_conv_price?: number | null
  conv_price?: number | null
  rate_clause?: string
  put_clause?: string
  maturity_call_price?: number | null
  call_clause?: string
  reset_clause?: string
  conv_clause?: string
  guarantor?: string
  guarantee_type?: string
  issue_rating?: string
  newest_rating?: string
  rating_comp?: string
}

export type CBHolder = {
  end_date: string
  holder_rank: number
  holder_name: string
  hold_amount: number | null
  hold_ratio: number | null
}

export type CBHolderSummary = {
  end_date: string
  top10_ratio: number
  top10_amount: number
  holder_count: number
}

export type CBDetailResponse = {
  basic: CBDetail
  holder_periods: string[]
  selected_period: string
  holders: CBHolder[]
  holder_summaries: CBHolderSummary[]
}

export function fetchCBList(params: {
  q?: string
  limit?: number
  offset?: number
  listedOnly?: boolean
  sortBy?: string
  sortDir?: 'asc' | 'desc'
}) {
  const qs = new URLSearchParams()
  if (params.q) qs.set('q', params.q)
  if (params.limit != null) qs.set('limit', String(params.limit))
  if (params.offset != null) qs.set('offset', String(params.offset))
  if (params.listedOnly) qs.set('listed_only', '1')
  if (params.sortBy) qs.set('sort_by', params.sortBy)
  if (params.sortDir) qs.set('sort_dir', params.sortDir)
  return getJSON<CBListResponse>(`/api/cb?${qs}`)
}

export function fetchCBDetail(code: string, period?: string) {
  const qs = new URLSearchParams()
  if (period) qs.set('period', period)
  const suffix = qs.toString() ? `?${qs}` : ''
  return getJSON<CBDetailResponse>(`/api/cb/${encodeURIComponent(code)}${suffix}`)
}

export type CBNeighborsResponse = {
  prev: string
  next: string
  index: number
  total: number
}

export function fetchCBNeighbors(
  code: string,
  params: {
    q?: string
    listedOnly?: boolean
    sortBy?: string
    sortDir?: 'asc' | 'desc'
  },
) {
  const qs = new URLSearchParams()
  if (params.q) qs.set('q', params.q)
  if (params.listedOnly) qs.set('listed_only', '1')
  if (params.sortBy) qs.set('sort_by', params.sortBy)
  if (params.sortDir) qs.set('sort_dir', params.sortDir)
  const suffix = qs.toString() ? `?${qs}` : ''
  return getJSON<CBNeighborsResponse>(
    `/api/cb/${encodeURIComponent(code)}/neighbors${suffix}`,
  )
}

export type CBReplaceHoldersItem = {
  holder_rank?: number
  holder_name: string
  hold_amount: number
  hold_ratio: number
}

export type CBReplaceHoldersResponse = {
  period: string
  count: number
  holders: CBHolder[]
}

export function saveCBHolders(
  code: string,
  body: { period: string; holders: CBReplaceHoldersItem[] },
) {
  return putJSON<CBReplaceHoldersResponse>(
    `/api/cb/${encodeURIComponent(code)}/holders`,
    body,
  )
}
