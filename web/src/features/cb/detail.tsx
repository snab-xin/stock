import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { Link, getRouteApi } from '@tanstack/react-router'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import {
  fetchCBDetail,
  fetchCBNeighbors,
  saveCBHolders,
  type CBDetailResponse,
  type CBNeighborsResponse,
} from '@/lib/api'
import { resolveCBListSearch, toCBListSearchParams } from './list-search'
import { parseHoldersPaste, type ParsedHolders } from './parse-holders'

const routeApi = getRouteApi('/_authenticated/cb/$code')

function fmtNum(n: number | null | undefined, digits = 2) {
  if (n == null || Number.isNaN(n)) return '—'
  return n.toLocaleString('zh-CN', { maximumFractionDigits: digits })
}

function fmtYuanYi(n: number | null | undefined) {
  if (n == null || Number.isNaN(n)) return '—'
  return `${(n / 1e8).toFixed(2)} 亿`
}

function Field({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className='space-y-0.5'>
      <div className='text-xs text-muted-foreground'>{label}</div>
      <div className='break-words text-sm'>{value || '—'}</div>
    </div>
  )
}

export function CBDetail() {
  const { code } = routeApi.useParams()
  const search = routeApi.useSearch()
  const listQuery = resolveCBListSearch(search)
  const listSearch = toCBListSearchParams({
    q: listQuery.q,
    listedOnly: listQuery.listedOnly,
    sortBy: listQuery.sortBy,
    sortDir: listQuery.sortDir,
  })

  const [period, setPeriod] = useState('')
  const [data, setData] = useState<CBDetailResponse | null>(null)
  const [neighbors, setNeighbors] = useState<CBNeighborsResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [pasteText, setPasteText] = useState('')
  const [pasteOpen, setPasteOpen] = useState(false)
  const [parsed, setParsed] = useState<ParsedHolders | null>(null)
  const [savingHolders, setSavingHolders] = useState(false)
  const [pasteError, setPasteError] = useState('')
  const [pasteOk, setPasteOk] = useState('')

  async function load(nextCode: string, nextPeriod?: string) {
    if (!nextCode) return
    setLoading(true)
    setError('')
    try {
      const res = await fetchCBDetail(nextCode, nextPeriod || undefined)
      setData(res)
      setPeriod(res.selected_period || '')
    } catch (e) {
      setError(e instanceof Error ? e.message : '加载失败')
      setData(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load(code)
  }, [code])

  useEffect(() => {
    let cancelled = false
    void fetchCBNeighbors(code, {
      q: listQuery.q,
      listedOnly: listQuery.listedOnly,
      sortBy: listQuery.sortBy,
      sortDir: listQuery.sortDir,
    })
      .then((res) => {
        if (!cancelled) setNeighbors(res)
      })
      .catch(() => {
        if (!cancelled) setNeighbors(null)
      })
    return () => {
      cancelled = true
    }
  }, [code, listQuery.q, listQuery.listedOnly, listQuery.sortBy, listQuery.sortDir])

  function resetPaste() {
    setPasteText('')
    setParsed(null)
    setPasteError('')
    setPasteOk('')
  }

  function parsePaste() {
    setPasteError('')
    setPasteOk('')
    try {
      setParsed(parseHoldersPaste(pasteText))
    } catch (e) {
      setParsed(null)
      setPasteError(e instanceof Error ? e.message : '解析失败')
    }
  }

  async function confirmSave() {
    if (!code || !parsed) {
      setPasteError('请先解析粘贴文本')
      return
    }
    setSavingHolders(true)
    setPasteError('')
    setPasteOk('')
    try {
      const res = await saveCBHolders(code, {
        period: parsed.end_date,
        holders: parsed.holders,
      })
      setPasteText('')
      setParsed(null)
      setPasteError('')
      setPasteOk(`已保存 ${res.period}，共 ${res.count} 名`)
      setPasteOpen(false)
      await load(code, res.period)
    } catch (e) {
      setPasteError(e instanceof Error ? e.message : '保存失败')
    } finally {
      setSavingHolders(false)
    }
  }

  const selectedSum = useMemo(() => {
    if (!data?.selected_period) return null
    return data.holder_summaries.find((s) => s.end_date === data.selected_period) ?? null
  }, [data])

  const b = data?.basic

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4'>
        <div className='sticky top-16 z-40 -mx-4 border-b bg-background/95 px-4 py-2.5 backdrop-blur supports-backdrop-filter:bg-background/80'>
          <div className='flex flex-wrap items-center gap-3'>
            <Link
              to='/cb'
              search={listSearch}
              className='text-sm text-muted-foreground hover:text-foreground'
            >
              ← 可转债列表
            </Link>
            <div className='flex flex-wrap items-center gap-1.5'>
              {neighbors?.prev ? (
                <Button size='sm' variant='outline' asChild>
                  <Link to='/cb/$code' params={{ code: neighbors.prev }} search={listSearch}>
                    上一个
                  </Link>
                </Button>
              ) : (
                <Button size='sm' variant='outline' disabled>
                  上一个
                </Button>
              )}
              {neighbors?.next ? (
                <Button size='sm' variant='outline' asChild>
                  <Link to='/cb/$code' params={{ code: neighbors.next }} search={listSearch}>
                    下一个
                  </Link>
                </Button>
              ) : (
                <Button size='sm' variant='outline' disabled>
                  下一个
                </Button>
              )}
              {neighbors && neighbors.index > 0 ? (
                <span className='text-xs text-muted-foreground'>
                  {neighbors.index} / {neighbors.total}
                </span>
              ) : null}
            </div>
            {b ? (
              <>
                <h1 className='text-lg font-bold tracking-tight sm:text-xl'>
                  {b.ts_code} {b.bond_short_name}
                </h1>
                {b.cb_type ? <Badge variant='secondary'>{b.cb_type}</Badge> : null}
                {b.delist_date ? (
                  <Badge variant='warning'>已摘牌 {b.delist_date}</Badge>
                ) : null}
                {loading ? <Badge variant='outline'>加载中…</Badge> : null}
              </>
            ) : (
              <h1 className='text-lg font-bold tracking-tight sm:text-xl'>{code}</h1>
            )}
          </div>
        </div>

        {error ? <p className='text-sm text-destructive'>{error}</p> : null}

        {b && (
          <>
            <Card>
              <CardHeader>
                <CardTitle>基础信息</CardTitle>
              </CardHeader>
              <CardContent className='grid gap-4 sm:grid-cols-2 lg:grid-cols-3'>
                <Field label='全称' value={b.bond_full_name} />
                <Field
                  label='正股'
                  value={[b.stk_code, b.stk_short_name].filter(Boolean).join(' ')}
                />
                <Field label='交易所' value={b.exchange} />
                <Field
                  label='上市 / 摘牌'
                  value={[b.list_date, b.delist_date || '—'].join(' ~ ')}
                />
                <Field
                  label='起息 / 到期'
                  value={[b.value_date || '—', b.maturity_date || '—'].join(' ~ ')}
                />
                <Field
                  label='转股区间'
                  value={[b.conv_start_date || '—', b.conv_end_date || '—'].join(' ~ ')}
                />
                <Field
                  label='初始 / 最新转股价'
                  value={`${fmtNum(b.first_conv_price)} / ${fmtNum(b.conv_price)}`}
                />
                <Field
                  label='票面利率 / 补偿利率'
                  value={`${fmtNum(b.coupon_rate)}% / ${fmtNum(b.add_rate)}%`}
                />
                <Field
                  label='面值 / 发行价'
                  value={`${fmtNum(b.par)} / ${fmtNum(b.issue_price)}`}
                />
                <Field label='发行总额' value={fmtYuanYi(b.issue_size)} />
                <Field label='债券余额' value={fmtYuanYi(b.remain_size)} />
                <Field label='到期赎回价' value={fmtNum(b.maturity_call_price)} />
                <Field
                  label='发行 / 最新评级'
                  value={[b.issue_rating || '—', b.newest_rating || '—'].join(' / ')}
                />
                <Field label='评级机构' value={b.rating_comp} />
                <Field label='担保人' value={b.guarantor} />
              </CardContent>
            </Card>

            <Card>
              <CardHeader className='flex flex-row flex-wrap items-center justify-between gap-2 space-y-0'>
                <CardTitle>十大持有人</CardTitle>
                <Button
                  size='sm'
                  variant={pasteOpen ? 'default' : 'outline'}
                  onClick={() => {
                    setPasteOpen((v) => !v)
                    setPasteError('')
                    setPasteOk('')
                  }}
                >
                  {pasteOpen ? '收起维护' : '粘贴维护'}
                </Button>
              </CardHeader>
              <CardContent className='space-y-4'>
                {pasteOk && !pasteOpen ? (
                  <p className='text-sm text-muted-foreground'>{pasteOk}</p>
                ) : null}
                {pasteOpen ? (
                  <div className='space-y-3 rounded-md border bg-muted/30 p-3'>
                    <p className='text-xs text-muted-foreground'>
                      粘贴网页表格文本后点「解析预览」，确认表格无误再保存。数量按万张入库；「合计」「环比」会忽略。同报告期会整期覆盖。
                    </p>
                    <Textarea
                      className='min-h-40 font-mono text-xs'
                      placeholder={`截止日期2026-06-30\n持有人名称\t持有数量(张)\t持有比例(%)\t环比\n某某基金\t33.77万\t14.59\t不变\n合计\t133.1万\t57.49\t--`}
                      value={pasteText}
                      onChange={(e) => {
                        setPasteText(e.target.value)
                        setParsed(null)
                        setPasteError('')
                        setPasteOk('')
                      }}
                    />
                    <div className='flex flex-wrap items-center gap-2'>
                      <Button size='sm' variant='secondary' onClick={parsePaste}>
                        解析预览
                      </Button>
                      <Button
                        size='sm'
                        disabled={savingHolders || !parsed}
                        onClick={() => void confirmSave()}
                      >
                        {savingHolders ? '保存中…' : '确认保存'}
                      </Button>
                      {parsed || pasteText ? (
                        <Button size='sm' variant='ghost' onClick={resetPaste}>
                          清空
                        </Button>
                      ) : null}
                      {pasteError ? (
                        <span className='text-sm text-destructive'>{pasteError}</span>
                      ) : null}
                      {pasteOk ? (
                        <span className='text-sm text-muted-foreground'>{pasteOk}</span>
                      ) : null}
                    </div>

                    {parsed ? (
                      <div className='space-y-2'>
                        <div className='flex flex-wrap gap-2 text-sm'>
                          <Badge variant='secondary'>报告期 {parsed.end_date}</Badge>
                          <Badge variant='outline'>{parsed.holders.length} 名</Badge>
                        </div>
                        <div className='overflow-x-auto rounded-md border bg-background'>
                          <table className='w-full text-left text-sm'>
                            <thead className='text-muted-foreground'>
                              <tr>
                                <th className='px-3 py-1.5 font-medium'>排名</th>
                                <th className='px-3 py-1.5 font-medium'>持有人</th>
                                <th className='px-3 py-1.5 font-medium'>数量(万张)</th>
                                <th className='px-3 py-1.5 font-medium'>比例%</th>
                              </tr>
                            </thead>
                            <tbody>
                              {parsed.holders.map((h) => (
                                <tr key={h.holder_rank} className='border-t'>
                                  <td className='px-3 py-1.5'>{h.holder_rank}</td>
                                  <td className='px-3 py-1.5'>{h.holder_name}</td>
                                  <td className='px-3 py-1.5'>{fmtNum(h.hold_amount, 3)}</td>
                                  <td className='px-3 py-1.5'>{fmtNum(h.hold_ratio)}</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    ) : null}
                  </div>
                ) : null}

                {data.holder_periods.length === 0 ? (
                  <p className='text-sm text-muted-foreground'>
                    暂无持有人数据。可点「粘贴维护」手工录入，或采集页同步（覆盖往往不全）。
                  </p>
                ) : (
                  <>
                    <div className='flex flex-wrap gap-2'>
                      {data.holder_periods.map((p) => (
                        <Button
                          key={p}
                          size='sm'
                          variant={p === period ? 'default' : 'outline'}
                          disabled={loading}
                          onClick={() => void load(code, p)}
                        >
                          {p}
                        </Button>
                      ))}
                    </div>

                    {selectedSum ? (
                      <div className='flex flex-wrap gap-2 text-sm'>
                        <Badge variant='secondary'>报告期 {selectedSum.end_date}</Badge>
                        <Badge variant='outline'>
                          前十大占比合计 {fmtNum(selectedSum.top10_ratio)}%
                        </Badge>
                        <Badge variant='outline'>
                          前十大数量合计 {fmtNum(selectedSum.top10_amount)} 万张
                        </Badge>
                        <Badge variant='outline'>{selectedSum.holder_count} 名</Badge>
                      </div>
                    ) : null}

                    <div className='overflow-x-auto'>
                      <table className='w-full text-left text-sm'>
                        <thead className='text-muted-foreground'>
                          <tr>
                            <th className='py-1.5 pr-3 font-medium'>排名</th>
                            <th className='py-1.5 pr-3 font-medium'>持有人</th>
                            <th className='py-1.5 pr-3 font-medium'>数量(万张)</th>
                            <th className='py-1.5 font-medium'>比例%</th>
                          </tr>
                        </thead>
                        <tbody>
                          {data.holders.map((h) => (
                            <tr key={`${h.end_date}-${h.holder_rank}`} className='border-t'>
                              <td className='py-1.5 pr-3'>{h.holder_rank}</td>
                              <td className='py-1.5 pr-3'>{h.holder_name}</td>
                              <td className='py-1.5 pr-3'>{fmtNum(h.hold_amount, 3)}</td>
                              <td className='py-1.5'>{fmtNum(h.hold_ratio)}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>

                    {data.holder_summaries.length > 1 ? (
                      <div>
                        <div className='mb-2 text-sm font-medium'>各期前十大集中度</div>
                        <div className='overflow-x-auto'>
                          <table className='w-full text-left text-sm'>
                            <thead className='text-muted-foreground'>
                              <tr>
                                <th className='py-1.5 pr-3 font-medium'>报告期</th>
                                <th className='py-1.5 pr-3 font-medium'>占比合计%</th>
                                <th className='py-1.5 pr-3 font-medium'>数量合计(万张)</th>
                                <th className='py-1.5 font-medium'>人数</th>
                              </tr>
                            </thead>
                            <tbody>
                              {data.holder_summaries.map((s) => (
                                <tr key={s.end_date} className='border-t'>
                                  <td className='py-1.5 pr-3'>{s.end_date}</td>
                                  <td className='py-1.5 pr-3'>{fmtNum(s.top10_ratio)}</td>
                                  <td className='py-1.5 pr-3'>{fmtNum(s.top10_amount, 3)}</td>
                                  <td className='py-1.5'>{s.holder_count}</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                        <p className='mt-2 text-xs text-muted-foreground'>
                          合计为前十大占比之和，不等于全体持有人 100%。用于对比集中度变化。
                        </p>
                      </div>
                    ) : null}
                  </>
                )}
              </CardContent>
            </Card>

            {(b.rate_clause ||
              b.put_clause ||
              b.call_clause ||
              b.reset_clause ||
              b.conv_clause) && (
              <Card>
                <CardHeader>
                  <CardTitle>条款摘要</CardTitle>
                </CardHeader>
                <CardContent className='space-y-3 text-sm text-muted-foreground'>
                  {b.rate_clause ? (
                    <p>
                      <span className='font-medium text-foreground'>利率：</span>
                      {b.rate_clause}
                    </p>
                  ) : null}
                  {b.put_clause ? (
                    <p>
                      <span className='font-medium text-foreground'>回售：</span>
                      {b.put_clause}
                    </p>
                  ) : null}
                  {b.call_clause ? (
                    <p>
                      <span className='font-medium text-foreground'>赎回：</span>
                      {b.call_clause}
                    </p>
                  ) : null}
                  {b.reset_clause ? (
                    <p>
                      <span className='font-medium text-foreground'>下修：</span>
                      {b.reset_clause}
                    </p>
                  ) : null}
                  {b.conv_clause ? (
                    <p>
                      <span className='font-medium text-foreground'>转股：</span>
                      {b.conv_clause}
                    </p>
                  ) : null}
                </CardContent>
              </Card>
            )}
          </>
        )}
      </Main>
    </>
  )
}
