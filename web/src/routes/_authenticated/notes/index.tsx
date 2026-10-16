import { createFileRoute } from '@tanstack/react-router'
import { NotesPage } from '@/features/notes'

export const Route = createFileRoute('/_authenticated/notes/')({
  component: NotesPage,
})
