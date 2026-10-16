import { useEffect, useState } from 'react'
import { PanelLeftClose, PanelLeftOpen, Trash2 } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { MarkdownEditor } from '@/components/markdown-editor'
import { Input } from '@/components/ui/input'
import {
  createNote,
  deleteNote,
  fetchAuthStatus,
  fetchNote,
  fetchNotes,
  loginUser,
  logoutUser,
  registerUser,
  updateNote,
  type AuthUser,
  type NoteItem,
  type NoteTargetType,
} from '@/lib/api'

const kinds: { value: NoteTargetType; label: string }[] = [
  { value: '', label: '不关联标的' },
  { value: 'stock', label: '股票' },
  { value: 'cb', label: '可转债' },
  { value: 'index', label: '指数' },
]

function todayISO() {
  const d = new Date()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${d.getFullYear()}-${m}-${day}`
}

export function NotesPage() {
  const [authLoading, setAuthLoading] = useState(true)
  const [needsSetup, setNeedsSetup] = useState(false)
  const [user, setUser] = useState<AuthUser | null>(null)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [authError, setAuthError] = useState('')

  const [items, setItems] = useState<NoteItem[]>([])
  const [editing, setEditing] = useState<number | null>(null)
  const [tradeDate, setTradeDate] = useState(todayISO())
  const [summary, setSummary] = useState('')
  const [content, setContent] = useState('')
  const [tsCode, setTsCode] = useState('')
  const [targetType, setTargetType] = useState<NoteTargetType>('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)
  const [listOpen, setListOpen] = useState(true)
  const [pendingDelete, setPendingDelete] = useState<NoteItem | null>(null)
  const [deleting, setDeleting] = useState(false)

  async function refreshAuth() {
    const st = await fetchAuthStatus()
    setNeedsSetup(st.needs_setup)
    setUser(st.user)
    return st.user
  }

  async function refreshNotes() {
    const res = await fetchNotes()
    setItems(res.items ?? [])
  }

  useEffect(() => {
    void refreshAuth()
      .then((u) => {
        if (u) return refreshNotes()
      })
      .catch((e) => setAuthError(e instanceof Error ? e.message : '加载失败'))
      .finally(() => setAuthLoading(false))
  }, [])

  function resetForm() {
    setEditing(null)
    setTradeDate(todayISO())
    setSummary('')
    setContent('')
    setTsCode('')
    setTargetType('')
  }

  async function onAuth(e: React.FormEvent) {
    e.preventDefault()
    setAuthError('')
    try {
      if (needsSetup) await registerUser(username, password)
      else await loginUser(username, password)
      setPassword('')
      const u = await refreshAuth()
      if (u) await refreshNotes()
    } catch (err) {
      setAuthError(err instanceof Error ? err.message : '登录失败')
    }
  }

  async function onSave(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    if (tsCode.trim() && !targetType) {
      setError('填写了代码时，请选择股票、可转债或指数')
      return
    }
    setSaving(true)
    const body = {
      trade_date: tradeDate,
      summary,
      content,
      ts_code: targetType ? tsCode : '',
      target_type: tsCode.trim() ? targetType : '',
    }
    try {
      if (editing) await updateNote(editing, body)
      else await createNote(body)
      resetForm()
      await refreshNotes()
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  function onEdit(it: NoteItem) {
    setEditing(it.id)
    setTradeDate(it.trade_date)
    setSummary(it.summary)
    setTsCode(it.ts_code || '')
    setTargetType((it.target_type || '') as NoteTargetType)
    setContent('')
    void fetchNote(it.id)
      .then((full) => setContent(full.content ?? ''))
      .catch((err) => setError(err instanceof Error ? err.message : '加载失败'))
  }

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        {user ? (
          <Button
            variant='outline'
            size='sm'
            onClick={() => {
              void logoutUser().then(() => {
                setUser(null)
                setItems([])
              })
            }}
          >
            退出
          </Button>
        ) : null}
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>笔记</h1>
          <p className='text-sm text-muted-foreground'>
            正文用 Markdown 保存。交易日必填；代码可选。有代码时需选择股票、可转债或指数，日 K 上会显示「记」标记。
          </p>
        </div>

        {authLoading ? (
          <p className='text-sm text-muted-foreground'>加载中…</p>
        ) : !user ? (
          <Card className='max-w-md'>
            <CardHeader>
              <CardTitle>{needsSetup ? '创建第一个用户' : '登录'}</CardTitle>
            </CardHeader>
            <CardContent>
              <form className='flex flex-col gap-3' onSubmit={(e) => void onAuth(e)}>
                <Input
                  placeholder='用户名'
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  autoComplete='username'
                />
                <Input
                  type='password'
                  placeholder='密码'
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  autoComplete={needsSetup ? 'new-password' : 'current-password'}
                />
                {authError ? <p className='text-sm text-destructive'>{authError}</p> : null}
                <Button type='submit'>{needsSetup ? '注册并登录' : '登录'}</Button>
              </form>
            </CardContent>
          </Card>
        ) : (
          <div className='flex items-start gap-3'>
            {listOpen ? (
              <Card className='w-52 shrink-0'>
                <CardHeader className='px-3'>
                  <CardTitle className='text-base'>笔记</CardTitle>
                  <CardDescription className='truncate'>{user.username}</CardDescription>
                  <CardAction>
                    <Button
                      type='button'
                      variant='ghost'
                      size='icon'
                      className='size-8'
                      aria-label='收起列表'
                      onClick={() => setListOpen(false)}
                    >
                      <PanelLeftClose />
                    </Button>
                  </CardAction>
                </CardHeader>
                <CardContent className='flex flex-col gap-2 px-3'>
                  {items.length === 0 ? (
                    <p className='text-sm text-muted-foreground'>还没有笔记。</p>
                  ) : (
                    items.map((it) => (
                      <div
                        key={it.id}
                        className='flex items-center justify-between gap-1 rounded-md border px-2 py-2'
                      >
                        <button
                          className='min-w-0 text-left'
                          onClick={() => onEdit(it)}
                          type='button'
                        >
                          <div className='truncate text-sm font-medium'>{it.summary}</div>
                          <div className='truncate text-xs text-muted-foreground'>
                            {it.trade_date}
                            {it.ts_code ? ` · ${it.ts_code}` : ''}
                          </div>
                        </button>
                        <Button
                          type='button'
                          variant='ghost'
                          size='icon'
                          className='size-8 shrink-0'
                          aria-label='删除'
                          onClick={() => setPendingDelete(it)}
                        >
                          <Trash2 />
                        </Button>
                      </div>
                    ))
                  )}
                </CardContent>
              </Card>
            ) : (
              <Button
                type='button'
                variant='outline'
                size='icon'
                className='shrink-0'
                aria-label='展开列表'
                onClick={() => setListOpen(true)}
              >
                <PanelLeftOpen />
              </Button>
            )}

            <Card className='min-w-0 flex-1'>
              <CardHeader>
                <CardTitle>{editing ? '编辑' : '新建'}</CardTitle>
              </CardHeader>
              <CardContent>
                <form className='flex flex-col gap-3' onSubmit={(e) => void onSave(e)}>
                  <Input
                    type='date'
                    value={tradeDate}
                    onChange={(e) => setTradeDate(e.target.value)}
                    required
                  />
                  <Input
                    placeholder='标题'
                    value={summary}
                    onChange={(e) => setSummary(e.target.value)}
                    required
                  />
                  <Input
                    placeholder='代码（可选）'
                    value={tsCode}
                    onChange={(e) => setTsCode(e.target.value)}
                  />
                  <select
                    className='h-9 rounded-md border bg-background px-3 text-sm'
                    value={targetType}
                    onChange={(e) => setTargetType(e.target.value as NoteTargetType)}
                  >
                    {kinds.map((k) => (
                      <option key={k.value || 'none'} value={k.value}>
                        {k.label}
                      </option>
                    ))}
                  </select>
                  <MarkdownEditor value={content} onChange={setContent} />
                  {error ? <p className='text-sm text-destructive'>{error}</p> : null}
                  <div className='flex gap-2'>
                    <Button type='submit' disabled={saving}>
                      {saving ? '保存中…' : '保存'}
                    </Button>
                    {editing ? (
                      <Button type='button' variant='outline' onClick={resetForm}>
                        取消
                      </Button>
                    ) : null}
                  </div>
                </form>
              </CardContent>
            </Card>
            <ConfirmDialog
              open={pendingDelete !== null}
              onOpenChange={(open) => {
                if (!open) setPendingDelete(null)
              }}
              title='删除笔记'
              desc={
                pendingDelete
                  ? `确定删除「${pendingDelete.summary}」（${pendingDelete.trade_date}）？删除后无法恢复。`
                  : '确定删除这篇笔记？删除后无法恢复。'
              }
              cancelBtnText='取消'
              confirmText={deleting ? '删除中…' : '删除'}
              destructive
              isLoading={deleting}
              handleConfirm={() => {
                if (!pendingDelete) return
                const id = pendingDelete.id
                setDeleting(true)
                void deleteNote(id)
                  .then(async () => {
                    if (editing === id) resetForm()
                    setPendingDelete(null)
                    await refreshNotes()
                  })
                  .catch((err) => setError(err instanceof Error ? err.message : '删除失败'))
                  .finally(() => setDeleting(false))
              }}
            />
          </div>
        )}
      </Main>
    </>
  )
}
