import { useCallback, useEffect, useState } from 'react'
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
import {
  fetchAppConfig,
  fetchIngestRuns,
  fetchIngestStatus,
  startIngest,
  type AppConfig,
  type IngestJobStatus,
  type IngestRun,
  type StartIngestRequest,
} from '@/lib/api'
import { formatISODateTimeStr } from '@/lib/date'

function statusVariant(status: string) {
  if (status === 'success') return 'success' as const
  if (status === 'failed') return 'warning' as const
  return 'secondary' as const
}

function formatParams(params: Record<string, unknown> | undefined) {
  if (!params || Object.keys(params).length === 0) return ''
  return JSON.stringify(params)
}

function todayYYYYMMDD() {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}${m}${day}`
}

export function Ingest() {
  const [rows, setRows] = useState<IngestRun[]>([])
  const [cfg, setCfg] = useState<AppConfig | null>(null)
  const [jobStatus, setJobStatus] = useState<IngestJobStatus | null>(null)
  const [error, setError] = useState('')
  const [formError, setFormError] = useState('')
  const [loading, setLoading] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const [job, setJob] = useState<StartIngestRequest['job']>('sync')
  const [start, setStart] = useState('20260101')
  const [end, setEnd] = useState(todayYYYYMMDD())
  const [period, setPeriod] = useState('20250630')
  const [tradeDate, setTradeDate] = useState(todayYYYYMMDD())
  const [code, setCode] = useState('000001.SZ')

  const loadRuns = useCallback(async () => {
    setLoading(true)
    setError('')
    try {
      setRows(await fetchIngestRuns())
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }, [])

  const refreshStatus = useCallback(async () => {
    try {
      const st = await fetchIngestStatus()
      setJobStatus(st)
      return st
    } catch {
      return null
    }
  }, [])

  useEffect(() => {
    void loadRuns()
    void fetchAppConfig()
      .then(setCfg)
      .catch(() => setCfg(null))
    void refreshStatus()
  }, [loadRuns, refreshStatus])

  useEffect(() => {
    if (!jobStatus?.running) return
    const t = window.setInterval(() => {
      void refreshStatus().then((st) => {
        if (st && !st.running) void loadRuns()
      })
    }, 2000)
    return () => window.clearInterval(t)
  }, [jobStatus?.running, refreshStatus, loadRuns])

  async function onSubmit() {
    setFormError('')
    if (job === 'cb-holders') {
      const hasPeriod = Boolean(period.trim())
      const hasRange = Boolean(start.trim() && end.trim())
      if (!hasPeriod && !hasRange) {
        setFormError('请填写报告期，或同时填写开始/结束日期（YYYYMMDD）')
        return
      }
    } else if (job === 'cb-daily') {
      if (!tradeDate.trim() && (!start.trim() || !end.trim())) {
        setFormError('请填写交易日 date，或同时填写开始/结束日期')
        return
      }
    } else if (needRange && (!start.trim() || !end.trim())) {
      setFormError('请填写开始/结束日期（YYYYMMDD）')
      return
    }
    setSubmitting(true)
    try {
      const body: StartIngestRequest = { job }
      if (needRange) {
        body.start = start.trim()
        body.end = end.trim()
      }
      if (job === 'cb-holders') {
        // 与 Tushare 一致：period / start_date / end_date / ts_code；code 可空=库内全部
        if (period.trim()) body.period = period.trim()
        if (start.trim() && end.trim()) {
          body.start = start.trim()
          body.end = end.trim()
        } else if (start.trim() || end.trim()) {
          setFormError('开始日期与结束日期需同时填写')
          setSubmitting(false)
          return
        }
      }
      if (job === 'cb-daily') {
        if (tradeDate.trim()) body.date = tradeDate.trim()
        if (start.trim() && end.trim()) {
          body.start = start.trim()
          body.end = end.trim()
        }
      }
      if (code.trim()) body.code = code.trim()
      if (
        (job === 'daily' ||
          job === 'sync' ||
          job === 'index-daily' ||
          job === 'sw-daily' ||
          job === 'etf-daily' ||
          job === 'etf-adj' ||
          job === 'etf-share' ||
          job === 'etf-sh-cons' ||
          job === 'etf-sz-cons') &&
        !body.code
      ) {
        const ok = window.confirm('未填写代码将按交易日拉全市场日线，可能很慢且消耗大量积分。继续？')
        if (!ok) {
          setSubmitting(false)
          return
        }
      }
      if (job === 'cb-holders' && !body.code) {
        const ok = window.confirm(
          '未填写可转债代码时，将按当前库中 cb_basic 的全部转债同步十大持有人（约需 5000 积分）。继续？',
        )
        if (!ok) {
          setSubmitting(false)
          return
        }
      }
      const st = await startIngest(body)
      setJobStatus(st)
      void loadRuns()
    } catch (e) {
      setFormError(e instanceof Error ? e.message : '启动失败')
    } finally {
      setSubmitting(false)
    }
  }

  const needRange =
    job !== 'instruments' &&
    job !== 'cb-basic' &&
    job !== 'cb-holders' &&
    job !== 'cb-daily' &&
    job !== 'index-basic' &&
    job !== 'index-classify' &&
    job !== 'etf-basic'
  const isCBHolders = job === 'cb-holders'
  const isCBDaily = job === 'cb-daily'
  const showCode =
    job === 'daily' ||
    job === 'sync' ||
    job === 'cb-holders' ||
    job === 'cb-daily' ||
    job === 'index-daily' ||
    job === 'sw-daily' ||
    job === 'etf-daily' ||
    job === 'etf-adj' ||
    job === 'etf-share' ||
    job === 'etf-sh-cons' ||
    job === 'etf-sz-cons'
  const running = Boolean(jobStatus?.running)

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4'>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">采集任务</h1>
          <p className="text-sm text-muted-foreground">
            在界面同步行情。限流：
            {cfg
              ? `默认 ${cfg.tushare_rpm} 次/分钟（约每 ${(60 / cfg.tushare_rpm).toFixed(2)}s 一次）`
              : '…'}
            {cfg?.tushare_apis && Object.keys(cfg.tushare_apis).length > 0
              ? `；覆盖 ${Object.entries(cfg.tushare_apis)
                  .map(([k, v]) => `${k}=${v}`)
                  .join(', ')}`
              : ''}
            {cfg && !cfg.has_tushare_token ? '（未配置 Token）' : ''}
          </p>
        </div>
        <Button variant="outline" onClick={() => void loadRuns()} disabled={loading || running}>
          刷新
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>同步行情</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex flex-wrap gap-2">
            {(
              [
                ['sync', '一键同步'],
                ['calendar', '交易日历'],
                ['instruments', '证券列表'],
                ['daily', '日线'],
                ['cb-basic', '转债基础'],
                ['cb-daily', '转债行情'],
                ['cb-holders', '转债十大持有人'],
                ['index-basic', '指数列表'],
                ['index-classify', '申万行业'],
                ['index-daily', '指数日线'],
                ['sw-daily', '申万日线'],
                ['etf-basic', 'ETF列表'],
                ['etf-daily', 'ETF日线'],
                ['etf-adj', 'ETF复权'],
                ['etf-share', 'ETF份额'],
                ['etf-sh-cons', 'ETF持仓沪'],
                ['etf-sz-cons', 'ETF持仓深'],
              ] as const
            ).map(([value, label]) => (
              <Button
                key={value}
                variant={job === value ? 'default' : 'outline'}
                size="sm"
                disabled={running}
                onClick={() => {
                  setJob(value)
                  if (value === 'cb-holders') {
                    setStart('')
                    setEnd('')
                    setCode('')
                  } else if (value === 'cb-daily') {
                    setTradeDate(todayYYYYMMDD())
                    setStart('')
                    setEnd('')
                    setCode('')
                  } else if (value === 'cb-basic') {
                    setCode('')
                  } else if (!start) {
                    setStart('20260101')
                    setEnd(todayYYYYMMDD())
                  }
                }}
              >
                {label}
              </Button>
            ))}
          </div>

          <div className="flex flex-wrap items-end gap-3">
            {isCBHolders && (
              <label className="space-y-1 text-sm">
                <span className="text-muted-foreground">报告期</span>
                <Input
                  className="w-36"
                  value={period}
                  disabled={running}
                  placeholder="如 20250630"
                  onChange={(e) => setPeriod(e.target.value)}
                />
              </label>
            )}
            {isCBDaily && (
              <label className="space-y-1 text-sm">
                <span className="text-muted-foreground">交易日</span>
                <Input
                  className="w-36"
                  value={tradeDate}
                  disabled={running}
                  placeholder="YYYYMMDD"
                  onChange={(e) => setTradeDate(e.target.value)}
                />
              </label>
            )}
            {(needRange || isCBHolders || isCBDaily) && (
              <>
                <label className="space-y-1 text-sm">
                  <span className="text-muted-foreground">开始日期</span>
                  <Input
                    className="w-36"
                    value={start}
                    disabled={running}
                    placeholder="YYYYMMDD"
                    onChange={(e) => setStart(e.target.value)}
                  />
                </label>
                <label className="space-y-1 text-sm">
                  <span className="text-muted-foreground">结束日期</span>
                  <Input
                    className="w-36"
                    value={end}
                    disabled={running}
                    placeholder="YYYYMMDD"
                    onChange={(e) => setEnd(e.target.value)}
                  />
                </label>
              </>
            )}
            {showCode && (
              <label className="space-y-1 text-sm">
                <span className="text-muted-foreground">
                  {isCBHolders ? '可转债编码' : '代码（可逗号分隔）'}
                </span>
                <Input
                  className="w-56"
                  value={code}
                  disabled={running}
                  placeholder={isCBHolders ? '留空=库内全部转债' : '000001.SZ'}
                  onChange={(e) => setCode(e.target.value)}
                />
              </label>
            )}
            <Button onClick={() => void onSubmit()} disabled={running || submitting}>
              {running ? '同步中…' : submitting ? '提交中…' : '开始同步'}
            </Button>
          </div>

          <p className="text-xs text-muted-foreground">
            {isCBHolders ? (
              <>
                四大参数：报告期、开始日期、结束日期、可转债编码。报告期与起止日至少填一种；编码留空则同步
                <code className="mx-1 rounded bg-muted px-1">cb_basic</code>
                中全部转债。接口约需 5000 积分。
              </>
            ) : isCBDaily ? (
              <>
                优先填「交易日」拉全市场截面；或用开始/结束日期（可加编码）。列表涨跌依赖此行情。约需 2000
                积分。
              </>
            ) : (
              <>
                推荐先用「一键同步」+ 单只股票试通。限流在项目根目录{' '}
                <code className="rounded bg-muted px-1">config.yaml</code> 的{' '}
                <code className="rounded bg-muted px-1">tushare.rpm</code>。
              </>
            )}
          </p>

          {formError ? <p className="text-sm text-destructive">{formError}</p> : null}

          {jobStatus && (jobStatus.running || jobStatus.finished_at) ? (
            <div className="rounded-md border bg-muted/40 px-3 py-2 text-sm">
              <div className="flex flex-wrap items-center gap-2">
                <span>当前任务</span>
                <Badge variant={jobStatus.running ? 'secondary' : jobStatus.error ? 'warning' : 'success'}>
                  {jobStatus.running ? 'running' : jobStatus.message || 'done'}
                </Badge>
                {jobStatus.job ? <Badge variant="outline">{jobStatus.job}</Badge> : null}
                {jobStatus.code ? <span className="text-muted-foreground">{jobStatus.code}</span> : null}
              </div>
              {jobStatus.start && jobStatus.end ? (
                <div className="mt-1 text-muted-foreground">
                  {jobStatus.start} ~ {jobStatus.end}
                </div>
              ) : null}
              {jobStatus.error ? <div className="mt-1 text-destructive">{jobStatus.error}</div> : null}
            </div>
          ) : null}
        </CardContent>
      </Card>

      <Card className="gap-0 py-0">
        <CardHeader className="px-4 py-3">
          <CardTitle>最近任务</CardTitle>
        </CardHeader>
        <CardContent className="px-0">
          {error ? (
            <p className="px-4 pb-4 text-sm text-destructive">{error}。请先启动：go run ./cmd/ashare web</p>
          ) : rows.length === 0 ? (
            <p className="px-4 pb-4 text-sm text-muted-foreground">还没有采集记录。</p>
          ) : (
            <div className="overflow-x-auto border-t">
              <table className="w-full text-left text-sm [&_td:first-child]:pl-4 [&_td:last-child]:pr-4 [&_th:first-child]:pl-4 [&_th:last-child]:pr-4">
                <thead className="text-muted-foreground">
                  <tr className="border-b">
                    <th className="py-2 pr-3 font-medium">任务</th>
                    <th className="py-2 pr-3 font-medium">状态</th>
                    <th className="py-2 pr-3 font-medium">行数</th>
                    <th className="py-2 pr-3 font-medium">开始</th>
                    <th className="py-2 pr-3 font-medium">结束</th>
                    <th className="py-2 pr-3 font-medium">参数</th>
                    <th className="py-2 pr-3 font-medium">说明</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row) => {
                    const paramsText = formatParams(row.params)
                    return (
                      <tr key={row.id} className="border-b last:border-0">
                        <td className="py-2 pr-3">{row.job_name}</td>
                        <td className="py-2 pr-3">
                          <Badge variant={statusVariant(row.status)}>{row.status}</Badge>
                        </td>
                        <td className="py-2 pr-3">{row.rows_affected}</td>
                        <td className="py-2 pr-3 whitespace-nowrap">
                          {formatISODateTimeStr(row.started_at)}
                        </td>
                        <td className="py-2 pr-3 whitespace-nowrap">
                          {row.finished_at ? formatISODateTimeStr(row.finished_at) : '—'}
                        </td>
                        <td className="max-w-xl py-2 pr-3 align-top">
                          {paramsText ? (
                            <pre className="overflow-x-auto rounded bg-muted/50 px-2 py-1 font-mono text-[11px] leading-5 text-muted-foreground">
                              {paramsText}
                            </pre>
                          ) : (
                            <span className="text-muted-foreground">—</span>
                          )}
                        </td>
                        <td className="max-w-md truncate py-2 pr-3 text-muted-foreground">{row.message}</td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
      </Main>
    </>
  )
}
