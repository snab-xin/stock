import { createFileRoute } from '@tanstack/react-router'
import { Backtest } from '@/features/backtest'

export const Route = createFileRoute('/_authenticated/backtest/')({
  component: Backtest,
})
