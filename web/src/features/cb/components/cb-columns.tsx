import { Link } from '@tanstack/react-router'
import { type ColumnDef } from '@tanstack/react-table'
import { DataTableColumnHeader } from '@/components/data-table'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import type { CBListItem } from '@/lib/api'
import type { CBListSearch } from '../list-search'

function fmtNum(n: number | null | undefined, digits = 2) {
  if (n == null || Number.isNaN(n)) return '—'
  return n.toLocaleString('zh-CN', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  })
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

export function cbColumns(listSearch: CBListSearch): ColumnDef<CBListItem>[] {
  return [
  {
    accessorKey: 'ts_code',
    header: ({ column }) => <DataTableColumnHeader column={column} title='代码' />,
    cell: ({ row }) => (
      <Link
        className='text-primary underline-offset-2 hover:underline'
        to='/cb/$code'
        params={{ code: row.original.ts_code }}
        search={listSearch}
      >
        {row.original.ts_code}
      </Link>
    ),
  },
  {
    accessorKey: 'bond_short_name',
    header: ({ column }) => <DataTableColumnHeader column={column} title='简称' />,
    cell: ({ row }) => row.original.bond_short_name || '—',
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
    accessorKey: 'cb_over_rate',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='溢价' className='justify-end' />
    ),
    cell: ({ row }) => (
      <div className='text-right tabular-nums'>{fmtPct(row.original.cb_over_rate)}</div>
    ),
  },
  {
    id: 'stk',
    accessorFn: (row) =>
      [row.stk_code, row.stk_short_name].filter(Boolean).join(' '),
    header: ({ column }) => <DataTableColumnHeader column={column} title='正股' />,
    cell: ({ getValue }) => getValue<string>() || '—',
  },
  {
    accessorKey: 'conv_price',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='转股价' className='justify-end' />
    ),
    cell: ({ row }) => (
      <div className='text-right tabular-nums'>{fmtNum(row.original.conv_price)}</div>
    ),
  },
  {
    accessorKey: 'maturity_call_price',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='到期价' className='justify-end' />
    ),
    cell: ({ row }) => (
      <div className='text-right tabular-nums'>
        {fmtNum(row.original.maturity_call_price)}
      </div>
    ),
  },
  {
    id: 'rating',
    accessorFn: (row) => row.newest_rating || row.issue_rating || '',
    header: ({ column }) => <DataTableColumnHeader column={column} title='评级' />,
    cell: ({ getValue }) => getValue<string>() || '—',
  },
  {
    accessorKey: 'quote_date',
    header: ({ column }) => <DataTableColumnHeader column={column} title='行情日' />,
    cell: ({ row }) => {
      const it = row.original
      return (
        <div>
          {it.quote_date || '—'}
          {it.delist_date ? (
            <Badge className='ms-2' variant='warning'>
              已摘牌
            </Badge>
          ) : null}
          {!it.quote_date && !it.delist_date ? (
            <span className='ms-2 text-xs text-muted-foreground'>无行情</span>
          ) : null}
        </div>
      )
    },
  },
]
}
