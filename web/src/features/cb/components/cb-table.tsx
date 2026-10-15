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
import type { CBListItem } from '@/lib/api'
import type { CBListSearch } from '../list-search'
import { cbColumns } from './cb-columns'

type CBTableProps = {
  data: CBListItem[]
  loading?: boolean
  listSearch: CBListSearch
  sorting: SortingState
  onSortingChange: OnChangeFn<SortingState>
}

export function CBTable({
  data,
  loading,
  listSearch,
  sorting,
  onSortingChange,
}: CBTableProps) {
  const columns = cbColumns(listSearch)
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
