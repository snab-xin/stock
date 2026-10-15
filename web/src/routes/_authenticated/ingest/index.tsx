import { createFileRoute } from '@tanstack/react-router'
import { Ingest } from '@/features/ingest'

export const Route = createFileRoute('/_authenticated/ingest/')({
  component: Ingest,
})
