import { createFileRoute } from '@tanstack/react-router'
import { Inspect } from '@/features/inspect'

export const Route = createFileRoute('/_authenticated/')({
  component: Inspect,
})
