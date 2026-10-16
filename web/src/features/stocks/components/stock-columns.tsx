import { Link } from '@tanstack/react-router'
import { type ColumnDef } from '@tanstack/react-table'
import { DataTableColumnHeader } from '@/components/data-table'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { StockListItem } from '@/lib/api'
import type { StockListSearch } from '../list-search'

function fmtNum(n: number | null | undefined, digits = 2) {
  if (n == null || Number.isNaN(n)) return '—'
  return n.toLocaleString('zh-CN', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  })
}

function fmtVol(n: number | null | undefined) {
  if (n == null || Number.isNaN(n)) return '—'
  if (n >= 1e8) return `${(n / 1e8).toFixed(2)}亿`
  if (n >= 1e4) return `${(n / 1e4).toFixed(2)}万`
  return n.toLocaleString('zh-CN')
}

function fmtAmount(n: number | null | undefined) {
  if (n == null || Number.isNaN(n)) return '—'
  if (n >= 1e8) return `${(n / 1e8).toFixed(2)}亿`
  if (n >= 1e4) return `${(n / 1e4).toFixed(2)}万`
  return fmtNum(n, 0)
}

function chgClass(n: number | null | undefined) {
  if (n == null || Number.isNaN(n) || n === 0) return 'text-muted-foreground'
  return n > 0 ? 'text-up' : 'text-down'
}

function fmtChg(n: number | null | undefined, digits = 2) {
  if (n == null || Number.isNaN(n)) return '—'
  const sign = n > 0 ? '+' : ''
  return `${sign}${n.toFixed(digits)}`
}

function fmtPct(n: number | null | undefined) {
  if (n == null || Number.isNaN(n)) return '—'
  const sign = n > 0 ? '+' : ''
  return `${sign}${n.toFixed(2)}%`
}

function statusLabel(status: string) {
  if (status === 'L') return '上市'
  if (status === 'D') return '退市'
  if (status === 'P') return '暂停'
  return status || '—'
}

export function stockColumns(listSearch: StockListSearch): ColumnDef<StockListItem>[] {
  return [
    {
      accessorKey: 'ts_code',
      header: ({ column }) => <DataTableColumnHeader column={column} title='代码' />,
      cell: ({ row }) => (
        <Link
          className='text-primary underline-offset-2 hover:underline'
          to='/'
          search={{ ...listSearch, code: row.original.ts_code }}
        >
          {row.original.ts_code}
        </Link>
      ),
    },
    {
      accessorKey: 'name',
      header: ({ column }) => <DataTableColumnHeader column={column} title='名称' />,
      cell: ({ row }) => (
        <div className='flex items-center gap-1.5'>
          <span>{row.original.name || '—'}</span>
          {row.original.is_st ? (
            <Badge variant='warning' className='text-[10px]'>
              ST
            </Badge>
          ) : null}
        </div>
      ),
    },
    {
      accessorKey: 'close',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='现价' className='justify-end' />
      ),
      cell: ({ row }) => (
        <div className={cn('text-right tabular-nums', chgClass(row.original.change))}>
          {fmtNum(row.original.close)}
        </div>
      ),
    },
    {
      accessorKey: 'change',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='涨跌' className='justify-end' />
      ),
      cell: ({ row }) => (
        <div className={cn('text-right tabular-nums', chgClass(row.original.change))}>
          {fmtChg(row.original.change)}
        </div>
      ),
    },
    {
      accessorKey: 'pct_chg',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='涨跌幅' className='justify-end' />
      ),
      cell: ({ row }) => (
        <div className={cn('text-right tabular-nums', chgClass(row.original.pct_chg))}>
          {fmtPct(row.original.pct_chg)}
        </div>
      ),
    },
    {
      accessorKey: 'volume',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='成交量' className='justify-end' />
      ),
      cell: ({ row }) => (
        <div className='text-right tabular-nums'>{fmtVol(row.original.volume)}</div>
      ),
    },
    {
      accessorKey: 'amount',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='成交额' className='justify-end' />
      ),
      cell: ({ row }) => (
        <div className='text-right tabular-nums'>{fmtAmount(row.original.amount)}</div>
      ),
    },
    {
      accessorKey: 'exchange',
      header: ({ column }) => <DataTableColumnHeader column={column} title='交易所' />,
      enableSorting: false,
      cell: ({ row }) => row.original.exchange || '—',
    },
    {
      accessorKey: 'status',
      header: ({ column }) => <DataTableColumnHeader column={column} title='状态' />,
      enableSorting: false,
      cell: ({ row }) => {
        const st = row.original.status
        return (
          <div className='flex items-center gap-1'>
            <span>{statusLabel(st)}</span>
            {row.original.is_paused ? (
              <Badge variant='outline' className='text-[10px]'>
                停牌
              </Badge>
            ) : null}
            {row.original.is_limit_up ? (
              <Badge variant='warning' className='text-[10px]'>
                涨停
              </Badge>
            ) : null}
            {row.original.is_limit_down ? (
              <Badge variant='success' className='text-[10px]'>
                跌停
              </Badge>
            ) : null}
          </div>
        )
      },
    },
    {
      accessorKey: 'quote_date',
      header: ({ column }) => <DataTableColumnHeader column={column} title='行情日' />,
      cell: ({ row }) => {
        const it = row.original
        return (
          <div>
            {it.quote_date || '—'}
            {!it.quote_date && it.status === 'L' ? (
              <span className='ms-2 text-xs text-muted-foreground'>无行情</span>
            ) : null}
          </div>
        )
      },
    },
  ]
}
