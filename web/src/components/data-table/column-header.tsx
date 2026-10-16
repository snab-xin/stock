import { ArrowDownIcon, ArrowUpIcon, CaretSortIcon } from '@radix-ui/react-icons'
import { type Column } from '@tanstack/react-table'
import { cn } from '@/lib/utils'

type DataTableColumnHeaderProps<TData, TValue> =
  React.HTMLAttributes<HTMLDivElement> & {
    column: Column<TData, TValue>
    title: string
  }

export function DataTableColumnHeader<TData, TValue>({
  column,
  title,
  className,
}: DataTableColumnHeaderProps<TData, TValue>) {
  if (!column.getCanSort()) {
    return <div className={cn(className)}>{title}</div>
  }

  const sorted = column.getIsSorted()

  return (
    <div className={cn('flex items-center', className)}>
      <button
        type='button'
        className='-mx-2 inline-flex h-8 items-center rounded-md px-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground'
        onClick={() => column.toggleSorting(sorted === 'asc')}
      >
        <span>{title}</span>
        {sorted === 'desc' ? (
          <ArrowDownIcon className='ms-2 h-4 w-4' />
        ) : sorted === 'asc' ? (
          <ArrowUpIcon className='ms-2 h-4 w-4' />
        ) : (
          <CaretSortIcon className='ms-2 h-4 w-4' />
        )}
      </button>
    </div>
  )
}
