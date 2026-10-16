import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type OnChangeFn,
  type SortingState,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { StockListItem } from '@/lib/api'
import type { StockListSearch } from '../list-search'
import { stockColumns } from './stock-columns'

type StockTableProps = {
  data: StockListItem[]
  loading?: boolean
  listSearch: StockListSearch
  sorting: SortingState
  onSortingChange: OnChangeFn<SortingState>
}

export function StockTable({
  data,
  loading,
  listSearch,
  sorting,
  onSortingChange,
}: StockTableProps) {
  const columns = stockColumns(listSearch)
  const table = useReactTable({
    data,
    columns,
    state: { sorting },
    onSortingChange,
    manualSorting: true,
    getCoreRowModel: getCoreRowModel(),
    getRowId: (row) => row.ts_code,
  })

  return (
    <div className='overflow-hidden rounded-md border'>
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableHead key={header.id} className='whitespace-nowrap'>
                  {header.isPlaceholder
                    ? null
                    : flexRender(header.column.columnDef.header, header.getContext())}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows.length ? (
            table.getRowModel().rows.map((row) => (
              <TableRow key={row.id} className='hover:bg-muted/40'>
                {row.getVisibleCells().map((cell) => (
                  <TableCell key={cell.id} className='py-1.5'>
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))
          ) : (
            <TableRow>
              <TableCell
                colSpan={columns.length}
                className='h-24 text-center text-muted-foreground'
              >
                {loading ? '加载中…' : '没有数据。'}
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
    </div>
  )
}
