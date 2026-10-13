import { CandlestickChart, Database } from 'lucide-react'
import { NavLink, Outlet } from 'react-router-dom'
import { cn } from '@/lib/utils'

const links = [
  { to: '/', label: '行情质检', icon: CandlestickChart, end: true },
  { to: '/ingest', label: '采集任务', icon: Database, end: false },
]

export function AppLayout() {
  return (
    <div className="flex min-h-svh">
      <aside className="flex w-56 shrink-0 flex-col bg-sidebar text-sidebar-foreground">
        <div className="border-b border-sidebar-border px-4 py-4">
          <div className="text-sm font-semibold tracking-wide">ashare</div>
          <div className="mt-0.5 text-xs text-sidebar-foreground/60">A 股数据看板</div>
        </div>
        <nav className="flex flex-1 flex-col gap-1 p-2">
          {links.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              end={link.end}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-2 rounded-md px-3 py-2 text-sm',
                  isActive
                    ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                    : 'text-sidebar-foreground/80 hover:bg-sidebar-accent/60',
                )
              }
            >
              <link.icon className="h-4 w-4" />
              {link.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="min-w-0 flex-1 overflow-auto p-6">
        <Outlet />
      </main>
    </div>
  )
}
