import { getCommands, getExtraCommands } from '@uiw/react-md-editor/commands-cn'
import MDEditor from '@uiw/react-md-editor/nohighlight'
import { useTheme } from '@/context/theme-provider'

export function MarkdownEditor({
  value,
  onChange,
}: {
  value: string
  onChange: (value: string) => void
}) {
  const { resolvedTheme } = useTheme()
  return (
    <div data-color-mode={resolvedTheme} className='rounded-md border'>
      <MDEditor
        value={value}
        height={420}
        preview='live'
        commands={getCommands()}
        extraCommands={getExtraCommands()}
        visibleDragbar
        textareaProps={{ placeholder: 'Markdown 正文' }}
        onChange={(next) => onChange(next ?? '')}
      />
    </div>
  )
}

export function MarkdownView({ source }: { source: string }) {
  const { resolvedTheme } = useTheme()
  if (!source.trim()) {
    return <p className='text-sm text-muted-foreground'>（无正文）</p>
  }
  return (
    <div data-color-mode={resolvedTheme} className='text-sm'>
      <MDEditor.Markdown source={source} />
    </div>
  )
}
