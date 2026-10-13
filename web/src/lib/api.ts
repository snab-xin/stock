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

export type BarsResponse = {
  instrument: Instrument
  qfq: boolean
  bars: Bar[]
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
  tushare_url: string
  http_addr: string
  has_tushare_token: boolean
}

export type StartIngestRequest = {
  job: 'calendar' | 'instruments' | 'daily' | 'sync'
  start?: string
  end?: string
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
  const res = await fetch(url)
  if (!res.ok) throw new Error(await parseError(res))
  return (await res.json()) as T
}

async function postJSON<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(await parseError(res))
  return (await res.json()) as T
}

export function searchInstruments(q: string) {
  return getJSON<Instrument[]>(`/api/instruments?q=${encodeURIComponent(q)}`)
}

export function fetchBars(code: string, qfq: boolean) {
  const qs = new URLSearchParams({ code, qfq: qfq ? '1' : '0' })
  return getJSON<BarsResponse>(`/api/bars?${qs}`)
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
