import { useCallback, useEffect, useState } from 'react'
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

function statusVariant(status: string) {
  if (status === 'success') return 'success' as const
  if (status === 'failed') return 'warning' as const
  return 'secondary' as const
}

function todayYYYYMMDD() {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}${m}${day}`
}

export function IngestPage() {
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
    if (job !== 'instruments' && (!start.trim() || !end.trim())) {
      setFormError('请填写开始/结束日期（YYYYMMDD）')
      return
    }
    if ((job === 'daily' || job === 'sync') && !code.trim()) {
      setFormError('建议填写证券代码；留空将拉取全市场，积分消耗很大')
      // still allow if user confirms? We'll block empty for safety in UI
      // User can clear and we warn - actually let them proceed with confirm
    }
    setSubmitting(true)
    try {
      const body: StartIngestRequest = { job }
      if (job !== 'instruments') {
        body.start = start.trim()
        body.end = end.trim()
      }
      if (code.trim()) body.code = code.trim()
      if ((job === 'daily' || job === 'sync') && !body.code) {
        const ok = window.confirm('未填写代码将按交易日拉全市场日线，可能很慢且消耗大量积分。继续？')
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

  const needRange = job !== 'instruments'
  const running = Boolean(jobStatus?.running)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">采集任务</h1>
          <p className="text-sm text-muted-foreground">
            在界面同步行情。限流：{cfg ? `${cfg.tushare_rpm} 次/分钟` : '…'}
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
              ] as const
            ).map(([value, label]) => (
              <Button
                key={value}
                variant={job === value ? 'default' : 'outline'}
                size="sm"
                disabled={running}
                onClick={() => setJob(value)}
              >
                {label}
              </Button>
            ))}
          </div>

          <div className="flex flex-wrap items-end gap-3">
            {needRange && (
              <>
                <label className="space-y-1 text-sm">
                  <span className="text-muted-foreground">开始</span>
                  <Input
                    className="w-36"
                    value={start}
                    disabled={running}
                    placeholder="YYYYMMDD"
                    onChange={(e) => setStart(e.target.value)}
                  />
                </label>
                <label className="space-y-1 text-sm">
                  <span className="text-muted-foreground">结束</span>
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
            {(job === 'daily' || job === 'sync') && (
              <label className="space-y-1 text-sm">
                <span className="text-muted-foreground">代码（可逗号分隔）</span>
                <Input
                  className="w-56"
                  value={code}
                  disabled={running}
                  placeholder="000001.SZ"
                  onChange={(e) => setCode(e.target.value)}
                />
              </label>
            )}
            <Button onClick={() => void onSubmit()} disabled={running || submitting}>
              {running ? '同步中…' : submitting ? '提交中…' : '开始同步'}
            </Button>
          </div>

          <p className="text-xs text-muted-foreground">
            推荐先用「一键同步」+ 单只股票试通。限流在项目根目录{' '}
            <code className="rounded bg-muted px-1">config.yaml</code> 的{' '}
            <code className="rounded bg-muted px-1">tushare.rpm</code>。
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

      <Card>
        <CardHeader>
          <CardTitle>最近任务</CardTitle>
        </CardHeader>
        <CardContent>
          {error ? (
            <p className="text-sm text-destructive">{error}。请先启动：go run ./cmd/ashare web</p>
          ) : rows.length === 0 ? (
            <p className="text-sm text-muted-foreground">还没有采集记录。</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead className="text-muted-foreground">
                  <tr className="border-b">
                    <th className="py-2 pr-3 font-medium">任务</th>
                    <th className="py-2 pr-3 font-medium">状态</th>
                    <th className="py-2 pr-3 font-medium">行数</th>
                    <th className="py-2 pr-3 font-medium">开始</th>
                    <th className="py-2 pr-3 font-medium">说明</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row) => (
                    <tr key={row.id} className="border-b last:border-0">
                      <td className="py-2 pr-3">{row.job_name}</td>
                      <td className="py-2 pr-3">
                        <Badge variant={statusVariant(row.status)}>{row.status}</Badge>
                      </td>
                      <td className="py-2 pr-3">{row.rows_affected}</td>
                      <td className="py-2 pr-3 whitespace-nowrap">
                        {row.started_at.replace('T', ' ').slice(0, 19)}
                      </td>
                      <td className="max-w-md truncate py-2 pr-3 text-muted-foreground">{row.message}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
